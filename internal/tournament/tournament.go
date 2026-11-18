package tournament

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/nchandur/ranga-tools/internal/registry"
	"github.com/nchandur/ranga-tools/internal/settings"
)

const DefaultConfigFile = "game-config.json"

type EngineSpec struct {
	Name    string         `json:"name"`
	Options map[string]any `json:"options"`
	Role    string         `json:"-"` // used in error messages
	Suffix  string         `json:"-"` // appended to the name fastchess sees
}

type Openings struct {
	File   string `json:"file"`
	Format string `json:"format"`
	Order  string `json:"order"`
}

// Common holds what every tournament type needs. It also supplies no-op
// Validate and Args, so a type only overrides what it actually has.
type Common struct {
	TC          string    `json:"tc"`
	Nodes       int       `json:"nodes"`
	Rounds      int       `json:"rounds"`
	Games       int       `json:"games"`
	Concurrency int       `json:"concurrency"`
	Openings    *Openings `json:"openings"`
	OutputDir   string    `json:"output_dir"`
	ExtraArgs   []string  `json:"extra_args"`

	dir string // config file's directory; relative paths resolve against it
}

func (c *Common) Shared() *Common { return c }
func (c *Common) Validate() error { return nil }
func (c *Common) Args() []string  { return nil }

// Tournament is what a type provides: Engines always, Validate and Args when needed.
type Tournament interface {
	Shared() *Common
	Engines() []EngineSpec
	Validate() error // type-specific checks
	Args() []string  // type-specific fastchess flags
}

func (c *Common) resolve(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(c.dir, p)
}

// prepare applies defaults and validates the shared fields.
// tc wins over nodes; with neither, nodes defaults to 5000.
func (c *Common) prepare() error {
	if min(c.Rounds, c.Games, c.Concurrency, c.Nodes) < 0 {
		return errors.New("rounds, games, concurrency and nodes must not be negative")
	}
	c.Rounds = cmp.Or(c.Rounds, 5000)
	c.Games = cmp.Or(c.Games, 2)
	c.Concurrency = cmp.Or(c.Concurrency, max(runtime.NumCPU()-1, 1))
	c.OutputDir = cmp.Or(c.OutputDir, "output")
	if c.TC != "" {
		c.Nodes = 0
	} else {
		c.Nodes = cmp.Or(c.Nodes, 5000)
	}

	if o := c.Openings; o != nil {
		if o.File == "" {
			return errors.New(`"openings.file" is required when "openings" is set`)
		}
		o.Format = cmp.Or(o.Format, strings.TrimPrefix(strings.ToLower(filepath.Ext(o.File)), "."))
		o.Order = cmp.Or(o.Order, "random")
		if o.Format != "epd" && o.Format != "pgn" {
			return fmt.Errorf("openings.format must be \"epd\" or \"pgn\" (got %q)", o.Format)
		}
		if o.Order != "random" && o.Order != "sequential" {
			return fmt.Errorf("openings.order must be \"random\" or \"sequential\" (got %q)", o.Order)
		}
	}
	return nil
}

// execute loads the config into t, validates it, and runs fastchess.
func execute(configPath, fastchessOverride string, t Tournament) error {
	abs, err := filepath.Abs(configPath)
	if err != nil {
		return err
	}
	f, err := os.Open(abs)
	if err != nil {
		return fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields() // typos fail loudly
	dec.UseNumber()             // keep option values like 128 intact
	if err := dec.Decode(t); err != nil {
		return fmt.Errorf("parse %s: %w", abs, err)
	}

	c := t.Shared()
	c.dir = filepath.Dir(abs)
	for _, check := range []func() error{
		c.prepare,
		func() error { return validateEngines(t.Engines()) },
		t.Validate,
	} {
		if err := check(); err != nil {
			return fmt.Errorf("%s: %w", abs, err)
		}
	}
	return run(t, fastchessOverride)
}

func validateEngines(specs []EngineSpec) error {
	seen := map[string]string{}
	for _, s := range specs {
		if strings.TrimSpace(s.Name) == "" {
			return fmt.Errorf(`"%s.name" is required`, s.Role)
		}
		slug := registry.Slug(s.Name + s.Suffix)
		if other, dup := seen[slug]; dup {
			return fmt.Errorf("%s and %s must be different engines", other, s.Role)
		}
		seen[slug] = s.Role
	}
	return nil
}

func run(t Tournament, fastchessOverride string) error {
	c, specs := t.Shared(), t.Engines()

	store, err := registry.Default()
	if err != nil {
		return err
	}
	args := []string{}
	for _, s := range specs {
		e, err := store.Get(s.Name)
		if errors.Is(err, registry.ErrNotFound) {
			return fmt.Errorf("%s engine %q is not registered", s.Role, s.Name)
		} else if err != nil {
			return err
		}
		args = append(args, "-engine", "cmd="+e.Path, "name="+e.Name+s.Suffix)
		for _, k := range slices.Sorted(maps.Keys(s.Options)) {
			args = append(args, fmt.Sprintf("option.%s=%v", k, s.Options[k]))
		}
	}

	bin, err := findFastchess(fastchessOverride)
	if err != nil {
		return err
	}
	outDir := c.resolve(c.OutputDir)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	if c.TC != "" {
		args = append(args, "-each", "tc="+c.TC)
	} else {
		args = append(args, "-each", "nodes="+strconv.Itoa(c.Nodes))
	}
	if o := c.Openings; o != nil {
		args = append(args, "-openings", "file="+c.resolve(o.File), "format="+o.Format, "order="+o.Order)
	}
	args = append(args,
		"-rounds", strconv.Itoa(c.Rounds),
		"-games", strconv.Itoa(c.Games),
		"-concurrency", strconv.Itoa(c.Concurrency))
	args = append(args, t.Args()...)
	args = append(args,
		"-pgnout", "file="+filepath.Join(outDir, "games.pgn"),
		"-log", "file="+filepath.Join(outDir, "games.log"), "engine=true",
		"-recover")
	args = append(args, c.ExtraArgs...)

	fmt.Fprintf(os.Stderr, "running: %s %q\n", bin, args)

	cmd := exec.Command(bin, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	// Ctrl-C also reaches fastchess; ignore it here so it can print its summary.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	defer signal.Stop(sigs)

	if err := cmd.Run(); err != nil {
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			return fmt.Errorf("fastchess exited with code %d", exitErr.ExitCode())
		}
		return fmt.Errorf("run fastchess: %w", err)
	}
	return nil
}

// findFastchess resolves the binary: flag, then stored setting, then PATH.
func findFastchess(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	s, err := settings.Load()
	if err != nil {
		return "", err
	}
	if s.FastchessPath != "" {
		return s.FastchessPath, nil
	}
	p, err := exec.LookPath("fastchess")
	if err != nil {
		return "", fmt.Errorf("fastchess not found: pass --fastchess, run `config set %s <path>`, or put it on PATH", settings.KeyFastchessPath)
	}
	return p, nil
}

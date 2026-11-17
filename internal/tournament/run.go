package tournament

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/nchandur/ranga-tools/internal/registry"
	"github.com/nchandur/ranga-tools/internal/settings"
)

type Tournament interface {
	Shared() *Common
	Engines() []EngineSpec
	Validate() error
	Args() []string
}

// execute loads the config into t, validates it, and runs fastchess
func execute(configPath, fastchessOverride string, t Tournament) error {
	abs, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("resolve config path: %w", err)
	}
	if err := load(abs, t); err != nil {
		return err
	}
	return run(t, fastchessOverride)
}

func load(abs string, t Tournament) error {
	f, err := os.Open(abs)
	if err != nil {
		return fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	dec.UseNumber()
	if err := dec.Decode(t); err != nil {
		return fmt.Errorf("parse %s: %w", abs, err)
	}

	c := t.Shared()
	c.dir = filepath.Dir(abs)

	for _, step := range []func() error{
		c.applyDefaults,
		c.validate,
		func() error { return validateEngines(t.Engines()) },
		t.Validate,
	} {
		if err := step(); err != nil {
			return fmt.Errorf("%s: %w", abs, err)
		}
	}
	return nil
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

		for k := range s.Options {
			if strings.TrimSpace(k) == "" {
				return fmt.Errorf("engine %q has an empty option name", s.Name)
			}
		}
	}
	return nil
}

func run(t Tournament, fastchessOverride string) error {
	c := t.Shared()
	specs := t.Engines()

	store, err := registry.Default()
	if err != nil {
		return err
	}
	entries := make([]registry.Entry, len(specs))
	for i, s := range specs {
		e, err := store.Get(s.Name)
		if errors.Is(err, registry.ErrNotFound) {
			return fmt.Errorf("%s engine %q is not registered", s.Role, s.Name)
		}
		if err != nil {
			return err
		}
		entries[i] = e
	}

	bin, err := findFastchess(fastchessOverride)
	if err != nil {
		return err
	}

	outDir := c.resolve(c.OutputDir)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	args := buildArgs(c, specs, entries, outDir, t.Args())
	fmt.Fprintln(os.Stderr, "running:", bin, quoteArgs(args))

	cmd := exec.Command(bin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

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

// resolves the binary
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

func buildArgs(c *Common, specs []EngineSpec, entries []registry.Entry, outDir string, typeArgs []string) []string {
	var args []string

	for i, e := range entries {
		args = append(args, engineArgs(e, specs[i])...)
	}

	if c.TC != "" {
		args = append(args, "-each", "tc="+c.TC)
	} else {
		args = append(args, "-each", "nodes="+strconv.Itoa(c.Nodes))
	}

	if o := c.Openings; o != nil {
		args = append(args, "-openings",
			"file="+c.resolve(o.File), "format="+o.Format, "order="+o.Order)
	}

	args = append(args,
		"-rounds", strconv.Itoa(c.Rounds),
		"-games", strconv.Itoa(c.Games),
		"-concurrency", strconv.Itoa(c.Concurrency),
	)
	args = append(args, typeArgs...)
	args = append(args,
		"-pgnout", "file="+filepath.Join(outDir, "games.pgn"),
		"-log", "file="+filepath.Join(outDir, "games.log"), "engine=true",
		"-recover",
	)

	return append(args, c.ExtraArgs...)
}

func engineArgs(e registry.Entry, spec EngineSpec) []string {
	args := []string{"-engine", "cmd=" + e.Path, "name=" + e.Name + spec.Suffix}

	keys := make([]string, 0, len(spec.Options))
	for k := range spec.Options {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		args = append(args, fmt.Sprintf("option.%s=%v", k, spec.Options[k]))
	}
	return args
}

func fmtFloat(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }

func quoteArgs(args []string) string {
	out := make([]string, len(args))
	for i, a := range args {
		if strings.ContainsAny(a, " \t") {
			a = strconv.Quote(a)
		}
		out[i] = a
	}
	return strings.Join(out, " ")
}

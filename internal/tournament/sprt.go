package tournament

import (
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

// loads the config, resolves both engines from the registry and runs fastchess
func RunSPRT(configPath, fastchessOverride string) error {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return err
	}

	store, err := registry.Default()
	if err != nil {
		return err
	}
	rc, err := lookup(store, cfg.RC.Name, "rc")
	if err != nil {
		return err
	}
	base, err := lookup(store, cfg.Base.Name, "base")
	if err != nil {
		return err
	}

	bin, err := findFastchess(fastchessOverride)
	if err != nil {
		return err
	}

	outDir := cfg.resolve(cfg.OutputDir)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	args := buildArgs(cfg, rc, base, outDir)
	fmt.Fprintln(os.Stderr, "running:", bin, quoteArgs(args))

	cmd := exec.Command(bin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	defer signal.Stop(sigs)

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("fastchess exited with code %d", exitErr.ExitCode())
		}
		return fmt.Errorf("run fastchess: %w", err)
	}
	return nil
}

func lookup(store *registry.Store, name, role string) (registry.Entry, error) {
	e, err := store.Get(name)
	if errors.Is(err, registry.ErrNotFound) {
		return e, fmt.Errorf("%s engine %q is not registered", role, name)
	}
	return e, err
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

func buildArgs(c *Config, rc, base registry.Entry, outDir string) []string {
	var args []string

	args = append(args, engineArgs(rc, c.RC)...)
	args = append(args, engineArgs(base, c.Base)...)

	if c.TC != "" {
		args = append(args, "-each", "tc="+c.TC)
	} else {
		args = append(args, "-each", "nodes="+strconv.Itoa(c.Nodes))
	}

	if o := c.Openings; o != nil {
		args = append(args, "-openings",
			"file="+c.resolve(o.File), "format="+o.Format, "order=random")
	}

	args = append(args,
		"-rounds", strconv.Itoa(c.Rounds),
		"-games", strconv.Itoa(c.Games),
		"-concurrency", strconv.Itoa(c.Concurrency),
		"-sprt",
		"elo0="+fmtFloat(*c.SPRT.Elo0),
		"elo1="+fmtFloat(*c.SPRT.Elo1),
		"alpha="+fmtFloat(*c.SPRT.Alpha),
		"beta="+fmtFloat(*c.SPRT.Beta),
		"-pgnout", "file="+filepath.Join(outDir, "games.pgn"),
		"-log", "file="+filepath.Join(outDir, "games.log"), "engine=true",
		"-recover",
	)

	return append(args, c.ExtraArgs...)
}

func engineArgs(e registry.Entry, spec EngineSpec) []string {
	args := []string{"-engine", "cmd=" + e.Path, "name=" + e.Name}

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

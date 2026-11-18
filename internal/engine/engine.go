package engine

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/nchandur/ranga-tools/internal/registry"
	"github.com/nchandur/ranga-tools/internal/uci"
)

// verifies the engine at path via the UCI handshake and records it
func Register(path, nameOverride string, force bool) (registry.Entry, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return registry.Entry{}, err
	}
	if fi, err := os.Stat(abs); err != nil {
		return registry.Entry{}, fmt.Errorf("engine path: %w", err)
	} else if fi.IsDir() {
		return registry.Entry{}, fmt.Errorf("engine path %s is a directory", abs)
	}

	info, err := uci.Verify(abs)
	if err != nil {
		return registry.Entry{}, fmt.Errorf("verify engine: %w", err)
	}
	name := strings.TrimSpace(nameOverride)
	if name == "" {
		name = info.Name
	}
	if name == "" {
		return registry.Entry{}, uci.ErrNoName
	}

	store, err := registry.Default()
	if err != nil {
		return registry.Entry{}, err
	}
	entry := registry.Entry{Name: name, Author: info.Author, Path: abs}
	if err := store.Add(entry, force); errors.Is(err, registry.ErrExists) {
		return registry.Entry{}, fmt.Errorf("engine %q is already registered; use --force to overwrite it", name)
	} else if err != nil {
		return registry.Entry{}, err
	}
	return entry, nil
}

func Deregister(name string) error {
	store, err := registry.Default()
	if err != nil {
		return err
	}
	return notRegistered(name, store.Remove(name))
}

func List() ([]registry.Entry, error) {
	store, err := registry.Default()
	if err != nil {
		return nil, err
	}
	return store.List()
}

func Lookup(name string) (registry.Entry, error) {
	store, err := registry.Default()
	if err != nil {
		return registry.Entry{}, err
	}
	entry, err := store.Get(name)
	return entry, notRegistered(name, err)
}

// runs registered engine attached to the terminal until it exits
func Use(name string) error {
	entry, err := Lookup(name)
	if err != nil {
		return err
	}
	cmd := exec.Command(entry.Path)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	defer signal.Stop(sigs)

	if err := cmd.Run(); err != nil {
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			return fmt.Errorf("engine %q exited with code %d", entry.Name, exitErr.ExitCode())
		}
		return fmt.Errorf("run engine %q: %w", entry.Name, err)
	}
	return nil
}

func notRegistered(name string, err error) error {
	if errors.Is(err, registry.ErrNotFound) {
		return fmt.Errorf("no engine named %q is registered", name)
	}
	return err
}

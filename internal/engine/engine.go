package engine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nchandur/ranga-tools/internal/registry"
	"github.com/nchandur/ranga-tools/internal/uci"
)

// verifies engine at path via UCI handshake and records it
func Register(path, nameOverride string, force bool) (registry.Entry, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return registry.Entry{}, fmt.Errorf("resolve path: %w", err)
	}
	fi, err := os.Stat(abs)
	if err != nil {
		return registry.Entry{}, fmt.Errorf("engine path: %w", err)
	}
	if fi.IsDir() {
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
	if err := store.Add(entry, force); err != nil {
		if errors.Is(err, registry.ErrExists) {
			return registry.Entry{}, fmt.Errorf("engine %q is already registered; use --force to overwrite it", name)
		}
		return registry.Entry{}, err
	}
	return entry, nil
}

// deregisters engine by name
func Deregister(name string) error {
	store, err := registry.Default()
	if err != nil {
		return err
	}
	if err := store.Remove(name); err != nil {
		if errors.Is(err, registry.ErrNotFound) {
			return fmt.Errorf("no engine named %q is registered", name)
		}
		return err
	}
	return nil
}

// returns all registered engines
func List() ([]registry.Entry, error) {
	store, err := registry.Default()
	if err != nil {
		return nil, err
	}
	return store.List()
}

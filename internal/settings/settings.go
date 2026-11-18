package settings

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nchandur/ranga-tools/internal/jsonfile"
	"github.com/nchandur/ranga-tools/internal/registry"
)

const KeyFastchessPath = "fastchess-path"

type Settings struct {
	FastchessPath string `json:"fastchess_path,omitempty"`
}

func path() (string, error) {
	dir, err := registry.Dir()
	return filepath.Join(dir, "config.json"), err
}

func Load() (Settings, error) {
	var s Settings
	p, err := path()
	if err != nil {
		return s, err
	}
	return s, jsonfile.Load(p, &s)
}

// stores a setting
func Set(key, value string) error {
	if key != KeyFastchessPath {
		return fmt.Errorf("unknown setting %q (known: %s)", key, KeyFastchessPath)
	}
	abs, err := filepath.Abs(value)
	if err != nil {
		return err
	}
	fi, err := os.Stat(abs)
	if err != nil {
		return fmt.Errorf("fastchess path: %w", err)
	}
	if fi.IsDir() {
		return fmt.Errorf("fastchess path %s is a directory", abs)
	}
	s, err := Load()
	if err != nil {
		return err
	}
	s.FastchessPath = abs
	p, err := path()
	if err != nil {
		return err
	}
	return jsonfile.Save(p, s)
}

// returns a setting's value, empty if unset.
func Get(key string) (string, error) {
	if key != KeyFastchessPath {
		return "", fmt.Errorf("unknown setting %q (known: %s)", key, KeyFastchessPath)
	}
	s, err := Load()
	return s.FastchessPath, err
}

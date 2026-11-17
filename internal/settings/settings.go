package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/nchandur/ranga-tools/internal/registry"
)

const fileName = "config.json"

const KeyFastchessPath = "fastchess-path"

type Settings struct {
	FastchessPath string `json:"fastchess_path,omitempty"`
}

func filePath() (string, error) {
	dir, err := registry.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fileName), nil
}

func Load() (Settings, error) {
	var s Settings
	p, err := filePath()
	if err != nil {
		return s, err
	}
	data, err := os.ReadFile(p)
	if errors.Is(err, fs.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return s, fmt.Errorf("read %s: %w", p, err)
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return s, fmt.Errorf("parse %s: %w", p, err)
	}
	return s, nil
}

func save(s Settings) error {
	p, err := filePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, p); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("replace %s: %w", p, err)
	}
	return nil
}

// stores a setting
func Set(key, value string) error {
	switch key {
	case KeyFastchessPath:
		abs, err := filepath.Abs(value)
		if err != nil {
			return fmt.Errorf("resolve path: %w", err)
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
		return save(s)
	default:
		return fmt.Errorf("unknown setting %q (known: %s)", key, KeyFastchessPath)
	}
}

// returns a setting's value
func Get(key string) (string, error) {
	switch key {
	case KeyFastchessPath:
		s, err := Load()
		return s.FastchessPath, err
	default:
		return "", fmt.Errorf("unknown setting %q (known: %s)", key, KeyFastchessPath)
	}
}

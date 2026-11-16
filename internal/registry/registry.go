package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	appDir   = "ranga-tools" // 	folder created inside os.UserConfigDir()
	fileName = "engines.json"
)

var (
	ErrExists   = errors.New("engine already registered")
	ErrNotFound = errors.New("engine not found")
	ErrBadName  = errors.New("engine name is empty")
)

// registered engine
type Entry struct {
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Author string `json:"author"`
	Path   string `json:"path"`
}

type file struct {
	Engines []Entry `json:"engines"`
}

// persists entries in a JSON file
type Store struct {
	path string
}

func Default() (*Store, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("locate user config dir: %w", err)
	}
	return &Store{path: filepath.Join(dir, appDir, fileName)}, nil
}

// lowercases name and replaces whitespace with a single hyphen
func Slug(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(name), "-"))
}

// add engine
func (s *Store) Add(e Entry, force bool) error {
	e.Slug = Slug(e.Name)
	if e.Slug == "" {
		return ErrBadName
	}

	entries, err := s.load()
	if err != nil {
		return err
	}

	for i := range entries {
		if entries[i].Slug == e.Slug {
			if !force {
				return ErrExists
			}
			entries[i] = e
			return s.save(entries)
		}
	}
	return s.save(append(entries, e))
}

// deletes the entry whose slug matches
func (s *Store) Remove(name string) error {
	slug := Slug(name)
	entries, err := s.load()
	if err != nil {
		return err
	}
	for i := range entries {
		if entries[i].Slug == slug {
			return s.save(append(entries[:i], entries[i+1:]...))
		}
	}
	return ErrNotFound
}

// returns all entries sorted by slug
func (s *Store) List() ([]Entry, error) {
	entries, err := s.load()
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Slug < entries[j].Slug })
	return entries, nil
}

func (s *Store) load() ([]Entry, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", s.path, err)
	}
	var f file
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse %s: %w", s.path, err)
	}
	return f.Engines, nil
}

func (s *Store) save(entries []Entry) error {
	if entries == nil {
		entries = []Entry{}
	}
	data, err := json.MarshalIndent(file{Engines: entries}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode registry: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("replace %s: %w", s.path, err)
	}
	return nil
}

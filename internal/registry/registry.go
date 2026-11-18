package registry

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/nchandur/ranga-tools/internal/jsonfile"
)

const appDir = "ranga-tools"

var (
	ErrExists   = errors.New("engine already registered")
	ErrNotFound = errors.New("engine not found")
	ErrBadName  = errors.New("engine name is empty")
)

type Entry struct {
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	Author string `json:"author"`
	Path   string `json:"path"`
}

type file struct {
	Engines []Entry `json:"engines"`
}

type Store struct{ path string }

// returns folder where all of the tool's files live
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate user config dir: %w", err)
	}
	return filepath.Join(base, appDir), nil
}

func Default() (*Store, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}
	return &Store{filepath.Join(dir, "engines.json")}, nil
}

// lowercases name and joins its words with hyphens
func Slug(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(name), "-"))
}

func (s *Store) load() ([]Entry, error) {
	var f file
	err := jsonfile.Load(s.path, &f)
	return f.Engines, err
}

func (s *Store) save(entries []Entry) error {
	return jsonfile.Save(s.path, file{append([]Entry{}, entries...)}) // never "null"
}

func find(entries []Entry, slug string) int {
	return slices.IndexFunc(entries, func(e Entry) bool { return e.Slug == slug })
}

// stores e, replacing an existing entry with the same slug only if force is set
func (s *Store) Add(e Entry, force bool) error {
	if e.Slug = Slug(e.Name); e.Slug == "" {
		return ErrBadName
	}
	entries, err := s.load()
	if err != nil {
		return err
	}
	if i := find(entries, e.Slug); i < 0 {
		entries = append(entries, e)
	} else if force {
		entries[i] = e
	} else {
		return ErrExists
	}
	return s.save(entries)
}

func (s *Store) Remove(name string) error {
	entries, err := s.load()
	if err != nil {
		return err
	}
	i := find(entries, Slug(name))
	if i < 0 {
		return ErrNotFound
	}
	return s.save(slices.Delete(entries, i, i+1))
}

// returns all entries sorted by slug
func (s *Store) List() ([]Entry, error) {
	entries, err := s.load()
	slices.SortFunc(entries, func(a, b Entry) int { return strings.Compare(a.Slug, b.Slug) })
	return entries, err
}

func (s *Store) Get(name string) (Entry, error) {
	entries, err := s.load()
	if err != nil {
		return Entry{}, err
	}
	i := find(entries, Slug(name))
	if i < 0 {
		return Entry{}, ErrNotFound
	}
	return entries[i], nil
}

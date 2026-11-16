package tournament

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/nchandur/ranga-tools/internal/registry"
)

const (
	defaultNodes     = 5000
	defaultRounds    = 10
	defaultGames     = 2
	defaultOutputDir = "output"
)

type EngineSpec struct {
	Name    string         `json:"name"`
	Options map[string]any `json:"options"`
}

type SPRT struct {
	Elo0  *float64 `json:"elo0"`
	Elo1  *float64 `json:"elo1"`
	Alpha *float64 `json:"alpha"`
	Beta  *float64 `json:"beta"`
}

type Openings struct {
	File   string `json:"file"`
	Format string `json:"format"`
}

type Config struct {
	RC   EngineSpec `json:"rc"`
	Base EngineSpec `json:"base"`

	TC    string `json:"tc"`
	Nodes int    `json:"nodes"`

	Rounds      int `json:"rounds"`
	Games       int `json:"games"`
	Concurrency int `json:"concurrency"`

	SPRT     *SPRT     `json:"sprt"`
	Openings *Openings `json:"openings"`

	OutputDir string   `json:"output_dir"`
	ExtraArgs []string `json:"extra_args"`

	dir string
}

// reads, defaults and validates the file at path.
func LoadConfig(path string) (*Config, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve config path: %w", err)
	}
	f, err := os.Open(abs)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	dec.UseNumber()

	var c Config
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", abs, err)
	}
	c.dir = filepath.Dir(abs)

	if err := c.applyDefaults(); err != nil {
		return nil, fmt.Errorf("%s: %w", abs, err)
	}
	if err := c.validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", abs, err)
	}
	return &c, nil
}

func (c *Config) resolve(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(c.dir, p)
}

func (c *Config) applyDefaults() error {
	if c.Rounds < 0 || c.Games < 0 || c.Concurrency < 0 || c.Nodes < 0 {
		return fmt.Errorf("rounds, games, concurrency and nodes must not be negative")
	}
	if c.Rounds == 0 {
		c.Rounds = defaultRounds
	}
	if c.Games == 0 {
		c.Games = defaultGames
	}
	if c.Concurrency == 0 {
		c.Concurrency = max(runtime.NumCPU()-1, 1)
	}
	if c.OutputDir == "" {
		c.OutputDir = defaultOutputDir
	}

	if c.TC != "" {
		c.Nodes = 0
	} else if c.Nodes == 0 {
		c.Nodes = defaultNodes
	}

	if c.Openings != nil {
		if c.Openings.Format == "" {
			c.Openings.Format = strings.TrimPrefix(strings.ToLower(filepath.Ext(c.Openings.File)), ".")
		}
	}
	return nil
}

func (c *Config) validate() error {
	if c.RC.Name == "" {
		return fmt.Errorf(`"rc.name" is required`)
	}
	if c.Base.Name == "" {
		return fmt.Errorf(`"base.name" is required`)
	}
	if registry.Slug(c.RC.Name) == registry.Slug(c.Base.Name) {
		return fmt.Errorf("rc and base must be different engines")
	}

	s := c.SPRT
	if s == nil || s.Elo0 == nil || s.Elo1 == nil || s.Alpha == nil || s.Beta == nil {
		return fmt.Errorf(`"sprt" requires elo0, elo1, alpha and beta`)
	}
	if *s.Elo0 >= *s.Elo1 {
		return fmt.Errorf("sprt.elo0 (%g) must be less than sprt.elo1 (%g)", *s.Elo0, *s.Elo1)
	}
	if *s.Alpha <= 0 || *s.Alpha >= 1 || *s.Beta <= 0 || *s.Beta >= 1 {
		return fmt.Errorf("sprt.alpha and sprt.beta must be between 0 and 1")
	}

	if o := c.Openings; o != nil {
		if o.File == "" {
			return fmt.Errorf(`"openings.file" is required when "openings" is set`)
		}
		if o.Format != "epd" && o.Format != "pgn" {
			return fmt.Errorf("openings.format must be \"epd\" or \"pgn\" (got %q)", o.Format)
		}
	}

	for _, spec := range []EngineSpec{c.RC, c.Base} {
		for k := range spec.Options {
			if strings.TrimSpace(k) == "" {
				return fmt.Errorf("engine %q has an empty option name", spec.Name)
			}
		}
	}
	return nil
}

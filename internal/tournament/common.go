package tournament

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	defaultNodes     = 5000
	defaultRounds    = 5000
	defaultGames     = 2
	defaultOutputDir = "output"
)

type EngineSpec struct {
	Name    string         `json:"name"`
	Options map[string]any `json:"options"`
	Role    string         `json:"-"`
	Suffix  string         `json:"-"`
}

type Openings struct {
	File   string `json:"file"`
	Format string `json:"format"`
	Order  string `json:"order"`
}

// holds only what every tournament type needs
type Common struct {
	TC    string `json:"tc"`
	Nodes int    `json:"nodes"`

	Rounds      int `json:"rounds"`
	Games       int `json:"games"`
	Concurrency int `json:"concurrency"`

	Openings *Openings `json:"openings"`

	OutputDir string   `json:"output_dir"`
	ExtraArgs []string `json:"extra_args"`

	dir string
}

func (c *Common) resolve(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(c.dir, p)
}

func (c *Common) applyDefaults() error {
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
		if c.Openings.Order == "" {
			c.Openings.Order = "random"
		}
	}
	return nil
}

func (c *Common) validate() error {
	if o := c.Openings; o != nil {
		if o.File == "" {
			return fmt.Errorf(`"openings.file" is required when "openings" is set`)
		}
		if o.Format != "epd" && o.Format != "pgn" {
			return fmt.Errorf("openings.format must be \"epd\" or \"pgn\" (got %q)", o.Format)
		}
		if o.Order != "random" && o.Order != "sequential" {
			return fmt.Errorf("openings.order must be \"random\" or \"sequential\" (got %q)", o.Order)
		}
	}
	return nil
}

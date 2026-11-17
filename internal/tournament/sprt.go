package tournament

import "fmt"

type sprtBounds struct {
	Elo0  *float64 `json:"elo0"`
	Elo1  *float64 `json:"elo1"`
	Alpha *float64 `json:"alpha"`
	Beta  *float64 `json:"beta"`
}

type sprtConfig struct {
	Common
	RC   EngineSpec  `json:"rc"`
	Base EngineSpec  `json:"base"`
	SPRT *sprtBounds `json:"sprt"`
}

func (c *sprtConfig) Shared() *Common { return &c.Common }

func (c *sprtConfig) Engines() []EngineSpec {
	rc, base := c.RC, c.Base
	rc.Role, base.Role = "rc", "base"
	return []EngineSpec{rc, base}
}

func (c *sprtConfig) Validate() error {
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
	return nil
}

func (c *sprtConfig) Args() []string {
	return []string{
		"-sprt",
		"elo0=" + fmtFloat(*c.SPRT.Elo0),
		"elo1=" + fmtFloat(*c.SPRT.Elo1),
		"alpha=" + fmtFloat(*c.SPRT.Alpha),
		"beta=" + fmtFloat(*c.SPRT.Beta),
	}
}

// runs an SPRT between the rc and base engines in the config file
func RunSPRT(configPath, fastchessOverride string) error {
	return execute(configPath, fastchessOverride, &sprtConfig{})
}

package tournament

import (
	"errors"
	"fmt"
)

type sprtConfig struct {
	Common
	RC   EngineSpec `json:"rc"`
	Base EngineSpec `json:"base"`
	SPRT *struct {
		Elo0, Elo1, Alpha, Beta *float64
	} `json:"sprt"`
}

func (c *sprtConfig) Engines() []EngineSpec {
	rc, base := c.RC, c.Base
	rc.Role, base.Role = "rc", "base"
	return []EngineSpec{rc, base}
}

func (c *sprtConfig) Validate() error {
	s := c.SPRT
	if s == nil || s.Elo0 == nil || s.Elo1 == nil || s.Alpha == nil || s.Beta == nil {
		return errors.New(`"sprt" requires elo0, elo1, alpha and beta`)
	}
	if *s.Elo0 >= *s.Elo1 {
		return fmt.Errorf("sprt.elo0 (%g) must be less than sprt.elo1 (%g)", *s.Elo0, *s.Elo1)
	}
	if *s.Alpha <= 0 || *s.Alpha >= 1 || *s.Beta <= 0 || *s.Beta >= 1 {
		return errors.New("sprt.alpha and sprt.beta must be between 0 and 1")
	}
	return nil
}

func (c *sprtConfig) Args() []string {
	s := c.SPRT
	return []string{"-sprt",
		fmt.Sprintf("elo0=%g", *s.Elo0), fmt.Sprintf("elo1=%g", *s.Elo1),
		fmt.Sprintf("alpha=%g", *s.Alpha), fmt.Sprintf("beta=%g", *s.Beta)}
}

// runs an SPRT between the rc and base engines in the config file
func RunSPRT(configPath, fastchessOverride string) error {
	return execute(configPath, fastchessOverride, &sprtConfig{})
}

package tournament

// pits one registered engine against itself
type selfTestConfig struct {
	Common
	Engine EngineSpec `json:"engine"`
}

func (c *selfTestConfig) Shared() *Common {
	return &c.Common
}

// returns the same engine twice, suffixed so fastchess sees two different names
func (c *selfTestConfig) Engines() []EngineSpec {
	a, b := c.Engine, c.Engine
	a.Role, a.Suffix = "engine", "-a"
	b.Role, b.Suffix = "engine", "-b"
	return []EngineSpec{a, b}
}

func (c *selfTestConfig) Validate() error {
	return nil
}

func (c *selfTestConfig) Args() []string {
	return nil
}

// plays the configured engine against itself
func RunSelfTest(configPath, fastchessOverride string) error {
	c := &selfTestConfig{Rounds: 50000, Games: 1}
	return execute(configPath, fastchessOverride, c)
}

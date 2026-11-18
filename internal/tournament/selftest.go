package tournament

type selfTestConfig struct {
	Common
	Engine EngineSpec `json:"engine"`
}

// returns the same engine twice, suffixed so fastchess sees two names
func (c *selfTestConfig) Engines() []EngineSpec {
	a, b := c.Engine, c.Engine
	a.Role, a.Suffix = "engine", "-a"
	b.Role, b.Suffix = "engine", "-b"
	return []EngineSpec{a, b}
}

// plays the configured engine against itself
func RunSelfTest(configPath, fastchessOverride string) error {
	return execute(configPath, fastchessOverride, &selfTestConfig{Common: Common{Rounds: 50000, Games: 1}})
}

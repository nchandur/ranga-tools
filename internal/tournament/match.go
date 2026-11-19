package tournament

// plain match between two different registered engines.
type matchConfig struct {
	Common
	Engine1 EngineSpec `json:"engine1"`
	Engine2 EngineSpec `json:"engine2"`
}

func (c *matchConfig) Engines() []EngineSpec {
	a, b := c.Engine1, c.Engine2
	a.Role, b.Role = "engine1", "engine2"
	return []EngineSpec{a, b}
}

// plays the two engines in the config file against each other
func RunMatch(configPath, fastchessOverride string) error {
	return execute(configPath, fastchessOverride, &matchConfig{})
}

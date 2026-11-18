package analysis

import "fmt"

type Strength string

const (
	Low    Strength = "low"
	Medium Strength = "medium"
	High   Strength = "high"
)

// engine settings behind strength level
type Preset struct{ Depth, HashMB int }

var presets = map[Strength]Preset{
	Low:    {8, 64},
	Medium: {12, 64},
	High:   {16, 64},
}

func ParseStrength(s string) (Strength, error) {
	if _, ok := presets[Strength(s)]; !ok {
		return "", fmt.Errorf("invalid strength %q (use low, medium or high)", s)
	}
	return Strength(s), nil
}

func (s Strength) Preset() Preset {
	return presets[s]
}

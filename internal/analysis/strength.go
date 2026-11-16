package analysis

import "fmt"

type Strength string

const (
	Low    Strength = "low"
	Medium Strength = "medium"
	High   Strength = "high"
)

// holds the engine settings behind a strength level
type Preset struct {
	Depth   int
	HashMB  int
	Threads int
}

// lists the accepted flag values
var Strengths = []string{string(Low), string(Medium), string(High)}

func ParseStrength(s string) (Strength, error) {
	switch st := Strength(s); st {
	case Low, Medium, High:
		return st, nil
	}
	return "", fmt.Errorf("invalid strength %q (use low, medium or high)", s)
}

// returns the depth/hash values for level
func (s Strength) Preset() (Preset, error) {
	switch s {
	case Low:
		return Preset{Depth: 8, HashMB: 64, Threads: 1}, nil
	case Medium:
		return Preset{Depth: 12, HashMB: 64, Threads: 1}, nil
	case High:
		return Preset{Depth: 16, HashMB: 64, Threads: 1}, nil
	}
	return Preset{}, fmt.Errorf("unknown strength %q", s)
}

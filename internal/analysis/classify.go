package analysis

import (
	"math"

	"github.com/nchandur/ranga-tools/internal/uci"
)

type Class string

const (
	Best       Class = "best"
	Excellent  Class = "excellent"
	Good       Class = "good"
	Inaccuracy Class = "inaccuracy"
	Mistake    Class = "mistake"
	Blunder    Class = "blunder"
)

var Classes = []Class{
	Blunder,
	Mistake,
	Inaccuracy,
	Good,
	Excellent,
	Best,
}

// most win-probability points a move can lose and still earn the class
var bands = []struct {
	class   Class
	maxLoss float64
}{
	{Best, 0.5},
	{Excellent, 2},
	{Good, 5},
	{Inaccuracy, 10},
	{Mistake, 20},
}

func Classify(loss float64) Class {
	for _, b := range bands {
		if loss < b.maxLoss {
			return b.class
		}
	}
	return Blunder
}

// converts a score to that side's win probability
func WinPercent(s uci.Score) float64 {
	if s.IsMate {
		if s.Mate > 0 {
			return 100
		}
		return 0
	}
	return 50 + 50*(2/(1+math.Exp(-0.00368208*float64(s.CP)))-1)
}

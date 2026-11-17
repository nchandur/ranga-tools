package analysis

import (
	"math"

	"github.com/nchandur/ranga-tools/internal/uci"
)

// move's quality label
type Class string

const (
	Best       Class = "best"
	Excellent  Class = "excellent"
	Good       Class = "good"
	Inaccuracy Class = "inaccuracy"
	Mistake    Class = "mistake"
	Blunder    Class = "blunder"
)

// bands lists each class with the most win-probability points
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

// maps lost win probability to label
func Classify(loss float64) Class {
	for _, b := range bands {
		if loss < b.maxLoss {
			return b.class
		}
	}
	return Blunder
}

// converts an engine score to that side's win probability
func WinPercent(s uci.Score) float64 {
	if s.IsMate {
		if s.Mate > 0 {
			return 100
		}
		return 0 // mated, or about to be
	}
	return 50 + 50*(2/(1+math.Exp(-0.00368208*float64(s.CP)))-1)
}

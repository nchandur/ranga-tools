package analysis

import (
	"math"

	"github.com/nchandur/ranga-tools/internal/uci"
)

// move's quality label
type Class string

const (
	Good    Class = "good"
	Ok      Class = "ok"
	Bad     Class = "bad"
	Blunder Class = "blunder"
)

// win-probability points lost by the mover
const (
	okLoss      = 2.0  // below this: good
	badLoss     = 8.0  // below this: ok
	blunderLoss = 20.0 // below this: bad; at or above: blunder
)

// maps lost win probability to label
func Classify(loss float64) Class {
	switch {
	case loss < okLoss:
		return Good
	case loss < badLoss:
		return Ok
	case loss < blunderLoss:
		return Bad
	default:
		return Blunder
	}
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

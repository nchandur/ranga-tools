package analysis

import "math"

// each player's game accuracy
type Accuracy struct {
	White float64 `json:"white"`
	Black float64 `json:"black"`
}

// converts win-probability points lost into score.
func moveAccuracy(loss float64) float64 {
	a := 103.1668*math.Exp(-0.04354*loss) - 3.1669
	return math.Min(100, math.Max(0, a))
}

// scores each side's moves
func computeAccuracy(moves []Move) Accuracy {
	n := len(moves)
	if n == 0 {
		return Accuracy{}
	}

	points := make([]float64, 0, n+1)
	for _, m := range moves {
		points = append(points, whiteWin(m.Side, m.WinBefore))
	}
	last := moves[n-1]
	points = append(points, whiteWin(last.Side, last.WinAfter))

	weights := volatilityWeights(points, n)

	var accs [2][]float64
	var ws [2][]float64
	for i, m := range moves {
		side := 0
		if m.Side == "black" {
			side = 1
		}
		accs[side] = append(accs[side], moveAccuracy(m.Loss))
		ws[side] = append(ws[side], weights[i])
	}
	return Accuracy{
		White: combine(accs[0], ws[0]),
		Black: combine(accs[1], ws[1]),
	}
}

func whiteWin(side string, moverWin float64) float64 {
	if side == "white" {
		return moverWin
	}
	return 100 - moverWin
}

// returns one weight per move
func volatilityWeights(points []float64, n int) []float64 {
	size := min(max(len(points)/10, 2), 8)

	var windows [][]float64
	first := points[:min(size, len(points))]
	for i := 0; i < size-2; i++ {
		windows = append(windows, first)
	}
	for i := 0; i+size <= len(points); i++ {
		windows = append(windows, points[i:i+size])
	}

	weights := make([]float64, n)
	for i := range weights {
		w := 0.5
		if i < len(windows) {
			w = math.Min(12, math.Max(0.5, stddev(windows[i])))
		}
		weights[i] = w
	}
	return weights
}

func stddev(xs []float64) float64 {
	var sum float64
	for _, x := range xs {
		sum += x
	}
	mean := sum / float64(len(xs))
	var sq float64
	for _, x := range xs {
		sq += (x - mean) * (x - mean)
	}
	return math.Sqrt(sq / float64(len(xs)))
}

// combine averages volatility-weighted mean and harmonic mean
func combine(accs, weights []float64) float64 {
	if len(accs) == 0 {
		return 0
	}
	var wSum, wTotal, inv float64
	for i, a := range accs {
		wSum += a * weights[i]
		wTotal += weights[i]
		inv += 1 / math.Max(a, 0.01)
	}
	weighted := wSum / wTotal
	harmonic := float64(len(accs)) / inv
	return math.Round((weighted+harmonic)/2*10) / 10
}

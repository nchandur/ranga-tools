package analysis

import "math"

type Accuracy struct {
	White float64 `json:"white"`
	Black float64 `json:"black"`
}

// converts win-probability points lost into a 0-100 score
func moveAccuracy(loss float64) float64 {
	return min(100, max(0, 103.1668*math.Exp(-0.04354*loss)-3.1669))
}

// expects at least one move
func computeAccuracy(moves []Move) Accuracy {
	n := len(moves)

	points := make([]float64, n+1)
	for i, m := range moves {
		points[i] = whiteWin(m.Side, m.WinBefore)
	}
	points[n] = whiteWin(moves[n-1].Side, moves[n-1].WinAfter)

	size := min(max(len(points)/10, 2), 8)

	var accs, weights [2][]float64
	for i, m := range moves {
		side := 0
		if m.Side == "black" {
			side = 1
		}
		// weight: volatility of the window of positions around the move
		start := max(i-(size-2), 0)
		w := min(12, max(0.5, stddev(points[start:start+size])))

		accs[side] = append(accs[side], moveAccuracy(m.Loss))
		weights[side] = append(weights[side], w)
	}
	return Accuracy{combine(accs[0], weights[0]), combine(accs[1], weights[1])}
}

func whiteWin(side string, moverWin float64) float64 {
	if side == "white" {
		return moverWin
	}
	return 100 - moverWin
}

func stddev(xs []float64) float64 {
	var sum, sq float64
	for _, x := range xs {
		sum += x
	}
	mean := sum / float64(len(xs))
	for _, x := range xs {
		sq += (x - mean) * (x - mean)
	}
	return math.Sqrt(sq / float64(len(xs)))
}

// combine averages the volatility-weighted mean and the harmonic mean.
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
	harmonic := float64(len(accs)) / inv
	return math.Round((wSum/wTotal+harmonic)/2*10) / 10
}

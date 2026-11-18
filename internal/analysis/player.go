package analysis

import "math"

type PlayerGame struct {
	Game     int     `json:"game"`
	Opponent string  `json:"opponent"`
	Color    string  `json:"color"`
	Result   string  `json:"result"` // win, draw, loss or unknown
	Plies    int     `json:"plies"`
	Accuracy float64 `json:"accuracy"`
	Summary  Counts  `json:"summary"`
}

type Split struct {
	Games    int     `json:"games"`
	Accuracy float64 `json:"accuracy"` // mean of per-game accuracies
	Summary  Counts  `json:"summary"`
}

type PlayerReport struct {
	Name    string       `json:"name"`
	Wins    int          `json:"wins"`
	Draws   int          `json:"draws"`
	Losses  int          `json:"losses"`
	Unknown int          `json:"unknown_results"`
	Overall Split        `json:"overall"`
	White   Split        `json:"as_white"`
	Black   Split        `json:"as_black"`
	Games   []PlayerGame `json:"games"`
}

// aggregates the reports of games the named player took part in
func NewPlayerReport(name string, reports []*Report) *PlayerReport {
	p := &PlayerReport{Name: name}
	splits := [3]*Split{&p.Overall, &p.White, &p.Black}
	var sums [3]float64
	for _, s := range splits {
		s.Summary = Counts{}
	}

	for _, r := range reports {
		color, slot := "white", 1
		acc, counts, opp := r.Accuracy.White, r.Summary.White, r.Black
		if r.White != name {
			color, slot = "black", 2
			acc, counts, opp = r.Accuracy.Black, r.Summary.Black, r.White
		}

		res := playerResult(r.Result, color)
		switch res {
		case "win":
			p.Wins++
		case "draw":
			p.Draws++
		case "loss":
			p.Losses++
		default:
			p.Unknown++
		}

		for _, i := range []int{0, slot} {
			splits[i].Games++
			sums[i] += acc
			for c, n := range counts {
				splits[i].Summary[c] += n
			}
		}
		p.Games = append(p.Games, PlayerGame{r.Game, opp, color, res, len(r.Moves), acc, counts})
	}

	for i, s := range splits {
		if s.Games > 0 {
			s.Accuracy = math.Round(sums[i]/float64(s.Games)*10) / 10
		}
	}
	return p
}

func playerResult(result, color string) string {
	switch {
	case result == "Draw":
		return "draw"
	case result == "White Wins" && color == "white", result == "Black Wins" && color == "black":
		return "win"
	case result == "White Wins", result == "Black Wins":
		return "loss"
	}
	return "unknown"
}

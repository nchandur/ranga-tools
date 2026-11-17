package analysis

import "math"

// one game seen from the player's side
type PlayerGame struct {
	Game     int     `json:"game"`
	Opponent string  `json:"opponent"`
	Color    string  `json:"color"`  // "white" or "black"
	Result   string  `json:"result"` // "win", "draw", "loss" or "unknown"
	Plies    int     `json:"plies"`
	Accuracy float64 `json:"accuracy"`
	Summary  Counts  `json:"summary"`
}

// player's aggregate over a set of games
type Split struct {
	Games    int     `json:"games"`
	Accuracy float64 `json:"accuracy"`
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

// accumulates one player's games
type PlayerAgg struct {
	rep  PlayerReport
	sums [3]float64
}

func NewPlayerAgg(name string) *PlayerAgg {
	return &PlayerAgg{rep: PlayerReport{Name: name}}
}

// folds in report where the player had given colour
func (p *PlayerAgg) Add(r *Report, color string) {
	acc, counts, opp := r.Accuracy.White, r.Summary.White, r.Black
	colorSplit, colorSlot := &p.rep.White, 1
	if color == "black" {
		acc, counts, opp = r.Accuracy.Black, r.Summary.Black, r.White
		colorSplit, colorSlot = &p.rep.Black, 2
	}

	res := playerResult(r.Result, color)
	switch res {
	case "win":
		p.rep.Wins++
	case "draw":
		p.rep.Draws++
	case "loss":
		p.rep.Losses++
	default:
		p.rep.Unknown++
	}

	for _, t := range []struct {
		sp   *Split
		slot int
	}{{&p.rep.Overall, 0}, {colorSplit, colorSlot}} {
		t.sp.Games++
		t.sp.Summary.merge(counts)
		p.sums[t.slot] += acc
	}

	p.rep.Games = append(p.rep.Games, PlayerGame{
		Game: r.Game, Opponent: opp, Color: color, Result: res,
		Plies: len(r.Moves), Accuracy: acc, Summary: counts,
	})
}

// returns the aggregates computed so far
func (p *PlayerAgg) Report() PlayerReport {
	r := p.rep
	for i, sp := range []*Split{&r.Overall, &r.White, &r.Black} {
		if sp.Games > 0 {
			sp.Accuracy = math.Round(p.sums[i]/float64(sp.Games)*10) / 10
		}
	}
	return r
}

func playerResult(result, color string) string {
	switch result {
	case "White Wins":
		if color == "white" {
			return "win"
		}
		return "loss"
	case "Black Wins":
		if color == "black" {
			return "win"
		}
		return "loss"
	case "Draw":
		return "draw"
	}
	return "unknown"
}

func (c *Counts) merge(o Counts) {
	c.Blunder += o.Blunder
	c.Mistake += o.Mistake
	c.Inaccuracy += o.Inaccuracy
	c.Good += o.Good
	c.Excellent += o.Excellent
}

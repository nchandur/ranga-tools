package analysis

import "fmt"

// score from White's point of view
type Eval struct {
	CP       *int   `json:"cp,omitempty"`
	Mate     *int   `json:"mate,omitempty"`
	Terminal string `json:"terminal,omitempty"`
}

func (e Eval) String() string {
	switch {
	case e.Terminal != "":
		return e.Terminal
	case e.Mate != nil:
		if *e.Mate < 0 {
			return fmt.Sprintf("-#%d", -*e.Mate)
		}
		return fmt.Sprintf("#%d", *e.Mate)
	case e.CP != nil:
		return fmt.Sprintf("%+.2f", float64(*e.CP)/100)
	}
	return "?"
}

// Move is the verdict on one played move.
type Move struct {
	Ply        int     `json:"ply"`
	Number     int     `json:"move_number"`
	Side       string  `json:"side"`
	SAN        string  `json:"san"`
	UCI        string  `json:"uci"`
	BestMove   string  `json:"best_move"`
	EvalBefore Eval    `json:"eval_before"`
	EvalAfter  Eval    `json:"eval_after"`
	WinBefore  float64 `json:"win_before"`
	WinAfter   float64 `json:"win_after"`
	Loss       float64 `json:"win_loss"`
	Class      Class   `json:"class"`
}

type Counts struct {
	Blunder int `json:"blunder"`
	Bad     int `json:"bad"`
	Ok      int `json:"ok"`
	Good    int `json:"good"`
}

func (c *Counts) add(cl Class) {
	switch cl {
	case Blunder:
		c.Blunder++
	case Bad:
		c.Bad++
	case Ok:
		c.Ok++
	case Good:
		c.Good++
	}
}

type Summary struct {
	White Counts `json:"white"`
	Black Counts `json:"black"`
}

// full result of analyzing one game
type Report struct {
	White    string   `json:"white"`
	Black    string   `json:"black"`
	StartFEN string   `json:"start_fen"`
	Engine   string   `json:"engine"`
	Strength Strength `json:"strength"`
	Depth    int      `json:"depth"`
	HashMB   int      `json:"hash_mb"`
	Moves    []Move   `json:"moves"`
	Summary  Summary  `json:"summary"`
	Accuracy Accuracy `json:"accuracy"`
	Result   string   `json:"result"`
}

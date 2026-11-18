package analysis

// score from White's point of view
type Eval struct {
	CP       *int   `json:"cp,omitempty"`
	Mate     *int   `json:"mate,omitempty"`
	Terminal string `json:"terminal,omitempty"`
}

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

// maps class to how many moves earned it
type Counts map[Class]int

type Summary struct {
	White Counts `json:"white"`
	Black Counts `json:"black"`
}

type Report struct {
	Game     int      `json:"game,omitempty"`
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

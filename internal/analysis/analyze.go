package analysis

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/nchandur/ranga-tools/internal/board"
	"github.com/nchandur/ranga-tools/internal/pgn"
	"github.com/nchandur/ranga-tools/internal/uci"
)

type Options struct {
	EnginePath string
	Strength   Strength
}

// owns one engine session
type Analyzer struct {
	sess     *uci.Session
	strength Strength
	preset   Preset
}

func NewAnalyzer(opts Options) (*Analyzer, error) {
	preset := opts.Strength.Preset()
	sess, err := uci.Start(opts.EnginePath, map[string]string{
		"Hash":    strconv.Itoa(preset.HashMB),
		"Threads": "1",
	})
	if err != nil {
		return nil, err
	}
	return &Analyzer{sess, opts.Strength, preset}, nil
}

func (a *Analyzer) Close() { a.sess.Close() }

// evaluation of one position, before comparison
type posEval struct {
	side   string
	number int
	win    float64 // side to move's win %
	eval   Eval
	best   string
}

var results = map[float32]string{1: "White Wins", 0.5: "Draw", 0: "Black Wins"}

// evaluates every position once and classifies each move by the win
func (a *Analyzer) Analyze(g *pgn.Game) (*Report, error) {
	if !g.IsStandard() {
		return nil, fmt.Errorf("variant %q is not supported (standard chess only)", g.Variant)
	}
	n := len(g.Moves)
	if n == 0 {
		return nil, errors.New("the game has no moves")
	}
	if err := a.sess.NewGame(); err != nil {
		return nil, err
	}

	b := g.Board.Preserve()
	evals := make([]posEval, 0, n+1)
	for k := 0; k <= n; k++ {
		ev, err := evaluate(a.sess, g, &b, k, a.preset.Depth)
		if err != nil {
			return nil, fmt.Errorf("position after %d plies: %w", k, err)
		}
		evals = append(evals, ev)
		if k == n {
			break
		}
		m := b.ParseMove(g.Moves[k])
		if m == board.NOMOVE || !b.MakeMove(m, false) {
			return nil, fmt.Errorf("ply %d (%s) is illegal", k+1, g.Moves[k])
		}
		b.Ply++
	}

	result, ok := results[g.Result]
	if !ok {
		result = "Unknown"
	}
	rep := &Report{
		White:    g.White,
		Black:    g.Black,
		StartFEN: g.FEN,
		Engine:   a.sess.Info.Name,
		Strength: a.strength,
		Depth:    a.preset.Depth,
		HashMB:   a.preset.HashMB,
		Result:   result,
		Moves:    make([]Move, 0, n),
		Summary:  Summary{Counts{}, Counts{}},
	}

	for k := range n {
		before, after := evals[k], evals[k+1]
		winAfter := 100 - after.win
		loss := max(0, before.win-winAfter)
		if g.Moves[k] == before.best {
			loss = 0
		}

		mv := Move{
			Ply: k + 1, Number: before.number, Side: before.side,
			SAN: g.SAN[k], UCI: g.Moves[k], BestMove: before.best,
			EvalBefore: before.eval, EvalAfter: after.eval,
			WinBefore: before.win, WinAfter: winAfter,
			Loss: loss, Class: Classify(loss),
		}
		rep.Moves = append(rep.Moves, mv)

		counts := rep.Summary.White
		if mv.Side == "black" {
			counts = rep.Summary.Black
		}
		counts[mv.Class]++
	}
	rep.Accuracy = computeAccuracy(rep.Moves)
	return rep, nil
}

// evaluate scores position after k moves
func evaluate(s *uci.Session, g *pgn.Game, b *board.Board, k, depth int) (posEval, error) {
	ev := posEval{side: "white", number: b.Ply/2 + 1}
	if b.Side == board.Black {
		ev.side = "black"
	}

	if !b.HasLegalMove() {
		if b.InCheck() {
			ev.eval = Eval{Terminal: "checkmate"}
		} else {
			ev.win, ev.eval = 50, Eval{Terminal: "stalemate"}
		}
		return ev, nil
	}

	res, err := s.Search(g.FEN, g.Moves[:k], depth)
	if err != nil {
		return ev, err
	}
	ev.win = WinPercent(res.Score)
	ev.eval = whiteEval(res.Score, b.Side)
	ev.best = res.BestMove
	return ev, nil
}

// flips a side-to-move score into White's point of view
func whiteEval(s uci.Score, side board.Color) Eval {
	sign := 1
	if side == board.Black {
		sign = -1
	}
	if s.IsMate {
		m := sign * s.Mate
		return Eval{Mate: &m}
	}
	cp := sign * s.CP
	return Eval{CP: &cp}
}

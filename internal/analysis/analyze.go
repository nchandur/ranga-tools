package analysis

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/nchandur/ranga-tools/internal/board"
	"github.com/nchandur/ranga-tools/internal/pgn"
	"github.com/nchandur/ranga-tools/internal/uci"
)

// configures one analysis run
type Options struct {
	EnginePath string
	Strength   Strength
	Progress   func(done, total int)
}

// evaluation of one position before comparison
type posEval struct {
	side   string
	number int
	win    float64
	eval   Eval
	best   string
}

// evaluates every position of game once and classifies each move by how much win probability the mover gave up compared with the best line
func Analyze(g *pgn.Game, opts Options) (*Report, error) {
	if !g.IsStandard() {
		return nil, fmt.Errorf("variant %q is not supported (standard chess only)", g.Variant)
	}
	n := len(g.Moves)
	if n == 0 {
		return nil, errors.New("the game has no moves")
	}
	preset, err := opts.Strength.Preset()
	if err != nil {
		return nil, err
	}

	sess, err := uci.Start(opts.EnginePath, map[string]string{
		"Hash":    strconv.Itoa(preset.HashMB),
		"Threads": strconv.Itoa(preset.Threads),
	})
	if err != nil {
		return nil, err
	}
	defer sess.Close()
	if err := sess.NewGame(); err != nil {
		return nil, err
	}

	b := g.Board.Preserve()
	evals := make([]posEval, 0, n+1)
	for k := 0; k <= n; k++ {
		ev, err := evaluate(sess, g, &b, k, preset.Depth)
		if err != nil {
			return nil, fmt.Errorf("position after %d plies: %w", k, err)
		}
		evals = append(evals, ev)
		if opts.Progress != nil {
			opts.Progress(k+1, n+1)
		}

		if k == n {
			break
		}
		m := b.ParseMove(g.Moves[k])
		if m == board.NOMOVE || !b.MakeMove(m, false) {
			return nil, fmt.Errorf("ply %d (%s) is illegal", k+1, g.Moves[k])
		}
		b.Ply++
	}

	rep := &Report{
		White:    g.White,
		Black:    g.Black,
		StartFEN: g.FEN,
		Engine:   sess.Info.Name,
		Strength: opts.Strength,
		Depth:    preset.Depth,
		HashMB:   preset.HashMB,
		Moves:    make([]Move, 0, n),
	}

	for k := range n {
		before, after := evals[k], evals[k+1]

		winBefore := before.win
		winAfter := 100 - after.win
		loss := math.Max(0, winBefore-winAfter)
		if g.Moves[k] == before.best {
			loss = 0
		}

		mv := Move{
			Ply:        k + 1,
			Number:     before.number,
			Side:       before.side,
			SAN:        g.SAN[k],
			UCI:        g.Moves[k],
			BestMove:   before.best,
			EvalBefore: before.eval,
			EvalAfter:  after.eval,
			WinBefore:  winBefore,
			WinAfter:   winAfter,
			Loss:       loss,
			Class:      Classify(loss),
		}
		rep.Moves = append(rep.Moves, mv)

		if mv.Side == "white" {
			rep.Summary.White.add(mv.Class)
		} else {
			rep.Summary.Black.add(mv.Class)
		}
	}
	rep.Accuracy = computeAccuracy(rep.Moves)
	return rep, nil
}

// evaluate scores from position
func evaluate(s *uci.Session, g *pgn.Game, b *board.Board, k, depth int) (posEval, error) {
	ev := posEval{side: "white", number: b.Ply/2 + 1}
	if b.Side == board.Black {
		ev.side = "black"
	}

	if !b.HasLegalMove() {
		if b.InCheck() {
			ev.win, ev.eval = 0, Eval{Terminal: "checkmate"}
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

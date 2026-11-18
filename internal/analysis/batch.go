package analysis

import (
	"sync"
	"sync/atomic"

	"github.com/nchandur/ranga-tools/internal/pgn"
)

// one game to analyze
type Job struct {
	Number int
	Text   string
}

type GameResult struct {
	Index  int
	Report *Report
	Err    error
}

// analyzes jobs with `workers` engines in parallel and returns one result per job
func AnalyzeBatch(jobs []Job, opts Options, workers int, onDone func(done int)) ([]GameResult, error) {
	workers = min(max(workers, 1), len(jobs))

	analyzers := make([]*Analyzer, 0, workers)
	defer func() {
		for _, a := range analyzers {
			a.Close()
		}
	}()
	for range workers {
		a, err := NewAnalyzer(opts)
		if err != nil {
			return nil, err
		}
		analyzers = append(analyzers, a)
	}

	out := make([]GameResult, len(jobs))
	next := make(chan int)
	var (
		wg   sync.WaitGroup
		done atomic.Int32
	)
	for _, a := range analyzers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range next {
				out[i] = analyzeOne(a, jobs[i])
				if onDone != nil {
					onDone(int(done.Add(1)))
				}
			}
		}()
	}
	for i := range jobs {
		next <- i
	}
	close(next)
	wg.Wait()
	return out, nil
}

func analyzeOne(a *Analyzer, j Job) GameResult {
	g, err := pgn.NewGame(j.Text)
	if err != nil {
		return GameResult{Index: j.Number, Err: err}
	}
	rep, err := a.Analyze(&g)
	if err != nil {
		return GameResult{Index: j.Number, Err: err}
	}
	rep.Game = j.Number
	return GameResult{Index: j.Number, Report: rep}
}

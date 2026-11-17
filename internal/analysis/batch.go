package analysis

import (
	"fmt"
	"sync"

	"github.com/nchandur/ranga-tools/internal/pgn"
)

// one game to analyze
type Job struct {
	Number int
	Text   string
}

// outcome of one game
type GameResult struct {
	Index  int
	Report *Report
	Err    error
}

// analyzes jobs with `workers` engine sessions in parallel
func AnalyzeBatch(jobs []Job, opts Options, workers int, emit func(GameResult)) error {
	if len(jobs) == 0 {
		return nil
	}
	workers = min(max(workers, 1), len(jobs))

	analyzers := make([]*Analyzer, 0, workers)
	defer func() {
		for _, a := range analyzers {
			a.Close()
		}
	}()
	for range workers {
		o := opts
		o.Progress = nil // per-position progress makes no sense across workers
		a, err := NewAnalyzer(o)
		if err != nil {
			return err
		}
		analyzers = append(analyzers, a)
	}

	type item struct {
		seq int
		res GameResult
	}
	seqs := make(chan int)
	results := make(chan item)

	var wg sync.WaitGroup
	for _, a := range analyzers {
		wg.Add(1)
		go func(a *Analyzer) {
			defer wg.Done()
			for i := range seqs {
				results <- item{i, analyzeOne(a, jobs[i])}
			}
		}(a)
	}
	go func() {
		for i := range jobs {
			seqs <- i
		}
		close(seqs)
	}()
	go func() {
		wg.Wait()
		close(results)
	}()

	pending := map[int]GameResult{}
	next := 0
	for it := range results {
		pending[it.seq] = it.res
		for {
			p, ok := pending[next]
			if !ok {
				break
			}
			delete(pending, next)
			emit(p)
			next++
		}
	}
	return nil
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

func (r GameResult) String() string {
	if r.Err != nil {
		return fmt.Sprintf("game %d: %v", r.Index, r.Err)
	}
	return fmt.Sprintf("game %d: ok", r.Index)
}

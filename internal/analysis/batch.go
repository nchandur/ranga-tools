package analysis

import (
	"fmt"
	"sync"

	"github.com/nchandur/ranga-tools/internal/pgn"
)

// outcome of one game in a batch
type GameResult struct {
	Index  int
	Report *Report
	Err    error
}

// analyzes every game using `workers` engine sessions in parallel
func AnalyzeBatch(games []string, opts Options, workers int, emit func(GameResult)) error {
	workers = min(max(workers, 1), len(games))

	analyzers := make([]*Analyzer, 0, workers)
	defer func() {
		for _, a := range analyzers {
			a.Close()
		}
	}()
	for range workers {
		o := opts
		o.Progress = nil
		a, err := NewAnalyzer(o)
		if err != nil {
			return err
		}
		analyzers = append(analyzers, a)
	}

	jobs := make(chan int)
	results := make(chan GameResult)

	var wg sync.WaitGroup
	for _, a := range analyzers {
		wg.Add(1)
		go func(a *Analyzer) {
			defer wg.Done()
			for i := range jobs {
				results <- analyzeOne(a, i+1, games[i])
			}
		}(a)
	}
	go func() {
		for i := range games {
			jobs <- i
		}
		close(jobs)
	}()
	go func() {
		wg.Wait()
		close(results)
	}()

	pending := map[int]GameResult{}
	next := 1
	for r := range results {
		pending[r.Index] = r
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

func analyzeOne(a *Analyzer, index int, text string) GameResult {
	g, err := pgn.NewGame(text)
	if err != nil {
		return GameResult{Index: index, Err: err}
	}
	rep, err := a.Analyze(&g)
	if err != nil {
		return GameResult{Index: index, Err: err}
	}
	rep.Game = index
	return GameResult{Index: index, Report: rep}
}

func (r GameResult) String() string {
	if r.Err != nil {
		return fmt.Sprintf("game %d: %v", r.Index, r.Err)
	}
	return fmt.Sprintf("game %d: ok", r.Index)
}

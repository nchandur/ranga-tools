package analyze

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nchandur/ranga-tools/internal/analysis"
	"github.com/nchandur/ranga-tools/internal/pgn"
	"github.com/spf13/cobra"
)

type gameRow struct {
	Index        int
	White, Black string
	Plies        int
	AccW, AccB   float64
	BlunW, BlunB int
}

func runBatch(cmd *cobra.Command, pgnPath, enginePath string, strength analysis.Strength, workers int, jsonPath, player string) error {
	games, err := pgn.ReadGames(pgnPath)
	if err != nil {
		return err
	}
	preset, err := strength.Preset()
	if err != nil {
		return err
	}

	jobs := make([]analysis.Job, len(games))
	for i, g := range games {
		jobs[i] = analysis.Job{Number: i + 1, Text: g}
	}
	var agg *analysis.PlayerAgg
	var playerName string
	if player != "" {
		if jobs, playerName, err = selectPlayerJobs(games, player); err != nil {
			return err
		}
		agg = analysis.NewPlayerAgg(playerName)
	}

	var jw *batchWriter
	if jsonPath != "" {
		if jw, err = newBatchWriter(jsonPath, strength, preset); err != nil {
			return err
		}
		defer jw.discard()
	}

	errw := cmd.ErrOrStderr()
	tty := isTerminal(errw)

	var (
		rows   []gameRow
		failed []analysis.GameResult
		engine string
		done   int
	)
	start := time.Now()
	err = analysis.AnalyzeBatch(jobs,
		analysis.Options{EnginePath: enginePath, Strength: strength},
		workers,
		func(r analysis.GameResult) {
			done++
			if tty {
				fmt.Fprintf(errw, "\r\033[Kanalyzing game %d/%d", done, len(jobs))
			}
			if r.Err != nil {
				failed = append(failed, r)
				return
			}
			if engine == "" {
				engine = r.Report.Engine
			}
			rows = append(rows, newRow(r.Report))
			if agg != nil {
				color := "black"
				if r.Report.White == playerName {
					color = "white"
				}
				agg.Add(r.Report, color)
			}
			if jw != nil {
				jw.add(r.Report)
			}
		})
	if tty {
		fmt.Fprint(errw, "\r\033[K")
	}
	if err != nil {
		return err
	}
	duration := time.Since(start)

	if len(rows) == 0 {
		return fmt.Errorf("all %d games failed; first error: %w", len(failed), failed[0].Err)
	}

	var pr *analysis.PlayerReport
	if agg != nil {
		r := agg.Report()
		pr = &r
	}

	if jw != nil {
		if err := jw.finish(failed, pr); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(),
			"Analysis complete. %d games analyzed, %d failed. Report written to %s\n",
			len(rows), len(failed), jsonPath)
		return nil
	}

	if pr != nil {
		printPlayer(cmd.OutOrStdout(), engine, strength, preset, len(failed), *pr, duration)
		return nil
	}
	printBatch(cmd.OutOrStdout(), engine, strength, preset, rows, failed)
	return nil
}

func newRow(r *analysis.Report) gameRow {
	return gameRow{
		Index: r.Game, White: r.White, Black: r.Black, Plies: len(r.Moves),
		AccW: r.Accuracy.White, AccB: r.Accuracy.Black,
		BlunW: r.Summary.White.Blunder, BlunB: r.Summary.Black.Blunder,
	}
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func printBatch(w io.Writer, engine string, strength analysis.Strength, p analysis.Preset,
	rows []gameRow, failed []analysis.GameResult) {

	const nameWidth = 16
	rule := strings.Repeat("─", 77)

	fmt.Fprintln(w)
	info := func(label, value string) {
		fmt.Fprintf(w, "  %s %s\n", padRight(label, labelWidth), value)
	}
	info("Engine", engine)
	info("Strength", fmt.Sprintf("%s (depth %d, hash %d MB)", strength, p.Depth, p.HashMB))
	info("Games", fmt.Sprintf("%d analyzed, %d failed", len(rows), len(failed)))
	fmt.Fprintln(w)

	fmt.Fprintf(w, "  %s  %s  %s  %s  %s  %s  %s\n",
		padLeft("#", 4), padRight("White", nameWidth), padRight("Black", nameWidth),
		padLeft("Plies", 5), padLeft("Acc W", 6), padLeft("Acc B", 6), "Blunders W/B")
	fmt.Fprintf(w, "  %s\n", rule)
	for _, r := range rows {
		fmt.Fprintf(w, "  %s  %s  %s  %s  %s  %s  %s\n",
			padLeft(strconv.Itoa(r.Index), 4),
			padRight(truncate(orUnknown(r.White), nameWidth), nameWidth),
			padRight(truncate(orUnknown(r.Black), nameWidth), nameWidth),
			padLeft(strconv.Itoa(r.Plies), 5),
			padLeft(fmt.Sprintf("%.1f%%", r.AccW), 6),
			padLeft(fmt.Sprintf("%.1f%%", r.AccB), 6),
			fmt.Sprintf("%d / %d", r.BlunW, r.BlunB))
	}

	if len(failed) > 0 {
		fmt.Fprintf(w, "\n  Failed\n  %s\n", rule)
		for _, f := range failed {
			fmt.Fprintf(w, "  %s  %v\n", padLeft(strconv.Itoa(f.Index), 4), f.Err)
		}
	}
	fmt.Fprintln(w)
}

// streams reports into one JSON file as they arrive
type batchWriter struct {
	f    *os.File
	w    *bufio.Writer
	n    int
	err  error
	done bool
	path string
}

func newBatchWriter(path string, s analysis.Strength, p analysis.Preset) (*batchWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	b := &batchWriter{f: f, w: bufio.NewWriter(f), path: path}
	fmt.Fprintf(b.w, `{"strength":%q,"depth":%d,"hash_mb":%d,"games":[`, s, p.Depth, p.HashMB)
	return b, nil
}

func (b *batchWriter) add(r *analysis.Report) {
	if b.err != nil {
		return
	}
	data, err := json.Marshal(r)
	if err != nil {
		b.err = err
		return
	}
	if b.n > 0 {
		b.w.WriteByte(',')
	}
	_, b.err = b.w.Write(data)
	b.n++
}

func (b *batchWriter) finish(failed []analysis.GameResult, player *analysis.PlayerReport) error {
	b.done = true
	type failure struct {
		Game  int    `json:"game"`
		Error string `json:"error"`
	}
	b.w.WriteString("]")
	if player != nil {
		data, _ := json.Marshal(player)
		b.w.WriteString(`,"player":`)
		b.w.Write(data)
	}
	b.w.WriteString("}\n")
	if b.err == nil {
		b.err = b.w.Flush()
	}
	if cerr := b.f.Close(); b.err == nil {
		b.err = cerr
	}
	if b.err != nil {
		return fmt.Errorf("write %s: %w", b.path, b.err)
	}
	return nil
}

// discard removes the half-written file if the run did not finish
func (b *batchWriter) discard() {
	if b.done {
		return
	}
	b.f.Close()
	os.Remove(b.path)
}

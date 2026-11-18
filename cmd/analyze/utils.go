package analyze

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	eng "github.com/nchandur/ranga-tools/internal/engine"

	"github.com/nchandur/ranga-tools/internal/analysis"
	"github.com/nchandur/ranga-tools/internal/pgn"
	"github.com/spf13/cobra"
)

func run(cmd *cobra.Command, _ []string) error {
	strength, err := analysis.ParseStrength(strengthName)
	if err != nil {
		return err
	}
	entry, err := eng.Lookup(engineName)
	if err != nil {
		return err
	}
	if gameNum < 1 || workers < 1 {
		return errors.New("--game and --workers must be at least 1")
	}
	cmd.SilenceUsage = true

	games, err := pgn.ReadGames(pgnPath)
	if err != nil {
		return err
	}
	jobs, name, err := selectJobs(games)
	if err != nil {
		return err
	}

	fi, _ := os.Stderr.Stat()
	tty := fi != nil && fi.Mode()&os.ModeCharDevice != 0

	start := time.Now()
	results, err := analysis.AnalyzeBatch(jobs,
		analysis.Options{EnginePath: entry.Path, Strength: strength}, workers,
		func(done int) {
			if tty && len(jobs) > 1 {
				n := done * 24 / len(jobs)
				fmt.Fprintf(os.Stderr, "\r  analyzing  %s%s  %d/%d",
					strings.Repeat("█", n), strings.Repeat("░", 24-n), done, len(jobs))
			}
		})
	if tty {
		fmt.Fprint(os.Stderr, "\r\033[K")
	}
	if err != nil {
		return err
	}
	took := time.Since(start).Round(100 * time.Millisecond)

	var (
		reports []*analysis.Report
		errs    []string
	)
	for _, r := range results {
		if r.Err != nil {
			errs = append(errs, fmt.Sprintf("game %d: %v", r.Index, r.Err))
		} else {
			reports = append(reports, r.Report)
		}
	}
	if len(reports) == 0 {
		return fmt.Errorf("all games failed: %s", errs[0])
	}

	var pr *analysis.PlayerReport
	if name != "" {
		pr = analysis.NewPlayerReport(name, reports)
	}
	batch := all || player != ""
	out := cmd.OutOrStdout()

	if jsonPath != "" {
		var v any = reports[0]
		if batch {
			v = struct {
				Games  []*analysis.Report     `json:"games"`
				Errors []string               `json:"errors"`
				Player *analysis.PlayerReport `json:"player,omitempty"`
			}{reports, errs, pr}
		}
		data, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(jsonPath, data, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(out, "Analysis complete. %d games analyzed, %d failed. Report written to %s\n",
			len(reports), len(errs), jsonPath)
		return nil
	}

	switch {
	case pr != nil:
		printPlayer(out, reports[0], pr, len(errs), took)
	case batch:
		printBatch(out, reports, len(errs), took)
	default:
		printGame(out, reports[0], took)
	}
	if len(errs) > 0 {
		fmt.Fprintf(out, "  Failed\n%s\n", rule(56))
		for _, e := range errs {
			fmt.Fprintf(out, "  %s\n", e)
		}
		fmt.Fprintln(out)
	}
	return nil
}

// picks the games to analyze
func selectJobs(games []string) (jobs []analysis.Job, name string, err error) {
	switch {
	case player != "":
		return playerJobs(games)
	case all:
		for i, g := range games {
			jobs = append(jobs, analysis.Job{Number: i + 1, Text: g})
		}
	case gameNum > len(games):
		err = fmt.Errorf("game %d does not exist: the file has %d games", gameNum, len(games))
	default:
		jobs = []analysis.Job{{Number: gameNum, Text: games[gameNum-1]}}
	}
	return
}

// first name matching the regex wins
func playerJobs(games []string) ([]analysis.Job, string, error) {
	re, err := regexp.Compile("(?i)" + player)
	if err != nil {
		return nil, "", fmt.Errorf("invalid --player pattern: %w", err)
	}
	var (
		jobs []analysis.Job
		name string
	)
	for i, text := range games {
		w, b := pgn.Players(text)
		if name == "" {
			switch {
			case w != "" && re.MatchString(w):
				name = w
			case b != "" && re.MatchString(b):
				name = b
			}
		}
		if name != "" && (w == name) != (b == name) {
			jobs = append(jobs, analysis.Job{Number: i + 1, Text: text})
		}
	}
	if len(jobs) == 0 {
		return nil, "", fmt.Errorf("no games found for player matching %q", player)
	}
	return jobs, name, nil
}

func rule(n int) string {
	return "  " + strings.Repeat("─", n)
}

func header(w io.Writer, title string, r *analysis.Report, took time.Duration, extra ...string) {
	fmt.Fprintf(w, "\n  %s\n%s\n", title, rule(56))
	fmt.Fprintf(w, "  %-10s %s\n", "Engine", r.Engine)
	fmt.Fprintf(w, "  %-10s %s\n", "Strength", r.Strength)
	for i := 0; i+1 < len(extra); i += 2 {
		fmt.Fprintf(w, "  %-10s %s\n", extra[i], extra[i+1])
	}
	fmt.Fprintf(w, "  %-10s %s\n\n", "Time", took)
}

func printGame(w io.Writer, r *analysis.Report, took time.Duration) {
	header(w, fmt.Sprintf("%s (White)  vs  %s (Black)", r.White, r.Black), r, took,
		"Length", fmt.Sprintf("%d plies", len(r.Moves)), "Result", r.Result)

	fmt.Fprintf(w, "  %-12s%9s%9s\n%s\n", "", "White", "Black", rule(30))
	fmt.Fprintf(w, "  %-12s%8.1f%%%8.1f%%\n%s\n", "Accuracy", r.Accuracy.White, r.Accuracy.Black, rule(30))
	for _, c := range analysis.Classes {
		fmt.Fprintf(w, "  %-12s%9d%9d\n", c, r.Summary.White[c], r.Summary.Black[c])
	}
	fmt.Fprintln(w)
}

func printBatch(w io.Writer, reports []*analysis.Report, failed int, took time.Duration) {
	header(w, "Batch analysis", reports[0], took,
		"Games", fmt.Sprintf("%d analyzed, %d failed", len(reports), failed))

	fmt.Fprintf(w, "  %4s  %-16s  %-16s  %5s  %7s  %7s  %s\n%s\n",
		"#", "White", "Black", "Plies", "Acc W", "Acc B", "Blunders", rule(76))
	for _, r := range reports {
		fmt.Fprintf(w, "  %4d  %-16.16s  %-16.16s  %5d  %6.1f%%  %6.1f%%  %d / %d\n",
			r.Game, r.White, r.Black, len(r.Moves), r.Accuracy.White, r.Accuracy.Black,
			r.Summary.White[analysis.Blunder], r.Summary.Black[analysis.Blunder])
	}
	fmt.Fprintln(w)
}

func printPlayer(w io.Writer, r0 *analysis.Report, pr *analysis.PlayerReport, failed int, took time.Duration) {
	header(w, pr.Name, r0, took,
		"Games", fmt.Sprintf("%d analyzed, %d failed", pr.Overall.Games, failed),
		"Record", fmt.Sprintf("+%d =%d -%d", pr.Wins, pr.Draws, pr.Losses))

	fmt.Fprintf(w, "  %-12s%9s%9s%9s\n%s\n", "", "Overall", "White", "Black", rule(39))
	fmt.Fprintf(w, "  %-12s%9d%9d%9d\n", "Games", pr.Overall.Games, pr.White.Games, pr.Black.Games)
	fmt.Fprintf(w, "  %-12s%8.1f%%%8.1f%%%8.1f%%\n%s\n", "Accuracy",
		pr.Overall.Accuracy, pr.White.Accuracy, pr.Black.Accuracy, rule(39))
	for _, c := range analysis.Classes {
		fmt.Fprintf(w, "  %-12s%9d%9d%9d\n", c, pr.Overall.Summary[c], pr.White.Summary[c], pr.Black.Summary[c])
	}

	fmt.Fprintf(w, "\n  %4s  %-18s  %-5s  %-6s  %5s  %7s  %s\n%s\n",
		"#", "Opponent", "Color", "Result", "Plies", "Acc", "Blunders", rule(66))
	for _, g := range pr.Games {
		fmt.Fprintf(w, "  %4d  %-18.18s  %-5s  %-6s  %5d  %6.1f%%  %d\n",
			g.Game, g.Opponent, g.Color, g.Result, g.Plies, g.Accuracy, g.Summary[analysis.Blunder])
	}
	fmt.Fprintln(w)
}

package analyze

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"text/tabwriter"
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

	start := time.Now()
	results, err := analysis.AnalyzeBatch(jobs,
		analysis.Options{EnginePath: entry.Path, Strength: strength}, workers,
		func(done int) {
			if len(jobs) > 1 {
				fmt.Fprintf(cmd.ErrOrStderr(), "analyzed %d/%d games\n", done, len(jobs))
			}
		})
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
	for _, e := range errs {
		fmt.Fprintln(out, "failed:", e)
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

// header writes the lines shared by every view
func header(tw io.Writer, r *analysis.Report, took time.Duration, extra ...string) {
	fmt.Fprintf(tw, "Engine\t%s\nStrength\t%s (depth %d, hash %d MB)\n", r.Engine, r.Strength, r.Depth, r.HashMB)
	for _, e := range extra {
		fmt.Fprintln(tw, e)
	}
	fmt.Fprintf(tw, "Time\t%s\n\n", took)
}

func printGame(w io.Writer, r *analysis.Report, took time.Duration) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	defer tw.Flush()
	fmt.Fprintf(tw, "\n%s (White) vs %s (Black)\n\n", r.White, r.Black)
	header(tw, r, took, fmt.Sprintf("Length\t%d plies", len(r.Moves)), "Result\t"+r.Result)
	fmt.Fprintln(tw, "\tWhite\tBlack")
	fmt.Fprintf(tw, "Accuracy\t%.1f%%\t%.1f%%\n", r.Accuracy.White, r.Accuracy.Black)
	for _, c := range analysis.Classes {
		fmt.Fprintf(tw, "%s\t%d\t%d\n", c, r.Summary.White[c], r.Summary.Black[c])
	}
}

func printBatch(w io.Writer, reports []*analysis.Report, failed int, took time.Duration) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	defer tw.Flush()
	fmt.Fprintln(tw)
	header(tw, reports[0], took, fmt.Sprintf("Games\t%d analyzed, %d failed", len(reports), failed))
	fmt.Fprintln(tw, "#\tWhite\tBlack\tPlies\tAcc W\tAcc B\tBlunders W/B")
	for _, r := range reports {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%d\t%.1f%%\t%.1f%%\t%d / %d\n",
			r.Game, r.White, r.Black, len(r.Moves), r.Accuracy.White, r.Accuracy.Black,
			r.Summary.White[analysis.Blunder], r.Summary.Black[analysis.Blunder])
	}
}

func printPlayer(w io.Writer, r0 *analysis.Report, pr *analysis.PlayerReport, failed int, took time.Duration) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	defer tw.Flush()
	fmt.Fprintf(tw, "\n%s\n\n", pr.Name)
	header(tw, r0, took,
		fmt.Sprintf("Games\t%d analyzed, %d failed", pr.Overall.Games, failed),
		fmt.Sprintf("Record\t+%d =%d -%d", pr.Wins, pr.Draws, pr.Losses))
	fmt.Fprintln(tw, "\tOverall\tWhite\tBlack")
	fmt.Fprintf(tw, "Games\t%d\t%d\t%d\n", pr.Overall.Games, pr.White.Games, pr.Black.Games)
	fmt.Fprintf(tw, "Accuracy\t%.1f%%\t%.1f%%\t%.1f%%\n", pr.Overall.Accuracy, pr.White.Accuracy, pr.Black.Accuracy)
	for _, c := range analysis.Classes {
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\n", c, pr.Overall.Summary[c], pr.White.Summary[c], pr.Black.Summary[c])
	}
	fmt.Fprintln(tw, "\n#\tOpponent\tColor\tResult\tPlies\tAcc\tBlunders")
	for _, g := range pr.Games {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%d\t%.1f%%\t%d\n",
			g.Game, g.Opponent, g.Color, g.Result, g.Plies, g.Accuracy, g.Summary[analysis.Blunder])
	}
}

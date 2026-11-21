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
	if strings.TrimSpace(player) == "" {
		return errors.New("--player is required: give the name of the player to analyze")
	}
	if workers < 1 {
		return errors.New("--workers must be at least 1")
	}
	strength, err := analysis.ParseStrength(strengthName)
	if err != nil {
		return err
	}
	entry, err := eng.Lookup(engineName)
	if err != nil {
		return err
	}
	cmd.SilenceUsage = true

	games, err := pgn.ReadGames(pgnPath)
	if err != nil {
		return err
	}
	jobs, name, err := playerJobs(games)
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
				fmt.Fprintf(os.Stderr, "\r  analyzing  [%s%s]  %d/%d",
					strings.Repeat("#", n), strings.Repeat("-", 24-n), done, len(jobs))
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

	pr := analysis.NewPlayerReport(name, reports)
	out := cmd.OutOrStdout()

	if jsonPath == "" && textPath == "" {
		writeReport(out, reports[0], pr, errs, took)
		return nil
	}

	if jsonPath != "" {
		r0 := reports[0]
		data, err := json.MarshalIndent(struct {
			Engine   string                 `json:"engine"`
			Strength analysis.Strength      `json:"strength"`
			Depth    int                    `json:"depth"`
			HashMB   int                    `json:"hash_mb"`
			Errors   []string               `json:"errors,omitempty"`
			Player   *analysis.PlayerReport `json:"player"`
		}{r0.Engine, r0.Strength, r0.Depth, r0.HashMB, errs, pr}, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(jsonPath, data, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(out, "Summary written to %s\n", jsonPath)
	}

	if textPath != "" {
		f, err := os.Create(textPath)
		if err != nil {
			return err
		}
		writeReport(f, reports[0], pr, errs, took)
		if err := f.Close(); err != nil {
			return err
		}
		fmt.Fprintf(out, "Report written to %s\n", textPath)
	}
	fmt.Fprintf(out, "Analysis complete. %d games analyzed, %d failed.\n", len(reports), len(errs))
	return nil
}

// writes the full terminal report, failures included
func writeReport(w io.Writer, r0 *analysis.Report, pr *analysis.PlayerReport, errs []string, took time.Duration) {
	printPlayer(w, r0, pr, len(errs), took)
	if len(errs) > 0 {
		fmt.Fprintf(w, "  Failed\n%s\n", rule(56))
		for _, e := range errs {
			fmt.Fprintf(w, "  %s\n", e)
		}
		fmt.Fprintln(w)
	}
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

package analyze

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nchandur/ranga-tools/internal/analysis"
	"github.com/nchandur/ranga-tools/internal/pgn"
)

// finds the first name matching the regex and returns that player's games
func selectPlayerJobs(games []string, pattern string) ([]analysis.Job, string, error) {
	re, err := regexp.Compile("(?i)" + pattern)
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
		if name == "" {
			continue
		}
		wm, bm := w == name, b == name
		if wm == bm { // neither plays, or the player faces themselves
			continue
		}
		jobs = append(jobs, analysis.Job{Number: i + 1, Text: text})
	}
	if name == "" {
		return nil, "", fmt.Errorf("no player in the PGN matches %q", pattern)
	}
	return jobs, name, nil
}

func printPlayer(w io.Writer, engine string, strength analysis.Strength, p analysis.Preset,
	failed int, pr analysis.PlayerReport, d time.Duration) {

	const nameWidth = 16
	rule := strings.Repeat("─", max(utf8.RuneCountInString(pr.Name), 63))

	fmt.Fprintln(w)
	fmt.Fprintf(w, "  %s\n", pr.Name)
	fmt.Fprintf(w, "  %s\n\n", rule)

	info := func(label, value string) {
		fmt.Fprintf(w, "  %s %s\n", padRight(label, labelWidth), value)
	}
	info("Engine", engine)
	info("Strength", fmt.Sprintf("%s (depth %d, hash %d MB)", strength, p.Depth, p.HashMB))
	info("Games", fmt.Sprintf("%d analyzed, %d failed", pr.Overall.Games, failed))
	info("Record", fmt.Sprintf("+%d =%d -%d", pr.Wins, pr.Draws, pr.Losses))
	info("Time", formatDuration(d))
	fmt.Fprintln(w)

	// Summary table
	row := func(label string, a, b, c string) {
		fmt.Fprintf(w, "  %s  %s  %s  %s\n",
			padRight(label, labelWidth),
			padLeft(a, cellWidth), padLeft(b, cellWidth), padLeft(c, cellWidth))
	}
	counts := func(label string, get func(analysis.Counts) int) {
		row(label,
			strconv.Itoa(get(pr.Overall.Summary)),
			strconv.Itoa(get(pr.White.Summary)),
			strconv.Itoa(get(pr.Black.Summary)))
	}

	row("", "Overall", "White", "Black")
	fmt.Fprintf(w, "  %s\n", rule)
	row("Games",
		strconv.Itoa(pr.Overall.Games), strconv.Itoa(pr.White.Games), strconv.Itoa(pr.Black.Games))
	row("Accuracy",
		fmt.Sprintf("%.1f%%", pr.Overall.Accuracy),
		fmt.Sprintf("%.1f%%", pr.White.Accuracy),
		fmt.Sprintf("%.1f%%", pr.Black.Accuracy))
	fmt.Fprintf(w, "  %s\n", rule)
	counts("Blunder", func(c analysis.Counts) int { return c.Blunder })
	counts("Mistake", func(c analysis.Counts) int { return c.Mistake })
	counts("Inaccuracy", func(c analysis.Counts) int { return c.Inaccuracy })
	counts("Good", func(c analysis.Counts) int { return c.Good })
	counts("Excellent", func(c analysis.Counts) int { return c.Excellent })
	fmt.Fprintln(w)

	// Per-game table, from the player's point of view
	fmt.Fprintf(w, "  %s  %s  %s  %s  %s  %s  %s\n",
		padLeft("#", 4), padRight("Opponent", nameWidth), padRight("Color", 5),
		padRight("Result", 7), padLeft("Plies", 5), padLeft("Acc", 6), "Blunders")
	fmt.Fprintf(w, "  %s\n", rule)
	for _, g := range pr.Games {
		fmt.Fprintf(w, "  %s  %s  %s  %s  %s  %s  %d\n",
			padLeft(strconv.Itoa(g.Game), 4),
			padRight(truncate(orUnknown(g.Opponent), nameWidth), nameWidth),
			padRight(g.Color, 5),
			padRight(g.Result, 7),
			padLeft(strconv.Itoa(g.Plies), 5),
			padLeft(fmt.Sprintf("%.1f%%", g.Accuracy), 6),
			g.Summary.Blunder)
	}
	fmt.Fprintln(w)
}

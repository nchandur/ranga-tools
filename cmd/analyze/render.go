package analyze

import (
	"cmp"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/nchandur/ranga-tools/internal/analysis"
)

// renders 10-cell gauge for a 0-100 value
func bar(pct float64) string {
	n := min(10, max(0, int(pct/10+0.5)))
	return "[" + strings.Repeat("*", n) + strings.Repeat("-", 10-n) + "]"
}

var resultLetter = map[string]string{"win": "W", "draw": "D", "loss": "L", "unknown": "?"}

func header(w io.Writer, title string, r *analysis.Report, took time.Duration, extra ...string) {
	fmt.Fprintf(w, "\n  %s\n%s\n", title, rule(56))
	fmt.Fprintf(w, "  %-12s %s\n", "Engine", r.Engine)
	fmt.Fprintf(w, "  %-12s %s\n", "Strength", r.Strength)
	for i := 0; i+1 < len(extra); i += 2 {
		fmt.Fprintf(w, "  %-12s %s\n", extra[i], extra[i+1])
	}
	fmt.Fprintf(w, "  %-12s %s\n\n", "Time", took)
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

func printGames(w io.Writer, title string, games []analysis.PlayerGame) {
	fmt.Fprintf(w, "\n  %s\n", title)
	for _, g := range games {
		fmt.Fprintf(w, "   %-4d vs %-16.16s  %s  %s  %4d plies  %5.1f%%  %s\n",
			g.Game, g.Opponent, strings.ToUpper(g.Color[:1]), resultLetter[g.Result],
			g.Plies, g.Accuracy, bar(g.Accuracy))
	}
}

func printPlayer(w io.Writer, r0 *analysis.Report, pr *analysis.PlayerReport, failed int, took time.Duration) {
	fmt.Fprintf(w, "\n  %s · %d games analyzed", pr.Name, pr.Overall.Games)
	if failed > 0 {
		fmt.Fprintf(w, " (%d failed)", failed)
	}
	fmt.Fprintf(w, "\n%s\n", rule(56))
	fmt.Fprintf(w, "  %-12s %s · %s · %s\n\n", "Engine", r0.Engine, r0.Strength, took)

	fmt.Fprintf(w, "  %-12s %5.1f%%  %s\n", "Accuracy", pr.Overall.Accuracy, bar(pr.Overall.Accuracy))
	fmt.Fprintf(w, "  %-12s %5.1f%%  %s  (%d games)\n", "As White", pr.White.Accuracy, bar(pr.White.Accuracy), pr.White.Games)
	fmt.Fprintf(w, "  %-12s %5.1f%%  %s  (%d games)\n", "As Black", pr.Black.Accuracy, bar(pr.Black.Accuracy), pr.Black.Games)

	blunders := pr.Overall.Summary[analysis.Blunder]
	fmt.Fprintf(w, "\n  %-12s %.1f per game | %d total\n", "Blunders", float64(blunders)/float64(pr.Overall.Games), blunders)

	mistakes := pr.Overall.Summary[analysis.Mistake]
	fmt.Fprintf(w, "  %-12s %.1f per game | %d total\n", "Mistakes", float64(mistakes)/float64(pr.Overall.Games), mistakes)

	innaccuracies := pr.Overall.Summary[analysis.Inaccuracy]
	fmt.Fprintf(w, "  %-12s %.1f per game | %d total\n", "Inaccuracy", float64(innaccuracies)/float64(pr.Overall.Games), innaccuracies)

	goods := pr.Overall.Summary[analysis.Good]
	fmt.Fprintf(w, "  %-12s %.1f per game | %d total\n", "Good", float64(goods)/float64(pr.Overall.Games), goods)

	excellents := pr.Overall.Summary[analysis.Excellent]
	fmt.Fprintf(w, "  %-12s %.1f per game | %d total\n", "Excellent", float64(excellents)/float64(pr.Overall.Games), excellents)

	bests := pr.Overall.Summary[analysis.Best]
	fmt.Fprintf(w, "  %-12s %.1f per game | %d total\n", "Best", float64(bests)/float64(pr.Overall.Games), bests)

	fmt.Fprintf(w, "\n  %-12s W: %d | D: %d | L: %d\n", "Record", pr.Wins, pr.Draws, pr.Losses)

	if full || len(pr.Games) <= 10 {
		printGames(w, "Games", pr.Games)
	} else {
		sorted := slices.Clone(pr.Games)
		slices.SortFunc(sorted, func(a, b analysis.PlayerGame) int { return cmp.Compare(a.Accuracy, b.Accuracy) })
		printGames(w, "Worst games", sorted[:5])
		slices.Reverse(sorted)
		printGames(w, "Best games", sorted[:5])
		fmt.Fprintf(w, "\n  Use --full to list all %d games.\n", len(pr.Games))
	}
	fmt.Fprintln(w)
}

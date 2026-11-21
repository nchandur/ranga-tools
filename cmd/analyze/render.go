package analyze

import (
	"cmp"
	"fmt"
	"io"
	"slices"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/nchandur/ranga-tools/internal/analysis"
)

// renders 10-cell gauge for a 0-100 value
func bar(pct float64) string {
	n := min(10, max(0, int(pct/10+0.5)))
	return "[" + strings.Repeat("#", n) + strings.Repeat("-", 10-n) + "]"
}

var resultLetter = map[string]string{"win": "W", "draw": "D", "loss": "L", "unknown": "?"}

func table(w io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
}

func printGames(w io.Writer, title string, games []analysis.PlayerGame) {
	fmt.Fprintf(w, "\n  %s\n", title)
	tw := table(w)
	fmt.Fprintln(tw, "  #\tOPPONENT\tCOLOR\tRESULT\tPLIES\tACCURACY\t")
	for _, g := range games {
		fmt.Fprintf(tw, "  #%d\t%.18s\t%s\t%s\t%d\t%.1f%%\t%s\n",
			g.Game, g.Opponent, g.Color, resultLetter[g.Result], g.Plies, g.Accuracy, bar(g.Accuracy))
	}
	tw.Flush()
}

func printPlayer(w io.Writer, r0 *analysis.Report, pr *analysis.PlayerReport, failed int, took time.Duration) {
	fmt.Fprintf(w, "\n  %s (%d games analyzed)", pr.Name, pr.Overall.Games)
	if failed > 0 {
		fmt.Fprintf(w, " (%d failed)", failed)
	}
	fmt.Fprintf(w, "\n%s\n", rule(56))

	tw := table(w)
	fmt.Fprintf(tw, "  Engine\t%s\n", r0.Engine)
	fmt.Fprintf(tw, "  Strength\t%s\n", r0.Strength)
	fmt.Fprintf(tw, "  Time\t%s\n", took)
	tw.Flush()

	fmt.Fprintln(w)
	tw = table(w)
	fmt.Fprintf(tw, "  Accuracy\t%.1f%%\t%s\n", pr.Overall.Accuracy, bar(pr.Overall.Accuracy))
	fmt.Fprintf(tw, "  As White\t%.1f%%\t%s\t(%d games)\n", pr.White.Accuracy, bar(pr.White.Accuracy), pr.White.Games)
	fmt.Fprintf(tw, "  As Black\t%.1f%%\t%s\t(%d games)\n", pr.Black.Accuracy, bar(pr.Black.Accuracy), pr.Black.Games)
	tw.Flush()

	fmt.Fprintln(w)
	tw = table(w)
	fmt.Fprintln(tw, "  MOVES\tPER GAME\tTOTAL")
	for _, c := range analysis.Classes {
		n := pr.Overall.Summary[c]
		name := strings.ToUpper(string(c)[:1]) + string(c)[1:]
		fmt.Fprintf(tw, "  %s\t%.1f\t%d\n", name, float64(n)/float64(pr.Overall.Games), n)
	}
	tw.Flush()

	fmt.Fprintf(w, "\n  Record   W: %d | D: %d | L: %d\n", pr.Wins, pr.Draws, pr.Losses)

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

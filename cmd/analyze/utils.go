package analyze

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/nchandur/ranga-tools/internal/analysis"
)

const (
	labelWidth = 10
	cellWidth  = 8
)

func writeJSON(path string, r *analysis.Report) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("encode report: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// draws a counter on stderr, but only when stderr is a terminal
func progressPrinter(w io.Writer) func(done, total int) {
	if !isTerminal(w) {
		return nil
	}
	return func(done, total int) {
		fmt.Fprintf(w, "\r\033[Kanalyzing position %d/%d", done, total)
	}
}

func clearProgress(w io.Writer) {
	if isTerminal(w) {
		fmt.Fprint(w, "\r\033[K")
	}
}

func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

func printReport(w io.Writer, r *analysis.Report) {
	title := fmt.Sprintf("%s (White)  vs  %s (Black)", orUnknown(r.White), orUnknown(r.Black))
	tableWidth := labelWidth + 2*(2+cellWidth)
	rule := strings.Repeat("─", max(utf8.RuneCountInString(title), tableWidth))

	fmt.Fprintln(w)
	fmt.Fprintf(w, "  %s\n", title)
	fmt.Fprintf(w, "  %s\n\n", rule)

	info := func(label, value string) {
		fmt.Fprintf(w, "  %s %s\n", padRight(label, labelWidth), value)
	}
	info("Engine", r.Engine)
	info("Strength", fmt.Sprintf("%s", r.Strength))
	info("Length", fmt.Sprintf("%d plies", len(r.Moves)))
	fmt.Fprintln(w)

	row := func(label, white, black string) {
		fmt.Fprintf(w, "  %s  %s  %s\n",
			padRight(label, labelWidth),
			padLeft(white, cellWidth),
			padLeft(black, cellWidth))
	}

	row("", "White", "Black")
	fmt.Fprintf(w, "  %s\n", rule)
	row("Accuracy",
		fmt.Sprintf("%.1f%%", r.Accuracy.White),
		fmt.Sprintf("%.1f%%", r.Accuracy.Black))
	fmt.Fprintf(w, "  %s\n", rule)

	sw, sb := r.Summary.White, r.Summary.Black
	row("Blunder", strconv.Itoa(sw.Blunder), strconv.Itoa(sb.Blunder))
	row("Bad", strconv.Itoa(sw.Bad), strconv.Itoa(sb.Bad))
	row("Ok", strconv.Itoa(sw.Ok), strconv.Itoa(sb.Ok))
	row("Good", strconv.Itoa(sw.Good), strconv.Itoa(sb.Good))
	fmt.Fprintln(w)
}

func orUnknown(s string) string {
	if s == "" {
		return "?"
	}
	return s
}

func padLeft(s string, n int) string {
	if d := n - utf8.RuneCountInString(s); d > 0 {
		return strings.Repeat(" ", d) + s
	}
	return s
}

func padRight(s string, n int) string {
	if d := n - utf8.RuneCountInString(s); d > 0 {
		return s + strings.Repeat(" ", d)
	}
	return s
}

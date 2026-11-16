package analyze

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/nchandur/ranga-tools/internal/analysis"
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
	fmt.Fprintf(w, "%s vs %s\n", orUnknown(r.White), orUnknown(r.Black))
	fmt.Fprintf(w, "Engine: %s | strength: %s (depth %d, hash %d MB)\n\n",
		r.Engine, r.Strength, r.Depth, r.HashMB)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "MOVE\tEVAL\tCLASS\tBEST")
	for _, m := range r.Moves {
		best := ""
		if m.Class != analysis.Good && m.BestMove != "" {
			best = m.BestMove
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", moveLabel(m), m.EvalAfter, m.Class, best)
	}
	tw.Flush()

	fmt.Fprintln(w)
	tw = tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "\tBLUNDER\tBAD\tOK\tGOOD")
	counts := func(name string, c analysis.Counts) {
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%d\n", name, c.Blunder, c.Bad, c.Ok, c.Good)
	}
	counts("White", r.Summary.White)
	counts("Black", r.Summary.Black)
	tw.Flush()
}

func moveLabel(m analysis.Move) string {
	if m.Side == "white" {
		return fmt.Sprintf("%d. %s", m.Number, m.SAN)
	}
	return fmt.Sprintf("%d... %s", m.Number, m.SAN)
}

func orUnknown(s string) string {
	if s == "" {
		return "?"
	}
	return s
}

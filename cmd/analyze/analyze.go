package analyze

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/nchandur/ranga-tools/internal/analysis"
	eng "github.com/nchandur/ranga-tools/internal/engine"
	"github.com/nchandur/ranga-tools/internal/pgn"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze the first game of a PGN with a registered engine",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		pgnPath, err := cmd.Flags().GetString("pgn")
		if err != nil {
			return err
		}
		engineName, err := cmd.Flags().GetString("engine")
		if err != nil {
			return err
		}
		strengthFlag, err := cmd.Flags().GetString("strength")
		if err != nil {
			return err
		}
		jsonPath, err := cmd.Flags().GetString("json")
		if err != nil {
			return err
		}

		strength, err := analysis.ParseStrength(strengthFlag)
		if err != nil {
			return err
		}
		entry, err := eng.Lookup(engineName)
		if err != nil {
			return err
		}

		cmd.SilenceUsage = true

		game, err := pgn.ParseFirstGame(pgnPath)
		if err != nil {
			return err
		}

		opts := analysis.Options{
			EnginePath: entry.Path,
			Strength:   strength,
			Progress:   progressPrinter(cmd.ErrOrStderr()),
		}
		report, err := analysis.Analyze(&game, opts)
		clearProgress(cmd.ErrOrStderr())
		if err != nil {
			return err
		}

		if jsonPath != "" {
			if err := writeJSON(jsonPath, report); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Analysis complete. Report written to %s\n", jsonPath)
			return nil
		}

		printReport(cmd.OutOrStdout(), report)
		return nil
	},
}

func Cmd() *cobra.Command { return analyzeCmd }

func init() {
	f := analyzeCmd.Flags()
	f.String("pgn", "", "path to the PGN file (only the first game is analyzed)")
	f.String("engine", "", "registered name of the engine to analyze with")
	f.String("strength", string(analysis.Medium), "analysis strength: low, medium or high")
	f.String("json", "", "write the full report to this JSON file instead of printing it")

	analyzeCmd.MarkFlagRequired("pgn")
	analyzeCmd.MarkFlagRequired("engine")
}

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

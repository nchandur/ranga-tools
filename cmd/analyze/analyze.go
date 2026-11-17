package analyze

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/nchandur/ranga-tools/internal/analysis"
	eng "github.com/nchandur/ranga-tools/internal/engine"
	"github.com/nchandur/ranga-tools/internal/pgn"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "analyze the first game of a PGN with a registered engine",
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
		start := time.Now()
		report, err := analysis.Analyze(&game, opts)
		duration := time.Since(start)

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

		printReport(cmd.OutOrStdout(), report, duration)
		return nil
	},
}

func Cmd() *cobra.Command { return analyzeCmd }

func init() {
	f := analyzeCmd.Flags()
	f.String("pgn", "", "path to the PGN file (only the first game is analyzed)")
	f.String("engine", "", "registered name of the engine to analyze with")
	f.String("strength", string(analysis.Low), "analysis strength: low, medium or high")
	f.String("json", "", "write the full report to this JSON file instead of printing it")

	analyzeCmd.MarkFlagRequired("pgn")
	analyzeCmd.MarkFlagRequired("engine")
}

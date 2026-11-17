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

		gameNum, err := cmd.Flags().GetInt("game")
		if err != nil {
			return err
		}
		all, err := cmd.Flags().GetBool("all")
		if err != nil {
			return err
		}
		workers, err := cmd.Flags().GetInt("workers")
		if err != nil {
			return err
		}
		if gameNum < 1 || workers < 1 {
			return fmt.Errorf("--game and --workers must be at least 1")
		}

		player, err := cmd.Flags().GetString("player")
		if err != nil {
			return err
		}

		cmd.SilenceUsage = true

		if all || player != "" {
			return runBatch(cmd, pgnPath, entry.Path, strength, workers, jsonPath, player)
		}

		game, err := pgn.ParseGameN(pgnPath, gameNum)
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
		fmt.Fprintf(cmd.OutOrStdout(), "Analysis Completed in %s", formatDuration(duration))
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
	f.Int("game", 1, "which game in the PGN to analyze (1-based)")
	f.Bool("all", false, "analyze every game in the PGN")
	f.Int("workers", 1, "engines to run in parallel with --all (each uses its own hash memory)")
	f.String("player", "", "analyze only this player's games (case-insensitive regex on the PGN names; implies --all)")
	analyzeCmd.MarkFlagsMutuallyExclusive("game", "player")
	analyzeCmd.MarkFlagsMutuallyExclusive("game", "all")

	analyzeCmd.MarkFlagRequired("pgn")
	analyzeCmd.MarkFlagRequired("engine")
}

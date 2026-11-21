package analyze

import (
	"github.com/spf13/cobra"

	"github.com/nchandur/ranga-tools/internal/analysis"
)

var (
	pgnPath, engineName, strengthName, jsonPath, textPath, player string
	workers                                                       int
	full                                                          bool
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze a player's games from a PGN with a registered engine",
	Args:  cobra.NoArgs,
	RunE:  run,
}

func Cmd() *cobra.Command { return analyzeCmd }

func init() {
	f := analyzeCmd.Flags()
	f.StringVar(&pgnPath, "pgn", "", "path to the PGN file")
	f.StringVar(&engineName, "engine", "", "registered name of the engine")
	f.StringVar(&strengthName, "strength", string(analysis.Low), "low, medium or high")
	f.StringVar(&jsonPath, "json", "", "write the summary to this JSON file")
	f.StringVar(&textPath, "text", "", "write the report to this text file (same as the terminal output)")
	f.StringVar(&player, "player", "", "player whose games are analyzed (case-insensitive regex)")
	f.BoolVar(&full, "full", false, "list every game, not just the best and worst")
	f.IntVar(&workers, "workers", 1, "engines to run in parallel")
	analyzeCmd.MarkFlagRequired("pgn")
	analyzeCmd.MarkFlagRequired("engine")
	analyzeCmd.MarkFlagRequired("player")
}

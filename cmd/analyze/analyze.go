package analyze

import (
	"github.com/spf13/cobra"

	"github.com/nchandur/ranga-tools/internal/analysis"
)

var (
	pgnPath, engineName, strengthName, jsonPath, player string
	gameNum, workers                                    int
	all                                                 bool
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze games from a PGN with a registered engine",
	Args:  cobra.NoArgs,
	RunE:  run,
}

func Cmd() *cobra.Command { return analyzeCmd }

func init() {
	f := analyzeCmd.Flags()
	f.StringVar(&pgnPath, "pgn", "", "path to the PGN file")
	f.StringVar(&engineName, "engine", "", "registered name of the engine")
	f.StringVar(&strengthName, "strength", string(analysis.Low), "low, medium or high")
	f.StringVar(&jsonPath, "json", "", "write the full report to this JSON file")
	f.StringVar(&player, "player", "", "analyze one player's games (case-insensitive regex; implies --all)")
	f.IntVar(&gameNum, "game", 1, "which game to analyze (1-based)")
	f.BoolVar(&all, "all", false, "analyze every game")
	f.IntVar(&workers, "workers", 1, "engines to run in parallel")
	analyzeCmd.MarkFlagRequired("pgn")
	analyzeCmd.MarkFlagRequired("engine")
	analyzeCmd.MarkFlagsMutuallyExclusive("game", "all")
	analyzeCmd.MarkFlagsMutuallyExclusive("game", "player")
}

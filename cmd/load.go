package cmd

import (
	"fmt"

	"github.com/nchandur/ranga-tools/internal/uci"
	"github.com/spf13/cobra"
)

var loadCmd = &cobra.Command{
	Use:   "load [path-to-engine]",
	Short: "Loads and verifies a UCI chess engine executable",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]

		fmt.Printf("Attempting to load engine from: %s\n", path)

		client, err := uci.Load(path)
		if err != nil {
			return fmt.Errorf("failed to load engine: %w", err)
		}

		appState.EngineClient = client

		fmt.Println("Engine successfully loaded and verified")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loadCmd)
}

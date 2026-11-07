package cmd

import (
	"fmt"
	"os"

	"github.com/nchandur/ranga-tools/internal/uci"
	"github.com/spf13/cobra"
)

// AppContext stores persistent state across commands
type AppContext struct {
	EngineClient *uci.Client
}

var appState = &AppContext{}

var rootCmd = &cobra.Command{
	Use:   "chess-cli",
	Short: "A CLI tool for analyzing PGN games with UCI engines",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

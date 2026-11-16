package cmd

import (
	"fmt"
	"os"

	"github.com/nchandur/ranga-tools/cmd/analyze"
	"github.com/nchandur/ranga-tools/cmd/config"
	"github.com/nchandur/ranga-tools/cmd/engine"
	"github.com/nchandur/ranga-tools/cmd/tournament"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ranga-tools",
	Short: "A CLI for a doing a bunch of stuff with UCI chess engines",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(engine.Cmd())
	rootCmd.AddCommand(tournament.Cmd())
	rootCmd.AddCommand(config.Cmd())
	rootCmd.AddCommand(analyze.Cmd())
}

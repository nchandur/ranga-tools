package config

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nchandur/ranga-tools/internal/settings"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage tool settings",
}

func Cmd() *cobra.Command { return configCmd }

var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a setting (e.g. fastchess-path)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := settings.Set(args[0], args[1]); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "set %s\n", args[0])
		return nil
	},
}

var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Print a setting",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, err := settings.Get(args[0])
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), v)
		return nil
	},
}

func init() {
	configCmd.AddCommand(setCmd, getCmd)
}

package engine

import (
	"github.com/spf13/cobra"
)

var engineCmd = &cobra.Command{
	Use:   "engine",
	Short: "manage UCI chess engines",
}

func Cmd() *cobra.Command {
	return engineCmd
}

func init() {
	engineCmd.AddCommand(registerCmd)
	engineCmd.AddCommand(deregisterCmd)
	engineCmd.AddCommand(listCmd)
}

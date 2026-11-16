package engine

import (
	"github.com/spf13/cobra"

	eng "github.com/nchandur/ranga-tools/internal/engine"
)

var useCmd = &cobra.Command{
	Use:   "use",
	Short: "run registered engine",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		cmd.SilenceUsage = true
		return eng.Use(name)

	},
}

func init() {
	useCmd.Flags().String("name", "", "name of the engine to run")
	useCmd.MarkFlagRequired("name")
}

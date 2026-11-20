package engine

import (
	"github.com/spf13/cobra"

	eng "github.com/nchandur/ranga-tools/internal/engine"
)

var benchCmd = &cobra.Command{
	Use:   "bench",
	Short: "run registered engine benchmark",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		cmd.SilenceUsage = true
		return eng.Bench(name)
	},
}

func init() {
	benchCmd.Flags().String("name", "", "name of the engine to benchmark")
	benchCmd.MarkFlagRequired("name")
}

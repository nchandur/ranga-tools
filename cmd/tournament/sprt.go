package tournament

import (
	"github.com/spf13/cobra"

	"github.com/nchandur/ranga-tools/internal/tournament"
)

var sprtCmd = &cobra.Command{
	Use:   "sprt",
	Short: "Run an SPRT between an RC engine and a base engine",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}
		fc, err := cmd.Flags().GetString("fastchess")
		if err != nil {
			return err
		}
		cmd.SilenceUsage = true
		return tournament.RunSPRT(cfg, fc)
	},
}

func init() {
	f := sprtCmd.Flags()
	f.String("config", "", "path to the game config file")
	f.String("fastchess", "", "path to the fastchess binary (overrides the stored setting)")
	sprtCmd.MarkFlagRequired("config")

}

package tournament

import (
	"github.com/spf13/cobra"
)

var tournamentCmd = &cobra.Command{
	Use:   "tournament",
	Short: "run engine tournaments through fastchess",
}

func Cmd() *cobra.Command { return tournamentCmd }

func init() {
	tournamentCmd.AddCommand(sprtCmd)
	tournamentCmd.AddCommand(selfTestCmd)
	tournamentCmd.AddCommand(matchCmd)
}

func newTournamentCmd(use, short string, run func(config, fastchess string) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
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
			return run(cfg, fc)
		},
	}
	f := cmd.Flags()
	f.String("config", "", "path to the game config file")
	f.String("fastchess", "", "path to the fastchess binary (overrides the stored setting)")
	return cmd
}

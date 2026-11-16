package tournament

import "github.com/spf13/cobra"

var tournamentCmd = &cobra.Command{
	Use:   "tournament",
	Short: "Run engine tournaments through fastchess",
}

func Cmd() *cobra.Command { return tournamentCmd }

func init() {
	tournamentCmd.AddCommand(sprtCmd)
}

package engine

import (
	"fmt"

	eng "github.com/nchandur/ranga-tools/internal/engine"
	"github.com/spf13/cobra"
)

var deregisterCmd = &cobra.Command{
	Use:   "deregister",
	Short: "remove registered engine",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		name, err := cmd.Flags().GetString("name")

		if err != nil {
			return err
		}

		if err := eng.Deregister(name); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Deleted %q\n", name)
		return nil
	},
}

func init() {

	var name string

	deregisterCmd.Flags().StringVar(&name, "name", "", "name of the engine to delete")
	deregisterCmd.MarkFlagRequired("name")
}

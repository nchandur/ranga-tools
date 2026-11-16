package engine

import (
	"fmt"

	eng "github.com/nchandur/ranga-tools/internal/engine"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered engines",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, err := eng.List()
		if err != nil {
			return err
		}
		if len(entries) == 0 {
			fmt.Fprintln(cmd.ErrOrStderr(), "No engines registered.")
			return nil
		}
		for _, e := range entries {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", e.Name, e.Author)
		}
		return nil
	},
}

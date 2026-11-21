package engine

import (
	"fmt"
	"text/tabwriter"

	eng "github.com/nchandur/ranga-tools/internal/engine"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list registered engines",
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

		tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
		defer tw.Flush()
		fmt.Fprintln(tw, "NAME\tAUTHOR\tREGISTERED")
		for _, e := range entries {
			author := e.Author
			if author == "" {
				author = "-"
			}
			fmt.Fprintf(tw, "%s\t%s\t%s\n", e.Name, author, e.Created.Format("2006-01-02"))
		}
		return nil
	},
}

package engine

import (
	"fmt"
	"os"

	eng "github.com/nchandur/ranga-tools/internal/engine"
	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "verify and register UCI engine",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		path, err := cmd.Flags().GetString("path")

		if err != nil {
			return err
		}

		name, err := cmd.Flags().GetString("name")

		if err != nil {
			return err
		}

		force, err := cmd.Flags().GetBool("force")

		if err != nil {
			return err
		}

		entry, err := eng.Register(path, name, force)

		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "registered %s (%s)\n", entry.Name, entry.Path)
		return nil
	},
}

func init() {

	var path, name string
	var force bool

	f := registerCmd.Flags()
	f.StringVar(&path, "path", "", "path to the engine executable")
	f.StringVar(&name, "name", "", "override the name reported by the engine")
	f.BoolVar(&force, "force", false, "overwrite an existing engine with the same name")
	registerCmd.MarkFlagRequired("path")
}

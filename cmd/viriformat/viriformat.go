package viriformat

import (
	"fmt"

	"github.com/spf13/cobra"

	vf "github.com/nchandur/ranga-tools/internal/viriformat"
)

var viriformatCmd = &cobra.Command{
	Use:   "viriformat",
	Short: "Utilities for viriformat (.vf) training data",
}

func Cmd() *cobra.Command { return viriformatCmd }

var dedupeCmd = &cobra.Command{
	Use:   "dedupe",
	Short: "Remove duplicate games from viriformat files",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		inputs, err := cmd.Flags().GetStringArray("input")
		if err != nil {
			return err
		}
		output, err := cmd.Flags().GetString("output")
		if err != nil {
			return err
		}
		check, err := cmd.Flags().GetBool("check")
		if err != nil {
			return err
		}

		cmd.SilenceUsage = true

		st, err := vf.Dedupe(inputs, output)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "kept %d, dropped %d, result conflicts %d\n",
			st.Kept, st.Dropped, st.Conflicts)

		if check {
			games, err := vf.Check(output)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ok, %d games\n", games)
		}
		return nil
	},
}

var shuffleCmd = &cobra.Command{
	Use:   "shuffle",
	Short: "Shuffle a viriformat file at game level",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		input, err := cmd.Flags().GetString("input")
		if err != nil {
			return err
		}
		output, err := cmd.Flags().GetString("output")
		if err != nil {
			return err
		}
		seed, err := cmd.Flags().GetInt64("seed")
		if err != nil {
			return err
		}
		expect := -1
		if cmd.Flags().Changed("expect-games") {
			if expect, err = cmd.Flags().GetInt("expect-games"); err != nil {
				return err
			}
		}

		cmd.SilenceUsage = true

		st, err := vf.Shuffle(input, output, seed, expect)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "parsed %d games, %d bytes\ndone\n", st.Games, st.Bytes)
		return nil
	},
}

func init() {
	df := dedupeCmd.Flags()
	df.StringArrayP("input", "i", nil, "input .vf file (repeat for several; duplicates are caught across all of them)")
	df.StringP("output", "o", "", "output .vf file (must not be one of the inputs)")
	df.BoolP("check", "c", false, "verify the output file's framing after writing")
	dedupeCmd.MarkFlagRequired("input")
	dedupeCmd.MarkFlagRequired("output")

	sf := shuffleCmd.Flags()
	sf.StringP("input", "i", "", "input .vf file")
	sf.StringP("output", "o", "", "output .vf file (must not already exist)")
	sf.Int64("seed", 0, "shuffle seed (same seed, same order)")
	sf.Int("expect-games", 0, "abort if the parsed game count differs (sanity check)")
	shuffleCmd.MarkFlagRequired("input")
	shuffleCmd.MarkFlagRequired("output")

	viriformatCmd.AddCommand(dedupeCmd, shuffleCmd)
}

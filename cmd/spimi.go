package cmd

import (
	"bufio"
	"os"

	"github.com/just-hms/pulse/pkg/spimi"
	"github.com/just-hms/pulse/pkg/spimi/readers"
	"github.com/spf13/cobra"
)

var spimiCmd = &cobra.Command{
	Use:     "spimi",
	Short:   "generate the indexes",
	Example: "`cat dataset | pulse spimi` or `pulse spimi ./path/to/dataset",
	Args:    cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f := os.Stdin
		if len(args) == 1 {
			var err error
			f, err = os.Open(args[1])
			if err != nil {
				return err
			}
		}

		r := readers.NewMsMarco(bufio.NewReader(f), 50_000)

		if err := os.RemoveAll("data/dump"); err != nil {
			return err
		}

		if err := spimi.Parse(r, 16, "data/dump"); err != nil {
			return err
		}

		if err := spimi.Merge("data/dump"); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(spimiCmd)
}

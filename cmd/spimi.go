package cmd

import (
	"bufio"
	"os"

	"github.com/just-hms/pulse/pkg/spimi"
	"github.com/just-hms/pulse/pkg/spimi/readers"
	"github.com/spf13/cobra"
)

var (
	workersFlag   uint
	chunkSizeFlag uint
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

		r := readers.NewMsMarco(bufio.NewReader(f), int(chunkSizeFlag))

		if err := os.RemoveAll("data/dump"); err != nil {
			return err
		}

		if err := spimi.Parse(r, int(workersFlag), "data/dump"); err != nil {
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

	spimiCmd.Flags().UintVarP(&workersFlag, "workers", "w", 16, "number of workers indexing the dataset")
	spimiCmd.Flags().UintVarP(&chunkSizeFlag, "chunk", "c", 50_000, "reader chunk size")
}

package cmd

import (
	"bufio"
	"os"

	"github.com/just-hms/pulse/config"
	"github.com/just-hms/pulse/pkg/spimi"
	"github.com/just-hms/pulse/pkg/spimi/readers"
	"github.com/just-hms/pulse/pkg/spimi/stats"
	"github.com/spf13/cobra"
)

var (
	workersFlag     uint
	chunkSizeFlag   uint
	noStopWordsFlag bool
	noStemmingFlag  bool
)

var spimiCmd = &cobra.Command{
	Use:     "spimi",
	Short:   "generate the indexes",
	Example: "cat dataset | pulse spimi\npulse spimi ./path/to/dataset\ntar xOf dataset.tar.gz | pulse spimi",
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

		if err := os.RemoveAll(config.DATA_FOLDER); err != nil {
			return err
		}

		msMarcoReader := readers.NewMsMarco(bufio.NewReader(f), int(chunkSizeFlag))
		settings := stats.IndexingSettings{
			Stemming:         !noStemmingFlag,
			StopWordsRemoval: !noStopWordsFlag,
		}

		if err := spimi.Parse(msMarcoReader, int(workersFlag), config.DATA_FOLDER, settings); err != nil {
			return err
		}

		if err := spimi.Merge(config.DATA_FOLDER); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(spimiCmd)

	spimiCmd.Flags().UintVarP(&workersFlag, "workers", "w", 16, "number of workers indexing the dataset")
	spimiCmd.Flags().UintVarP(&chunkSizeFlag, "chunk", "c", 50_000, "reader chunk size")

	spimiCmd.Flags().BoolVar(&noStopWordsFlag, "no-stopwords", false, "remove stopwords")
	spimiCmd.Flags().BoolVar(&noStemmingFlag, "no-stemming", false, "remove stemming")
}

package cmd

import (
	"bufio"
	"os"

	"github.com/just-hms/pulse/config"
	"github.com/just-hms/pulse/pkg/spimi"
	"github.com/just-hms/pulse/pkg/spimi/reader"
	"github.com/just-hms/pulse/pkg/spimi/stats"
	"github.com/spf13/cobra"
)

var (
	workersFlag         uint
	chunkSizeFlag       uint
	memoryThresholdFlag uint
	noStopWordsFlag     bool
	noStemmingFlag      bool
	noCompressionFlag   bool
)

var spimiCmd = &cobra.Command{
	Use:   "spimi",
	Short: "generate the indexes",

	Example: `  cat dataset | pulse spimi
  pulse spimi ./path/to/dataset
  tar xOf dataset.tar.gz | pulse spimi`,

	Args: cobra.RangeArgs(0, 1),
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

		msMarcoReader := reader.NewMsMarco(bufio.NewReader(f), int(chunkSizeFlag))
		settings := stats.IndexingSettings{
			Stemming:          !noStemmingFlag,
			StopWordsRemoval:  !noStopWordsFlag,
			Compression:       !noCompressionFlag,
			MemoryThresholdMB: int(memoryThresholdFlag),
		}

		spimiBuilder := spimi.NewBuilder(settings)

		if err := spimiBuilder.Parse(msMarcoReader, int(workersFlag), config.DATA_FOLDER); err != nil {
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
	spimiCmd.Flags().UintVarP(&memoryThresholdFlag, "max-memory", "m", spimi.DefaulMemoryThreshold, "max memory used during indexing [MB]")

	spimiCmd.Flags().BoolVar(&noCompressionFlag, "no-compression", false, "compress the posting list and term frequencies")
	spimiCmd.Flags().BoolVar(&noStopWordsFlag, "no-stopwords", false, "remove stopwords removal")
	spimiCmd.Flags().BoolVar(&noStemmingFlag, "no-stemming", false, "remove stemming")
}

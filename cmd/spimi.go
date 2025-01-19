package cmd

import (
	"bufio"
	"os"

	"github.com/just-hms/pulse/config"
	"github.com/just-hms/pulse/pkg/preprocess"
	"github.com/just-hms/pulse/pkg/spimi"
	"github.com/just-hms/pulse/pkg/spimi/reader"
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

// spimiCmd is the subcommand responsible to index a given dataset
var spimiCmd = &cobra.Command{
	Use:   "spimi",
	Short: "generate the indexes",

	Example: `  cat dataset | pulse spimi
  pulse spimi ./path/to/dataset
  tar xOf dataset.tar.gz | pulse spimi`,

	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {

		// read from stdin or from file if 1 argument is present
		f := os.Stdin
		if len(args) == 1 {
			var err error
			f, err = os.Open(args[1])
			if err != nil {
				return err
			}
		}

		// clean up the data folder
		if err := os.RemoveAll(config.DATA_FOLDER); err != nil {
			return err
		}

		// create a msMarco reader
		msMarcoReader := reader.NewMsMarco(bufio.NewReader(f), int(chunkSizeFlag))

		// create the indexing settings
		settings := spimi.IndexingSettings{
			PreprocessSettings: preprocess.Settings{
				Stemming:         !noStemmingFlag,
				StopWordsRemoval: !noStopWordsFlag,
			},
			Compression:       !noCompressionFlag,
			MemoryThresholdMB: int(memoryThresholdFlag),
			NumWorkers:        int(workersFlag),
		}

		spimiBuilder := spimi.NewBuilder(settings)

		// parse the dataset using the created reader and create indexing files inside the config.DATA_FOLDER
		if err := spimiBuilder.Parse(msMarcoReader, config.DATA_FOLDER); err != nil {
			return err
		}

		// spimi merge phase
		if err := spimi.Merge(config.DATA_FOLDER); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(spimiCmd)

	// configuration/performance flags
	{
		spimiCmd.Flags().UintVarP(&workersFlag, "workers", "w", 16, "number of workers indexing the dataset")
		spimiCmd.Flags().UintVarP(&chunkSizeFlag, "chunk", "c", 50_000, "reader chunk size")
		spimiCmd.Flags().UintVarP(&memoryThresholdFlag, "max-memory", "m", spimi.DefaulMemoryThreshold, "max memory used during indexing [MB]")
	}

	// preprocessing flags
	{
		spimiCmd.Flags().BoolVar(&noCompressionFlag, "no-compression", false, "compress the posting list and term frequencies")
		spimiCmd.Flags().BoolVar(&noStopWordsFlag, "no-stopwords", false, "remove stopwords removal")
		spimiCmd.Flags().BoolVar(&noStemmingFlag, "no-stemming", false, "remove stemming")
	}
}

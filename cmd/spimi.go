package cmd

import (
	"bufio"
	"os"
	"unicode"

	"github.com/just-hms/pulse/config"
	"github.com/just-hms/pulse/pkg/spimi"
	"github.com/just-hms/pulse/pkg/spimi/readers"
	"github.com/spf13/cobra"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	workersFlag   uint
	chunkSizeFlag uint
)

var unicodeNormalizer = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

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

		// given the file descriptor create a buffered reader, with a chanined unicode normalizer
		r := bufio.NewReader(transform.NewReader(f, unicodeNormalizer))
		msMarcoReader := readers.NewMsMarco(r, int(chunkSizeFlag))

		if err := spimi.Parse(msMarcoReader, int(workersFlag), config.DATA_FOLDER); err != nil {
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
}

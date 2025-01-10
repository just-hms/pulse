package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"iter"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/just-hms/pulse/config"
	"github.com/just-hms/pulse/pkg/engine"
	"github.com/spf13/cobra"
)

var (
	kFlag           uint
	metricFlag      string
	interactiveFlag bool
)

func InteractiveArgs() iter.Seq2[int, string] {
	i := 0
	reader := bufio.NewReader(os.Stdin)

	return func(yield func(int, string) bool) {
		defer func() {
			i++
		}()
		for {
			fmt.Print("> ")
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimSpace(line)
			if !yield(i, line) {
				return
			}
		}
	}
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "do a query",
	Args:  cobra.RangeArgs(0, 1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 && !interactiveFlag {
			return errors.New("either write a query or use -i flag")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		// Parse the metric
		metric, err := engine.ParseMetric(metricFlag)
		if err != nil {
			return err
		}

		e, err := engine.Load(config.DATA_FOLDER)
		if err != nil {
			return err
		}

		queries := slices.All(args)
		if interactiveFlag {
			queries = InteractiveArgs()
		}

		// Process each search argument
		for _, arg := range queries {
			start := time.Now()
			// Perform the search
			res, err := e.Search(arg, &engine.Settings{
				K:      int(kFlag),
				Metric: metric,
			})
			if err != nil {
				return err
			}

			// Print the first result
			for _, doc := range res {
				fmt.Println(doc)
			}
			log.Println(time.Since(start))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().UintVarP(&kFlag, "doc2ret", "k", 10, "number of documents to be returned")
	searchCmd.Flags().BoolVarP(&interactiveFlag, "interactive", "i", false, "interactive search")
	searchCmd.Flags().StringVarP(&metricFlag, "metric", "m", "BM25",
		fmt.Sprintf("score metric to be used [%s]", strings.Join(engine.AllowedMetrics, "|")),
	)
}

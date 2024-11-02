package cmd

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"
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

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "do a query",
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

		if profileFlag {
			f, err := os.Create("pulse.prof")
			if err != nil {
				return err
			}
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}

		// Process each search argument
		for _, arg := range args {

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

	searchCmd.Flags().BoolVarP(&profileFlag, "profile", "p", false, "profile the execution")
	searchCmd.Flags().UintVarP(&kFlag, "doc2ret", "k", 10, "number of documents to be returned")
	searchCmd.Flags().BoolVarP(&interactiveFlag, "interactive", "i", false, "interactive search")
	searchCmd.Flags().StringVarP(&metricFlag, "metric", "m", "BM25",
		fmt.Sprintf("score metric to be used [%s]", strings.Join(engine.AllowedMetrics, "|")),
	)
}

package cmd

import (
	"fmt"
	"iter"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/just-hms/pulse/config"
	"github.com/just-hms/pulse/pkg/engine"
	"github.com/just-hms/pulse/pkg/query"
	"github.com/spf13/cobra"
)

var (
	kFlag           uint
	metricFlag      string
	interactiveFlag bool
	conjuctiveFlag  bool
	singleQueryFlag string
	fileFlag        string
)

// searchCmd is the subcommand responsible to execute one or more query
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "do one or more queries",

	Example: `  pulse search -q "who is the president right now" -m TFIDF
  pulse search -i`,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if fileFlag != "" {
			f, err := os.Stat(fileFlag)
			if err != nil {
				return err
			}
			if f.IsDir() {
				return fmt.Errorf("provided path %s is a folder", fileFlag)
			}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		// Parse the metric
		metric, err := engine.ParseMetric(metricFlag)
		if err != nil {
			return err
		}

		// Load the indexing data
		e, err := engine.Load(config.DATA_FOLDER)
		if err != nil {
			return err
		}

		// create a query iterator depending on the query mode flags
		var queries iter.Seq2[int, query.Query]

		switch {
		case interactiveFlag:
			queries = query.InteractiveQueries()
		case fileFlag != "":
			fqueries, err := query.FileQueries(fileFlag)
			if err != nil {
				return err
			}
			queries = slices.All(fqueries)
		case singleQueryFlag != "":
			queries = slices.All([]query.Query{{Value: singleQueryFlag}})
		}

		// process each query
		for _, query := range queries {

			// perform the search
			start := time.Now()
			res, err := e.Search(query.Value, &engine.SearchSettings{
				K:           int(kFlag),
				Metric:      metric,
				Conjunctive: conjuctiveFlag,
			})
			if err != nil {
				return err
			}
			end := time.Since(start)

			// print each result
			for rank, doc := range res {
				// result format:
				// <query_id> 0 <doc_id> <rank> <score> <run_id>
				fmt.Printf("%d\tQ0\t%s\t%d\t%.4f\tRANDOMID\n", query.ID, doc.No(), rank, doc.Score)
			}

			// print the query execution time
			// elapsed time format:
			// # <query_id> <elapsed_time> <elapsed_time_microseconds>
			fmt.Printf("#\t%d\t%v\t%d\n", query.ID, end, end.Microseconds())
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// query input flags (they determine how the queries are read by the program)
	{
		searchCmd.Flags().BoolVarP(&interactiveFlag, "interactive", "i", false, "interactive search")
		searchCmd.Flags().StringVarP(&fileFlag, "file", "f", "", "add a queries file")
		searchCmd.Flags().StringVarP(&singleQueryFlag, "query", "q", "", "single search")
	}

	// search flags
	{
		searchCmd.Flags().UintVarP(&kFlag, "doc2ret", "k", 10, "number of documents to be returned")
		searchCmd.Flags().BoolVarP(&conjuctiveFlag, "conjunctive", "c", false, "search in conjunctive mode")
		searchCmd.Flags().StringVarP(&metricFlag, "metric", "m", "BM25",
			fmt.Sprintf("score metric to be used [%s]", strings.Join(engine.AllowedMetrics, "|")),
		)
	}

	searchCmd.MarkFlagsMutuallyExclusive("file", "interactive", "query")
	searchCmd.MarkFlagsOneRequired("file", "interactive", "query")
}

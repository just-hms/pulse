package cmd

import (
	"fmt"
	"strings"

	"github.com/just-hms/pulse/config"
	"github.com/just-hms/pulse/pkg/engine"
	"github.com/spf13/cobra"
)

var (
	kFlag      uint
	metricFlag string
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "do a query",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		metric, err := engine.ParseMetric(metricFlag)
		if err != nil {
			return err
		}
		e, err := engine.Load(config.DATA_FOLDER)
		if err != nil {
			return err
		}
		res, err := e.Search(args[0], &engine.Settings{
			K:      int(kFlag),
			Metric: metric,
		})
		if err != nil {
			return err
		}
		for _, doc := range res {
			fmt.Println(doc)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().UintVarP(&kFlag, "doc2ret", "k", 10, "number of documents to be returned")
	searchCmd.Flags().StringVarP(&metricFlag, "metric", "m", "BM25",
		fmt.Sprintf("score metric to be used [%s]", strings.Join(engine.AllowedMetrics, "|")),
	)
}

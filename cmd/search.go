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
		m, err := engine.ParseMetric(metricFlag)
		if err != nil {
			return err
		}
		res, err := engine.Search(args[0], config.DATA_FOLDER, &engine.Settings{
			K:      int(kFlag),
			Metric: m,
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

package cmd

import (
	"bufio"
	"fmt"
	"iter"
	"log"
	"os"
	"slices"
	"strconv"
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
	singleQueryFlag string
	fileFlag        string
)

type query struct {
	id    uint32
	value string
}

func FileQueries(path string) ([]query, error) {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var queries []query

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Split the line into parts
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			log.Printf("format error discarding line: %s\n", line)
			continue
		}

		// Parse query_id and relevance_score
		queryID, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			panic(err)
		}
		queries = append(queries, query{id: uint32(queryID), value: parts[1]})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return queries, nil
}

func InteractiveQueries() iter.Seq2[int, query] {
	i := 0
	reader := bufio.NewReader(os.Stdin)

	return func(yield func(int, query) bool) {
		defer func() {
			i++
		}()
		for {
			log.Print("> ")
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimSpace(line)
			q := query{
				value: line,
				id:    uint32(i),
			}
			if !yield(i, q) {
				return
			}
		}
	}
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "do one or more queries",
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

		e, err := engine.Load(config.DATA_FOLDER)
		if err != nil {
			return err
		}

		var queries iter.Seq2[int, query]

		switch {
		case interactiveFlag:
			queries = InteractiveQueries()
			break
		case fileFlag != "":
			fqueries, err := FileQueries(fileFlag)
			if err != nil {
				return err
			}
			queries = slices.All(fqueries)
			break
		case singleQueryFlag != "":
			queries = slices.All([]query{{value: singleQueryFlag}})
		}

		// Process each search argument
		for _, query := range queries {
			start := time.Now()
			// Perform the search
			res, err := e.Search(query.value, &engine.Settings{
				K:      int(kFlag),
				Metric: metric,
			})
			if err != nil {
				return err
			}

			// Print the first result
			for rank, doc := range res {
				// result file format:
				// <query_id> 0 <doc_id> <rank> <score> <run_id>
				fmt.Printf("%d\tQ0\t%s\t%d\t%.4f\tRANDOMID\n", query.id, doc.No(), rank, doc.Score)
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
	searchCmd.Flags().StringVarP(&singleQueryFlag, "query", "q", "", "single search")
	searchCmd.Flags().StringVarP(&fileFlag, "file", "f", "", "add a queries file")

	searchCmd.Flags().StringVarP(&metricFlag, "metric", "m", "BM25",
		fmt.Sprintf("score metric to be used [%s]", strings.Join(engine.AllowedMetrics, "|")),
	)

	searchCmd.MarkFlagsMutuallyExclusive("file", "interactive", "query")
	searchCmd.MarkFlagsOneRequired("file", "interactive", "query")
}

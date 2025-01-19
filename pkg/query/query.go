package query

import (
	"bufio"
	"fmt"
	"iter"
	"log"
	"os"
	"strconv"
	"strings"
)

type Query struct {
	ID    uint32
	Value string
}

// FileQueries return a list of query given a file path
func FileQueries(path string) ([]Query, error) {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var queries []Query

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// split the line into parts
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			log.Printf("format error discarding line: %s\n", line)
			continue
		}

		// parse query_id and relevance_score
		queryID, err := strconv.ParseUint(parts[0], 10, 32)
		if err != nil {
			panic(err)
		}
		queries = append(queries, Query{ID: uint32(queryID), Value: parts[1]})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return queries, nil
}

// InteractiveQueries returns an iterator of queries
//
// each query is given in input by the user by pressing `Enter` after the `>`
func InteractiveQueries() iter.Seq2[int, Query] {
	i := 0
	reader := bufio.NewReader(os.Stdin)

	return func(yield func(int, Query) bool) {
		defer func() {
			i++
		}()
		for {
			fmt.Fprintf(os.Stderr, "> ")
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimSpace(line)
			q := Query{
				Value: line,
				ID:    uint32(i),
			}
			if !yield(i, q) {
				return
			}
		}
	}
}

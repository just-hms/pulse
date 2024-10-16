package cmd

import (
	"fmt"

	"github.com/just-hms/pulse/pkg/engine"
	"github.com/spf13/cobra"
)

var (
	k uint
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "do a query",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := engine.Search(args[0], "data/dump", int(k))
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
	searchCmd.Flags().UintVarP(&k, "doc2ret", "k", 10, "number of documents to be returned")
}

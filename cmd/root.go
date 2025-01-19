/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"runtime/pprof"

	"github.com/spf13/cobra"
)

var (
	profileFlag bool
)

// rootCmd is the base command which every https://github.com/spf13/cobra cli program has.
// The subcommands are added to this one and are called by the end user writing:
//
//	the rootCmd.Use followed by the specificCmd.Use
//
// ex:
//
//	pulse search
//
// calls the subcommand search
var rootCmd = &cobra.Command{
	Use:   "pulse",
	Short: "this is the pulse search engine",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if profileFlag {
			f, err := os.Create(".pulse.prof")
			if err != nil {
				return err
			}
			pprof.StartCPUProfile(f)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		pprof.StopCPUProfile()
		return nil
	},
}

func init() {
	rootCmd.Root().CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().BoolVarP(&profileFlag, "profile", "p", false, "profile the execution")
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

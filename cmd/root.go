package cmd

import (
	"os"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cfr",
	Short: "CFR CLI tool",
	Long:  `CFR is a CLI tool for loading and testing problems.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

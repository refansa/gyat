package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gyat",
	Short: "Git Your Ass Together — a git submodule manager",
	Long: `gyat is a git submodule manager that aggregates multiple related
repositories under one umbrella repository, making them easy to manage
without wrestling with raw git submodule commands.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(syncCmd)
}

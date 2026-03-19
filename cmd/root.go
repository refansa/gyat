package cmd

import (
	"github.com/spf13/cobra"
)

// Version is the current gyat version. It defaults to "dev" when built
// without -ldflags; release builds inject it via:
//
//	go build -ldflags "-X github.com/refansa/gyat/cmd.Version=v0.2.0" .
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "gyat",
	Short:   "Git Your Ass Together — a git submodule manager",
	Version: Version,
	Long: `gyat is a git submodule manager that aggregates multiple related
repositories under one umbrella repository, making them easy to manage
without wrestling with raw git submodule commands.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Version = Version
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(trackCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(syncCmd)
}

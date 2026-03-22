package cmd

import (
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Version is the current gyat version. It defaults to "dev" when built
// without -ldflags; release builds inject it via:
//
//	go build -ldflags "-X github.com/refansa/gyat/cmd.Version=v0.2.0" .
//
// When installed via "go install module@version", the version is read
// automatically from the embedded build info as a fallback.
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
	if Version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok {
			if v := info.Main.Version; v != "" && v != "(devel)" {
				Version = v
			}
		}
	}
	rootCmd.Version = Version
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(trackCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(untrackCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(commitCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(statusCmd)
}

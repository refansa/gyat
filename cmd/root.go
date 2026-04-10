package cmd

import (
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Version is the current gyat version. It defaults to "dev" when built
// without -ldflags; release builds inject it via:
//
//	go build -ldflags "-X github.com/refansa/gyat/v2/cmd.Version=v0.2.0" .
//
// When installed via "go install module@version", the version is read
// automatically from the embedded build info as a fallback.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "gyat",
	Short:   "Git Your Ass Together — an umbrella workspace manager",
	Version: Version,
	Long: `gyat is an umbrella workspace manager for multi-repository projects.

It helps you organize related repositories under one root workspace so they can
be managed together without stressing on the chaos that you've created.`,
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
	registerBuiltins(rootCmd)
}

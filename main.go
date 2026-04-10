package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/refansa/gyat/v2/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if shouldShowHelpHint(err) {
			fmt.Fprintln(os.Stderr, "hint: use 'gyat help' or 'gyat -h' for usage")
		}
		os.Exit(1)
	}
}

func shouldShowHelpHint(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	patterns := []string{
		"unknown command",
		"unknown flag",
		"unknown shorthand flag",
		"accepts ",
		"requires at least",
		"requires at most",
		"requires exactly",
		"required flag",
		"cannot be combined",
		"does not support",
	}

	for _, pattern := range patterns {
		if strings.Contains(message, pattern) {
			return true
		}
	}

	return false
}

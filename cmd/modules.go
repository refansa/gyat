package cmd

import (
	gyatruntime "github.com/refansa/gyat/v2/internal/runtime"
	"github.com/spf13/cobra"
)

// modules.go centralises registration of gyat's built-in CLI commands.
//
// The file provides a small Module implementation (builtinModule) that
// registers every built-in subcommand in one place and a helper
// (registerBuiltins) that wires it into the runtime Registry. Keeping the
// registration together improves discoverability and makes it easy to inject
// additional modules in tests or alternative builds.

// builtinModule implements gyatruntime.Module and registers the set of
// built-in commands on the root cobra.Command. This keeps the command list
// centralised and makes it straightforward to add/remove built-ins.
type builtinModule struct{}

func (builtinModule) Register(root *cobra.Command) {
	root.AddCommand(initCmd)
	root.AddCommand(execCmd)
	root.AddCommand(statusCmd)
	root.AddCommand(syncCmd)
	root.AddCommand(updateCmd)
	root.AddCommand(trackCmd)
	root.AddCommand(untrackCmd)
	root.AddCommand(listCmd)
}

// registerBuiltins constructs a Registry containing the builtin module and
// registers it on the provided root command. Tests or other entry points can
// create alternate registries if they need to inject or replace modules.
func registerBuiltins(root *cobra.Command) {
	gyatruntime.NewRegistry(builtinModule{}).Register(root)
}

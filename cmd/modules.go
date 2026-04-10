package cmd

import (
	gyatruntime "github.com/refansa/gyat/v2/internal/runtime"
	"github.com/spf13/cobra"
)

type builtinModule struct{}

func (builtinModule) Register(root *cobra.Command) {
	root.AddCommand(initCmd)
	root.AddCommand(execCmd)
	root.AddCommand(trackCmd)
	root.AddCommand(addCmd)
	root.AddCommand(untrackCmd)
	root.AddCommand(updateCmd)
	root.AddCommand(listCmd)
	root.AddCommand(commitCmd)
	root.AddCommand(syncCmd)
	root.AddCommand(pullCmd)
	root.AddCommand(pushCmd)
	root.AddCommand(rmCmd)
	root.AddCommand(statusCmd)
}

func registerBuiltins(root *cobra.Command) {
	gyatruntime.NewRegistry(builtinModule{}).Register(root)
}

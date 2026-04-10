package runtime

import "github.com/spf13/cobra"

// Module registers one or more commands on the root command.
type Module interface {
	Register(root *cobra.Command)
}

// Registry wires built-in modules into the root command.
type Registry struct {
	modules []Module
}

// NewRegistry creates a registry with the provided modules.
func NewRegistry(modules ...Module) Registry {
	return Registry{modules: append([]Module(nil), modules...)}
}

// Register installs all modules on root.
func (registry Registry) Register(root *cobra.Command) {
	for _, module := range registry.modules {
		if module == nil {
			continue
		}
		module.Register(root)
	}
}

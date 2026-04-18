package runtime

import "github.com/spf13/cobra"

// Module is a small abstraction used to register one or more cobra commands
// onto the root command. This allows grouping related commands into modules
// and makes it easy to inject or replace modules during tests or alternate
// builds.
type Module interface {
	Register(root *cobra.Command)
}

// Registry holds a set of Module implementations and installs them onto a
// root cobra.Command when Register is called. The Registry itself is tiny
// and intentionally simple — it only iterates modules and invokes Register.
type Registry struct {
	modules []Module
}

// NewRegistry returns a Registry pre-populated with the given modules. The
// slice copy ensures callers can safely reuse the input slice without the
// Registry mutating it.
func NewRegistry(modules ...Module) Registry {
	return Registry{modules: append([]Module(nil), modules...)}
}

// Register iterates the configured modules and asks each to register its
// commands on the provided root. Nil modules are ignored for convenience.
func (registry Registry) Register(root *cobra.Command) {
	for _, module := range registry.modules {
		if module == nil {
			continue
		}
		module.Register(root)
	}
}

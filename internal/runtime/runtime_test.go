package runtime

import (
	"testing"

	"github.com/spf13/cobra"
)

type testModule struct {
	use string
}

func (module testModule) Register(root *cobra.Command) {
	root.AddCommand(&cobra.Command{Use: module.use})
}

func TestRegistryRegister(t *testing.T) {
	t.Parallel()

	root := &cobra.Command{Use: "gyat"}
	registry := NewRegistry(testModule{use: "alpha"}, testModule{use: "beta"})

	registry.Register(root)

	if root.CommandPath() != "gyat" {
		t.Fatalf("root command path = %q, want %q", root.CommandPath(), "gyat")
	}
	if root.Commands()[0].Name() != "alpha" && root.Commands()[1].Name() != "alpha" {
		t.Fatalf("registered commands = %#v, want alpha and beta", root.Commands())
	}
	if root.Commands()[0].Name() != "beta" && root.Commands()[1].Name() != "beta" {
		t.Fatalf("registered commands = %#v, want alpha and beta", root.Commands())
	}
}
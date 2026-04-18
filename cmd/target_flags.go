package cmd

import (
	"fmt"
	"strings"

	"github.com/refansa/gyat/v2/internal/workspace"
	"github.com/spf13/cobra"
)

type workspaceTargetFlags struct {
	repoSelectors   []string
	groups          []string
	noRoot          bool
	rootOnly        bool
	parallel        bool
	continueOnError bool
}

var sharedTargetFlags workspaceTargetFlags

// bindWorkspaceTargetFlags attaches the shared workspace selection flags to
// the provided command. These flags control which repositories/groups are
// targeted and whether the umbrella repository is included.
func bindWorkspaceTargetFlags(command *cobra.Command) {
	command.PersistentFlags().StringSliceVar(&sharedTargetFlags.repoSelectors, "repo", nil, "Run only in the specified repo name or path (repeatable)")
	command.PersistentFlags().StringSliceVar(&sharedTargetFlags.groups, "group", nil, "Run only in repos belonging to the specified group (repeatable)")
	command.PersistentFlags().BoolVar(&sharedTargetFlags.noRoot, "no-root", false, "Exclude the umbrella repository from execution")
	command.PersistentFlags().BoolVar(&sharedTargetFlags.rootOnly, "root-only", false, "Run only in the umbrella repository")
	command.PersistentFlags().BoolVar(&sharedTargetFlags.continueOnError, "continue-on-error", false, "Continue running in remaining targets after a failure")
}

// bindWorkspaceParallelFlag attaches the per-command "parallel" flag which
// controls whether per-repository work runs concurrently (while preserving
// output ordering).
func bindWorkspaceParallelFlag(command *cobra.Command) {
	command.Flags().BoolVarP(&sharedTargetFlags.parallel, "parallel", "p", false, "Run per-repository work in parallel while preserving output order")
}

// hasSelection reports whether any selection-related flags are set
// (--repo, --group, --no-root, --root-only).
func (flags workspaceTargetFlags) hasSelection() bool {
	return len(flags.repoSelectors) > 0 || len(flags.groups) > 0 || flags.noRoot || flags.rootOnly
}

// hasAny reports whether any workspace target flags are present. This
// includes selection flags as well as execution flags like --parallel and
// --continue-on-error.
func (flags workspaceTargetFlags) hasAny() bool {
	return flags.hasSelection() || flags.parallel || flags.continueOnError
}

// runOptions converts the workspaceTargetFlags into a workspace.RunOptions
// value suitable for passing to workspace.RunTargets.
func (flags workspaceTargetFlags) runOptions() workspace.RunOptions {
	return workspace.RunOptions{
		ContinueOnError: flags.continueOnError,
		Parallel:        flags.parallel,
	}
}

// targetOptions builds workspace.TargetOptions from the flags. The
// includeRoot parameter is used as a default that the flags can override
// (flags.rootOnly / flags.noRoot). extraRepoSelectors are merged with the
// repo selectors already present in the flags.
func (flags workspaceTargetFlags) targetOptions(includeRoot bool, extraRepoSelectors []string) workspace.TargetOptions {
	if flags.rootOnly {
		includeRoot = true
	}
	if flags.noRoot {
		includeRoot = false
	}

	repoSelectors := append([]string{}, extraRepoSelectors...)
	repoSelectors = append(repoSelectors, flags.repoSelectors...)

	return workspace.TargetOptions{
		IncludeRoot:   includeRoot,
		RootOnly:      flags.rootOnly,
		RepoSelectors: repoSelectors,
		Groups:        append([]string{}, flags.groups...),
	}
}

// validateUnsupported returns an error when any workspace targeting flags are
// set but the named command does not support them. This is used by commands
// that intentionally do not accept selection or execution flags.
func (flags workspaceTargetFlags) validateUnsupported(command string) error {
	if !flags.hasAny() {
		return nil
	}
	return fmt.Errorf("%s does not support --repo, --group, --no-root, --root-only, --parallel, or --continue-on-error", command)
}

// commandFailures collects non-fatal error messages when a command is run
// with --continue-on-error. The helpers below either record an error string
// or return it immediately depending on the continueOnError flag.
type commandFailures []string

// handle formats an error message and either returns it or records it inside
// the failures slice depending on continueOnError.
func (failures *commandFailures) handle(continueOnError bool, format string, args ...any) error {
	err := fmt.Errorf(format, args...)
	return failures.handleErr(continueOnError, err)
}

// handleErr either returns err immediately (when continueOnError is false)
// or records the error's message and returns nil.
func (failures *commandFailures) handleErr(continueOnError bool, err error) error {
	if !continueOnError {
		return err
	}
	*failures = append(*failures, err.Error())
	return nil
}

// err converts any accumulated failures into a single error value. If there
// are no failures nil is returned. If there is one failure it is returned
// directly; otherwise a summarized error listing the failures is returned.
func (failures commandFailures) err(summary string) error {
	switch len(failures) {
	case 0:
		return nil
	case 1:
		return fmt.Errorf("%s", failures[0])
	default:
		return fmt.Errorf("%s in %d target(s): %s", summary, len(failures), strings.Join(failures, "; "))
	}
}

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

func bindWorkspaceTargetFlags(command *cobra.Command) {
	command.PersistentFlags().StringSliceVar(&sharedTargetFlags.repoSelectors, "repo", nil, "Run only in the specified repo name or path (repeatable)")
	command.PersistentFlags().StringSliceVar(&sharedTargetFlags.groups, "group", nil, "Run only in repos belonging to the specified group (repeatable)")
	command.PersistentFlags().BoolVar(&sharedTargetFlags.noRoot, "no-root", false, "Exclude the umbrella repository from execution")
	command.PersistentFlags().BoolVar(&sharedTargetFlags.rootOnly, "root-only", false, "Run only in the umbrella repository")
	command.PersistentFlags().BoolVar(&sharedTargetFlags.continueOnError, "continue-on-error", false, "Continue running in remaining targets after a failure")
}

func bindWorkspaceParallelFlag(command *cobra.Command) {
	command.Flags().BoolVarP(&sharedTargetFlags.parallel, "parallel", "p", false, "Run per-repository work in parallel while preserving output order")
}

func (flags workspaceTargetFlags) hasSelection() bool {
	return len(flags.repoSelectors) > 0 || len(flags.groups) > 0 || flags.noRoot || flags.rootOnly
}

func (flags workspaceTargetFlags) hasAny() bool {
	return flags.hasSelection() || flags.parallel || flags.continueOnError
}

func (flags workspaceTargetFlags) runOptions() workspace.RunOptions {
	return workspace.RunOptions{
		ContinueOnError: flags.continueOnError,
		Parallel:        flags.parallel,
	}
}

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

func (flags workspaceTargetFlags) validateUnsupported(command string) error {
	if !flags.hasAny() {
		return nil
	}
	return fmt.Errorf("%s does not support --repo, --group, --no-root, --root-only, --parallel, or --continue-on-error", command)
}

type commandFailures []string

func (failures *commandFailures) handle(continueOnError bool, format string, args ...any) error {
	err := fmt.Errorf(format, args...)
	return failures.handleErr(continueOnError, err)
}

func (failures *commandFailures) handleErr(continueOnError bool, err error) error {
	if !continueOnError {
		return err
	}
	*failures = append(*failures, err.Error())
	return nil
}

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

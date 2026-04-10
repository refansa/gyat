package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/refansa/gyat/v2/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	execRepoSelectors   []string
	execGroups          []string
	execNoRoot          bool
	execRootOnly        bool
	execContinueOnError bool
)

var execCmd = &cobra.Command{
	Use:   "exec [flags] -- <command> [args...]",
	Short: "Run a command across the umbrella workspace and managed repos",
	Long: `Run an external command across the current gyat workspace.

By default the command runs in the umbrella repository first and then in every
managed repo listed in .gyat. Use --repo and --group to narrow the repo set,
--no-root to exclude the umbrella repository, or --root-only to target only
the umbrella repository.

When command arguments contain flags of their own, place "--" before the
command so gyat stop parsing exec flags.`,
	Example: `  # Run in the umbrella root and every managed repo
  gyat exec -- git status --short

  # Run only in repos tagged "backend"
  gyat exec --group backend -- go test ./...

  # Run only in the auth repo, excluding the umbrella root
  gyat exec --repo auth --no-root -- git status`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get current directory: %w", err)
		}
		return runExec(dir, execRepoSelectors, execGroups, execNoRoot, execRootOnly, execContinueOnError, cmd, args)
	},
}

func runExec(dir string, repoSelectors, groups []string, noRoot, rootOnly, continueOnError bool, cmd *cobra.Command, args []string) error {
	ws, err := workspace.Load(dir)
	if err != nil {
		return err
	}

	targets, err := ws.ResolveTargets(workspace.TargetOptions{
		IncludeRoot:   !noRoot,
		RootOnly:      rootOnly,
		RepoSelectors: repoSelectors,
		Groups:        groups,
	})
	if err != nil {
		return err
	}

	command := workspace.Command{Name: args[0], Args: args[1:]}
	fmt.Fprintf(cmd.ErrOrStderr(), "running '%s' in %d target(s)...\n", command.Display(), len(targets))

	results, runErr := workspace.RunCommand(targets, command, workspace.RunOptions{ContinueOnError: continueOnError})
	printExecResults(cmd.OutOrStdout(), results)

	return runErr
}

func printExecResults(out io.Writer, results []workspace.RunResult) {
	if len(results) == 1 {
		if results[0].Output != "" {
			fmt.Fprintln(out, results[0].Output)
		}
		return
	}

	for index, result := range results {
		header := result.Target.Label
		sep := strings.Repeat("─", utf8.RuneCountInString(header))
		fmt.Fprintln(out, header)
		fmt.Fprintln(out, sep)
		if result.Output != "" {
			fmt.Fprintln(out, result.Output)
		}
		if index < len(results)-1 {
			fmt.Fprintln(out)
		}
	}
}

func init() {
	execCmd.Flags().StringSliceVar(&execRepoSelectors, "repo", nil, "Run only in the specified repo name or path (repeatable)")
	execCmd.Flags().StringSliceVar(&execGroups, "group", nil, "Run only in repos belonging to the specified group (repeatable)")
	execCmd.Flags().BoolVar(&execNoRoot, "no-root", false, "Exclude the umbrella repository from execution")
	execCmd.Flags().BoolVar(&execRootOnly, "root-only", false, "Run only in the umbrella repository")
	execCmd.Flags().BoolVar(&execContinueOnError, "continue-on-error", false, "Continue running in remaining targets after a failure")
}

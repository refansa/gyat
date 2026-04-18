package cmd

import (
	"testing"

	"github.com/refansa/gyat/v2/internal/manifest"
)

func TestMatchTrackedRepo(t *testing.T) {
	repos := []manifest.Repo{
		{Name: "auth", Path: "services/auth"},
		{Name: "billing", Path: "services/billing"},
		{Name: "shared", Path: "services"},
	}

	tests := []struct {
		name           string
		arg            string
		wantRepoPath   string
		wantRepoArg   string
		wantStageAll bool
		wantMatched  bool
	}{
		{
			name:           "exact match repo",
			arg:            "services/auth",
			wantRepoPath:   "services/auth",
			wantRepoArg:    "",
			wantStageAll:   true,
			wantMatched:    true,
		},
		{
			name:           "nested file path",
			arg:            "services/auth/handler.go",
			wantRepoPath:   "services/auth",
			wantRepoArg:    "handler.go",
			wantStageAll:   false,
			wantMatched:    true,
		},
		{
			name:           "nested directory path",
			arg:            "services/billing/internal/api",
			wantRepoPath:   "services/billing",
			wantRepoArg:    "internal/api",
			wantStageAll:   false,
			wantMatched:    true,
		},
		{
			name:           "no match",
			arg:            "README.md",
			wantRepoPath:   "",
			wantRepoArg:   "",
			wantStageAll:  false,
			wantMatched:   false,
		},
		{
			name:           "shorter prefix takes precedence over longer",
			arg:            "services/foo.go",
			wantRepoPath:   "services",
			wantRepoArg:    "foo.go",
			wantStageAll:   false,
			wantMatched:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoPath, repoArg, stageAll, matched := matchTrackedRepo(repos, tt.arg)
			if repoPath != tt.wantRepoPath {
				t.Errorf("matchTrackedRepo() repoPath = %v, want %v", repoPath, tt.wantRepoPath)
			}
			if repoArg != tt.wantRepoArg {
				t.Errorf("matchTrackedRepo() repoArg = %v, want %v", repoArg, tt.wantRepoArg)
			}
			if stageAll != tt.wantStageAll {
				t.Errorf("matchTrackedRepo() stageAll = %v, want %v", stageAll, tt.wantStageAll)
			}
			if matched != tt.wantMatched {
				t.Errorf("matchTrackedRepo() matched = %v, want %v", matched, tt.wantMatched)
			}
		})
	}
}

func TestMatchTrackedRepoNestedRepos(t *testing.T) {
	repos := []manifest.Repo{
		{Name: "services", Path: "services"},
		{Name: "servicesAuth", Path: "services/auth"},
		{Name: "servicesAuthApi", Path: "services/auth/api"},
	}

	tests := []struct {
		name           string
		arg            string
		wantRepoPath   string
		wantRepoArg    string
		wantStageAll   bool
		wantMatched    bool
	}{
		{
			name:           "most specific match wins",
			arg:            "services/auth/api/client.go",
			wantRepoPath:   "services/auth/api",
			wantRepoArg:    "client.go",
			wantStageAll:  false,
			wantMatched:   true,
		},
		{
			name:           "exact match for nested repo",
			arg:            "services/auth",
			wantRepoPath:   "services/auth",
			wantRepoArg:   "",
			wantStageAll:  true,
			wantMatched:   true,
		},
		{
			name:           "services prefix goes to services not services/auth",
			arg:            "services/utils.go",
			wantRepoPath:   "services",
			wantRepoArg:   "utils.go",
			wantStageAll:  false,
			wantMatched:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoPath, repoArg, stageAll, matched := matchTrackedRepo(repos, tt.arg)
			if repoPath != tt.wantRepoPath {
				t.Errorf("matchTrackedRepo() repoPath = %v, want %v", repoPath, tt.wantRepoPath)
			}
			if repoArg != tt.wantRepoArg {
				t.Errorf("matchTrackedRepo() repoArg = %v, want %v", repoArg, tt.wantRepoArg)
			}
			if stageAll != tt.wantStageAll {
				t.Errorf("matchTrackedRepo() stageAll = %v, want %v", stageAll, tt.wantStageAll)
			}
			if matched != tt.wantMatched {
				t.Errorf("matchTrackedRepo() matched = %v, want %v", matched, tt.wantMatched)
			}
		})
	}
}

func TestNormalizeWorkspaceArg(t *testing.T) {
	root := "C:/workspace"
	startDir := "C:/workspace"

	tests := []struct {
		name       string
		arg       string
		wantRel   string
		wantErr   bool
	}{
		{
			name:     "simple relative path",
			arg:      "README.md",
			wantRel:  "README.md",
			wantErr:  false,
		},
		{
			name:     "nested relative path",
			arg:      "services/auth/handler.go",
			wantRel:  "services/auth/handler.go",
			wantErr:  false,
		},
		{
			name:     "path with dots",
			arg:     "./README.md",
			wantRel:  "README.md",
			wantErr:  false,
		},
		{
			name:     "escaping workspace is rejected",
			arg:     "../outside/file.txt",
			wantRel:  "",
			wantErr:  true,
		},
		{
			name:     "empty path returns error",
			arg:     "   ",
			wantRel:  "",
			wantErr:  true,
		},
		{
			name:     "path exactly outside workspace",
			arg:     "..",
			wantRel:  "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel, err := normalizeWorkspaceArg(root, startDir, tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("normalizeWorkspaceArg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && rel != tt.wantRel {
				t.Errorf("normalizeWorkspaceArg() rel = %v, want %v", rel, tt.wantRel)
			}
		})
	}
}

func TestClassifyWorkspaceArgs(t *testing.T) {
	root := "C:/workspace"
	startDir := "C:/workspace"
	repos := []manifest.Repo{
		{Name: "auth", Path: "services/auth"},
		{Name: "billing", Path: "services/billing"},
	}

	tests := []struct {
		name           string
		args           []string
		wantRootArgs   []string
		wantRepoTargets map[string]*repoStage
	}{
		{
			name:         "all root args",
			args:         []string{"README.md", ".gitignore"},
			wantRootArgs: []string{"README.md", ".gitignore"},
			wantRepoTargets: map[string]*repoStage{},
		},
		{
			name:         "all repo args",
			args:         []string{"services/auth", "services/billing"},
			wantRootArgs: []string{},
			wantRepoTargets: map[string]*repoStage{
				"services/auth":    {stageAll: true},
				"services/billing": {stageAll: true},
			},
		},
		{
			name:         "mixed root and repo args",
			args:         []string{"README.md", "services/auth"},
			wantRootArgs: []string{"README.md"},
			wantRepoTargets: map[string]*repoStage{
				"services/auth": {stageAll: true},
			},
		},
		{
			name:         "nested file in repo",
			args:         []string{"services/auth/handler.go"},
			wantRootArgs: []string{},
			wantRepoTargets: map[string]*repoStage{
				"services/auth": {stageAll: false, files: []string{"handler.go"}},
			},
		},
		{
			name:         "multiple files in same repo",
			args:         []string{"services/auth/handler.go", "services/auth/middleware.go"},
			wantRootArgs: []string{},
			wantRepoTargets: map[string]*repoStage{
				"services/auth": {stageAll: false, files: []string{"handler.go", "middleware.go"}},
			},
		},
		{
			name:         "repo path then file in same repo converts to stageAll",
			args:         []string{"services/auth", "services/auth/handler.go"},
			wantRootArgs: []string{},
			wantRepoTargets: map[string]*repoStage{
				"services/auth": {stageAll: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootArgs, repoTargets, err := classifyWorkspaceArgs(root, repos, startDir, tt.args)
			if err != nil {
				t.Fatalf("classifyWorkspaceArgs() unexpected error: %v", err)
			}

			if len(rootArgs) != len(tt.wantRootArgs) {
				t.Errorf("classifyWorkspaceArgs() rootArgs = %v, want %v", rootArgs, tt.wantRootArgs)
			}
			for i, want := range tt.wantRootArgs {
				if i >= len(rootArgs) || rootArgs[i] != want {
					t.Errorf("classifyWorkspaceArgs() rootArgs[%d] = %v, want %v", i, rootArgs[i], want)
				}
			}

			if len(repoTargets) != len(tt.wantRepoTargets) {
				t.Errorf("classifyWorkspaceArgs() repoTargets count = %d, want %d", len(repoTargets), len(tt.wantRepoTargets))
			}
			for path, stage := range tt.wantRepoTargets {
				got, ok := repoTargets[path]
				if !ok {
					t.Errorf("classifyWorkspaceArgs() missing repo target for %s", path)
					continue
				}
				if got.stageAll != stage.stageAll {
					t.Errorf("classifyWorkspaceArgs() stageAll for %s = %v, want %v", path, got.stageAll, stage.stageAll)
				}
				if len(got.files) != len(stage.files) {
					t.Errorf("classifyWorkspaceArgs() files for %s = %v, want %v", path, got.files, stage.files)
				}
			}
		})
	}
}

func TestAppendUnique(t *testing.T) {
	tests := []struct {
		name      string
		values    []string
		value     string
		want      []string
	}{
		{
			name:   "appends new value",
			values: []string{"a", "b"},
			value:  "c",
			want:   []string{"a", "b", "c"},
		},
		{
			name:   "does not append duplicate",
			values: []string{"a", "b"},
			value:  "a",
			want:   []string{"a", "b"},
		},
		{
			name:   "empty slice appends value",
			values: []string{},
			value:  "a",
			want:   []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appendUnique(tt.values, tt.value)
			if len(got) != len(tt.want) {
				t.Errorf("appendUnique() len = %d, want %d", len(got), len(tt.want))
				return
			}
			for i, v := range tt.want {
				if got[i] != v {
					t.Errorf("appendUnique()[%d] = %v, want %v", i, got[i], v)
				}
			}
		})
	}
}

func TestHasWorkingTreeChanges(t *testing.T) {
	tests := []struct {
		name       string
		statusOut  string
		want       bool
	}{
		{
			name:       "has unstaged changes",
			statusOut: " M README.md",
			want:      true,
		},
		{
			name:       "has untracked file",
			statusOut: "?? newfile.txt",
			want:      true,
		},
		{
			name:       "has deleted file",
			statusOut: " D deleted.go",
			want:      true,
		},
		{
			name:       "only staged changes returns false",
			statusOut: "M  staged.txt",
			want:      false,
		},
		{
			name:       "empty output returns false",
			statusOut: "",
			want:      false,
		},
		{
			name:       "multiple unstaged changes",
			statusOut: "M  a.txt\n?? b.txt\nD  c.go",
			want:      true,
		},
		{
			name:       "only staged",
			statusOut: "M  a.txt\nM  b.txt",
			want:      false,
		},
		{
			name:       "ignores carriage return",
			statusOut: "M  file.txt\r\n?? new.txt",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasWorkingTreeChanges(tt.statusOut)
			if got != tt.want {
				t.Errorf("hasWorkingTreeChanges() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepoStage(t *testing.T) {
	t.Run("stageAll true", func(t *testing.T) {
		stage := &repoStage{stageAll: true}
		if !stage.stageAll {
			t.Error("expected stageAll to be true")
		}
		if stage.files != nil {
			t.Error("expected files to be nil when stageAll is true")
		}
	})

	t.Run("stageAll false with files", func(t *testing.T) {
		stage := &repoStage{stageAll: false, files: []string{"a.go", "b.go"}}
		if stage.stageAll {
			t.Error("expected stageAll to be false")
		}
		if len(stage.files) != 2 {
			t.Errorf("expected 2 files, got %d", len(stage.files))
		}
	})
}
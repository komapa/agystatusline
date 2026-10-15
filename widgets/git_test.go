package widgets

import (
	"fmt"
	"testing"

	"github.com/yuys13/agystatusline/types"
)

func TestGitBranchWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("git-branch")
	if w == nil {
		t.Fatalf("Git branch widget not found")
	}
	if w.GetDefaultColor() != "brightMagenta" {
		t.Errorf("Expected default color %q, got %q", "brightMagenta", w.GetDefaultColor())
	}
	if w.GetDisplayName() != "Git Branch" {
		t.Errorf("Expected display name %q, got %q", "Git Branch", w.GetDisplayName())
	}

	settings := types.DefaultSettings()
	dirtyTrue := true
	dirtyFalse := false

	tests := []struct {
		name          string
		gitRunner     func(cmd string, ctx CwdResolver, ttl int) (string, error)
		item          types.WidgetItem
		ctx           types.RenderContext
		expectedTitle string
		expectedBody  string
		expectedColor string
	}{
		{
			name: "Normal git repository branch from command",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "symbolic-ref --short HEAD" {
					return "feature/tdd", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-branch"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/dummy/repo"}},
			expectedTitle: "",
			expectedBody:  "⎇ feature/tdd",
			expectedColor: "brightBlue",
		},
		{
			name: "Custom symbol",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "symbolic-ref --short HEAD" {
					return "main", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-branch", Symbol: "🌿 "},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/dummy/repo"}},
			expectedTitle: "",
			expectedBody:  "🌿 main",
			expectedColor: "brightBlue",
		},
		{
			name: "Branch from VCS telemetry with dirty",
			item: types.WidgetItem{Type: "git-branch"},
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					VCS: &types.VCSInfo{Branch: "main", Dirty: &dirtyTrue},
				},
			},
			expectedTitle: "",
			expectedBody:  "⎇ main*",
			expectedColor: "brightRed",
		},
		{
			name: "Branch from VCS telemetry clean with Raw=true",
			item: types.WidgetItem{Type: "git-branch", Raw: true},
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					VCS: &types.VCSInfo{Branch: "main", Dirty: &dirtyFalse},
				},
			},
			expectedTitle: "",
			expectedBody:  "main",
			expectedColor: "brightBlue",
		},
		{
			name: "Branch from VCS telemetry dirty with Raw=true",
			item: types.WidgetItem{Type: "git-branch", Raw: true},
			ctx: types.RenderContext{
				Data: types.StatusJSON{
					VCS: &types.VCSInfo{Branch: "main", Dirty: &dirtyTrue},
				},
			},
			expectedTitle: "",
			expectedBody:  "main*",
			expectedColor: "brightRed",
		},
		{
			name: "Non-git directory",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				return "false", nil
			},
			item:          types.WidgetItem{Type: "git-branch"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/tmp/non-git-dir"}},
			expectedTitle: "",
			expectedBody:  "⎇ no git",
			expectedColor: "brightBlue",
		},
		{
			name:          "Preview mode fallback to main",
			item:          types.WidgetItem{Type: "git-branch"},
			ctx:           types.RenderContext{IsPreview: true},
			expectedTitle: "",
			expectedBody:  "⎇ main",
			expectedColor: "brightBlue",
		},
	}

	oldRunner := runGitCommand
	defer func() { runGitCommand = oldRunner }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.gitRunner != nil {
				runGitCommand = tt.gitRunner
			}
			title, body, err := w.Render(tt.item, tt.ctx, settings)
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}
			if title != tt.expectedTitle || body != tt.expectedBody {
				t.Errorf("Expected title %q and body %q, got title %q and body %q", tt.expectedTitle, tt.expectedBody, title, body)
			}
			if tt.expectedColor != "" {
				if color := w.GetBodyColor(tt.item, tt.ctx); color != tt.expectedColor {
					t.Errorf("Expected body color %q, got %q", tt.expectedColor, color)
				}
			}
		})
	}
}

func TestGitChangesWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("git-changes")
	if w == nil {
		t.Fatalf("Git changes widget not found")
	}

	if w.GetDefaultColor() != "yellow" {
		t.Errorf("Expected default color %q, got %q", "yellow", w.GetDefaultColor())
	}
	if w.GetDisplayName() != "Git Changes" {
		t.Errorf("Expected display name %q, got %q", "Git Changes", w.GetDisplayName())
	}

	settings := types.DefaultSettings()

	tests := []struct {
		name          string
		gitRunner     func(cmd string, ctx CwdResolver, ttl int) (string, error)
		item          types.WidgetItem
		ctx           types.RenderContext
		expectedTitle string
		expectedBody  string
	}{
		{
			name: "Normal git changes",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "diff --shortstat" {
					return " 2 files changed, 10 insertions(+), 5 deletions(-)", nil
				}
				if cmd == "diff --cached --shortstat" {
					return " 1 file changed, 3 insertions(+)", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-changes"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/dummy/repo"}},
			expectedTitle: "",
			expectedBody:  "(+13,-5)",
		},
		{
			name:          "Preview mode changes",
			item:          types.WidgetItem{Type: "git-changes"},
			ctx:           types.RenderContext{IsPreview: true},
			expectedTitle: "",
			expectedBody:  "(+42,-10)",
		},
		{
			name: "Non-git directory",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				return "false", fmt.Errorf("not a git repository")
			},
			item:          types.WidgetItem{Type: "git-changes"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/tmp/non-git-dir"}},
			expectedTitle: "",
			expectedBody:  "(no git)",
		},
	}

	oldRunner := runGitCommand
	defer func() { runGitCommand = oldRunner }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.gitRunner != nil {
				runGitCommand = tt.gitRunner
			}
			title, body, err := w.Render(tt.item, tt.ctx, settings)
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}
			if title != tt.expectedTitle || body != tt.expectedBody {
				t.Errorf("Expected title %q and body %q, got title %q and body %q", tt.expectedTitle, tt.expectedBody, title, body)
			}
			if color := w.GetBodyColor(tt.item, tt.ctx); color != "yellow" {
				t.Errorf("Expected body color 'yellow', got %q", color)
			}
		})
	}
}

func TestParseShortStat(t *testing.T) {
	tests := []struct {
		name               string
		statInput          string
		expectedInsertions int
		expectedDeletions  int
	}{
		{
			name:               "Empty stat input",
			statInput:          "",
			expectedInsertions: 0,
			expectedDeletions:  0,
		},
		{
			name:               "Insertions only",
			statInput:          "5 insertions(+)",
			expectedInsertions: 5,
			expectedDeletions:  0,
		},
		{
			name:               "Deletions only",
			statInput:          "10 deletions(-)",
			expectedInsertions: 0,
			expectedDeletions:  10,
		},
		{
			name:               "Insertions and deletions combined",
			statInput:          "2 files changed, 10 insertions(+), 5 deletions(-)",
			expectedInsertions: 10,
			expectedDeletions:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ins, del := parseShortStat(tt.statInput)
			if ins != tt.expectedInsertions || del != tt.expectedDeletions {
				t.Errorf("For %q expected (%d, %d), got (%d, %d)", tt.statInput, tt.expectedInsertions, tt.expectedDeletions, ins, del)
			}
		})
	}
}

func TestGitAheadBehindWidget(t *testing.T) {
	RegisterAll()
	w := GetWidget("git-ahead-behind")
	if w == nil {
		t.Fatalf("Git ahead-behind widget not found")
	}

	if w.GetDefaultColor() != "brightCyan" {
		t.Errorf("Expected default color %q, got %q", "brightCyan", w.GetDefaultColor())
	}
	if w.GetDisplayName() != "Git Ahead/Behind" {
		t.Errorf("Expected display name %q, got %q", "Git Ahead/Behind", w.GetDisplayName())
	}

	settings := types.DefaultSettings()

	tests := []struct {
		name          string
		gitRunner     func(cmd string, ctx CwdResolver, ttl int) (string, error)
		item          types.WidgetItem
		ctx           types.RenderContext
		expectedTitle string
		expectedBody  string
		expectedColor string
	}{
		{
			name:          "Preview mode ahead-behind",
			item:          types.WidgetItem{Type: "git-ahead-behind"},
			ctx:           types.RenderContext{IsPreview: true},
			expectedTitle: "",
			expectedBody:  "↑2 ↓1",
			expectedColor: "brightCyan",
		},
		{
			name: "Both ahead and behind",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "rev-list --left-right --count HEAD...@{upstream}" {
					return "2\t1\n", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-ahead-behind"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/dummy/repo"}},
			expectedTitle: "",
			expectedBody:  "↑2 ↓1",
			expectedColor: "brightCyan",
		},
		{
			name: "Ahead only",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "rev-list --left-right --count HEAD...@{upstream}" {
					return "3\t0\n", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-ahead-behind"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/dummy/repo"}},
			expectedTitle: "",
			expectedBody:  "↑3",
			expectedColor: "brightCyan",
		},
		{
			name: "Behind only",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "rev-list --left-right --count HEAD...@{upstream}" {
					return "0\t5\n", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-ahead-behind"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/dummy/repo"}},
			expectedTitle: "",
			expectedBody:  "↓5",
			expectedColor: "brightCyan",
		},
		{
			name: "Up-to-date with upstream (identical)",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "rev-list --left-right --count HEAD...@{upstream}" {
					return "0\t0\n", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-ahead-behind"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/dummy/repo"}},
			expectedTitle: "",
			expectedBody:  "≡",
			expectedColor: "brightCyan",
		},
		{
			name: "Custom symbol",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "rev-list --left-right --count HEAD...@{upstream}" {
					return "1\t0\n", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-ahead-behind", Symbol: "⇅ "},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/dummy/repo"}},
			expectedTitle: "",
			expectedBody:  "⇅ ↑1",
			expectedColor: "brightCyan",
		},
		{
			name: "Custom symbol with Raw=true",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "rev-list --left-right --count HEAD...@{upstream}" {
					return "1\t0\n", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-ahead-behind", Symbol: "⇅ ", Raw: true},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/dummy/repo"}},
			expectedTitle: "",
			expectedBody:  "↑1",
			expectedColor: "brightCyan",
		},
		{
			name: "Non-git directory",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				return "false", fmt.Errorf("not a git repository")
			},
			item:          types.WidgetItem{Type: "git-ahead-behind"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/tmp/non-git-dir"}},
			expectedTitle: "",
			expectedBody:  "(no git)",
			expectedColor: "brightCyan",
		},
		{
			name: "No upstream configured",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "rev-list --left-right --count HEAD...@{upstream}" {
					return "", fmt.Errorf("fatal: no upstream configured for branch 'feat'")
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-ahead-behind"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/dummy/repo"}},
			expectedTitle: "",
			expectedBody:  "(no upstream)",
			expectedColor: "brightCyan",
		},
		{
			name: "Malformed rev-list output",
			gitRunner: func(cmd string, ctx CwdResolver, ttl int) (string, error) {
				if cmd == "rev-parse --is-inside-work-tree" {
					return "true", nil
				}
				if cmd == "rev-list --left-right --count HEAD...@{upstream}" {
					return "corrupt data", nil
				}
				return "", nil
			},
			item:          types.WidgetItem{Type: "git-ahead-behind"},
			ctx:           types.RenderContext{Data: types.StatusJSON{CWD: "/dummy/repo"}},
			expectedTitle: "",
			expectedBody:  "(no upstream)",
			expectedColor: "brightCyan",
		},
	}

	oldRunner := runGitCommand
	defer func() { runGitCommand = oldRunner }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.gitRunner != nil {
				runGitCommand = tt.gitRunner
			}
			title, body, err := w.Render(tt.item, tt.ctx, settings)
			if err != nil {
				t.Fatalf("Render error: %v", err)
			}
			if title != tt.expectedTitle || body != tt.expectedBody {
				t.Errorf("Expected title %q and body %q, got title %q and body %q", tt.expectedTitle, tt.expectedBody, title, body)
			}
			if color := w.GetBodyColor(tt.item, tt.ctx); color != tt.expectedColor {
				t.Errorf("Expected body color %q, got %q", tt.expectedColor, color)
			}
		})
	}
}

func TestParseAheadBehind(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedAhead  int
		expectedBehind int
		expectErr      bool
	}{
		{
			name:           "Valid tab-separated count",
			input:          "2\t1\n",
			expectedAhead:  2,
			expectedBehind: 1,
			expectErr:      false,
		},
		{
			name:           "Valid space-separated count",
			input:          "0 4",
			expectedAhead:  0,
			expectedBehind: 4,
			expectErr:      false,
		},
		{
			name:           "Both zero",
			input:          "0\t0",
			expectedAhead:  0,
			expectedBehind: 0,
			expectErr:      false,
		},
		{
			name:           "Empty string",
			input:          "",
			expectedAhead:  0,
			expectedBehind: 0,
			expectErr:      true,
		},
		{
			name:           "Single number only",
			input:          "2",
			expectedAhead:  0,
			expectedBehind: 0,
			expectErr:      true,
		},
		{
			name:           "Non-numeric tokens",
			input:          "foo bar",
			expectedAhead:  0,
			expectedBehind: 0,
			expectErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ahead, behind, err := parseAheadBehind(tt.input)
			if (err != nil) != tt.expectErr {
				t.Fatalf("For %q expected err=%v, got err=%v", tt.input, tt.expectErr, err)
			}
			if ahead != tt.expectedAhead || behind != tt.expectedBehind {
				t.Errorf("For %q expected (%d, %d), got (%d, %d)", tt.input, tt.expectedAhead, tt.expectedBehind, ahead, behind)
			}
		})
	}
}

func TestGitWidgets_CommandTimeoutAndFallback(t *testing.T) {
	RegisterAll()
	settings := types.DefaultSettings()

	oldRunner := runGitCommand
	defer func() { runGitCommand = oldRunner }()

	// Mock a timing out or slow runner
	runGitCommand = func(command string, ctx CwdResolver, ttlSeconds int) (string, error) {
		if command == "rev-parse --is-inside-work-tree" {
			return "", fmt.Errorf("git command timed out after 200ms")
		}
		return "", fmt.Errorf("git error")
	}

	gb := GetWidget("git-branch")
	ctx := types.RenderContext{
		Data: types.StatusJSON{CWD: "/some/path"},
	}

	title, body, err := gb.Render(types.WidgetItem{Type: "git-branch"}, ctx, settings)
	if err != nil {
		t.Fatalf("GitBranch render error: %v", err)
	}
	if title != "" || body != "⎇ no git" {
		t.Errorf("Expected fallback '⎇ no git', got title %q, body %q", title, body)
	}

	gc := GetWidget("git-changes")
	title, body, err = gc.Render(types.WidgetItem{Type: "git-changes"}, ctx, settings)
	if err != nil {
		t.Fatalf("GitChanges render error: %v", err)
	}
	if title != "" || body != "(no git)" {
		t.Errorf("Expected fallback '(no git)', got title %q, body %q", title, body)
	}

	gah := GetWidget("git-ahead-behind")
	title, body, err = gah.Render(types.WidgetItem{Type: "git-ahead-behind"}, ctx, settings)
	if err != nil {
		t.Fatalf("GitAheadBehind render error: %v", err)
	}
	if title != "" || body != "(no git)" {
		t.Errorf("Expected fallback '(no git)', got title %q, body %q", title, body)
	}
}

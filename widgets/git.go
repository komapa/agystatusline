package widgets

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/yuys13/agystatusline/types"
)

// GitBranchWidget displays the current Git branch name.
type GitBranchWidget struct{}

func (g *GitBranchWidget) GetDefaultColor() string { return "brightMagenta" }
func (g *GitBranchWidget) GetDisplayName() string  { return "Git Branch" }
func (g *GitBranchWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	if ctx.Data.VCS != nil && ctx.Data.VCS.Dirty != nil && *ctx.Data.VCS.Dirty {
		return "brightRed"
	}
	return "brightBlue"
}

func (g *GitBranchWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	symbol := "⎇ "
	if item.Symbol != "" {
		symbol = item.Symbol
	}

	// Try to get branch from VCS telemetry first
	var branch string
	var dirty bool
	if ctx.Data.VCS != nil {
		branch = ctx.Data.VCS.Branch
		if ctx.Data.VCS.Dirty != nil {
			dirty = *ctx.Data.VCS.Dirty
		}
	}

	if branch == "" && !ctx.IsPreview {
		// Fallback to git command
		isGit, err := runGitCommand("rev-parse --is-inside-work-tree", ctx, ctx.GitCacheTTLSeconds)
		if err == nil && strings.TrimSpace(isGit) == "true" {
			b, _ := runGitCommand("symbolic-ref --short HEAD", ctx, ctx.GitCacheTTLSeconds)
			branch = strings.TrimSpace(b)
			status, _ := runGitCommand("status --porcelain", ctx, ctx.GitCacheTTLSeconds)
			dirty = strings.TrimSpace(status) != ""
		}
	}

	if ctx.IsPreview && branch == "" {
		branch = "main"
	}

	if branch == "" {
		return "", symbol + "no git", nil
	}

	bodyStr := branch
	if !item.Raw {
		bodyStr = symbol + branch
	}
	if dirty {
		bodyStr += "*"
	}

	return "", bodyStr, nil
}

// GitChangesWidget displays the counts of Git insertions and deletions.
type GitChangesWidget struct{}

func (g *GitChangesWidget) GetDefaultColor() string { return "yellow" }
func (g *GitChangesWidget) GetDisplayName() string  { return "Git Changes" }
func (g *GitChangesWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	return "yellow"
}

func (g *GitChangesWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	if ctx.IsPreview {
		return "", "(+42,-10)", nil
	}

	// Check if inside git tree
	isGit, err := runGitCommand("rev-parse --is-inside-work-tree", ctx, ctx.GitCacheTTLSeconds)
	if err != nil || strings.TrimSpace(isGit) != "true" {
		return "", "(no git)", nil
	}

	unstagedStat, _ := runGitCommand("diff --shortstat", ctx, ctx.GitCacheTTLSeconds)
	stagedStat, _ := runGitCommand("diff --cached --shortstat", ctx, ctx.GitCacheTTLSeconds)

	uIns, uDel := parseShortStat(unstagedStat)
	sIns, sDel := parseShortStat(stagedStat)

	insertions := uIns + sIns
	deletions := uDel + sDel

	return "", fmt.Sprintf("(+%d,-%d)", insertions, deletions), nil
}

func parseShortStat(stat string) (int, int) {
	insertMatch := regexp.MustCompile(`(\d+)\s+insertions?`).FindStringSubmatch(stat)
	deleteMatch := regexp.MustCompile(`(\d+)\s+deletions?`).FindStringSubmatch(stat)

	ins := 0
	del := 0

	if len(insertMatch) > 1 {
		ins, _ = strconv.Atoi(insertMatch[1])
	}
	if len(deleteMatch) > 1 {
		del, _ = strconv.Atoi(deleteMatch[1])
	}

	return ins, del
}

// GitAheadBehindWidget displays Git ahead/behind commits count against remote upstream.
type GitAheadBehindWidget struct{}

func (g *GitAheadBehindWidget) GetDefaultColor() string { return "brightCyan" }
func (g *GitAheadBehindWidget) GetDisplayName() string  { return "Git Ahead/Behind" }
func (g *GitAheadBehindWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	return "brightCyan"
}

func (g *GitAheadBehindWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	if ctx.IsPreview {
		body := "↑2 ↓1"
		if item.Symbol != "" && !item.Raw {
			body = item.Symbol + body
		}
		return "", body, nil
	}

	// Check if inside git tree
	isGit, err := runGitCommand("rev-parse --is-inside-work-tree", ctx, ctx.GitCacheTTLSeconds)
	if err != nil || strings.TrimSpace(isGit) != "true" {
		body := "(no git)"
		if item.Symbol != "" && !item.Raw {
			body = item.Symbol + body
		}
		return "", body, nil
	}

	out, err := runGitCommand("rev-list --left-right --count HEAD...@{upstream}", ctx, ctx.GitCacheTTLSeconds)
	if err != nil {
		body := "(no upstream)"
		if item.Symbol != "" && !item.Raw {
			body = item.Symbol + body
		}
		return "", body, nil
	}

	ahead, behind, err := parseAheadBehind(out)
	if err != nil {
		body := "(no upstream)"
		if item.Symbol != "" && !item.Raw {
			body = item.Symbol + body
		}
		return "", body, nil
	}

	var body string
	if ahead > 0 && behind > 0 {
		body = fmt.Sprintf("↑%d ↓%d", ahead, behind)
	} else if ahead > 0 {
		body = fmt.Sprintf("↑%d", ahead)
	} else if behind > 0 {
		body = fmt.Sprintf("↓%d", behind)
	} else {
		body = "≡"
	}

	if item.Symbol != "" && !item.Raw {
		body = item.Symbol + body
	}

	return "", body, nil
}

func parseAheadBehind(out string) (int, int, error) {
	fields := strings.Fields(out)
	if len(fields) < 2 {
		return 0, 0, fmt.Errorf("unexpected rev-list count output: %q", out)
	}
	ahead, err1 := strconv.Atoi(fields[0])
	behind, err2 := strconv.Atoi(fields[1])
	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("invalid rev-list count: %q", out)
	}
	return ahead, behind, nil
}

// GitRepoWidget displays the current Git repository name.
type GitRepoWidget struct{}

func (g *GitRepoWidget) GetDefaultColor() string { return "white" }
func (g *GitRepoWidget) GetDisplayName() string  { return "Git Repository" }
func (g *GitRepoWidget) GetBodyColor(item types.WidgetItem, ctx types.RenderContext) string {
	return "white"
}

func (g *GitRepoWidget) Render(item types.WidgetItem, ctx types.RenderContext, settings types.Settings) (string, string, error) {
	if ctx.IsPreview {
		body := "agystatusline"
		if item.Symbol != "" && !item.Raw {
			body = item.Symbol + body
		}
		return "", body, nil
	}

	// Check if inside git tree
	isGit, err := runGitCommand("rev-parse --is-inside-work-tree", ctx, ctx.GitCacheTTLSeconds)
	if err != nil || strings.TrimSpace(isGit) != "true" {
		body := "(no git)"
		if item.Symbol != "" && !item.Raw {
			body = item.Symbol + body
		}
		return "", body, nil
	}

	topLevel, err := runGitCommand("rev-parse --show-toplevel", ctx, ctx.GitCacheTTLSeconds)
	if err != nil || strings.TrimSpace(topLevel) == "" {
		body := "(no git)"
		if item.Symbol != "" && !item.Raw {
			body = item.Symbol + body
		}
		return "", body, nil
	}

	repoName := filepath.Base(strings.TrimSpace(topLevel))
	body := repoName
	if item.Symbol != "" && !item.Raw {
		body = item.Symbol + body
	}

	return "", body, nil
}

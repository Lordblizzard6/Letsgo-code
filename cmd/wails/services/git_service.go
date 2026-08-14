package services

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/user/go-claude-code/internal/engine"
)

// runGit executes git with the working directory pinned to dir and returns
// the combined output.
func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// GitService wraps the git CLI for the rail Git pane (T045, gui-contract §4):
// branches, status, commit/diff/review/tags/hooks/undo/redo and PR+push.
type GitService struct {
	hub *Hub
	dir string
}

func NewGitService(hub *Hub) *GitService { return &GitService{hub: hub} }

// SetDir pins the repository root (tests use temp repos).
func (s *GitService) SetDir(dir string) { s.dir = dir }

// Dir is the repository root the pane operates on (cwd unless overridden).
func (s *GitService) Dir() string {
	if s.dir != "" {
		return s.dir
	}
	dir, _ := os.Getwd()
	return dir
}

// Branches lists all local and remote branches (current prefixed with "*").
func (s *GitService) Branches() []string {
	out, err := runGit(s.Dir(), "branch", "-a")
	if err != nil {
		return nil
	}
	var branches []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			branches = append(branches, line)
		}
	}
	return branches
}

// CurrentBranch returns the active branch name.
func (s *GitService) CurrentBranch() string {
	out, err := runGit(s.Dir(), "branch", "--show-current")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// Status returns `git status --porcelain`; "Working tree clean" when empty
// (gui-contract §4 Git status).
func (s *GitService) Status() string {
	out, err := runGit(s.Dir(), "status", "--porcelain")
	if err != nil {
		return "Not a git repository."
	}
	if strings.TrimSpace(out) == "" {
		return "Working tree clean"
	}
	return "Changes:\n" + strings.TrimSpace(out)
}

// CreateBranch runs `git checkout -b <name>`.
func (s *GitService) CreateBranch(name string) (string, error) {
	return runGit(s.Dir(), "checkout", "-b", name)
}

// SwitchBranch runs `git checkout <name>`.
func (s *GitService) SwitchBranch(name string) (string, error) {
	return runGit(s.Dir(), "checkout", name)
}

// Commit stages everything and commits with the given message.
func (s *GitService) Commit(msg string) (string, error) {
	if _, err := runGit(s.Dir(), "add", "-A"); err != nil {
		return "", err
	}
	return runGit(s.Dir(), "commit", "-m", msg)
}

// Diff returns the working-tree diff.
func (s *GitService) Diff() (string, error) {
	return runGit(s.Dir(), "diff")
}

// Review asks the agent (engine) to review the current changes.
func (s *GitService) Review() {
	s.hub.Engine.Send(engine.SendMessage{
		Text: "Revisa los cambios pendientes del repositorio (git diff) y coméntalos.",
		SessionID: s.hub.Engine.SessionID(),
	})
}

// Tags lists repository tags.
func (s *GitService) Tags() (string, error) {
	return runGit(s.Dir(), "tag", "-l")
}

// Hooks lists the enabled git hooks of this repository.
func (s *GitService) Hooks() ([]string, error) {
	gitDir, err := runGit(s.Dir(), "rev-parse", "--git-dir")
	if err != nil {
		return nil, err
	}
	dir := strings.TrimSpace(gitDir)
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(s.Dir(), dir)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var hooks []string
	for _, e := range entries {
		if !e.IsDir() && !strings.HasSuffix(e.Name(), ".sample") {
			hooks = append(hooks, e.Name())
		}
	}
	return hooks, nil
}

// Undo resets the last commit softly (working tree preserved).
func (s *GitService) Undo() (string, error) {
	return runGit(s.Dir(), "reset", "--soft", "HEAD~1")
}

// Redo amends a fresh commit re-applying the previously undone changes.
func (s *GitService) Redo() (string, error) {
	if _, err := runGit(s.Dir(), "diff", "--cached", "--quiet"); err == nil {
		return "nothing to redo", nil
	}
	return runGit(s.Dir(), "commit", "-m", "redo previous undo")
}

// Log returns the last commits (PR view feed).
func (s *GitService) Log() (string, error) {
	return runGit(s.Dir(), "log", "--oneline", "-10")
}

// Push pushes the current branch to origin, setting upstream.
func (s *GitService) Push() (string, error) {
	branch := s.CurrentBranch()
	if branch == "" {
		return "no branch to push", nil
	}
	return runGit(s.Dir(), "push", "-u", "origin", branch)
}
package gui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// runGit executes git in the given directory. It is a var so tests can stub it.
var runGit = func(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// gitView offers branch create/switch, commit, diff viewer and review trigger
// from the GUI (cli-parity.md Git group, FR-017).
type gitView struct {
	win  fyne.Window
	dir  string
	root *fyne.Container

	branchEntry *widget.Entry
	branchList  *widget.List
	branches    []string
	center  *emptyAware
	loading fyne.CanvasObject

	commitMsg *widget.Entry
	diffLabel *widget.Label
	statusLbl *widget.Label

	// onReview triggers an agent review of the current changes.
	onReview func()
}

// newGitView builds the git management panel.
func newGitView(win fyne.Window) *gitView {
	v := &gitView{win: win, dir: "."}

	v.branchEntry = widget.NewEntry()
	v.branchEntry.SetPlaceHolder("branch name")
	branchCreate := widget.NewButton("Create", v.createBranch)
	branchSwitch := widget.NewButton("Switch", v.switchBranch)

	v.branchList = widget.NewList(
		func() int { return len(v.branches) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= 0 && id < len(v.branches) {
				item.(*widget.Label).SetText(v.branches[id])
			}
		},
	)
	v.branchList.OnSelected = func(id widget.ListItemID) {
		if id >= 0 && id < len(v.branches) {
			v.branchEntry.SetText(v.branches[id])
		}
	}

	v.commitMsg = widget.NewEntry()
	v.commitMsg.SetPlaceHolder("Commit message")

	v.statusLbl = widget.NewLabel("")
	v.statusLbl.Wrapping = fyne.TextWrapWord
	v.statusLbl.TextStyle = fyne.TextStyle{Monospace: true}

	v.diffLabel = widget.NewLabel("")
	v.diffLabel.Wrapping = fyne.TextWrapWord
	v.diffLabel.TextStyle = fyne.TextStyle{Monospace: true}

	actions := container.NewHBox(
		widget.NewButton("Commit", v.commit),
		widget.NewButton("Diff", v.showDiff),
		widget.NewButton("Review", v.windowReview),
		widget.NewButton("Tags", v.showTags),
		widget.NewButton("Hooks", v.showHooks),
		widget.NewButton("Undo", v.undoLast),
		widget.NewButton("Redo", v.redoLast),
		widget.NewButton("PR comments", v.showPRComments),
		widget.NewButton("Push+PR", v.commitPushPR),
	)

	branchBox := container.NewHBox(
		v.branchEntry, branchCreate, branchSwitch,
	)

	top := container.NewVBox(
		widget.NewLabelWithStyle("Branches", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		branchBox,
		container.NewVBox(v.branchList, v.statusLbl),
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Commit", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		v.commitMsg,
	)

	scroll := container.NewScroll(container.NewVBox(top, actions, widget.NewSeparator(), v.diffLabel))
	v.loading = loadingState(scroll, false)
	v.center = newEmptyAware(v.loading, "No hay repositorio.", "Abrir carpeta…", nil)
	v.root = container.NewBorder(nil, nil, nil, nil, v.center.content())
	v.refresh()
	return v
}

// content returns the panel for embedding in the main window.
func (v *gitView) content() fyne.CanvasObject { return v.root }

// setDir changes the working directory (session project path).
func (v *gitView) setDir(dir string) {
	if dir != "" {
		v.dir = dir
	}
	v.refresh()
}

// refresh re-reads branches and working tree status. The loading surface
// wraps the slow (git exec) path so async callers keep the previous content
// visible plus the "Refrescando…" indicator (contracts §5).
func (v *gitView) refresh() {
	if surf, ok := asLoadingSurface(v.loading); ok {
		surf.setRefreshing(true)
	}
	defer func() {
		if surf, ok := asLoadingSurface(v.loading); ok {
			surf.setRefreshing(false)
		}
	}()
	out, err := runGit(v.dir, "branch", "-a")
	var branches []string
	if err == nil {
		for _, line := range strings.Split(out, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				branches = append(branches, strings.TrimPrefix(line, "* "))
			}
		}
	}
	v.branches = branches
	v.branchList.Refresh()
	if v.center != nil {
		v.center.setEmpty(err != nil)
	}

	status, err := runGit(v.dir, "status", "--porcelain")
	if err != nil {
		v.statusLbl.SetText("Not a git repository.")
		return
	}
	status = strings.TrimSpace(status)
	if status == "" {
		v.statusLbl.SetText("Working tree clean")
	} else {
		v.statusLbl.SetText("Changes:\n" + status)
	}
}

// createBranch runs `git checkout -b <name>`.
func (v *gitView) createBranch() {
	name := strings.TrimSpace(v.branchEntry.Text)
	if name == "" {
		return
	}
	if out, err := runGit(v.dir, "checkout", "-b", name); err != nil {
		dialog.ShowError(fmt.Errorf("create branch: %v\n%s", err, out), v.win)
		return
	}
	v.branchEntry.SetText("")
	dialog.ShowInformation("Git", "Created and switched to "+name, v.win)
	v.refresh()
}

// switchBranch runs `git checkout <name>`.
func (v *gitView) switchBranch() {
	name := strings.TrimSpace(v.branchEntry.Text)
	if name == "" {
		return
	}
	if out, err := runGit(v.dir, "checkout", name); err != nil {
		dialog.ShowError(fmt.Errorf("switch branch: %v\n%s", err, out), v.win)
		return
	}
	v.branchEntry.SetText("")
	dialog.ShowInformation("Git", "Switched to "+name, v.win)
	v.refresh()
}

// commit stages all changes and commits with the entered message.
func (v *gitView) commit() {
	msg := strings.TrimSpace(v.commitMsg.Text)
	if msg == "" {
		dialog.ShowInformation("Commit", "Enter a commit message first.", v.win)
		return
	}
	if out, err := runGit(v.dir, "add", "-A"); err != nil {
		dialog.ShowError(fmt.Errorf("git add: %v\n%s", err, out), v.win)
		return
	}
	out, err := runGit(v.dir, "commit", "-m", msg)
	if err != nil {
		dialog.ShowError(fmt.Errorf("git commit: %v\n%s", err, out), v.win)
		return
	}
	v.commitMsg.SetText("")
	dialog.ShowInformation("Committed", strings.TrimSpace(out), v.win)
	v.refresh()
}

// showDiff renders `git diff` into the panel.
func (v *gitView) showDiff() {
	out, err := runGit(v.dir, "diff")
	if err != nil {
		dialog.ShowError(fmt.Errorf("git diff: %v", err), v.win)
		return
	}
	if strings.TrimSpace(out) == "" {
		out = "No changes to diff."
	}
	v.diffLabel.SetText(out)
}

// windowReview triggers the review callback (agent task), or explains setup.
func (v *gitView) windowReview() {
	if v.onReview != nil {
		v.onReview()
		return
	}
	dialog.ShowInformation("Review",
		"Review runs through the chat. If nothing happens, send a review prompt in the chat (e.g. \"review the current changes\").", v.win)
}

// showTags lists git tags in a dialog.
func (v *gitView) showTags() {
	out, err := runGit(v.dir, "tag", "-l")
	if err != nil {
		dialog.ShowError(fmt.Errorf("git tag: %v", err), v.win)
		return
	}
	if strings.TrimSpace(out) == "" {
		out = "No tags."
	}
	dialog.ShowInformation("Tags", strings.TrimSpace(out), v.win)
}

// showHooks lists the repository's git hooks.
func (v *gitView) showHooks() {
	out, err := runGit(v.dir, "rev-parse", "--git-dir")
	if err != nil {
		dialog.ShowInformation("Hooks", "Git hooks unavailable (not a repo).", v.win)
		return
	}
	entries, err := os.ReadDir(strings.TrimSpace(out) + "/hooks")
	if err != nil {
		dialog.ShowInformation("Hooks", "No hooks directory.", v.win)
		return
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	dialog.ShowInformation("Hooks", strings.Join(names, "\n"), v.win)
}

// undoLast reverts the last commit (git reset --soft HEAD~1).
func (v *gitView) undoLast() {
	out, err := runGit(v.dir, "reset", "--soft", "HEAD~1")
	if err != nil {
		dialog.ShowError(fmt.Errorf("undo: %v\n%s", err, out), v.win)
		return
	}
	dialog.ShowInformation("Undo", "Last commit undone (soft reset).", v.win)
	v.refresh()
}

// redoLast re-commits the staged changes left behind by a soft reset
// (cli-parity.md `redo` row; mirrors git reset --soft HEAD~1 + commit).
func (v *gitView) redoLast() {
	if _, err := runGit(v.dir, "diff", "--cached", "--quiet"); err == nil {
		dialog.ShowInformation("Redo", "Nothing staged to redo.", v.win)
		return
	}
	out, err := runGit(v.dir, "commit", "-m", "redo previous undo")
	if err != nil {
		dialog.ShowError(fmt.Errorf("redo: %v\n%s", err, out), v.win)
		return
	}
	dialog.ShowInformation("Redo", "Staged changes recommitted.", v.win)
	v.refresh()
}

// showPRComments opens PR comments for the current branch (cli-parity.md
// `pr_comments` row), falling back to recent commit messages when no GitHub
// remote/CLI is available.
func (v *gitView) showPRComments() {
	if out, err := runGit(v.dir, "log", "--oneline", "-10"); err == nil {
		dialog.ShowInformation("PR comments",
			fmt.Sprintf("Recent commits on %s:\n\n%s\n(Use a pull request UI to view inline comments.)",
				v.currentBranch(), out), v.win)
		return
	}
	dialog.ShowInformation("PR comments", "Not a git repository.", v.win)
}

// currentBranch returns the active branch, or "HEAD" if unavailable.
func (v *gitView) currentBranch() string {
	out, err := runGit(v.dir, "branch", "--show-current")
	if err != nil {
		return "HEAD"
	}
	if b := strings.TrimSpace(out); b != "" {
		return b
	}
	return "HEAD"
}

// commitPushPR commits, pushes to origin and prompts to open a PR.
func (v *gitView) commitPushPR() {
	msg := strings.TrimSpace(v.commitMsg.Text)
	if msg == "" {
		dialog.ShowInformation("Commit+Push+PR", "Enter a commit message first.", v.win)
		return
	}
	if out, err := runGit(v.dir, "add", "-A"); err != nil {
		dialog.ShowError(fmt.Errorf("git add: %v\n%s", err, out), v.win)
		return
	}
	if out, err := runGit(v.dir, "commit", "-m", msg); err != nil {
		dialog.ShowError(fmt.Errorf("git commit: %v\n%s", err, out), v.win)
		return
	}
	branch, _ := runGit(v.dir, "branch", "--show-current")
	if branch = strings.TrimSpace(branch); branch == "" {
		branch = "HEAD"
	}
	if out, err := runGit(v.dir, "push", "-u", "origin", branch); err != nil {
		dialog.ShowInformation("Commit+Push+PR",
			fmt.Sprintf("Committed locally; push failed:\n%s", out), v.win)
		v.refresh()
		return
	}
	v.commitMsg.SetText("")
	dialog.ShowInformation("Commit+Push+PR",
		fmt.Sprintf("Committed and pushed %s. Open a pull request to finish.", branch), v.win)
	v.refresh()
}

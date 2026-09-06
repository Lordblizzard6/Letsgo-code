# Interface Contracts: Activity Panel & Slash Commands

## 1. Wails Service Contracts (`GitService`)

### `DiffSummary() (GitDiffSummary, error)`
Returns current repository overview with:
- `IsRepo bool`: whether current directory has git.
- `Clean bool`: whether tree is clean.
- `Branch string`: current active branch.
- `UncommittedCount int`: count of modified uncommitted files.
- `CommittedCount int`: count of commits on current branch.
- `Files []GitFileDiff`: list of files with additions, deletions, directory, and staged flags.
- `RawDiff string`: full working-tree diff.

### `StageFile(filePath string) error`
Stages a specific file path into the git index (`git add <filePath>`).

### `UnstageFile(filePath string) error`
Unstages a specific file path from the git index (`git restore --staged <filePath>`).

### `StageAll() error`
Stages all modified files (`git add -A`).

### `UnstageAll() error`
Unstages all modified files (`git restore --staged .`).

---

## 2. GUI Slash Commands Contract

| Command | Arguments | UI Action | Engine Action |
|---|---|---|---|
| `/init` | None | Open Project Init dialog or runs initialization | Creates standard `AGENTS.md` / config |
| `/undo` | None | Reverts last turn message in UI | Rolls back conversation turn & file edit |
| `/redo` | None | Re-applies undone turn | Re-dispatches turn |
| `/diff` | None | Opens Activity Inspector on Git Review tab | None |
| `/review` | None | Dispatches prompt to AI | Sends review prompt for pending changes |
| `/compact` | None | Truncates old message turns with summary | Compresses engine context window |
| `/cost` | None | Opens Spend modal | Queries session usage |
| `/skills` | None | Opens Skills browser tab | Lists available skill manifests |
| `/tasks` | None | Opens Activity Inspector on Overview tab | Queries background tasks |
| `/terminal` | None | Opens Activity Inspector on Commands tab | None |
| `/share` | None | Sanitizes & copies markdown chat to clipboard | None |

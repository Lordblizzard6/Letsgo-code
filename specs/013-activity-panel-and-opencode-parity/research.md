# Phase 0 Research: Opencode Parity & Antigravity Activity Panel

## Research Topics & Findings

### 1. Antigravity CLI Activity Inspector Architecture
- **Structure**:
  - Docked on the right side of the main conversation workspace.
  - Can be collapsed/toggled via header action or keyboard shortcut (`Ctrl+B` / `Ctrl+D`).
  - Maximize button (`🗖` / `🗗`) that expands the panel to occupy the main canvas, giving developers a dedicated, unobstructed environment to inspect code diffs and outputs.
  - Close button (`✕`) to dismiss the panel.
  - Top tab navigation:
    - **Tab 1: Overview**: Top cards with Uncommitted & Committed badges; Changed files list with `+Lines` (green badge), `-Lines` (red badge), filename, and directory path in grey (`cmd.go  cmd/`); Background Tasks with execution status; Skills used; Active/recent terminals.
    - **Tab 2: Git Review**: Staged vs Unstaged groups, Stage/Unstage buttons, Commit message composer, Center diff viewer with syntax colors and hunk badges.
    - **Tab 3: Executed Commands & Tools**: Execution cards showing command/tool name, status, arguments, duration, and embedded inline git diff of changes produced.
- **Decision**: Replace both the isolated `#git-diff-panel` and `#right-pane` with a single unified component `<ActivityInspector />` that implements these three tabs, center diff viewer, and maximize toggle.

### 2. Git Stats & Staging in `GitService`
- **Current implementation**: `GitService.DiffSummary()` returns file paths and statuses from `git status --porcelain`.
- **Finding**: Running `git diff --numstat` provides exact additions and deletions per unstaged file. Running `git diff --cached --numstat` provides additions and deletions for staged files.
- **Decision**: Extend `GitFileDiff` with `Additions int` and `Deletions int`, and `IsStaged bool`. Add `StageFile(path string)`, `UnstageFile(path string)`, `StageAll()`, and `UnstageAll()` to `GitService`.

### 3. OpenCode Parity Slash Commands
- **Parity Commands required**:
  - `/init`: Initializes configuration, `.specify`, or `AGENTS.md`.
  - `/undo`: Reverts the last assistant message and its modified files.
  - `/redo`: Redoes the last undone message.
  - `/diff`: Opens the activity inspector directly on the Git Review tab.
  - `/review`: Sends a structured request to review uncommitted repository changes.
  - `/compact`: Condenses conversation history to reduce context window usage.
  - `/cost`: Opens the Spend modal with token usage and cost metrics.
  - `/skills`: Opens the Skills browser or displays active skills list.
  - `/tasks`: Opens the activity inspector on the Overview / Tasks section.
  - `/terminal`: Opens the activity inspector on the Commands / Terminals section.
  - `/share`: Copies a clean markdown export of the chat to the system clipboard.
- **Decision**: Implement all 11 commands in `App.tsx` command handler and autocomplete popup, as well as in `internal/tui/slash_commands.go` for terminal parity.

### 4. Center File / Diff Viewer Ergonomics
- **Requirements**: Line numbers, color-coded rows (`+` addition in green, `-` deletion in red, `@@` hunks in indigo/blue), clear copy action, and file header with maximize toggle.
- **Decision**: Build a dedicated `<DiffViewer />` sub-component that accepts a diff string, parses hunks and lines, and renders line numbers with full syntax color themes conforming to LetsGo's dark/light/letsgo themes.

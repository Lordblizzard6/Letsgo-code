# Implementation Plan: Opencode Command Parity & Antigravity CLI Activity Panel

**Branch**: `013-activity-panel-and-opencode-parity` | **Date**: 2026-09-04 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/013-activity-panel-and-opencode-parity/spec.md`

## Summary

This feature delivers complete parity with OpenCode slash commands and introduces a unified, 1:1 Antigravity CLI-style collapsible Activity Inspector Panel. The panel consolidates system overview, git review with a center file/diff viewer, staged/unstaged file management, background tasks, skills, and execution history with embedded command diffs.

## Technical Context

**Language/Version**: Go 1.26, TypeScript 5.4, React 18
**Primary Dependencies**: Wails v3 beta, Lucide-style SVG icons, Vite, Tailwind/CSS variables
**Storage**: SQLite (`internal/db`), Git CLI (`git` execution via `GitService`)
**Testing**: `vitest` (frontend), `go test` (backend services)
**Target Platform**: Windows 10/11 Desktop (`letsgo.exe gui`)
**Project Type**: Desktop GUI + CLI hybrid
**Performance Goals**: Activity panel toggle < 16ms, Maximize transition smooth 60fps, diff parsing < 50ms
**Constraints**: No GPL3 libraries, responsive UI, no regressions in existing chat and composer flows

## Constitution Check

*GATE: Passed*
- **Test-Last Verification (Principle I)**: Automated unit tests will run at the end of each implementation milestone.
- **Contract-Only Core Access (Principle II)**: Frontend interacts with `GitService`, `ChatService`, and `SettingsService` exclusively via bound RPC methods.
- **One Source of Truth (Principle III)**: Git state and command histories originate from Go engine and Git CLI.
- **Secure Rendering (Principle IV)**: Diff content sanitized and escaped, no raw `innerHTML`.
- **Simplicity & YAGNI (Principle V)**: Minimal, clean component architecture replacing two disjoint panels with one unified inspector.

## Project Structure

### Documentation (this feature)

```text
specs/013-activity-panel-and-opencode-parity/
├── spec.md                                       # Feature specification
├── plan.md                                       # This implementation plan
├── research.md                                   # Phase 0 research findings
├── data-model.md                                 # Phase 1 data entities and types
├── quickstart.md                                 # Phase 1 validation walkthrough
├── contracts/
│   └── activity-panel-and-commands.md            # Interface and RPC contracts
└── checklists/
    └── requirements.md                           # Specification quality checklist
```

### Source Code

```text
cmd/wails/
├── services/
│   ├── git_service.go                            # Stage/Unstage, diff numstat additions/deletions
│   └── services_test.go                          # Backend tests for git staging and stats
└── frontend/src/
    ├── App.tsx                                   # Parity slash commands, inspector panel state & shortcuts
    ├── ActivityInspector.tsx                     # Unified Antigravity CLI activity panel component
    ├── DiffViewer.tsx                            # Center file diff viewer with additions, deletions, hunks
    ├── app.css                                   # Modern Antigravity styles for inspector, diffs, badges
    └── i18n.ts                                   # Localization strings for overview, tabs, and commands
```

## Implementation Phases

### Phase 1: Backend `GitService` Extensions
1. Update `GitDiffSummary` and `GitFileDiff` structs in `cmd/wails/services/git_service.go` to parse `git diff --numstat` and `git diff --cached --numstat`.
2. Add `StageFile`, `UnstageFile`, `StageAll`, `UnstageAll` methods to `GitService`.
3. Add unit test coverage in `cmd/wails/services/services_test.go`.

### Phase 2: OpenCode Parity Slash Commands
1. Extend autocomplete command list in `App.tsx` with `/init`, `/undo`, `/redo`, `/diff`, `/review`, `/compact`, `/cost`, `/skills`, `/tasks`, `/terminal`, `/share`.
2. Implement command handlers in `App.tsx`:
   - `/diff`: Opens activity inspector on Git Review tab.
   - `/cost`: Opens Spend modal.
   - `/share`: Copies markdown chat transcript to clipboard with confirmation toast.
   - `/review`: Dispatches code review prompt.
   - `/tasks`: Opens activity inspector on Overview tab.
   - `/terminal`: Opens activity inspector on Commands tab.

### Phase 3: Antigravity CLI Activity Inspector Component
1. Create `DiffViewer.tsx`:
   - Line numbers, additions (`+`), deletions (`-`), hunk headers (`@@`).
2. Create `ActivityInspector.tsx`:
   - Header with `Overview`, `Git Review`, `Commands & Tools` tabs, center active file title, Maximize toggle (`🗖` / `🗗`), and Close button (`✕`).
   - Tab 1 Overview: Uncommitted/Committed counters, file list with `+Lines` / `-Lines` badges and grey directory path, background tasks, skills used, terminals.
   - Tab 2 Git Review: Staged & Unstaged lists with `+`/`-` stage buttons, commit box, center `DiffViewer`.
   - Tab 3 Commands & Tools: Execution history with duration, status, arguments, and embedded `DiffViewer` for file-modifying tools.
3. Integrate `ActivityInspector` into `App.tsx`, replacing the legacy `#git-diff-panel` and `#right-pane`.

### Phase 4: Styling & Keyboard Shortcuts
1. Add CSS rules in `app.css` for maximized layout, diff lines, badges, and smooth drawer transitions.
2. Bind `Ctrl+B` and `Ctrl+D` shortcuts to toggle the activity inspector.
3. Add i18n translation keys in `i18n.ts` for English and Spanish.

### Phase 5: Verification & Quality Gates
1. Run `npm test` and `npm run build` in `cmd/wails/frontend`.
2. Run `go test ./cmd/...`.
3. Run `go build -o letsgo.exe .`.
4. Validate all scenarios described in `quickstart.md`.

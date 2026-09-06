# Tasks: Opencode Command Parity & Antigravity CLI Activity Panel

**Input**: Design documents from `specs/013-activity-panel-and-opencode-parity/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

## Phase 1: Setup & Foundational Tasks

- [X] T001 Extend `GitFileDiff` and `GitDiffSummary` structs with additions, deletions, directory path, and staged status in `cmd/wails/services/git_service.go`
- [X] T002 Implement `git diff --numstat` and `git diff --cached --numstat` parsers in `cmd/wails/services/git_service.go`
- [X] T003 Implement `StageFile`, `UnstageFile`, `StageAll`, and `UnstageAll` methods in `cmd/wails/services/git_service.go`
- [X] T004 Add TypeScript interfaces `GitFileDiff`, `GitDiffSummary`, `ActivityTab`, and `CommandLogItem` in `cmd/wails/frontend/src/types.ts`

---

## Phase 2: User Story 1 - Antigravity Activity Inspector Panel & Overview (Priority: P1) 🎯 MVP

**Goal**: Deliver a collapsible, maximizable Antigravity CLI Activity Inspector on the right side with an Overview tab showing files changed (`+Lines` / `-Lines` & grey directories), background tasks, skills, and terminals.

- [X] T005 [US1] Create `DiffViewer.tsx` component with syntax colors, additions, deletions, hunk headers, and line numbers in `cmd/wails/frontend/src/DiffViewer.tsx`
- [X] T006 [US1] Create base `ActivityInspector.tsx` component with header tabs (`Overview`, `Git Review`, `Commands & Tools`), center file title, Maximize toggle (`🗖`/`🗗`), and Close button (`✕`) in `cmd/wails/frontend/src/ActivityInspector.tsx`
- [X] T007 [US1] Implement the Overview tab view in `ActivityInspector.tsx` rendering uncommitted/committed counters, file list with `+Lines` (green) / `-Lines` (red) badges and grey directory path, background tasks, skills, and terminals
- [X] T008 [US1] Integrate `ActivityInspector` into `App.tsx`, replacing the legacy `#git-diff-panel` and `#right-pane`
- [X] T009 [US1] Add CSS styles for `ActivityInspector`, maximized overlay, file list badges, and responsive drawer in `cmd/wails/frontend/src/app.css`
- [X] T010 [US1] Bind `Ctrl+B` and `Ctrl+D` keyboard shortcuts to toggle the activity inspector in `cmd/wails/frontend/src/App.tsx`

---

## Phase 3: User Story 2 - Git Review Tab & Staging (Priority: P1)

**Goal**: Deliver the Git Review tab in `ActivityInspector` with staged/unstaged file lists, staging/unstaging controls, commit composer, and center diff viewer.

- [X] T011 [US2] Implement Staged and Unstaged file sections in `ActivityInspector.tsx` with quick stage (`+`) and unstage (`-`) buttons
- [X] T012 [US2] Connect file click to load and display unified diff in center `DiffViewer` within `ActivityInspector.tsx`
- [X] T013 [US2] Add commit message input and commit execution calling `GitService.Commit` in `ActivityInspector.tsx`
- [X] T014 [US2] Add staging styles and git review layouts in `cmd/wails/frontend/src/app.css`

---

## Phase 4: User Story 3 - Executed Commands & Tool Calls Tab with Embedded Git Diff (Priority: P2)

**Goal**: Deliver the Commands & Tools tab logging executed commands, tool calls, and MCP calls with timings, status, arguments, and embedded diffs for modifying actions.

- [X] T015 [US3] Capture command, tool call, and MCP execution history in `App.tsx` and pass to `ActivityInspector.tsx`
- [X] T016 [US3] Implement the Commands & Tools tab in `ActivityInspector.tsx` rendering execution cards with status indicators, timings, arguments, and output accordions
- [X] T017 [US3] Render embedded `DiffViewer` within command cards for file-modifying executions in `ActivityInspector.tsx`

---

## Phase 5: User Story 4 - OpenCode Slash Command Parity (Priority: P2)

**Goal**: Implement essential OpenCode slash commands in autocomplete and command execution.

- [X] T018 [US4] Register `/init`, `/undo`, `/redo`, `/diff`, `/review`, `/compact`, `/cost`, `/skills`, `/tasks`, `/terminal`, `/share` in composer autocomplete list in `cmd/wails/frontend/src/App.tsx`
- [X] T019 [US4] Implement command execution handlers for `/diff` (opens Git Review tab), `/cost` (opens Spend modal), `/tasks` (opens Overview tab), `/terminal` (opens Commands tab), and `/share` (copies markdown chat to clipboard) in `cmd/wails/frontend/src/App.tsx`
- [X] T020 [US4] Add i18n translation strings for all new commands and inspector tabs in `cmd/wails/frontend/src/i18n.ts`

---

## Phase 6: Verification & Quality Gates

- [X] T021 Add backend test cases for `GitService` staging and numstat parsing in `cmd/wails/services/services_test.go`
- [X] T022 Run `npm test` in `cmd/wails/frontend` and verify all tests pass
- [X] T023 Run `npm run build` in `cmd/wails/frontend` and verify zero build errors
- [X] T024 Run `go test ./cmd/...` and verify all Go tests pass
- [X] T025 Build final executable `go build -o letsgo.exe .` and verify clean compilation

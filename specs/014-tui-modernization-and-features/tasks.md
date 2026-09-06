# Tasks: TUI Modernization & Advanced Features

**Input**: Design documents from `specs/014-tui-modernization-and-features/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

## Phase 1: Setup & Foundational Tasks

- [x] T001 Define `tuiOverlay` state machine enum (`overlayNone`, `overlayModelPicker`, `overlayGitReview`, `overlayActivityInspector`, `overlaySessionPicker`) and data structs (`AutocompleteState`, `TUIGitSummary`, `TUIActivityItem`, `TUIModelItem`) in `internal/tui/ui.go`
- [x] T002 Implement engine event listeners in `internal/tui/ui.go` to capture tool execution and command history into `activityHistory`
- [x] T003 [P] Add Git status and numstat parsing helpers in `internal/tui/git_view.go`

---

## Phase 2: User Story 1 - Modern Interactive Layout & Real-time Status Bar (Priority: P1) 🎯 MVP

**Goal**: Deliver a polished, modern terminal layout with styled header (repo, branch, status badge, session ID), Glamour markdown styling, and 2-tier real-time status bar with mode toggle (`BUILD`/`PLAN`/`RESEARCH`).

- [x] T004 [US1] Implement modern top header with project path, git branch, clean/dirty badge, session short ID, and theme styling in `internal/tui/ui.go`
- [x] T005 [US1] Implement 2-tier bottom status bar rendering mode badge, active model, token usage, cost, and keyboard hints in `internal/tui/ui.go`
- [x] T006 [US1] Support `Tab` key cycling between modes (`BUILD` → `PLAN` → `RESEARCH` → `BUILD`) when input is empty in `internal/tui/ui.go`
- [x] T007 [US1] Configure Glamour renderer with auto-styling and adaptive word wrapping on terminal resize in `internal/tui/ui.go`

---

## Phase 3: User Story 2 - Dynamic Provider & Model Switcher with Discovery (Priority: P1)

**Goal**: Replace static model lists with dynamic discovery across all configured providers (Gemini, Anthropic, OpenAI, Groq, OpenRouter, DeepSeek, Ollama) with interactive search and instant switching.

- [x] T008 [US2] Integrate `internal/api/discovery.go` and `config.AppConfig` to dynamically discover models across providers in `internal/tui/ui.go`
- [x] T009 [US2] Build interactive model picker overlay with provider tabs, search/filter query, and model descriptions in `internal/tui/ui.go`
- [x] T010 [US2] Bind `Ctrl+S`, `/model`, and `/models` to open/navigate the dynamic model picker and persist selection to `config.AppConfig.Model` in `internal/tui/ui.go`

---

## Phase 4: User Story 3 - Interactive Slash Command & Mention Autocomplete Overlay (Priority: P2)

**Goal**: Deliver an interactive popup suggestion box when typing `/` or `@` with arrow navigation, fuzzy filtering, and Tab auto-completion.

- [x] T011 [US3] Implement real-time trigger detection for `/` (slash commands) and `@` (context mentions) in `internal/tui/ui.go`
- [x] T012 [US3] Build floating suggestion box above textarea displaying matching items with icons, titles, and descriptions in `internal/tui/ui.go`
- [x] T013 [US3] Implement `Up`, `Down`, `Tab`, `Enter`, and `Esc` keyboard navigation for autocomplete popup in `internal/tui/ui.go`

---

## Phase 5: User Story 4 - TUI Git Review, Staging & Diff Overlay (Priority: P2)

**Goal**: Provide a full-featured Git Review view in the terminal with staged/unstaged file lists, numstat changes, colored diff viewer, and commit composer.

- [x] T014 [US4] Create `internal/tui/git_view.go` rendering Staged and Unstaged file sections with additions/deletions badges
- [x] T015 [US4] Implement stage (`+` / `s`) and unstage (`-` / `u`) file actions in `internal/tui/git_view.go`
- [x] T016 [US4] Implement unified syntax-colored diff preview pane with line numbers and hunk headers in `internal/tui/git_view.go`
- [x] T017 [US4] Implement commit composer prompt (`c`) inside Git Review in `internal/tui/git_view.go`
- [x] T018 [US4] Bind `Ctrl+D`, `/diff`, and `/review` to toggle the Git Review overlay in `internal/tui/ui.go`

---

## Phase 6: User Story 5 - Activity & Tool Execution Inspector (Priority: P2)

**Goal**: Deliver an Activity Inspector overlay tracking tool executions, commands, timings, outputs, and embedded diff previews.

- [x] T019 [US5] Create `internal/tui/activity_view.go` rendering timeline of executed tools, commands, and background tasks
- [x] T020 [US5] Implement expandable tool parameters and output inspection with embedded diff previews for file-modifying tools in `internal/tui/activity_view.go`
- [x] T021 [US5] Bind `Ctrl+B`, `/tasks`, and `/terminal` to toggle Activity Inspector overlay in `internal/tui/ui.go`

---

## Phase 7: User Story 6 - Session & Project Switcher Overlay (Priority: P3)

**Goal**: Enable switching between recent chat sessions and workspace directories directly from the TUI.

- [x] T022 [US6] Implement interactive session and project picker overlay querying `db.ListSessions()` and `db.ListProjects()` in `internal/tui/ui.go`
- [x] T023 [US6] Connect `/sessions`, `/open`, and `/new` to interactive session switching in `internal/tui/slash_commands.go` and `internal/tui/ui.go`

---

## Phase 8: Polish, Command Parity & Verification

- [x] T024 [P] Ensure full slash command parity across `AvailableSlashCommands` and handlers in `internal/tui/slash_commands.go`
- [x] T025 Add automated tests for overlay switching, autocomplete popup filtering, mode cycling, and git review in `internal/tui/interaction_test.go` and `internal/tui/slash_commands_test.go`
- [x] T026 Run full verification suite (`go test ./internal/tui/...`, `go test ./cmd/...`, and `go build -o letsgo.exe .`)

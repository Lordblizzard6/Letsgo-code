# Implementation Plan: TUI Modernization & Advanced Features

## 1. Overview
Modernize the terminal user interface (TUI) of LetsGO Code built on Bubble Tea, Lipgloss, and Glamour to achieve 1:1 functional parity with modern AI coding assistants (OpenCode, Claude Code, Antigravity).

## 2. Technical Approach
1. **Model & State Machine Refactor**:
   - Refactor `model` in `internal/tui/ui.go` to introduce clean overlay management (`overlayNone`, `overlayModelPicker`, `overlayGitReview`, `overlayActivityInspector`, `overlaySessionPicker`).
   - Add `TUIActivityItem`, `TUIGitSummary`, `AutocompleteState`.
2. **Modern Layout & Rendering**:
   - Redesign header with git branch, repo name, session ID, and status badge.
   - Redesign footer with 2-tier status bar: Mode pills, model name, tokens, cost, and keyboard hints.
   - Improve markdown rendering and syntax highlighting line wrapping.
3. **Dynamic Model Discovery Integration**:
   - Replace static arrays in `ui.go` with calls to `internal/api/discovery.go` and `internal/config/config.go`.
   - Render searchable model menu categorized by provider.
4. **Interactive Autocomplete Popup**:
   - Monitor textarea input for `/` and `@` triggers.
   - Render styled suggestions box above composer with keyboard navigation.
5. **Git Review & Diff View**:
   - Parse `git diff --numstat` and `git status --porcelain`.
   - Render Staged and Unstaged sections with stage/unstage hotkeys.
   - Colored diff viewer with line numbers and hunk headers.
   - In-terminal commit composer.
6. **Activity Inspector**:
   - Capture tool start/completion events from engine.
   - Render timeline with duration, status, arguments, and embedded diff previews.
7. **Session & Project Switcher**:
   - Interactive table of sessions from `db.ListSessions()` and projects from `db.ListProjects()`.

## 3. Implementation Phases
- **Phase 1**: Foundational state refactor, overlay enum, activity tracking, and git status helpers in `internal/tui/`.
- **Phase 2 (US1)**: Header, 2-tier status bar, Tab mode cycling (`BUILD`/`PLAN`/`RESEARCH`), and Glamour renderer tuning.
- **Phase 3 (US2)**: Dynamic provider & model picker overlay (`Ctrl+S`, `/model`, `/models`).
- **Phase 4 (US3)**: Interactive slash command and mention autocomplete popup.
- **Phase 5 (US4)**: Git Review, staging, unified diff viewer, and commit composer (`Ctrl+D`, `/diff`, `/review`).
- **Phase 6 (US5)**: Activity & Tool Inspector overlay (`Ctrl+B`, `/tasks`, `/terminal`).
- **Phase 7 (US6)**: Session & project switcher (`/sessions`, `/open`, `/new`).
- **Phase 8**: Verification & test coverage with `go test ./internal/tui/...` and compilation of `letsgo.exe`.

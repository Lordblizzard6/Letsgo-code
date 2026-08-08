# Implementation Plan: Fyne Desktop GUI for LetsGO Code

**Branch**: `Lets-go-Fyne` | **Date**: 2026-08-02 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/001-fyne-gui/spec.md` (clarified: full 1:1 parity with Codex, dark terminal aesthetic + Codex interaction patterns, shared modular core consumed by GUI and CLI)

## Summary

Build a desktop GUI for the LetsGO Code AI coding assistant using Fyne, reaching 1:1 capability parity with the reference assistant (Codex): chat, agent mode with tool approvals, sessions, git operations, parallel agents, MCP integration, and plugin management — without being tied to a single AI provider.

The existing codebase already separates a UI-agnostic core (`internal/api` multi-provider client, `internal/tools` with permission manager and cost tracker, `internal/db` SQLite storage, `internal/config`, `internal/mcp`, `internal/plugins`) from the presentation layer (`internal/tui`, Bubble Tea). The plan therefore:

1. Extracts the conversation/agent loop currently embedded in the TUI into a new presentation-agnostic `internal/engine` package (event-driven: typed events out, commands in, permission prompts behind a driver interface). Both the TUI (refactored) and the GUI consume it — this is what makes FR-021/FR-026 (shared core, no duplication) true.
2. Adds a new Fyne presentation layer `internal/gui` with a dark terminal-inspired theme (monospace typography), chat view with inline tool panels, approval dialogs, sessions sidebar, settings, usage view, git view, and keyboard-first navigation.
3. Registers the GUI as a new entry point (e.g., `letsgo gui` command in `cmd/`) while the CLI keeps working unchanged.

## Technical Context

**Language/Version**: Go 1.26 (per `go.mod`), existing module `github.com/user/go-claude-code`

**Primary Dependencies**:
- NEW: `fyne.io/fyne/v2` v2.8.x (current stable line as of Jul 2026; requires a C compiler on Windows — gcc via TDM-GCC or MSYS2)
- Existing (reused): `github.com/spf13/cobra` (CLI), `github.com/charmbracelet/bubbletea` (CLI TUI), `github.com/spf13/viper` (config), `modernc.org/sqlite` (storage via `internal/db`), `github.com/yuin/goldmark` (markdown → RichText rendering in GUI), `github.com/google/uuid`, `golang.org/x/term`
- NOT used by the GUI: `bubbletea`, `lipgloss`, `glamour` (TUI-only presentation concerns)

**Storage**: Existing SQLite via `internal/db` — sessions, messages, context files, memory, compact history, search. No schema change required for v1; GUI reads/writes through the same package.

**Testing**: Go stdlib testing for `internal/engine` (unit: loop orchestration, permission driver, event sequencing). `fyne.io/fyne/v2/test` for `internal/gui` widget tests (rendering chat messages, approval dialog state, theme). Existing tests for core packages stay green (CLI regression gate).

**Target Platform**: Windows 10/11 (primary dev machine), macOS, Linux desktop. Cross-compile with the `fyne` CLI (`fyne package`) / `fyne-cross` later; no mobile scope.

**Project Type**: desktop-app + shared core modules (single Go module, two presentation layers)

**Performance Goals**:
- SC-001: response begins ≤5 s on broadband
- SC-004: window interactive ≤3 s from launch
- Streaming deltas pushed to UI at ≤100 ms granularity (not per-token UI refresh)
- UI never blocks during agent work (engine runs off the UI goroutine; UI updates via `fyne.Do`)

**Constraints**:
- UI mutation only on the Fyne main thread (`fyne.Do`/`fyne.DoAndWait`)
- C compiler toolchain required for Fyne builds on Windows
- CLI behavior must not change (FR-026): TUI refactor is behavior-preserving
- Provider abstraction (`internal/api`) must not be bypassed by the GUI (FR-022)
- Every sensitive tool action flows through the existing `PermissionManager` + engine approval (FR-014/FR-016)

**Scale/Scope**: single-user desktop app; conversations can grow unbounded → chat view uses incremental append + scroll lock with lazy loading of old sessions (or windowed history) to stay responsive; parity surface ≈ 75 CLI commands mapped to GUI (contract: `contracts/cli-parity.md`).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

`.specify/memory/constitution.md` is an unfilled template (placeholder sections only, no ratified principles, no version). **GATE PASSED — no constitution violations.** Default quality bar applies:
- New packages (`internal/engine`, `internal/gui`) MUST ship with tests
- No new top-level package outside `internal/` for this feature
- Core packages (`api`, `tools`, `db`, `config`) must not be modified except where extracting the engine loop requires it; extraction must be behavior-preserving (verified by existing CLI tests)

Re-checked after Phase 1: still no violations; design stays within a single Go module with shared core (no new projects, no duplicated logic).

## Project Structure

### Documentation (this feature)

```text
specs/001-fyne-gui/
├── plan.md              # This file
├── research.md          # Phase 0 output (resolved technical decisions)
├── data-model.md        # Phase 1 output (entities mapped to existing schema)
├── quickstart.md        # Phase 1 output (validation guide)
├── contracts/           # Phase 1 output
│   ├── engine-contract.md   # Engine event/command protocol (shared by GUI + TUI)
│   ├── gui-contract.md      # GUI views, actions, keyboard map
│   └── cli-parity.md        # CLI command → GUI surface parity checklist
└── tasks.md             # Phase 2 output (/speckit.tasks - NOT created here)
```

### Source Code (repository root)

```text
cmd/
├── root.go              # existing — add `gui` subcommand entry point
└── gui.go               # NEW — cobra command that launches internal/gui

internal/
├── api/                 # existing — multi-provider client (UNTOUCHED)
├── tools/               # existing — tool registry, PermissionManager (UNTOUCHED)
├── db/                  # existing — SQLite storage (UNTOUCHED)
├── config/              # existing — viper config (UNTOUCHED)
├── mcp/  plugins/  analytics/  github/  lsp/  updater/   # existing (UNTOUCHED)
├── engine/              # NEW — presentation-agnostic conversation/agent loop
│   ├── engine.go            # loop orchestration (stream, tool cycle, approvals)
│   ├── events.go            # typed events (StreamDelta, ToolRequested, ToolResult, Done, Error…)
│   ├── commands.go          # commands (SendMessage, Cancel, ApproveTool, RejectTool, SetAutoApprove…)
│   ├── permission_driver.go # interface: presentations implement approval UI
│   └── engine_test.go
├── tui/                 # existing — REFACTORED to consume internal/engine (loop moved out)
└── gui/                 # NEW — Fyne presentation layer
    ├── app.go               # bootstrap, window, theme wiring
    ├── theme.go             # dark terminal theme (palette, monospace font)
    ├── chatview.go          # conversation view: markdown RichText, streaming append
    ├── toolpanel.go         # inline tool panels: diffs, command output, approval buttons
    ├── sessions.go          # session sidebar: list, create, resume, rename
    ├── settings.go          # provider/model/API key + auto-approve settings
    ├── usage.go             # cost/token usage view
    ├── gitview.go           # git operations view (branch/commit/diff/review)
    ├── keyboard.go          # shortcut map (FR-025)
    └── gui_test.go
```

**Structure Decision**: Single Go module; two presentation layers (TUI + GUI) over one shared core. The only new architectural piece is `internal/engine`, which lifts the conversation loop out of the TUI so neither presentation owns the loop logic. The GUI lives entirely under `internal/gui`, keeping the Fyne dependency contained (it must not leak into `api`, `tools`, `db`, or `engine`).

## Complexity Tracking

No constitution violations — table not applicable.

## Phase 0: Research (see `research.md`)

Resolved decisions: Fyne v2.8.x target; dark terminal theme approach; streaming/threading model (`fyne.Do`); markdown rendering via goldmark → RichText; engine extraction boundaries from the TUI; approval driver interface; Fyne testing strategy; Windows toolchain.

## Phase 1: Design (see `data-model.md`, `contracts/`, `quickstart.md`)

- Data model: maps spec entities (Session, Message, Tool Call, Agent Task, MCP Configuration, Plugin, Provider Configuration, Usage Record) onto the existing `internal/db` schema — no new storage engine
- Contracts: engine event/command protocol (the GUI/TUI boundary), GUI navigation & keyboard contract, CLI parity checklist
- Quickstart: end-to-end validation scenarios (launch, chat, approval flow, session persistence, provider switch, parity smoke test)

## Done When

- [X] `internal/engine` exists with tests; TUI refactored to consume it (existing CLI tests green)
- [ ] `internal/gui` renders chat + tool approvals with dark terminal theme (GUI tests pass) — code implemented; GUI tests pass with MSYS2 toolchain (see note in Done When). 1 flaky session test remains (updated_at second-granularity race)
- [ ] Sessions, settings, usage, git, MCP, plugins reachable from the GUI (parity checklist checked) — sessions+settings done; usage/git/MCP/plugins pending
- [ ] `cmd/gui.go` launches the app; `go build ./...` green on Windows — GREEN: `go build ./...` y `go vet ./...` pasan con `D:\msys64\ucrt64\bin` + CGO_ENABLED=1

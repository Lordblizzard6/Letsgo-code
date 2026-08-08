# Research: Fyne Desktop GUI for LetsGO Code

**Date**: 2026-08-02 | **Plan**: [plan.md](plan.md) | **Spec**: [spec.md](spec.md)

Phase 0 output. Resolves every technical unknown from the plan's Technical Context.

---

## 1. GUI toolkit version and module

- **Decision**: Use `fyne.io/fyne/v2` v2.8.x (latest stable v2.8.0, released 11 Jul 2026; v2.7.x is the mature patch line).
- **Rationale**: Current stable release with the improved threading model (introduced in 2.6) and expanded `RichText` capabilities needed for markdown chat rendering; single codebase for desktop platforms.
- **Alternatives considered**: `wails` (webview-based, heavier runtime, needs Node tooling), `gioui` (lower-level, more manual layout), GTK/Qt bindings (non-idiomatic for this codebase). Fyne keeps the "pure Go + existing internal packages" story intact.

## 2. Windows build toolchain

- **Decision**: Require a C compiler (TDM-GCC or MSYS2/MinGW `gcc`) in `PATH`; verify with `gcc -v` before `go build`.
- **Rationale**: Fyne uses cgo-backed OpenGL bindings on desktop; the first Windows compile can take several minutes, later builds are fast.
- **Alternatives considered**: GLFW pure-Go fallback does not exist for the full toolkit; `fyne-cross` Docker images are a later cross-compile option, out of v1 scope.

## 3. Dark terminal aesthetic (FR-023)

- **Decision**: Implement a custom `fyne.Theme` in `internal/gui/theme.go`:
  - Force `theme.VariantDark` on the app (`app.NewWithID` + `a.Settings().SetTheme(customTheme)`), with a terminal-style palette (near-black background, green/amber accents, ANSI-ish token colors).
  - Override `Font(TextStyle)` so `TextStyle{Monospace: true}` (or the default chat font) resolves to a bundled monospaced font (e.g., JetBrains Mono / Cascadia Mono / DejaVu Sans Mono embedded via `fyne.StaticResource`).
  - Custom widget renderers only where the stock look is unavoidable (chat bubbles, tool panels); otherwise standard widgets + theme colors.
- **Rationale**: Matches the Codex terminal look with the least custom rendering code; theme colors flow automatically to all stock widgets.
- **Alternatives considered**: Custom `widget.Widget` from scratch for every element (too much code, less testable); light-theme stock (rejected — the spec demands dark terminal style as default).

## 4. Streaming and threading model

- **Decision**: Engine runs on its own goroutine; UI never blocks.
  - Engine emits typed events to a buffered channel; a GUI pump goroutine forwards them to the main thread via `fyne.Do` (safe from any goroutine since Fyne 2.6's thread model).
  - Chat view appends in batches (throttle to ~10 refreshes/sec) rather than per-token, to keep rendering cheap on long streams.
  - Cancellation is a command (`Cancel`) processed by the engine loop; it aborts the in-flight provider stream and tool cycle.
- **Rationale**: `fyne.Do` is the supported cross-goroutine path; batching keeps `RichText` rebuilds bounded (performance goals SC-001/SC-004; UI responsiveness during agent work).
- **Alternatives considered**: Driving UI updates directly from engine goroutine (rejected — unsafe/racy); per-token refresh (rejected — measurable jank on long messages).

## 5. Markdown rendering in chat

- **Decision**: Parse assistant text with the existing `github.com/yuin/goldmark` dependency, walk the AST, and build `widget.RichText` segments (`TextSegment`, `HyperlinkSegment`, heading/bold/italic styles, `CodeSegment` for inline code, block code via monospace `TextSegment`).
- **Rationale**: goldmark is already in `go.mod` (used by the CLI) — zero new dependencies, one consistent renderer for both presentations.
- **Alternatives considered**: Third-party fyne-markdown libs (extra dependency, often unmaintained); plain text (rejected — degraded UX vs. Codex).

## 6. Engine extraction from the TUI

- **Decision**: Move the conversation loop out of `internal/tui/ui.go` into `internal/engine`:
  - `engine.Engine` owns: provider stream calls (`internal/api`), tool cycle (`internal/tools.ExecuteToolWithPermission`), history persistence (`internal/db`), compact/summarize triggers, cost tracking (`internal/tools.CostTracker`).
  - The TUI becomes a thin consumer: translate Bubble Tea messages → engine commands, engine events → Bubble Tea messages. Existing CLI behavior (inputs, prompts, output formatting) is preserved exactly; CLI tests are the regression gate.
  - Permission prompts go through `engine.PermissionDriver` (interface `Prompt(toolName, input) (allow, auto, err)`) — the TUI implements it with the existing stdin prompt; the GUI with a dialog (see contract `contracts/engine-contract.md`).
  - Approval state machine lives in the engine so both presentations get FR-014/FR-016 semantics for free.
- **Rationale**: This is the only way FR-021/FR-026 (shared core, GUI↔CLI parity without duplication) can hold; the tools/api/db packages already expose everything needed.
- **Alternatives considered**: Reimplementing the loop inside the GUI (rejected — guaranteed drift from CLI, violates FR-026); threading a shared state object through the TUI (rejected — TUI owns the loop today, extraction is cleaner).

## 7. Approval UX in the GUI

- **Decision**: Inline tool panel + modal dialog:
  - Each pending tool call renders an inline panel in the chat (tool name, input summary, diff preview for edits) with Approve / Reject buttons; a keyboard-driven modal `dialog.NewCustomConfirm` blocks further agent progress while pending (keyboard: Enter approve, Esc reject — FR-025).
  - Auto-approve mode is a per-category toggle in Settings (file edits, commands, web), persisted via `internal/config`; active auto-approve is visible in a status bar so the user always knows the agent can act without prompts.
- **Rationale**: Codex interaction pattern (explicit, visible, cancellable); zero silent modifications (SC-007); the existing `PermissionManager.CheckPermission` already classifies tool categories.
- **Alternatives considered**: Toast notifications (rejected — too easy to miss approvals); always-modal on every tool (rejected — auto-approve would be pointless).

## 8. GUI testing strategy

- **Decision**:
  - `internal/engine`: table-driven unit tests — stream→event mapping, tool cycle with mocked `PermissionDriver` (approve/reject/timeout), cancellation mid-stream, event ordering. No Fyne import.
  - `internal/gui`: `fyne.io/fyne/v2/test` — window construction, chat append renders text, approval dialog appears on `ToolRequested` event, shortcut map, theme variant is dark.
  - Parity gate: existing core + CLI tests must stay green after the TUI refactor (run `go test ./...`).
- **Rationale**: Standard Go testing + Fyne's test harness (headless, no display needed); keeps GUI logic testable without launching a window.
- **Alternatives considered**: UI automation frameworks (out of scope for v1).

## 9. Sessions and long conversations

- **Decision**: Reuse `internal/db` as-is. Chat view appends incrementally; when reopening a large session, render the last N messages first and load older history on scroll-up (windowed loading), so SC-002 (100% resume) holds without unbounded rendering cost.
- **Rationale**: Existing `GetHistory`, `ListSessions`, `ResumeSession` cover the requirements; windowing is a view concern.
- **Alternatives considered**: Pagination UI (rejected — breaks the chat metaphor).

## 10. Multi-provider parity (FR-022)

- **Decision**: The GUI only talks to `internal/api` (same multi-provider client the CLI uses: Anthropic, OpenAI, Groq, OpenRouter, Ollama) and surfaces provider/model selection in Settings. No provider-specific code in the GUI.
- **Rationale**: Provider-agnostic guarantees come from the shared client, not from a second implementation.
- **Alternatives considered**: GUI-native provider adapters (rejected — duplicates core, violates FR-022/FR-026).

---

## Consolidated decisions table

| # | Unknown | Decision |
|---|---------|----------|
| 1 | Toolkit version | fyne.io/fyne/v2 v2.8.x |
| 2 | Windows build | gcc (TDM-GCC/MSYS2) in PATH |
| 3 | Dark terminal look | Custom `fyne.Theme`, bundled monospace font, dark variant |
| 4 | Streaming | Engine goroutine → buffered channel → `fyne.Do`, batched refresh |
| 5 | Markdown | goldmark → `widget.RichText` segments |
| 6 | Shared loop | `internal/engine` extracted from TUI; TUI refactored to consume it |
| 7 | Approvals | Inline tool panel + keyboard-driven confirm dialog; per-category auto-approve |
| 8 | Testing | stdlib for engine; `fyne/test` for GUI; CLI tests as parity gate |
| 9 | Long sessions | Incremental append + windowed history load |
| 10 | Providers | GUI uses `internal/api` only, never provider-specific code |

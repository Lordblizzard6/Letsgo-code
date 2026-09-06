# Quickstart: Fyne Desktop GUI for LetsGO Code

**Date**: 2026-08-02 | **Plan**: [plan.md](plan.md) | **Contracts**: [engine](contracts/engine-contract.md) · [gui](contracts/gui-contract.md) · [parity](contracts/cli-parity.md) | **Data model**: [data-model.md](data-model.md)

Runnable validation scenarios proving the feature works end-to-end. Implementation details live in `tasks.md`.

## Prerequisites

- Go 1.26 (per `go.mod`) and a C compiler in `PATH` (Windows: TDM-GCC or MSYS2/MinGW `gcc`; verify `gcc -v`)
- Fyne dependency present: `go get fyne.io/fyne/v2@v2.8.0` (first build on Windows can take minutes)
- Existing app config with at least one provider/API key, or follow scenario S1 to configure one
- A scratch git repo for the agent scenarios (e.g., `git init` in a temp folder)

## Build & unit gate

```powershell
go get fyne.io/fyne/v2@v2.8.0
go mod tidy
go test ./...          # ALL tests green — CLI regression gate (parity, FR-026)
go build ./...         # compiles, incl. cmd/gui
go vet ./...
```

**Expected**: exit code 0; no failures in `internal/engine`, `internal/gui`, or existing core packages.

## Scenario S1 — First launch & provider configuration (FR-007/FR-008/FR-011, SC-003)

1. Delete/rename the local config so no provider is set (`$env:USERPROFILE\.letsGo\config.*` or CLI-equivalent).
2. Run the GUI (`go run . gui`).
3. **Expect**: app opens directly into the Settings flow; chat is blocked with a hint.
4. Select a provider, enter a valid API key, pick a model, save.
5. Restart the app. **Expect**: config retained; chat enabled; status bar shows provider + model.

## Scenario S2 — Chat round trip (FR-001..FR-003, SC-001)

1. Type "Say hello in one line" and press Enter.
2. **Expect**: user message appended immediately; assistant response begins streaming within 5 s on broadband; input re-enables only after the response completes (Idle).
3. Press Esc during a long response. **Expect**: stream stops; last complete message preserved.

## Scenario S3 — Agent mode with approvals (FR-013..FR-016, SC-007)

1. Open a scratch git repo in the app's working directory; enable Agent mode.
2. Prompt: "create a file hello.txt with the text hi, then run `dir` to verify".
3. **Expect**: a tool panel appears for the file edit with a diff preview and Approve/Reject buttons; dialog is keyboard-driven (Enter/Esc).
4. Approve the edit. **Expect**: edit executes; a second panel asks to run the command; its output renders inline.
5. Start a second task; reject its first tool action. **Expect**: action not applied; agent continues and reports adjustment.
6. Enable auto-approve for file edits in Settings; start a third task that edits a file. **Expect**: edit executes without a prompt and appears in the activity log; status bar shows auto-approve is active.
7. Leave an approval dialog open > timeout. **Expect**: treated as rejected with a notice (edge case).

## Scenario S4 — Session persistence (FR-005/FR-006/FR-012, SC-002)

1. Have a conversation; close the window mid-stream.
2. Relaunch. **Expect**: session listed; reopening shows full history up to the last complete message.
3. Create a new session and switch back with `Ctrl+N` / sidebar. **Expect**: both sessions intact.

## Scenario S5 — Provider independence (FR-022, SC-008)

1. In Settings, switch provider (e.g., Anthropic → OpenAI-compatible or Ollama local).
2. Repeat S2 and S3. **Expect**: identical feature set and flows; no capability disappears.

## Scenario S6 — Advanced capabilities (FR-017..FR-020)

1. **Git**: create a branch, edit a file via agent, commit with a message, open diff. **Expect**: all within the Git view.
2. **MCP**: add a test MCP server to config; request its tool in a conversation. **Expect**: tool listed and called through the same approval flow.
3. **Plugins**: install a test plugin; restart. **Expect**: plugin loaded; its contributions available in menus.
4. **Parallel agents**: ask for a large task split into sub-agents. **Expect**: Agent tasks panel shows progress; results report back into the conversation.

## Scenario S7 — Keyboard-only walkthrough (FR-025, SC-009)

Without touching the mouse: `Ctrl+L` focus input → send message → approve a tool with Enter → `Ctrl+N` new session → `Ctrl+1` switch back → `Esc` cancel. **Expect**: every step works; focus never lost.

## Scenario S8 — Parity smoke test (FR-021, SC-006)

Using [cli-parity.md](contracts/cli-parity.md): walk the checklist rows group by group and confirm each capability is reachable from its GUI target. **Expect**: no row is unreachable; count fully-verified rows and record the number against the checklist (100% = parity gate passed).

## Non-goals (guard against scope creep)

- No mobile builds, no packaging/distribution pipelines, no light theme in v1 — see [plan.md](plan.md) Technical Context.

---

description: "Task list for Fyne Desktop GUI implementation"
---

# Tasks: Fyne Desktop GUI for LetsGO Code

**Input**: Design documents from `/specs/001-fyne-gui/`

**Prerequisites**: [plan.md](plan.md) (required), [spec.md](spec.md) (user stories), [research.md](research.md) (decisions), [data-model.md](data-model.md) (entities), [contracts/](contracts/) (engine/gui/parity contracts)

**Tests**: Test tasks ARE included — plan.md and research.md #8 require unit tests for `internal/engine` and `fyne/test` tests for `internal/gui`, plus a CLI regression gate (`go test ./...`).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1..US5)
- Include exact file paths in descriptions

## Path Conventions

- Single Go module at repository root; presentation layers under `internal/` (see plan.md structure decision)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization for the GUI + engine packages

- [X] T001 Add Fyne dependency: run `go get fyne.io/fyne/v2@v2.8.0` then `go mod tidy` (files: `go.mod`, `go.sum`); verify `gcc -v` exists on PATH
- [X] T002 Register the GUI entry point: create `cmd/gui.go` with a cobra command `gui` that calls the GUI bootstrap, and register it in `cmd/root.go`
- [X] T003 Create `internal/gui/app.go` skeleton: `app.NewWithID`, main window creation with title, `window.ShowAndRun()`, empty content container
- [X] T004 [P] Create the dark terminal theme in `internal/gui/theme.go`: implement `fyne.Theme` (dark variant only, near-black background, ANSI-style accent palette) and bundle a monospaced font as `fyne.StaticResource` used for all text styles

**Checkpoint**: `go build ./...` compiles; `go run . gui` opens an empty dark window

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Extract the conversation/agent loop into `internal/engine` — blocks ALL user stories (per engine-contract.md)

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T005 Define engine types in `internal/engine/events.go` and `internal/engine/commands.go`: all event structs (StreamDelta, ToolRequested, ToolResult, Idle, Error…) and command structs (SendMessage, Cancel, ApproveTool…) per `contracts/engine-contract.md`
- [X] T006 Implement `internal/engine/engine.go` core loop: `SendMessage` → provider stream via `internal/api` → event emission (StreamStart/StreamDelta/StreamDone) with ≤100ms batching; persistence via `internal/db`
- [X] T007 Implement `internal/engine/permission_driver.go`: `PermissionDriver` interface + approval state machine (requested → pending → approved/rejected/timeout-as-rejected) with per-category auto-approve support
- [X] T008 Implement the tool cycle in `internal/engine/engine.go`: on tool_use → `ToolRequested` event → driver prompt (or auto-approve) → `tools.ExecuteToolWithPermission` → `ToolResult` event; guarantee no execution without approval (SC-007)
- [X] T009 Implement cancellation and shutdown in `internal/engine/engine.go`: `Cancel` aborts in-flight stream/tool cycle (last complete message preserved), `Stop` for graceful close
- [X] T010 [P] Write engine unit tests in `internal/engine/engine_test.go`: stream→event mapping, tool cycle with mocked `PermissionDriver` (approve/reject/timeout), cancel mid-stream, event ordering guarantees, `Idle` after every turn
- [X] T011 Refactor `internal/tui/ui.go` to consume `internal/engine` (replace inline loop with command/event translation; keep existing prompt text and output behavior identical) — TUI's PermissionDriver uses the existing stdin prompt flow
- [X] T012 Run `go test ./...` — CLI regression gate: ALL existing core + TUI tests must pass after the refactor (FR-026)

**Checkpoint**: Engine foundation ready; CLI unchanged in behavior; user story implementation can begin

---

## Phase 3: User Story 1 - Chat in a desktop window (Priority: P1) 🎯 MVP

**Goal**: Launch the GUI, send a message, see the AI response stream in a dark terminal-styled chat view (spec.md US1)

**Independent Test**: `go run . gui` → type a message → response streams in ≤5s; Esc cancels; errors show inline (quickstart S1/S2)

### Tests for User Story 1 (required per plan.md)

- [X] T013 [P] [US1] Write GUI tests in `internal/gui/gui_test.go` with `fyne/test`: chat append renders markdown text, Send issues the SendMessage command, Esc cancels an active stream (tests must fail until implementation exists)

### Implementation for User Story 1

- [X] T014 [P] [US1] Implement markdown renderer in `internal/gui/markdown.go`: parse with `github.com/yuin/goldmark`, build `widget.RichText` segments (text, bold/italic/heading, hyperlink, inline code, block code monospace)
- [X] T015 [P] [US1] Implement chat view in `internal/gui/chatview.go`: scrollable conversation container, message append, multiline input composer with Send button, Enter=send / Ctrl+Enter=newline / Esc=cancel (FR-001/002/004)
- [X] T016 [P] [US1] Implement the event pump in `internal/gui/stream.go`: goroutine reading engine events → `fyne.Do` batched UI refresh (≤100ms), input disabled until `Idle`/`Error`
- [X] T017 [US1] Wire the controller in `internal/gui/app.go`: engine creation, SendMessage/Cancel commands, error surfacing per FR-010
- [X] T018 [US1] Implement minimal settings in `internal/gui/settings.go`: provider selector, API key field, model selector via `internal/config` (FR-007/008); if no provider configured, open Settings first and block chat (FR-011)

**Checkpoint**: quickstart S1 (first launch config) and S2 (chat round trip) pass

---

## Phase 4: User Story 2 - Agent mode with tool approvals (Priority: P1)

**Goal**: Assign coding tasks; agent edits files, runs commands, searches web — every sensitive action approved, inline panels show diffs/output (spec.md US2)

**Independent Test**: quickstart S3 — task requiring a file edit + command run shows approval panels; reject path never applies; auto-approve executes without prompts and logs activity

### Tests for User Story 2 (required per plan.md)

- [X] T019 [P] [US2] Write GUI tests in `internal/gui/gui_test.go`: ToolRequested event shows the approval dialog, Enter approves, Esc rejects, timeout treats as rejected

### Implementation for User Story 2

- [X] T020 [P] [US2] Implement tool panel widget in `internal/gui/toolpanel.go`: renders tool name, input summary, diff preview for edits, command output, execution status per `contracts/gui-contract.md` (FR-024)
- [X] T021 [US2] Implement the approval driver in `internal/gui/approvals.go`: `PermissionDriver` backed by a keyboard-driven confirm dialog (Enter=approve, Esc=reject), timeout → rejected with notice (engine-contract.md)
- [X] T022 [US2] Implement status bar in `internal/gui/statusbar.go`: provider/model, auto-approve state with toggle (Ctrl+Shift+A) persisted via `internal/config`, usage summary
- [X] T023 [US2] Wire SendTask command and approval commands (ApproveTool/RejectTool/SetAutoApprove) into the controller in `internal/gui/app.go`

**Checkpoint**: quickstart S3 (approve, reject, auto-approve, timeout) passes; zero unapproved modifications (SC-007)

---

## Phase 5: User Story 3 - Session management (Priority: P2)

**Goal**: New conversations, session list, resume after restart, rename/delete/search (spec.md US3)

**Independent Test**: quickstart S4 — conversation survives restart (SC-002); switching sessions preserves both histories

### Tests for User Story 3 (required per plan.md)

- [X] T024 [P] [US3] Write GUI tests in `internal/gui/gui_test.go`: session list renders from `db.ListSessions`, new session creates and switches, closing window mid-stream issues engine `Stop`

### Implementation for User Story 3

- [X] T025 [P] [US3] Implement sessions sidebar in `internal/gui/sessions.go`: list (name/updated), create (`db.CreateSession`), resume (`engine` SwitchSession), rename (`db.RenameSession`), delete (`db.DeleteSession`)
- [X] T026 [US3] Implement windowed history loading in `internal/gui/chatview.go`: reopen renders last N messages, older messages load on scroll-up (data-model.md scale assumptions)
- [X] T027 [US3] Implement session search in `internal/gui/sessions.go` using `db.SearchMessages`
- [X] T028 [US3] Wire window-close handling in `internal/gui/app.go`: on close, send engine `Stop` (FR-012: last complete message persisted)

**Checkpoint**: quickstart S4 passes; sessions sidebar fully functional

---

## Phase 6: User Story 5 - Providers, settings & usage monitoring (Priority: P3)

**Goal**: Full settings surface + usage/cost view + parity for config commands (spec.md US5, cli-parity.md Config group)

**Independent Test**: quickstart S5 — switch provider and repeat S2/S3 with identical feature set (FR-022, SC-008); usage view shows tokens/cost for the period

### Implementation for User Story 5

- [X] T029 [P] [US5] Implement usage view in `internal/gui/usage.go`: aggregate `db` message token/cost columns per provider + `internal/tools.CostTracker` for the current run (FR-009)
- [X] T030 [P] [US5] Extend settings in `internal/gui/settings.go` with parity config rows: effort, output style, permissions levels, sandbox toggle, privacy settings, rate limit options, env vars (cli-parity.md)
- [X] T031 [P] [US5] Implement provider auth actions in `internal/gui/settings.go`: login/logout/oauth-refresh wired to the CLI's existing auth logic
- [X] T032 [US5] Implement app/meta actions menu in `internal/gui/app.go`: doctor, update/upgrade, version, release notes (cli-parity.md)

**Checkpoint**: quickstart S5 passes; usage view shows real data after S2 runs

---

## Phase 7: User Story 4 - Advanced capabilities: git, parallel agents, MCP, plugins (Priority: P3)

**Goal**: git operations, parallel sub-agents, MCP servers, plugin management from the GUI (spec.md US4)

**Independent Test**: quickstart S6 — git branch/commit/diff flow; MCP tool called through approval flow; plugin loads after restart; sub-agent progress panel

### Tests for User Story 4 (required per plan.md)

- [X] T033 [P] [US4] Write GUI tests in `internal/gui/gui_test.go`: git view actions call the existing git logic, MCP view lists configured servers

### Implementation for User Story 4

- [X] T034 [P] [US4] Implement git view in `internal/gui/gitview.go`: branch create/switch, commit (message + files), diff viewer, review trigger, tags/hooks/undo-redo/commit-push-pr via existing git commands (cli-parity.md Git group, FR-017)
- [X] T035 [P] [US4] Implement MCP view in `internal/gui/mcpview.go`: server list + connection status via `internal/mcp.Manager`, discovered tools (FR-019)
- [X] T036 [P] [US4] Implement plugins view in `internal/gui/pluginsview.go`: install/list/remove via `internal/plugins` registry; loaded on restart (FR-020)
- [X] T037 [P] [US4] Implement agent tasks panel in `internal/gui/agenttasks.go`: render AgentTaskUpdate events, parallel sub-agent progress and results (FR-018)
- [X] T038 [US4] Implement remaining parity rows in `internal/gui/app.go` menus: memory browser, skills, share, export, onboarding, plan mode toggle, fast mode, insights, review/ultrareview/security-review/advisor, thinkback, debug tool call (cli-parity.md — resolves all TBD rows)
- [X] T039 [US4] Implement global shortcut map in `internal/gui/keyboard.go` per `contracts/gui-contract.md` (Ctrl+N new session, Ctrl+1..9 switch, Ctrl+, settings, Ctrl+U usage, Ctrl+L focus input)

**Checkpoint**: quickstart S6 passes; cli-parity.md has no remaining TBD rows (SC-006)

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Quality gates, performance, and full validation

- [X] T040 Run `go vet ./...` and `go test ./...`; fix all failures (full regression: engine + gui + existing CLI)
- [X] T041 Run every scenario in `quickstart.md` (S1–S8) and record results against cli-parity.md
- [X] T042 Keyboard-only walkthrough: verify all primary actions without a mouse (SC-009); fix focus issues in `internal/gui/keyboard.go`
- [X] T043 Performance pass: verify batched streaming refresh and windowed history on a 1000+ message session (SC-004, no UI freeze)
- [X] T044 Security pass: API keys only ever via `internal/config`, never logged; tool approvals cannot be bypassed (FR-014, SC-007)
- [X] T045 [P] Update documentation: GUI section in `README.md` (run instructions, toolchain requirement, screenshot placeholder)

**Checkpoint**: parity gate passed (SC-006), all quality gates green

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **User Stories (Phases 3–7)**: All depend on Foundational completion; proceed in priority order (US1 → US2 → US3 → US5 → US4)
- **Polish (Phase 8)**: Depends on all user stories complete

### User Story Dependencies

- **US1 (P1)**: Engine only — no story dependencies
- **US2 (P1)**: Engine approval machinery (T007/T008) + US1 view/pump patterns; independently testable via quickstart S3
- **US3 (P2)**: Engine `SwitchSession`/`Stop` (T006/T009) + US1 chat view
- **US5 (P3)**: US1 minimal settings + US2 status bar
- **US4 (P3)**: US1 + US2 (approval flow reused by MCP tools) + US3 (sessions for agents); listed last because it composes the rest

### Within Each User Story

- Tests (where included) are written FIRST and must FAIL before implementation
- View widgets → event pump → controller wiring → story integration
- Story complete before moving to next priority

### Parallel Opportunities

- Phase 1: T004 [P] parallel with T001–T003
- Phase 2: T010 [P] (tests) can be written once T005–T006 exist; T005/T007 [P]-style splits within engine files are possible but keep sequential for coherence
- Phase 3: T013 (tests) first, then T014/T015/T016 all [P]
- Phase 4: T019 tests first; T020/T021 [P]
- Phase 5: T024 tests first; T025 [P] independent of US1 code
- Phase 6: T029/T030/T031 all [P]
- Phase 7: T034/T035/T036/T037 all [P]; T038 after
- After Foundational: US3 and US5 (P2/P3, non-blocking) can run in parallel with US1/US2 if staffed

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 first (must fail):
Task: "Write GUI tests in internal/gui/gui_test.go: chat append, Send command, Esc cancel"

# Launch all US1 implementation tasks together:
Task: "Implement markdown renderer in internal/gui/markdown.go"
Task: "Implement chat view in internal/gui/chatview.go"
Task: "Implement event pump in internal/gui/stream.go"
```

## Parallel Example: User Story 4

```bash
Task: "Implement git view in internal/gui/gitview.go"
Task: "Implement MCP view in internal/gui/mcpview.go"
Task: "Implement plugins view in internal/gui/pluginsview.go"
Task: "Implement agent tasks panel in internal/gui/agenttasks.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001–T004)
2. Complete Phase 2: Foundational (T005–T012) — CRITICAL, blocks everything
3. Complete Phase 3: User Story 1 (T013–T018)
4. **STOP and VALIDATE**: run quickstart S1/S2; demo the chat window
5. Deploy/demo if ready

### Incremental Delivery (recommended)

1. Setup + Foundational → engine ready, CLI unchanged
2. US1 chat → test independently (S1/S2) → **demo = MVP**
3. US2 agent + approvals → test independently (S3) → **demo = Codex-like core** (this is the 1:1 differentiator)
4. US3 sessions → test (S4)
5. US5 settings/usage → test (S5)
6. US4 advanced (git/MCP/plugins/agents) → test (S6); complete parity (S8)
7. Polish → full gates (S7, performance, security)

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: US1 → US2 (chat + agent loop)
   - Developer B: US3 (sessions sidebar, independent of A's view work beyond chatview)
   - Developer C: US5 (settings/usage)
3. Then Developer A or B: US4 (composes everything)
4. All stories integrate; parity gate + Polish by the whole team

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to the spec user story for traceability
- Each user story is independently completable and testable via its quickstart scenario
- Tests (T010, T013, T019, T024, T033) must fail before their implementation exists
- Commit after each task or logical group
- Stop at any checkpoint to validate the story independently
- Avoid: vague tasks, same-file conflicts (respect [P] markers), cross-story dependencies that break independence
- Final authority on contracts: `contracts/engine-contract.md` (event/command protocol) and `contracts/gui-contract.md` (views/shortcuts); parity verified via `contracts/cli-parity.md`

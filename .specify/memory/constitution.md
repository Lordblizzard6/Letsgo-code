<!--
  SYNC IMPACT REPORT (generated 2026-08-13 by /speckit.constitution)
  Version change: 0.0.0 (empty template) → 1.0.0
  Report rationale for bump: first materialization of the constitution from the
  scaffolded template; establishes the test-last governance policy requested by
  the user ("evitar esa intensividad de test, dejarlas al final siempre").

  Modified principles (template slot → filled title):
    - [PRINCIPLE_1_NAME] → I. Test-Last Verification (NON-NEGOTIABLE)
    - [PRINCIPLE_2_NAME] → II. Contract-Only Core Access
    - [PRINCIPLE_3_NAME] → III. One Source of Truth
    - [PRINCIPLE_4_NAME] → IV. Secure Rendering
    - [PRINCIPLE_5_NAME] → V. Simplicity & YAGNI
  Added sections:
    - Technology & Platform Constraints
    - Development Workflow & Quality Gates
    - Governance (amendment procedure, versioning, compliance)
  Removed sections: none (all template slots materialized).

  Templates requiring updates:
    ✅ .specify/templates/tasks-template.md       (test ordering → tests last)
    ✅ .specify/templates/plan-template.md        (Constitution Check gate note)
    ✅ .opencode/commands/speckit.tasks.md        (task ordering → tests last)
    ✅ .opencode/commands/speckit.implement.md    (TDD section → tests last)
    ⚠  specs/005-bubbletea-tui-wails/tasks.md    (feature task ordering; manual
         follow-up when that story is re-scoped with /speckit.tasks)

  Follow-up TODOs: none (no intentional deferred placeholders).
-->

# LetsGO Code Constitution

## Core Principles

### I. Test-Last Verification (NON-NEGOTIABLE)
All automated tests (unit, Vitest, Playwright, teatest) MUST be written and executed
AFTER the corresponding implementation, batched at the end of each user story or
feature phase. Test-first / red-green TDD cycles are NOT mandatory and MUST NOT gate
implementation progress. Each story's end-of-phase test suite and manual validation
(J1-J4 journeys) are the verification gate; a story is complete only when its
implementation exists AND its end-of-phase suite passes. Rationale: minimize testing
overhead during construction per explicit user directive while retaining a regression
gate at every delivery point.

### II. Contract-Only Core Access
TUI, GUI Wails, and any future interface MUST consume `internal/engine`,
`internal/db`, and `internal/config` exclusively through the published frontend
contract surface (commands §1, events §2, sessions/config §3, conformance
C-001..C-006 in `specs/005-bubbletea-tui-wails/contracts/`). Direct access to core
internals from `internal/tui` or `cmd/wails/services` outside that surface is
prohibited, and `internal/engine/contract_test.go` enforces it. Interfaces are swappable.

### III. One Source of Truth
All domain behavior and persistence live in `internal/engine`, `internal/db`, and
`internal/config`. No duplication of core logic in interfaces (TUI or GUI). Client-side
state in the Wails frontend is presentation state only and MUST be re-derived from
contract events (`chat:delta`, `session:*`, `tool:*`, `theme:changed`, `config:changed`).
Schema additions require a data-model amendment before implementation.

### IV. Secure Rendering
The GUI MUST NEVER set raw `innerHTML` from untrusted content. Markdown rendering
combines `streaming-markdown` with `DOMPurify` (double sanitize over the accumulated
text, never `innerHTML`) per contract §2 and decision D3. Both themes (dark/light)
MUST maintain WCAG contrast AA (text ≥4.5:1, non-text ≥3:1) per SC-003. Focus MUST be
visible in both themes.

### V. Simplicity & YAGNI
Implementation starts minimal and avoids speculative features. Complexity beyond the
obvious solution MUST be justified in the plan's Complexity Tracking table before
inclusion. Prefer existing contracts and shared patterns over new abstractions.

## Technology & Platform Constraints

- Core in Go pinned to `1.26` (go.mod), CLI via `main.go` + `cmd/`.
- GUI is Wails v3 pinned exactly to `v3.0.0-beta.7` (`github.com/wailsapp/wails/v3`);
  WebView2 embedded for distribution (`-webview2 embed`, decision D4). Fyne is retired
  after US3 gates G-1..G-3 (FR-018) — `go mod why fyne.io/systray` must be empty.
- Frontend: Svelte + TypeScript + Vite under `cmd/wails/frontend/`; bindings generated
  via `wails3 generate bindings`; runtime `@wailsio/runtime`.
- Markdown/security deps: `streaming-markdown`, `dompurify`.
- TUI: Bubble Tea (`charmbracelet/bubbletea`) formalized over the same contract.
- Testing tooling: Go `go test`, `teatest` for TUI, Vitest + jsdom + Playwright for FE.
- Storage: SQLite via `internal/db` (no new tables without data-model amendment).

## Development Workflow & Quality Gates

- Order within each user story: Setup → core/interface implementation → end-of-phase
  tests written and run (Test-Last, Principle I) → manual validation J1-J4.
- Commit after each task or logical group. Stop at each checkpoint to validate the story.
- US3 gates run in order: G-1 (ported regression) → G-2 (Fyne code removed,
  `internal/gui` deleted) → G-3 (go.mod tidy, assets migrated to `cmd/wails/build/`).
- Final Gate B: `go test ./...` + `go build ./...` green without Fyne, CI (Go + Vitest
  + Playwright) passing, and quickstart journeys J1-J4 executed.
- Regression suites stay green as the baseline reference; a failing regression suite at
  any gate blocks that gate only, never mid-story construction.

## Governance

This constitution supersedes all other development practices in this repository.
Amendments require documentation of the change, user approval, and propagation to the
dependent artifacts (templates and speckit commands listed in the Sync Impact Report).
Versioning follows SemVer: MAJOR for principle removals/redefinitions, MINOR for new
principles/sections, PATCH for clarifications. Compliance is reviewed at each story
checkpoint and at Gate B; `plan.md`'s Constitution Check gate must be re-validated on
every plan amendment.

**Version**: 1.0.0 | **Ratified**: 2026-08-07 | **Last Amended**: 2026-08-13
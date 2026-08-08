# Contract: GUI Surface (`internal/gui`)

**Feature**: Fyne Desktop GUI for LetsGO Code | **Date**: 2026-08-02

The Fyne presentation layer's navigable surface, actions, and keyboard map. Implementation details (widget internals, layout specifics) belong to `tasks.md`; this is the behavior contract for acceptance tests.

## Layout

```text
┌──────────────────────────────────────────────────────────────┐
│ Menu bar: File · Session · Git · Tools · Help                  │
├──────────┬───────────────────────────────────────────────────┤
│ Sessions │  Conversation (chat view + inline tool panels)    │
│ sidebar  │                                                   │
│ (list,   │                                                   │
│  new,    │                                                   │
│  search) ├───────────────────────────────────────────────────┤
│          │  Input box (multiline)  [Send] [Cancel] [Agent]   │
├──────────┴───────────────────────────────────────────────────┤
│ Status bar: provider · model · auto-approve state · usage    │
└──────────────────────────────────────────────────────────────┘
```

## Views

| View | Trigger | Contents |
|------|---------|----------|
| Chat | default | Conversation thread (markdown-rendered messages), inline tool panels, input box |
| Sessions | sidebar / menu | List sessions (name, updated), create new, resume, rename, delete, search (`db.SearchMessages`) |
| Settings | Tools menu | Provider selector, API key field, model selector, auto-approve toggles per category |
| Usage | status bar click / menu | Aggregated tokens + cost per provider for current period |
| Git | Git menu | Branch create/switch, commit (message + files), diff viewer, trigger review |
| Plugins | Tools menu | Installed plugin list, install/remove, load status |
| MCP | Tools menu | Server list, connection status, discovered tools |
| Agent tasks | menu / inline | Parallel sub-agent progress panel (FR-018) |

## Actions (mouse and/or keyboard)

| Action | Keyboard | Notes |
|--------|----------|-------|
| Send message | `Enter` (input), `Ctrl+Enter` = newline | FR-002 |
| Cancel response | `Esc` (while streaming) | FR-004 |
| Approve tool | `Enter` (dialog open) | FR-014 |
| Reject tool | `Esc` (dialog open) | FR-016 |
| New session | `Ctrl+N` | FR-006 |
| Switch session | `Ctrl+1..9` or sidebar click | FR-006 |
| Focus input | `Ctrl+L` | keyboard-only flows (FR-025) |
| Settings | `Ctrl+,` | FR-007 |
| Usage | `Ctrl+U` | FR-009 |
| Auto-approve toggle | `Ctrl+Shift+A` (status bar reflects state) | FR-014 |

## Appearance (FR-023)

- Default theme: dark, terminal-inspired; near-black background; ANSI-ish accent palette; monospaced font everywhere (bundled).
- Tool calls, diffs, and command output render in distinct inline panels, never as plain text (FR-024).
- No light theme in v1 (dark is the default and only shipped variant).

## Behavior guarantees (mapped to spec FRs)

- Launch with no provider configured → Settings flow first, chat blocked until valid (FR-011).
- Closing window mid-stream → engine receives `Stop`; last complete message persisted (FR-012).
- All primary actions reachable without a mouse (FR-025); verified by keyboard-only walkthrough (SC-009).
- Parity: every capability in `contracts/cli-parity.md` reachable from a view or menu (FR-021, SC-006).

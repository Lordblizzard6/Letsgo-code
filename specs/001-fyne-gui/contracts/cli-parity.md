# Contract: CLI → GUI Parity Checklist

**Feature**: Fyne Desktop GUI for LetsGO Code | **Date**: 2026-08-02

FR-021: every capability of the existing command-line application must be reachable from the GUI (SC-006 verified against this checklist). Grouped by CLI command (`cmd/*.go`); "GUI target" is where the capability lands in the GUI surface (see [gui-contract.md](gui-contract.md)).

Legend: **View** = dedicated GUI view/panel · **Menu** = menu or toolbar action · **Settings** = settings screen · **Chat** = slash command typed in the chat input · **TBD** = mapping assigned in tasks.md

## Core & messaging

| CLI command | GUI target | Notes |
|-------------|-----------|-------|
| `chat` | Chat view | main window |
| `clear` | Chat | context clear action |
| `copy` | Chat | copy-to-clipboard on messages |
| `context` | Chat | context panel |
| `files` | Chat | file list panel |
| `compact` | Chat | compact button / engine command |
| `exit` | Chat | window close |
| `status` | Status bar | provider/model/usage |
| `summary` | Chat | session summary action |
| `btw`, `thinkback`, `thinkback_play` | Menu | insight/thinkback views |
| `skills` | Menu | skills browser |

## Sessions

| CLI command | GUI target | Notes |
|-------------|-----------|-------|
| `session` | Sessions view | list |
| `resume` | Sessions view | resume on click |
| `rename` | Sessions view | inline rename |
| `reset` | Sessions view | clear history |
| `history` | Sessions view | history list |
| `rewind` | Sessions view | rewind conversation |
| `teleport` | Sessions view | jump to point in time |

## Git & GitHub

| CLI command | GUI target | Notes |
|-------------|-----------|-------|
| `init` | Git view | init repo |
| `branch` | Git view | create/switch |
| `tag` | Git view | tag list/create |
| `commit` | Git view | commit flow |
| `commit-push-pr` | Git view | commit + push + PR |
| `diff` | Git view | diff viewer |
| `hooks` | Git view | hooks manager |
| `undo` / `redo` | Git view | undo/redo |
| `github` | Git view | GitHub actions |
| `pr_comments` | Git view | PR comments |
| `install_github_app` | Menu | GitHub app install |

## Review & analysis

| CLI command | GUI target | Notes |
|-------------|-----------|-------|
| `review` | Menu | launch review |
| `ultrareview` | Menu | deep review |
| `security-review` | Menu | security review |
| `advisor` | Menu | advisor |
| `insights` | Menu | insights view |

## Configuration, providers & settings

| CLI command | GUI target | Notes |
|-------------|-----------|-------|
| `config` | Settings | same viper keys as CLI |
| `login` / `logout` / `oauth_refresh` | Settings | provider auth |
| `env` | Settings | environment variables |
| `effort` | Settings | effort level selector |
| `output_style` | Settings | output style |
| `theme` | Settings | theme (dark default) |
| `color` | Settings | color config |
| `rate_limit_options` | Settings | rate limit options |
| `privacy_settings` | Settings | privacy toggles |
| `permissions` | Settings | permission levels |
| `sandbox_toggle` | Settings | sandbox mode |
| `vim` | Settings | editor mode |
| `keybindings` | Menu | shortcut reference |
| `statusline` | Settings | status line content |
| `mobile` / `desktop` / `ide` / `chrome` | Settings | platform mode |
| `doctor` | Menu | diagnostics |
| `update` / `upgrade` / `version` / `release_notes` | Menu | app info/updates |

## Extensions & tooling

| CLI command | GUI target | Notes |
|-------------|-----------|-------|
| `mcp` | MCP view | server management |
| `plugin` / `reload_plugins` | Plugins view | install/list/remove |
| `memory` | Menu | memory browser |
| `lsp` | Chat | LSP status in context |
| `tasks` | Chat / Agent tasks | task management |
| `plan` | Chat | plan mode toggle |
| `share` | Chat | share session |
| `fast` | Chat | fast mode |
| `export` | Menu | export session |
| `onboarding` | Chat | onboarding flow |
| `install_slack_app` | Menu | Slack app install |
| `debug_tool_call` | Menu | tool call debug view |

## Acceptance rule

A GUI release only counts as parity-complete (SC-006) when **every row** has a working GUI target and is exercised in the quickstart smoke test ([quickstart.md](../quickstart.md)). Rows marked TBD in `tasks.md` must be resolved before the parity gate passes.

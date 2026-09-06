# Interface Contract: TUI Modernization & Keyboard Shortcuts

## Keyboard Mappings & Overlays

| Keybinding | Scope | Action | Expected Outcome |
|---|---|---|---|
| `Tab` | Input empty | Mode Toggle | Cycles `BUILD` → `PLAN` → `RESEARCH` → `BUILD` and syncs engine mode |
| `Tab` / `Shift+Tab` | Autocomplete active | Cycle selection | Moves selection through suggestions |
| `Ctrl+S` | Global | Toggle Model Picker | Opens/closes dynamic model & provider switcher overlay |
| `Ctrl+D` | Global | Toggle Git Review | Opens/closes interactive Git staging & diff overlay |
| `Ctrl+B` | Global | Toggle Activity Inspector | Opens/closes tool execution history & background tasks overlay |
| `Ctrl+Z` | Chat | Rollback | Reverts last assistant turn and restores prompt in textarea |
| `Esc` | Overlay active | Dismiss Overlay | Closes any open overlay and returns to chat focus |
| `Alt+Enter` / `Ctrl+J` | Chat | Newline | Inserts newline in multiline composer |

## Slash Command Parity Contract

All commands must be supported in both autocomplete and command execution:
- `/help`: Display formatted command reference.
- `/init`: Initialize `AGENTS.md` and workspace rules.
- `/model`: Open model picker or switch to model (`/model <name>`).
- `/models`: List available models.
- `/cost`: Display token spend and cost breakdown.
- `/tokens`: Display token counts.
- `/compact`: Compact conversation history.
- `/save`: Save current session.
- `/rename`: Rename current active session.
- `/sessions`: Open session switcher overlay.
- `/open <id>`: Switch directly to session ID.
- `/new`: Create fresh session.
- `/files`: List active context files.
- `/diff` / `/review`: Open Git Review overlay.
- `/tasks`: Open Activity Inspector overlay.
- `/terminal`: Open Commands execution history overlay.
- `/undo` & `/redo`: Soft undo/redo commit/turn actions.
- `/clear`: Clear conversation viewport.
- `/quit` / `/exit`: Exit cleanly.

# Contract: UI Composer and Window Layout

## 1. Window Initialization (`cmd/wails/gui/gui.go`)
- Initial window width: `1280`
- Initial window height: `850`
- Minimum width: `960`
- Minimum height: `600`

## 2. Composer Layout & Interactions (`App.tsx` & `app.css`)
- **Top Corner**:
  - Segmented control displaying all 4 modes: Plan, Código, Arch, Research.
  - Active mode receives highlighted background, distinct text color, and icon.
  - Shortcut cue: `Tab ⇥` label.
- **Keyboard Handling**:
  - `Tab`: if autocomplete is active, complete suggestion. Otherwise, advance to next mode.
  - `Shift+Tab`: if autocomplete is active, do nothing special. Otherwise, reverse to previous mode.
  - `Enter`: if autocomplete is active, pick selection. Otherwise, send message.
- **Bottom Toolbar**:
  - Left: `[📎 Attach]` + `[🤖 Model Selector ▾]`.
  - Right: `[↑ Enviar]`.
- **Sidebar**:
  - `[+ Nueva conversación]` button prominently placed above the conversation tab list.

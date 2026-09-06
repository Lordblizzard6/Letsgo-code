# Research: Composer Layout and Window Sizing Antigravity Parity

## Architectural Decisions

### Decision 1: Window Dimensions & Constraints
- **Decision**: Update Wails window options in `cmd/wails/gui/gui.go` to:
  ```go
  Width: 1280,
  Height: 850,
  MinWidth: 960,
  MinHeight: 600,
  ```
- **Rationale**: Matches standard dimensions for Antigravity IDE and modern developer tools on 1080p, 2K, and 4K displays with standard DPI scaling.

### Decision 2: Tab-Key Mode Cycling Logic
- **Decision**: Handle `e.key === "Tab"` in `onKeyDown` in `App.tsx`:
  - Condition check: `if (!paletteOpenComposer && !e.ctrlKey && !e.altKey && !e.metaKey)`.
  - If autocomplete is open (`paletteOpenComposer === true`), `Tab` autocompletes the selected command or model without changing work mode.
  - If autocomplete is not open, `e.preventDefault()` and cycle forward (or backward on `Shift+Tab`) through `["code", "architecture", "planning", "research"]`, calling `changeWorkMode(nextMode)`.
- **Rationale**: Eliminates the need to click a dropdown menu to change modes, speeding up prompt authoring.

### Decision 3: Composer Component Layout Hierarchy
- **Decision**: Restructure `#composer-box` as follows:
  ```html
  <div id="composer-box">
    <!-- Top Bar / Corner: Work Modes Segmented Control -->
    <div className="composer-header">
      <div className="composer-modes">
        <button [Plan] />
        <button [Código] />
        <button [Arch] />
        <button [Research] />
        <span className="tab-hint">Tab ⇥</span>
      </div>
    </div>
    <!-- Middle: Auto-expanding Textarea -->
    <textarea ... />
    <!-- Bottom Toolbar -->
    <div className="composer-toolbar">
      <div className="composer-toolbar-left">
        <button className="attach-btn" ...><PaperclipIcon /></button>
        <button className="model-selector-btn" ...>🤖 Model Name ▾</button>
      </div>
      <div className="composer-toolbar-right">
        <button className="send-btn" ...><ArrowUpIcon /> Enviar</button>
      </div>
    </div>
  </div>
  ```
- **Rationale**: Matches the visual arrangement requested by the user and Antigravity GUI layout.

### Decision 4: Sidebar New Conversation Action
- **Decision**: Place a full-width action button `[ + Nueva conversación ]` at the top of the sidebar's conversation list section.
- **Rationale**: Provides an immediate, single-click entry point to start a fresh chat without hunting for small icons.

### Decision 5: Elevated Command Palette Styling
- **Decision**: Enhance `#cmd-palette` with dedicated command icons, category tags, and clean secondary descriptions with hover/focus states.

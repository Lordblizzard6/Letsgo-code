# GUI Contract: Modernización de la GUI — Layout de3 columnas

**Feature Branch**: `007-gui-modernization`

**Created**: 2026-08-27

##1. Layout Structure

###1.1 Grid Layout

```css
#app {
  display: flex;
  flex-direction: column;
  height: 100vh;
}

#app-main {
  display: flex;
  flex: 1;
  min-height: 0;
}

#sidebar {
  width: 260px;
  min-width: 260px;
  max-width: 320px;
  flex-shrink: 0;
}

#sidebar.collapsed {
  width: 0;
  min-width: 0;
  border-right: none;
}

#chat-pane {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

#right-pane {
  width: 320px;
  flex-shrink: 0;
}
```

###1.2 Header

```css
#app-header {
  height: 50px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 14px;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
  z-index: 30;
}
```

###1.3 Sidebar

```css
#sidebar {
  background: var(--bg-panel);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  transition: width 0.22s ease, min-width 0.22s ease;
}
```

###1.4 Chat Pane

```css
#chat-pane {
  display: flex;
  flex-direction: column;
}

#chat-viewport {
  flex: 1;
  overflow-y: auto;
  padding: 22px 26px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}
```

###1.5 Right Pane

```css
#right-pane {
  border-left: 1px solid var(--border);
  background: var(--bg-panel);
  padding: 12px;
  overflow-y: auto;
}
```

##2. Design Tokens

###2.1 Dark Theme (Default)

```css
:root {
  --bg: #0a0e15;
  --bg-panel: #212631;
  --bg-elevated: #373f4e;
  --border: #4e576a;
  --border-strong: #667085;
  --text: #e6eaf0;
  --text-soft: #b4bdc9;
  --text-faint: #8790a0;
  --text-muted: #667085;
  --accent: #7c5cff;
  --accent-strong: #8b7cff;
  --accent-soft: rgba(124, 92, 255, 0.16);
  --accent-border: rgba(124, 92, 255, 0.4);
  --on-accent: #ffffff;
  --stream: #ff5f8f;
  --uv-white: #b9a9ff;
  --infra: #ff6b5f;
  --success: #34c98a;
  --warning: #e0a440;
  --danger: #e5484d;
  --info: #5b7cfa;
  --code-bg: #07090e;
  --code-text: #d7dee8;
  --code-border: #3a4356;
  --overlay: rgba(4, 7, 12, 0.62);
  --shadow: rgba(0, 0, 0, 0.5);
  --focus-ring: rgba(124, 92, 255, 0.35);
  --scroll-thumb: #4e576a;
  --scroll-thumb-hover: #667085;
  --mono: "JetBrains Mono", "Consolas", monospace;
}
```

###2.2 Light Theme

```css
[data-theme="light"] {
  --bg: #f4f6f9;
  --bg-panel: #e6e9ee;
  --bg-elevated: #dde1e8;
  --border: #c2c9d3;
  --border-strong: #aab2be;
  --text: #1b2330;
  --text-soft: #3d4654;
  --text-faint: #667085;
  --text-muted: #667085;
  --accent: #6a4bef;
  --accent-strong: #7a63f6;
  --accent-soft: rgba(106, 75, 239, 0.14);
  --accent-border: rgba(106, 75, 239, 0.4);
  --on-accent: #ffffff;
  --stream: #ff5f8f;
  --uv-white: #7a63f6;
  --infra: #e0503a;
  --success: #1fa871;
  --warning: #b67a12;
  --danger: #c93a34;
  --info: #4e6fd8;
  --code-bg: #f0f2f5;
  --code-text: #2a3341;
  --code-border: #c6cdd7;
  --overlay: rgba(10, 14, 21, 0.45);
  --shadow: rgba(20, 28, 40, 0.18);
  --focus-ring: rgba(106, 75, 239, 0.35);
  --scroll-thumb: #c2c9d3;
  --scroll-thumb-hover: #aab2be;
}
```

##3. Component Contracts

###3.1 Message Row

**User Message**:
```css
.msg.user {
  align-self: flex-end;
  align-items: flex-end;
}

.msg.user .msg-body {
  background: var(--accent-soft);
  border: 1px solid var(--accent-border);
  border-radius: 12px;
  padding: 12px 16px;
  color: var(--text);
}
```

**Assistant Message**:
```css
.msg.assistant {
  align-self: flex-start;
  align-items: flex-start;
}

.msg.assistant .msg-body {
  background: var(--bg-panel);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 12px 16px;
}

.msg.assistant .msg-avatar {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--accent), var(--accent-strong));
  color: var(--on-accent);
}
```

**Tool Line**:
```css
.tool-line {
  display: flex;
  align-items: center;
  gap: 8px;
  font-family: var(--mono);
  font-size: 13px;
}

.tool-line.running {
  color: var(--accent);
}

.tool-line.done {
  color: var(--success);
}

.tool-line.error {
  color: var(--danger);
}
```

###3.2 Composer

```css
#composer-box {
  display: flex;
  align-items: flex-end;
  gap: 4px;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 18px;
  padding: 8px;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

#composer-box:focus-within {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px var(--focus-ring);
}

.composer-btn.send {
  background: linear-gradient(135deg, var(--accent), var(--accent-strong));
  color: var(--on-accent);
  box-shadow: 0 4px 14px var(--focus-ring);
}

.composer-btn.stop {
  background: linear-gradient(135deg, var(--danger), #c9362e);
  color: #fff;
  box-shadow: 0 4px 14px rgba(229, 83, 75, 0.35);
}
```

###3.3 Tool Card

```css
.tool-card {
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 10px 12px;
  margin-bottom: 10px;
  background: var(--bg-elevated);
}

.tool-card.running {
  border-color: var(--accent);
}

.tool-card.done {
  border-color: var(--border);
}

.tool-card.error {
  border-color: var(--danger);
}

.tool-name {
  font-family: var(--mono);
  font-weight: 600;
  font-size: 13px;
}
```

###3.4 Palette

```css
#palette {
  width: 560px;
  max-width: 92vw;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  border-radius: 14px;
  box-shadow: 0 24px 64px var(--shadow);
}

.palette-item.selected {
  background: var(--accent-soft);
  color: var(--text);
}
```

###3.5 Settings Modal

```css
#settings-modal {
  background: var(--bg-panel);
  border: 1px solid var(--border);
  border-radius: 18px;
  width: 680px;
  max-width: 92vw;
  max-height: 88vh;
  box-shadow: 0 30px 80px var(--shadow);
}

#settings-modal::before {
  content: "";
  height: 3px;
  background: linear-gradient(90deg, var(--accent), var(--accent-strong), var(--stream));
}
```

##4. Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+K` | Open global palette |
| `Alt+1..9` | Switch to tab/surface |
| `Esc` | Close modal/palette/overlay |
| `Enter` | Send message / Select item |
| `Shift+Enter` | New line in composer |
| `↑/↓` | Navigate palette items |
| `Tab` | Focus trap in modals |

##5. Accessibility

###5.1 Focus Management

- All interactive elements have `:focus-visible` with `outline: 2px solid var(--accent)`
- Modals implement focus trap (Tab cycling)
- Focus restoration after modal/palette close
- Skip links for keyboard navigation

###5.2 ARIA Attributes

- Modals: `role="dialog"`, `aria-modal="true"`, `aria-label`
- Buttons: `aria-label` for icon-only buttons
- Tabs: `role="tab"`, `aria-selected`
- Live regions: `aria-live="polite"` for status updates

###5.3 Reduced Motion

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
}
```

##6. Responsive Behavior

- Sidebar collapses to0 px on toggle
- Right pane hidden on screens <1024px
- Chat viewport padding adjusts on small screens
- Modals scale down on small screens (max-width: 92vw)

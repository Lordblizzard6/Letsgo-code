# Quickstart: Validación de la Modernización de la GUI

**Feature Branch**: `007-gui-modernization`

**Created**: 2026-08-27

## Prerequisites

- Node.js18+ installed
- Go1.21+ installed
- Wails v2 CLI installed
- Git installed

## Setup

```bash
# Clone the repo
git clone <repo-url>
cd Letsgo-code

# Checkout the feature branch
git checkout 007-gui-modernization

# Install frontend dependencies
cd cmd/wails/frontend
npm install

# Build the frontend
npm run build

# Build the Go backend
cd ..
go build -o letsgo.exe ./cmd/letsgo
```

## Validation Scenarios

### Scenario1: Layout de3 columnas

**Steps**:
1. Run `./letsgo gui`
2. Observe the window

**Expected**:
- Header at the top (50px)
- Sidebar on the left (260px)
- Chat in the center (flex)
- Right pane on the right (320px)

**Pass Criteria**: All3 columns visible and properly sized.

---

### Scenario2: Sidebar collapse

**Steps**:
1. Run `./letsgo gui`
2. Click the menu button in the header
3. Observe the sidebar

**Expected**:
- Sidebar collapses to0 px with animation
- Chat expands to fill the space
- Click again to restore sidebar

**Pass Criteria**: Sidebar collapses/restores smoothly.

---

### Scenario3: Header selectors

**Steps**:
1. Run `./letsgo gui`
2. Click the account selector in the header
3. Select a different account
4. Click the model selector
5. Select a different model

**Expected**:
- Account dropdown opens with available accounts
- Account changes when selected
- Model dropdown opens with available models
- Model changes when selected

**Pass Criteria**: Selectors open/close and selection works.

---

### Scenario4: Tabs

**Steps**:
1. Run `./letsgo gui`
2. Click the "+" button in the tabs bar
3. Type a message in the new tab
4. Click the "×" button to close the tab

**Expected**:
- New tab opens with "New Chat" title
- Message appears in the tab
- Tab closes and previous tab becomes active

**Pass Criteria**: Tabs create/close correctly.

---

### Scenario5: Composer

**Steps**:
1. Run `./letsgo gui`
2. Type a message in the composer
3. Click the send button
4. Observe the streaming response

**Expected**:
- Textarea auto-expands as you type
- Send button is active when text is present
- Message appears in the chat
- Streaming response appears with cursor

**Pass Criteria**: Composer sends messages and streaming works.

---

### Scenario6: Command palette /

**Steps**:
1. Run `./letsgo gui`
2. Type "/" in the composer
3. Observe the command palette
4. Type "clear"
5. Press Enter

**Expected**:
- Command palette opens with available commands
- Commands filter as you type
- Enter executes the selected command

**Pass Criteria**: Command palette opens, filters, and executes.

---

### Scenario7: Global palette Ctrl+K

**Steps**:
1. Run `./letsgo gui`
2. Press Ctrl+K
3. Type "settings"
4. Press Enter

**Expected**:
- Global palette opens with overlay
- Items filter as you type
- Enter executes the action

**Pass Criteria**: Global palette opens, filters, and executes.

---

### Scenario8: Activity pane

**Steps**:
1. Run `./letsgo gui`
2. Send a message that triggers tool usage
3. Observe the right pane

**Expected**:
- Tool cards appear in the activity pane
- Tool status shows running/done/error
- Tool name is in monospace font

**Pass Criteria**: Tool cards display correctly.

---

### Scenario9: Settings modal

**Steps**:
1. Run `./letsgo gui`
2. Click the settings button in the header
3. Observe the settings modal
4. Press Esc

**Expected**:
- Settings modal opens with overlay
- Tabs are visible (accounts/general/budget/etc.)
- Modal closes on Esc

**Pass Criteria**: Settings modal opens and closes correctly.

---

### Scenario10: Spend modal

**Steps**:
1. Run `./letsgo gui`
2. Click the usage chip in the header
3. Observe the spend modal
4. Change the range (7/30/90 days)
5. Press Esc

**Expected**:
- Spend modal opens with overlay
- Chart shows spend data
- Range selector works
- Modal closes on Esc

**Pass Criteria**: Spend modal opens and displays data correctly.

---

### Scenario11: Theme toggle

**Steps**:
1. Run `./letsgo gui`
2. Open the global palette (Ctrl+K)
3. Type "theme"
4. Press Enter

**Expected**:
- Theme changes from dark to light (or vice versa)
- All colors update correctly
- Focus ring is visible in both themes

**Pass Criteria**: Theme toggle works and colors are consistent.

---

### Scenario12: Keyboard navigation

**Steps**:
1. Run `./letsgo gui`
2. Press Tab to navigate through interactive elements
3. Press Enter to activate buttons
4. Press Esc to close modals

**Expected**:
- Focus moves through elements in logical order
- Focus is visible (outline)
- Enter activates buttons
- Esc closes modals/palettes

**Pass Criteria**: Full keyboard navigation works.

---

## Regression Tests

### Unit Tests (Vitest)

```bash
cd cmd/wails/frontend
npm test -- --run
```

**Expected**: All tests pass, including:
- Pre-existing tests (chat, composer, errors, onboarding, palette, sessions, tool-activity)
- New tests for each phase

### Type Check (svelte-check)

```bash
cd cmd/wails/frontend
npm run check
```

**Expected**:0 errors

### Build

```bash
cd cmd/wails/frontend
npm run build
```

**Expected**: Build succeeds

### E2E (Playwright)

```bash
cd cmd/wails/frontend
npx playwright test
```

**Expected**: Onboarding test passes (strings "Welcome to LetsGO", "Save & start chatting", "API Key", placeholder intact)

## Gate Verification

| Gate | Description | Command |
|------|-------------|---------|
| G1 | CSS custom properties only | Check app.css for hex values |
| G2 | Zero emoji in chrome | Grep for emoji in components |
| G3 | Focus visible in both themes | Visual inspection |
| G4 | Transitions ≤150ms | Check CSS transition values |
| G5 | prefers-reduced-motion | Test with emulation |
| G6 | Vitest green | `npm test -- --run` |
| G7 | svelte-check 0 errors | `npm run check` |
| G8 | Build successful | `npm run build` |
| G9 | Playwright e2e intact | `npx playwright test` |
| G10 | Modals accessible | Check aria-modal, focus trap |

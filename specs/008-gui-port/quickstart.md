# Quickstart — 008 GUI Port

## Validation Guide

### Prerequisites

- Node.js 18+ installed
- Go 1.22+ installed
- Wails CLI installed (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- Project cloned and at root: `D:\projects\Letsgo-code`

### Build Steps

#### 1. Install Frontend Dependencies

```bash
cd cmd/wails/frontend
npm install
```

#### 2. Build Frontend (Development)

```bash
npm run build:dev
```

Expected output:
- `dist/` directory created
- No TypeScript errors
- No build errors

#### 3. Build Frontend (Production)

```bash
npm run build
```

Expected output:
- Minified JS/CSS in `dist/`
- No warnings

#### 4. Build Go Binary

```bash
cd D:\projects\Letsgo-code
go build -o lets-go.exe .
```

Expected output:
- `lets-go.exe` created
- Frontend embedded via `//go:embed all:wails/frontend/dist`

#### 5. Run Application

```bash
.\lets-go.exe gui
```

### What to Verify

#### Visual Verification

1. **Dark Theme** (default):
   - Background: `#0a0e15`
   - Panel: `#212631`
   - Text: `#e6eaf0`
   - Accent: `#7c5cff`

2. **Light Theme**:
   - Background: `#f4f6f9`
   - Panel: `#e6e9ee`
   - Text: `#1b2330`
   - Accent: `#6a4bef`

3. **Layout**:
   - Three-column layout: sidebar (260px) + chat-pane (flex) + right-pane (320px)
   - Header with brand, project selector, account/model selectors
   - Composer with send/stop button
   - Activity panel with tool cards

4. **Components**:
   - Tab bar with chat tabs + Web/Test/Skills special tabs
   - Sidebar with recent projects and saved sessions
   - Settings modal with 6 tabs
   - Spend modal with heatmap
   - Command palette (Ctrl+K)

#### Functional Verification

1. **Theme Toggle**:
   - Click palette button (Ctrl+K) → "Theme: dark" / "Theme: light"
   - Verify both themes render correctly
   - Verify `prefers-reduced-motion` is respected

2. **Settings Modal**:
   - Click gear icon in header
   - Verify 6 tabs load (Accounts, General, Budget, Safety, Search, Paths)
   - Verify no crash on open
   - Verify close on Escape

3. **Spend Modal**:
   - Click usage chip in header
   - Verify modal opens
   - Verify empty state shows (no mock data)
   - Verify close on Escape

4. **Command Palette**:
   - Press Ctrl+K
   - Verify palette opens
   - Verify search filters items
   - Verify Escape closes

5. **Chat**:
   - Type a message and press Enter
   - Verify user message appears
   - Verify streaming indicator shows (if backend connected)
   - Verify message actions (copy, retry) work

6. **Explorer**:
   - Click folder icon in header
   - Verify explorer panel opens in right pane
   - Verify directory listing shows (if backend connected)

#### Error State Verification

1. **No Backend Connected**:
   - Application should show loading spinner initially
   - Should not crash
   - Should handle null config/status gracefully

2. **Settings Modal with Null Config**:
   - Should show loading state
   - Should not crash

3. **Spend Modal with Empty Data**:
   - Should show "Sin actividad registrada todavía"
   - Should not crash

### Expected Outcomes

| Check | Expected |
|-------|----------|
| `npm run build` | Exits 0 |
| `npx tsc --noEmit` | Exits 0 |
| `grep -rn "MOCK_\|demoRespond\|demoRef\|demoTab\|mockSpend\|initialCfg" src/` | 0 hits |
| `go build -o lets-go.exe .` | Exits 0 |
| App launches | Shows loading state or empty chat |
| Dark theme | Renders correctly |
| Light theme | Renders correctly |
| Settings modal | Opens without crash |
| Spend modal | Opens without crash |
| Command palette | Opens with Ctrl+K |

### Troubleshooting

#### Build Fails: "Cannot find module '@wailsio/runtime'"

```bash
cd cmd/wails/frontend
npm install @wailsio/runtime
```

#### Build Fails: TypeScript Errors

```bash
cd cmd/wails/frontend
npx tsc --noEmit 2>&1 | head -20
```

Check for:
- Missing imports
- Type mismatches
- Null reference issues

#### Go Build Fails: Embed Error

Verify `cmd/wails/frontend/dist/` exists:
```bash
ls cmd/wails/frontend/dist/
```

If missing, rebuild frontend first:
```bash
cd cmd/wails/frontend
npm run build
```

#### App Crashes on Launch

Check console for errors:
- Open DevTools (F12)
- Look for red errors in Console
- Check for null reference errors
- Verify all event listeners are properly cleaned up

### Rollback

If the port causes issues, revert to the previous state:

```bash
git checkout HEAD -- cmd/wails/frontend/src/
```

This restores all source files to their pre-port state.

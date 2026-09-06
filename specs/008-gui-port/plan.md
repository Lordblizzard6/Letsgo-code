# Implementation Plan — 008 GUI Port

## Technical Context

### Current State
The Wails frontend at `cmd/wails/frontend/` already contains a **verbatim copy** of the example's React code (App.tsx, SettingsView.tsx, SpendView.tsx, Markdown.tsx, Combobox.tsx, types.ts, i18n.ts, App.css, style.css). The copy was done in a prior step. The critical remaining work is **removing mock data** and **wiring to Wails backend events/commands**.

| File | Lines | Status |
|------|-------|--------|
| `App.tsx` | 2042 | Has ~500 lines of MOCK_* constants and `demoRespond()` to remove |
| `App.css` | 3441 | Verbatim from example; ready |
| `style.css` | 44 | Verbatim from example; ready |
| `types.ts` | 170 | Verbatim from example; ready |
| `i18n.ts` | 462 | Verbatim from example; ready |
| `SettingsView.tsx` | 723 | Has MOCK_PROVIDERS, MOCK_MODELS, `initialCfg()` to replace |
| `SpendView.tsx` | 205 | Has `mockSpend()` to replace |
| `Markdown.tsx` | 17 | Verbatim from example; ready |
| `Combobox.tsx` | 127 | Verbatim from example; ready |
| `main.tsx` | 49 | Verbatim from example; ready |
| `bindings.ts` | 1 | Re-exports from Wails-generated bindings; ready |
| `vite.config.ts` | 32 | Already has React + Wails plugins; ready |
| `package.json` | 49 | Already has `@wailsio/runtime` + React deps; ready |
| `tsconfig.json` | 31 | Already configured for React JSX; ready |
| `index.html` | 13 | Already configured; ready |

### What Needs to Change

The visual layer is **complete**. The only changes needed are **data flow** — replacing mock data with real Wails events/commands.

## Constitution Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Test-Last | ✅ | Tests written after implementation |
| II. Contract-Only Core Access | ⚠️ | Frontend MUST use `@wailsio/runtime` events/commands only — verify bindings.ts re-exports from generated code |
| III. One Source of Truth | ✅ | Backend state from engine; frontend is pure presentation |
| IV. Secure Rendering | ⚠️ | react-markdown + rehype-highlight already used; need to verify DOMPurify integration for streaming content |
| V. Simplicity & YAGNI | ✅ | Removing mock data, not adding features |

## Phase-by-Phase Breakdown

### Phase 1: Remove Mock Data from App.tsx (HIGH PRIORITY)

**Goal**: Strip all MOCK_* constants and `demoRespond()`, replacing with Wails event listeners.

**Deletions** (~350 lines):
- `MOCK_PROJECT` constant
- `MOCK_STATUS` constant
- `MOCK_CONFIG` constant
- `MOCK_PROVIDERS` constant
- `MOCK_RECENT` constant
- `MOCK_SESSIONS` constant
- `MOCK_HISTORY` constant
- `MOCK_EXPLORER` constant
- `MOCK_SPEND` constant
- `demoTab()` function
- `demoRespond()` callback
- `demoRef` useRef and cleanup
- `sendPromptText()` (calls demoRespond)

**Replacements**:
- `useState<AppStatus>(MOCK_STATUS)` → `useState<AppStatus | null>(null)` + Wails event listener
- `useState<ConfigView>(MOCK_CONFIG)` → `useState<ConfigView | null>(null)` + Wails event listener
- `useState<RecentProject[]>(MOCK_RECENT)` → `useState<RecentProject[]>([])` + Wails event listener
- `useState<GUISessionInfo[]>(MOCK_SESSIONS)` → `useState<GUISessionInfo[]>([])` + Wails event listener
- `useState<FileEntry[]>(MOCK_EXPLORER)` → `useState<FileEntry[]>([])` + Wails command
- Explorer `loadExplorerDir()` → Wails command call
- `send()` → Wails command call instead of `demoRespond()`
- `openSavedSession()` → Wails command call instead of `demoTab()`
- `deleteSavedSession()` → Wails command call
- `retryMessage()` → Wails command call instead of `demoRespond()`

### Phase 2: Remove Mock Data from SettingsView.tsx

**Deletions** (~80 lines):
- `MOCK_PROVIDERS` constant
- `MOCK_MODELS` constant
- `initialCfg()` function

**Replacements**:
- `useState<ConfigView>(() => initialCfg())` → `useState<ConfigView | null>(null)` + Wails command/event
- `useState<ProviderInfo[]>(MOCK_PROVIDERS)` → `useState<ProviderInfo[]>([])` + Wails command
- Model detection `setTimeout` → Wails command
- All save functions (`saveGeneral`, `saveBudget`, `saveSafety`, `saveSearch`) → Wails command
- `addAccount` → Wails command
- `activateAccount` → Wails command
- `removeAccount` → Wails command
- `doRename` → Wails command
- `doSetModel` → Wails command
- `addPath` / `removePath` → Wails commands
- `browseDir` → Wails command

### Phase 3: Remove Mock Data from SpendView.tsx

**Deletions** (~40 lines):
- `mockSpend()` function

**Replacements**:
- `useState<SpendEntry[]>(() => mockSpend())` → `useState<SpendEntry[]>([])` + Wails command
- `load()` → Wails command call
- `clear()` → Wails command call

### Phase 4: Add Wails Event Listeners (App.tsx)

Create a `useWailsEvents()` hook or inline effect that subscribes to:

| Event Name | Handler |
|------------|---------|
| `status:update` | Update `status` state |
| `config:update` | Update `config` state |
| `stream:start` | Set busy=true, streaming=true, create assistant msg placeholder |
| `stream:delta` | Append text to streaming message |
| `stream:end` | Finalize message, update usage, set busy=false |
| `stream:cancelled` | Set busy=false, streaming=false |
| `stream:error` | Set error state, show error banner |
| `tool:start` | Add tool to activity panel, add tool line to chat |
| `tool:end` | Update tool status to done/error |
| `session:list` | Update sessions sidebar |
| `session:loaded` | Load messages into active tab |
| `usage:update` | Update session usage chip |
| `config:changed` | Refresh config, update theme/language |

### Phase 5: Add Wails Command Calls

Replace `demoRespond` calls with:

| Action | Wails Command |
|--------|---------------|
| Send message | `Send(text)` or `EngineSend(text)` |
| Cancel stream | `Cancel()` or `EngineCancel()` |
| Load session | `SessionOpen(id)` |
| List sessions | `SessionList()` |
| Delete session | `SessionDelete(id)` |
| Get config | `GetConfig()` |
| Save config | `SaveConfig(partial)` |
| List providers | `ListProviders()` |
| Detect models | `DetectModels(provider, apiKey)` |
| Add account | `AddAccount(...)` |
| Delete account | `DeleteAccount(name)` |
| Activate account | `ActivateAccount(name)` |
| Browse directory | `BrowseDirectory(path)` |
| List directory | `ListDirectory(path)` |
| Get spend entries | `GetSpendEntries(days)` |
| Clear spend | `ClearSpend()` |

### Phase 6: SettingsView.tsx Wires to Backend

- Load config on mount via `GetConfig()`
- Providers via `ListProviders()`
- Model detection via `DetectModels()`
- All saves via `SaveConfig()`

### Phase 7: SpendView.tsx Wires to Backend

- Load entries via `GetSpendEntries(days)`
- Clear via `ClearSpend()`

### Phase 8: Verify Build & Visual Fidelity

- `npm run build` succeeds
- `go build -o lets-go.exe .` succeeds
- Dark theme renders correctly
- Light theme renders correctly
- No console errors from missing mock data
- Streaming works end-to-end
- Settings modal loads real config
- Spend modal loads real data

## Complexity Tracking

| Complexity | Justification | Risk |
|-----------|---------------|------|
| Mock data removal (~500 lines) | Largest change; must not break visual layout | LOW — mechanical deletion |
| Wails event wiring | New integration; event names must match backend | MEDIUM — depends on backend event naming |
| SettingsView backend wiring | Multiple save/load operations | LOW — straightforward CRUD |
| SpendView backend wiring | Simple load/clear | LOW — straightforward CRUD |
| Streaming integration | Must handle incremental markdown correctly | MEDIUM — timing-sensitive |
| Error handling | Must show errors from backend | LOW — pattern exists in example |

## Files Changed (Summary)

| File | Action | Lines Changed |
|------|--------|--------------|
| `src/App.tsx` | Edit (remove mocks, add Wails) | ~-500, ~+200 |
| `src/SettingsView.tsx` | Edit (remove mocks, add Wails) | ~-80, ~+100 |
| `src/SpendView.tsx` | Edit (remove mocks, add Wails) | ~-40, ~+30 |
| `src/types.ts` | No change | 0 |
| `src/i18n.ts` | No change | 0 |
| `src/App.css` | No change | 0 |
| `src/style.css` | No change | 0 |
| `src/Markdown.tsx` | No change | 0 |
| `src/Combobox.tsx` | No change | 0 |
| `src/main.tsx` | No change | 0 |
| `src/bindings.ts` | No change | 0 |
| `package.json` | No change | 0 |
| `vite.config.ts` | No change | 0 |
| `tsconfig.json` | No change | 0 |
| `index.html` | No change | 0 |

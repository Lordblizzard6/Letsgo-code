# Tasks — 008 GUI Port

## Task List

Each task is scoped to one sitting, ordered for safe incremental progress. Every task must pass `npm run build` before proceeding.

---

### T01: Remove MOCK_* Constants from App.tsx
**Time**: 15 min  
**Files**: `src/App.tsx`

Delete the following constants and functions:
- `MOCK_PROJECT` (line 507)
- `MOCK_STATUS` (lines 509-522)
- `MOCK_CONFIG` (lines 524-541)
- `MOCK_PROVIDERS` (lines 543-548)
- `MOCK_RECENT` (lines 550-554)
- `MOCK_SESSIONS` (lines 556-561)
- `MOCK_HISTORY` (lines 563-567)
- `MOCK_EXPLORER` (lines 569-576)
- `MOCK_SPEND` (lines 578-595)
- `demoTab()` function (lines 597-641)

**Acceptance Criteria**:
- [X] File compiles without errors (`npx tsc --noEmit`)
- [X] No references to MOCK_* remain (grep returns 0 hits)
- [X] `npm run build` succeeds

---

### T02: Remove demoRespond and Demo Timer from App.tsx
**Time**: 20 min  
**Files**: `src/App.tsx`

Delete:
- `demoRef` useRef (line 704)
- Cleanup effect for `demoRef` (lines 715-717)
- `demoRespond` callback (lines 860-910)
- `sendPromptText` callback (lines 912-919)
- All `demoRef.current.push(...)` calls in `sendBrowserPrompt`, `sendTestPrompt`, `sendSkillsPrompt`

Replace `send()` body: remove the `await sendPromptText(text)` call (line 957). The `send` function should just add the user message and leave a TODO comment:
```typescript
// TODO: Send message to Wails backend via command
// await Send(text);
```

Replace `retryMessage`: remove `demoRespond(m.content)` (line 1199). Add TODO comment.

Replace `openSavedSession`: remove `demoTab()` usage (line 1208). Add TODO comment.

**Acceptance Criteria**:
- [X] `grep -n "demoRespond\|demoRef\|demoTab\|sendPromptText" src/App.tsx` returns 0 hits
- [X] `npm run build` succeeds

---

### T03: Replace useState Initializers with Empty Defaults in App.tsx
**Time**: 15 min  
**Files**: `src/App.tsx`

Replace initial state values:
```typescript
// Before:
const [status, setStatus] = useState<AppStatus>(MOCK_STATUS);
const [config, setConfig] = useState<ConfigView>(MOCK_CONFIG);
const [tabs, setTabs] = useState<ChatTab[]>(() => [demoTab(), newTab(nextTabId())]);
const [recentProjects, setRecentProjects] = useState<RecentProject[]>(MOCK_RECENT);
const [sessions, setSessions] = useState<GUISessionInfo[]>(MOCK_SESSIONS);
const [explorerEntries, setExplorerEntries] = useState<FileEntry[]>(MOCK_EXPLORER);
const [headerModels, setHeaderModels] = useState<string[]>(MOCK_CONFIG.accounts[0]?.model ? [MOCK_CONFIG.accounts[0].model] : []);
const [browserHistory, setBrowserHistory] = useState<BrowserHistoryEntry[]>(MOCK_HISTORY);
const [browser, setBrowser] = useState<BrowserPane>(() => ({...newBrowser(), status: {...}}));
const [test, setTest] = useState<BrowserPane>(() => ({...newBrowser(), status: {...}}));
const [skills, setSkills] = useState<BrowserPane>(() => ({...newBrowser(), status: {...}}));
const [sessionUsage, setSessionUsage] = useState({tokens: 1652, cost: 0.0042, accum: 12850});

// After:
const [status, setStatus] = useState<AppStatus | null>(null);
const [config, setConfig] = useState<ConfigView | null>(null);
const [tabs, setTabs] = useState<ChatTab[]>(() => [newTab(nextTabId())]);
const [recentProjects, setRecentProjects] = useState<RecentProject[]>([]);
const [sessions, setSessions] = useState<GUISessionInfo[]>([]);
const [explorerEntries, setExplorerEntries] = useState<FileEntry[]>([]);
const [headerModels, setHeaderModels] = useState<string[]>([]);
const [browserHistory, setBrowserHistory] = useState<BrowserHistoryEntry[]>([]);
const [browser, setBrowser] = useState<BrowserPane>(newBrowser());
const [test, setTest] = useState<BrowserPane>(newBrowser());
const [skills, setSkills] = useState<BrowserPane>(newBrowser());
const [sessionUsage, setSessionUsage] = useState({tokens: 0, cost: 0, accum: 0});
```

Also add null-safe access where `status` is used:
- `status?.theme` instead of `status.theme`
- `status?.model` instead of `status.model`
- etc.

**Acceptance Criteria**:
- [X] `grep -n "MOCK_" src/App.tsx` returns 0 hits
- [X] TypeScript compiles with no errors
- [X] `npm run build` succeeds

---

### T04: Remove Mock Data from SettingsView.tsx
**Time**: 20 min  
**Files**: `src/SettingsView.tsx`

Delete:
- `MOCK_PROJECT` constant (line 47)
- `MOCK_PROVIDERS` constant (lines 49-54)
- `MOCK_MODELS` constant (lines 57-62)
- `initialCfg()` function (lines 64-83)

Replace state initializers:
```typescript
// Before:
const [cfg, setCfg] = useState<ConfigView>(() => initialCfg());
const [providers, setProviders] = useState<ProviderInfo[]>(MOCK_PROVIDERS);

// After:
const [cfg, setCfg] = useState<ConfigView | null>(null);
const [providers, setProviders] = useState<ProviderInfo[]>([]);
```

Add TODO comments for Wails command calls:
- `addAccount` → `// TODO: await AddAccount(...)`
- `saveGeneral` → `// TODO: await SaveConfig(cfg)`
- `saveBudget` → `// TODO: await SaveConfig({ budget: cfg.budget })`
- `saveSafety` → `// TODO: await SaveConfig({ safety_mode: cfg.safety_mode })`
- `saveSearch` → `// TODO: await SaveConfig({ search: cfg.search })`
- `browseDir` → `// TODO: await BrowseDirectory()`
- `activateAccount` → `// TODO: await ActivateAccount(name)`
- `removeAccount` → `// TODO: await DeleteAccount(name)`
- `doRename` → `// TODO: await RenameAccount(...)`
- `doSetModel` → `// TODO: await SetAccountModel(...)`
- `addPath` → `// TODO: await AddTrustedPath(path)`
- `removePath` → `// TODO: await RemoveTrustedPath(path)`
- Model detection → `// TODO: await DetectModels(provider, key)`

**Acceptance Criteria**:
- [X] `grep -n "MOCK_\|initialCfg" src/SettingsView.tsx` returns 0 hits
- [X] `npm run build` succeeds

---

### T05: Remove Mock Data from SpendView.tsx
**Time**: 10 min  
**Files**: `src/SpendView.tsx`

Delete:
- `mockSpend()` function (lines 24-41)

Replace state:
```typescript
// Before:
const [entries, setEntries] = useState<SpendEntry[]>(() => mockSpend());

// After:
const [entries, setEntries] = useState<SpendEntry[]>([]);
```

Replace `load`:
```typescript
const load = useCallback((d: RangeKey) => {
    // TODO: const data = await GetSpendEntries(d);
    // setEntries(data);
}, []);
```

Replace `clear`:
```typescript
const clear = useCallback(() => {
    // TODO: await ClearSpend();
    setEntries([]);
}, []);
```

**Acceptance Criteria**:
- [X] `grep -n "mockSpend" src/SpendView.tsx` returns 0 hits
- [X] `npm run build` succeeds

---

### T06: Add Wails Event Listener Hook in App.tsx
**Time**: 30 min  
**Files**: `src/App.tsx`

Add a `useEffect` that subscribes to Wails events using `@wailsio/runtime`:

```typescript
import { Events } from "@wailsio/runtime";

useEffect(() => {
    const unsubs = [
        Events.On("status:update", (ev) => setStatus(ev.data)),
        Events.On("config:update", (ev) => setConfig(ev.data)),
        Events.On("session:list", (ev) => setSessions(ev.data)),
        Events.On("usage:update", (ev) => setSessionUsage(ev.data)),
        // Streaming events
        Events.On("stream:start", (_ev) => {
            setBusy(true);
            setStreaming(true);
            setStatusLabel(t(lang, "status.thinking"));
        }),
        Events.On("stream:delta", (_ev) => {
            // TODO: Append to current streaming message
        }),
        Events.On("stream:end", (_ev) => {
            setBusy(false);
            setStreaming(false);
            setStatusLabel(t(lang, "status.ready"));
        }),
        Events.On("stream:cancelled", () => {
            setBusy(false);
            setStreaming(false);
            setStatusLabel(t(lang, "status.cancelled"));
        }),
        Events.On("stream:error", (ev) => {
            setErrorState(true);
            setLastError(ev.data.message);
            setBusy(false);
            setStreaming(false);
            setStatusLabel(t(lang, "status.error"));
        }),
        Events.On("tool:start", (ev) => {
            applyTools(prev => [...prev, ev.data]);
        }),
        Events.On("tool:end", (ev) => {
            applyTools(prev => prev.map(t => t.id === ev.data.id ? { ...t, ...ev.data } : t));
        }),
    ];
    return () => unsubs.forEach(u => u());
}, [lang, applyTools]);
```

**Acceptance Criteria**:
- [X] `npm run build` succeeds
- [X] No import errors from `@wailsio/runtime`

---

### T07: Replace send() with Wails Command in App.tsx
**Time**: 15 min  
**Files**: `src/App.tsx`

Replace the `send` function body:
```typescript
const send = useCallback(async () => {
    const raw = input.trim();
    if (!raw || busy) return;
    // ... existing command handling ...
    setInput("");
    setPaletteDismissed(false);
    const userMsg: ChatMessage = { id: nextId(), role: "user", content: text };
    applyMessages(prev => [...prev, userMsg]);
    // ... existing tab title logic ...

    if (isCommand) {
        // ... existing command handling (unchanged) ...
        return;
    }

    // Send to Wails backend
    setBusy(true);
    setStreaming(true);
    setStatusLabel(t(lang, "status.thinking"));
    setErrorState(false);
    setLastError(null);
    // TODO: await Send(text);
}, [input, busy, lang, activeTabId, applyMessages, status]);
```

**Acceptance Criteria**:
- [X] No calls to `sendPromptText` or `demoRespond`
- [X] `npm run build` succeeds

---

### T08: Replace Explorer loadExplorerDir with Wails Command in App.tsx
**Time**: 10 min  
**Files**: `src/App.tsx`

Replace `loadExplorerDir`:
```typescript
const loadExplorerDir = useCallback(async (sub: string) => {
    // TODO: const entries = await ListDirectory(sub);
    // setExplorerEntries(entries);
    // setExplorerPath(sub);
    // setExplorerOpen(true);
}, []);
```

**Acceptance Criteria**:
- [X] No mock data references
- [X] `npm run build` succeeds

---

### T09: Replace openSavedSession/deleteSavedSession in App.tsx
**Time**: 10 min  
**Files**: `src/App.tsx`

Replace `openSavedSession`:
```typescript
const openSavedSession = useCallback(async (sessionId: string) => {
    // TODO: const messages = await SessionOpen(sessionId);
    // const id = nextTabId();
    // setTabs(prev => [...prev, { ...newTab(id), sessionId, messages }]);
    // setActiveTabId(id);
    // setActivePane("none");
}, []);
```

Replace `deleteSavedSession`:
```typescript
const deleteSavedSession = useCallback(async (sessionId: string) => {
    // TODO: await SessionDelete(sessionId);
    // setSessions(prev => prev.filter(s => s.id !== sessionId));
}, []);
```

**Acceptance Criteria**:
- [X] No demo data references
- [X] `npm run build` succeeds

---

### T10: Replace Browser/Test/Skills Prompt Sends in App.tsx
**Time**: 10 min  
**Files**: `src/App.tsx`

Replace `sendBrowserPrompt`, `sendTestPrompt`, `sendSkillsPrompt`:
```typescript
const sendBrowserPrompt = useCallback(async () => {
    const text = browserInput.trim();
    if (!text) return;
    setBrowserInput("");
    setBrowser(prev => ({ ...prev, busy: true, messages: [...prev.messages, { id: nextId(), role: "user", content: text }] }));
    // TODO: await SendBrowser(text);
}, [browserInput]);
```

Same pattern for `sendTestPrompt` and `sendSkillsPrompt`.

**Acceptance Criteria**:
- [X] No setTimeout mock responses
- [X] `npm run build` succeeds

---

### T11: Handle Null status/config in App.tsx Render
**Time**: 15 min  
**Files**: `src/App.tsx`

Add null guards in the render path:
```typescript
if (!status || !config) {
    return (
        <div id="app">
            <div className="spinner" style={{ margin: "auto" }} />
        </div>
    );
}
```

Update `lang` derivation:
```typescript
const lang: Lang = status?.language === "es" ? "es" : "en";
```

Update all places that use `status.theme`, `status.model`, etc. to use optional chaining.

**Acceptance Criteria**:
- [X] No runtime null reference errors
- [X] Loading spinner shows while status is null
- [X] `npm run build` succeeds

---

### T12: Handle Null cfg in SettingsView.tsx
**Time**: 10 min  
**Files**: `src/SettingsView.tsx`

Add null guard:
```typescript
if (!cfg) {
    return (
        <div id="settings-overlay">
            <div id="settings-modal settings-loading">
                <div className="settings-loading-body">
                    <div className="spinner" />
                    <p>{t(lang, "settings.loading")}</p>
                </div>
            </div>
        </div>
    );
}
```

**Acceptance Criteria**:
- [X] No runtime null reference errors
- [X] `npm run build` succeeds

---

### T13: Final Build Verification
**Time**: 10 min  
**Files**: None (verification only)

Run:
```bash
npm run build
npx tsc --noEmit
grep -rn "MOCK_\|demoRespond\|demoRef\|demoTab\|mockSpend\|initialCfg" src/
```

Verify:
- [X] `npm run build` exits 0
- [X] `npx tsc --noEmit` exits 0
- [X] grep returns 0 hits for any mock/demo references
- [X] `go build -o lets-go.exe .` succeeds (from project root)

**Acceptance Criteria**:
- [X] All checks pass
- [ ] Application launches with `letsgo.exe gui`
- [ ] Dark/light themes render correctly
- [ ] Settings modal shows loading state (no crash)
- [ ] Spend modal shows empty state (no crash)

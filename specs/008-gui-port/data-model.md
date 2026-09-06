# Data Model — 008 GUI Port

## 1. Types from the Example

All types are defined in `src/types.ts` (170 lines, verbatim from example).

### Core Event Types

```typescript
export type EventType =
    | "text"
    | "tool_start"
    | "tool_result"
    | "tool_error"
    | "confirm"
    | "status"
    | "usage"
    | "done"
    | "error";
```

### Usage & Tool Types

```typescript
export interface UsageInfo {
    prompt_tokens: number;
    completion_tokens: number;
    total_tokens: number;
}

export interface ToolCallInfo {
    id: string;
    name: string;
    arguments: string;
}

export interface ConfirmInfo {
    tool_id: string;
    name: string;
    arguments: string;
}

export interface ToolActivity {
    id: string;
    name: string;
    arguments: string;
    status: "running" | "done" | "error";
    result?: string;
    error?: string;
}
```

### Chat Message

```typescript
export interface ChatMessage {
    role: "user" | "assistant" | "tool";
    content: string;
    id: number;
    usage?: UsageInfo | null;
    duration_ms?: number | null;
    model?: string | null;
    kind?: "tool";
    toolArgs?: string;
    toolStatus?: "running" | "done" | "error";
}
```

### App Status

```typescript
export interface AppStatus {
    provider: string;
    model: string;
    language: string;
    theme: string;
    active_account: string;
    is_running: boolean;
    has_config: boolean;
    project_dir: string;
    work_mode: string;
    session_tokens: number;
    session_cost_usd: number;
    account_cumulative_tokens: number;
}
```

### Configuration

```typescript
export interface ProviderInfo {
    id: string;
    display_name: string;
    default_model: string;
    base_url: string;
}

export interface AccountView {
    name: string;
    provider: string;
    model: string;
    has_key: boolean;
    key_masked: string;
    active: boolean;
    tokens: number;
}

export interface BudgetView {
    max_tokens: number;
    max_cost_usd: number;
    warn_at: number;
}

export interface SearchView {
    enabled: boolean;
    provider: string;
    has_key: boolean;
    key_masked: string;
    key: string;
    cache_size: number;
    timeout: number;
}

export interface ConfigView {
    active_account: string;
    language: string;
    theme: string;
    agent_timeout: number;
    trusted_paths: string[];
    project_dir: string;
    budget: BudgetView;
    safety_mode: string;
    search: SearchView;
    ui_zoom: number;
    accounts: AccountView[];
    config_path: string;
    has_config: boolean;
}

export type SettingsTab = "accounts" | "general" | "budget" | "safety" | "search" | "paths";
```

### Projects & Sessions

```typescript
export interface RecentProject {
    path: string;
    name: string;
    last_opened: string;
}

export interface GUISessionInfo {
    id: string;
    kind?: string;
    title: string;
    created_at: string;
    updated_at: string;
    message_count: number;
}

export type SpendCategory = "normal" | "web" | "test" | "skills";

export interface SpendEntry {
    ts: number;
    category: SpendCategory;
    model: string;
    tokens: number;
    cost_usd: number;
    label?: string;
}

export interface FileEntry {
    name: string;
    path: string;
    is_dir: boolean;
    size: number;
}

export interface BrowserHistoryEntry {
    query: string;
    at: number;
}
```

### Attachment (UI-only)

```typescript
export interface Attachment {
    id: number;
    name: string;
    size: number;
    content?: string;
}
```

## 2. Mapping to Wails Events/Commands

### Backend → Frontend Events

| Event Name | Payload | Frontend State Updated |
|------------|---------|----------------------|
| `status:update` | `AppStatus` | `status` state |
| `config:update` | `ConfigView` | `config` state |
| `stream:start` | `{ sessionId: string }` | `busy`, `streaming`, `statusLabel` |
| `stream:delta` | `{ text: string }` | Active tab's streaming message |
| `stream:end` | `{ id: number; usage: UsageInfo; model: string; duration_ms: number }` | `busy`, `streaming`, active tab messages |
| `stream:cancelled` | `{}` | `busy`, `streaming`, `statusLabel` |
| `stream:error` | `{ message: string }` | `errorState`, `lastError`, `busy` |
| `tool:start` | `ToolActivity` | Active tab's tools |
| `tool:end` | `{ id: string; status: "done" | "error"; result?: string; error?: string }` | Active tab's tools |
| `session:list` | `GUISessionInfo[]` | `sessions` |
| `session:loaded` | `{ sessionId: string; messages: ChatMessage[] }` | Active tab's messages |
| `usage:update` | `{ tokens: number; cost: number; accum: number }` | `sessionUsage` |
| `config:changed` | `ConfigView` | `config`, theme, language |
| `explorer:entries` | `FileEntry[]` | `explorerEntries` |
| `spend:entries` | `SpendEntry[]` | SpendView entries |

### Frontend → Backend Commands

| Command | Input | Output | Description |
|---------|-------|--------|-------------|
| `Send` | `text: string` | `void` | Send chat message |
| `Cancel` | — | `void` | Cancel streaming |
| `GetConfig` | — | `ConfigView` | Load configuration |
| `SaveConfig` | `partial: Partial<ConfigView>` | `ConfigView` | Save configuration |
| `SessionList` | — | `GUISessionInfo[]` | List sessions |
| `SessionOpen` | `id: string` | `ChatMessage[]` | Load session messages |
| `SessionDelete` | `id: string` | `void` | Delete session |
| `ListProviders` | — | `ProviderInfo[]` | List available providers |
| `DetectModels` | `provider: string; apiKey: string` | `string[]` | Detect models for key |
| `AddAccount` | `{ name: string; provider: string; model: string; key: string }` | `AccountView` | Add account |
| `DeleteAccount` | `name: string` | `void` | Delete account |
| `ActivateAccount` | `name: string` | `void` | Activate account |
| `RenameAccount` | `oldName: string; newName: string` | `void` | Rename account |
| `SetAccountModel` | `name: string; model: string` | `void` | Set account model |
| `ListDirectory` | `path: string` | `FileEntry[]` | List directory contents |
| `BrowseDirectory` | — | `string` | Open native directory picker |
| `GetSpendEntries` | `days: number` | `SpendEntry[]` | Get spend entries |
| `ClearSpend` | — | `void` | Clear spend history |
| `AddTrustedPath` | `path: string` | `void` | Add trusted path |
| `RemoveTrustedPath` | `path: string` | `void` | Remove trusted path |
| `GetBrowserHistory` | — | `BrowserHistoryEntry[]` | Get browser search history |

## 3. Frontend State Shape

```typescript
// App.tsx state
interface AppState {
    status: AppStatus | null;          // null until loaded
    config: ConfigView | null;         // null until loaded
    tabs: ChatTab[];
    activeTabId: string;
    input: string;
    attachments: Attachment[];
    busy: boolean;
    streaming: boolean;
    statusLabel: string;
    errorState: boolean;
    lastError: string | null;
    settingsOpen: boolean;
    spendOpen: boolean;
    sidebarCollapsed: boolean;
    recentProjects: RecentProject[];
    sessions: GUISessionInfo[];
    projectMenuOpen: boolean;
    explorerOpen: boolean;
    explorerEntries: FileEntry[];
    explorerPath: string;
    headerModelMenu: boolean;
    headerModels: string[];
    headerAccountMenu: boolean;
    modeMenuOpen: boolean;
    paletteOpen: boolean;
    paletteQuery: string;
    paletteSel: number;
    cmdSel: number;
    paletteDismissed: boolean;
    activePane: "none" | "browser" | "test" | "skills";
    browser: BrowserPane;
    browserInput: string;
    browserModelMenu: boolean;
    browserModels: string[];
    browserHistory: BrowserHistoryEntry[];
    test: BrowserPane;
    testInput: string;
    skills: BrowserPane;
    skillsInput: string;
    sessionUsage: { tokens: number; cost: number; accum: number };
    stickToBottom: boolean;
    showJumpDown: boolean;
    activityStick: boolean;
    zoom: number;
    copiedId: number | null;
}

// ChatTab (internal)
interface ChatTab {
    id: string;
    title: string;
    messages: ChatMessage[];
    tools: ToolActivity[];
    streaming: {
        msgId: number;
        text: string;
        usage: UsageInfo | null;
        started: number;
    } | null;
    saved: boolean;
    sessionId: string | null;
}

// BrowserPane (internal)
interface BrowserPane {
    messages: ChatMessage[];
    tools: ToolActivity[];
    streaming: {
        msgId: number;
        text: string;
        usage: UsageInfo | null;
        started: number;
    } | null;
    busy: boolean;
    status: { provider: string; model: string; account: string } | null;
    sessionId: string | null;
    title: string;
}
```

## 4. Data Flow Diagram

```
┌─────────────────────────────────────────────────────┐
│                    Go Backend                        │
│  ┌──────────┐  ┌──────────┐  ┌──────────────────┐  │
│  │ Engine   │  │ Config   │  │ Session Manager  │  │
│  └────┬─────┘  └────┬─────┘  └────────┬─────────┘  │
│       │              │                  │            │
│       └──────────────┼──────────────────┘            │
│                      │                               │
│              ┌───────▼────────┐                      │
│              │  Wails Bindings │                      │
│              └───────┬────────┘                      │
└──────────────────────┼──────────────────────────────┘
                       │
          ┌────────────┼────────────┐
          │  Events    │  Commands  │
          │  (B→F)     │  (F→B)     │
          │            │            │
┌─────────▼────────────▼────────────▼──────────────────┐
│                  React Frontend                       │
│  ┌─────────────────────────────────────────────────┐ │
│  │  useEffect(() => Events.on(...))                │ │
│  │  └─ Updates: status, config, sessions, etc.     │ │
│  └─────────────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────────────┐ │
│  │  Event Handlers (send, cancel, etc.)            │ │
│  │  └─ Calls: Send(), Cancel(), GetConfig(), etc.  │ │
│  └─────────────────────────────────────────────────┘ │
│  ┌─────────────────────────────────────────────────┐ │
│  │  React State → UI Render                        │ │
│  │  └─ App.tsx → App.css → Browser                 │ │
│  └─────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────┘
```

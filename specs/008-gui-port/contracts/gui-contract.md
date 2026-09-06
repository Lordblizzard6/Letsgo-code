# GUI Contract — 008 GUI Port

**Purpose**: Define the exact interface between the React frontend and the Go Wails backend. This contract specifies all events the frontend listens to, all commands the frontend calls, and the state management approach.

## 1. Events (Backend → Frontend)

### 1.1 Status Events

| Event | Payload | Frontend Action |
|-------|---------|-----------------|
| `status:update` | `AppStatus` | Replace `status` state entirely |
| `config:update` | `ConfigView` | Replace `config` state entirely |
| `config:changed` | `ConfigView` | Replace `config` + apply theme/language changes |

### 1.2 Streaming Events

| Event | Payload | Frontend Action |
|-------|---------|-----------------|
| `stream:start` | `{ sessionId: string; model: string }` | Set `busy=true`, `streaming=true`, create empty assistant message placeholder |
| `stream:delta` | `{ text: string }` | Append `text` to current streaming message content |
| `stream:end` | `{ id: number; usage: UsageInfo; model: string; duration_ms: number }` | Finalize message, set `busy=false`, `streaming=false`, update usage |
| `stream:cancelled` | `{}` | Set `busy=false`, `streaming=false`, remove placeholder if empty |
| `stream:error` | `{ message: string }` | Set `errorState=true`, `lastError=message`, `busy=false` |

### 1.3 Tool Events

| Event | Payload | Frontend Action |
|-------|---------|-----------------|
| `tool:start` | `ToolActivity` | Add tool to active tab's tools array |
| `tool:end` | `{ id: string; status: "done" \| "error"; result?: string; error?: string }` | Update tool status in active tab's tools array |

### 1.4 Session Events

| Event | Payload | Frontend Action |
|-------|---------|-----------------|
| `session:list` | `GUISessionInfo[]` | Replace `sessions` state |
| `session:loaded` | `{ sessionId: string; messages: ChatMessage[] }` | Load messages into new/existing tab |

### 1.5 Usage Events

| Event | Payload | Frontend Action |
|-------|---------|-----------------|
| `usage:update` | `{ tokens: number; cost: number; accum: number }` | Update `sessionUsage` state |

### 1.6 Data Events

| Event | Payload | Frontend Action |
|-------|---------|-----------------|
| `explorer:entries` | `FileEntry[]` | Update `explorerEntries` state |
| `spend:entries` | `SpendEntry[]` | Update SpendView entries state |
| `browser:history` | `BrowserHistoryEntry[]` | Update `browserHistory` state |

## 2. Commands (Frontend → Backend)

### 2.1 Chat Commands

```typescript
// Send a chat message
Send(text: string): Promise<void>

// Cancel current streaming
Cancel(): Promise<void>
```

### 2.2 Config Commands

```typescript
// Get full configuration
GetConfig(): Promise<ConfigView>

// Save partial configuration
SaveConfig(partial: Partial<ConfigView>): Promise<ConfigView>

// List available providers
ListProviders(): Promise<ProviderInfo[]>

// Detect models for a provider+key
DetectModels(provider: string, apiKey: string): Promise<string[]>
```

### 2.3 Account Commands

```typescript
// Add a new account
AddAccount(params: {
    name: string;
    provider: string;
    model: string;
    key: string;
}): Promise<AccountView>

// Delete an account
DeleteAccount(name: string): Promise<void>

// Activate an account
ActivateAccount(name: string): Promise<void>

// Rename an account
RenameAccount(oldName: string, newName: string): Promise<void>

// Set account model
SetAccountModel(name: string, model: string): Promise<void>
```

### 2.4 Session Commands

```typescript
// List all sessions
SessionList(): Promise<GUISessionInfo[]>

// Open a session and get messages
SessionOpen(id: string): Promise<ChatMessage[]>

// Delete a session
SessionDelete(id: string): Promise<void>
```

### 2.5 Explorer Commands

```typescript
// List directory contents
ListDirectory(path: string): Promise<FileEntry[]>

// Open native directory picker
BrowseDirectory(): Promise<string>
```

### 2.6 Spend Commands

```typescript
// Get spend entries for N days
GetSpendEntries(days: number): Promise<SpendEntry[]>

// Clear spend history
ClearSpend(): Promise<void>
```

### 2.7 Trusted Paths Commands

```typescript
// Add a trusted path
AddTrustedPath(path: string): Promise<void>

// Remove a trusted path
RemoveTrustedPath(path: string): Promise<void>
```

### 2.8 Browser Commands

```typescript
// Get browser search history
GetBrowserHistory(): Promise<BrowserHistoryEntry[]>
```

## 3. State Management Approach

### 3.1 Pattern: useState + Events

The frontend uses React `useState` for all state, with `useEffect` subscribing to Wails events:

```typescript
// Initial load
useEffect(() => {
    GetConfig().then(setConfig);
    SessionList().then(setSessions);
}, []);

// Event subscriptions
useEffect(() => {
    const unsubs = [
        Events.on("status:update", (_, data) => setStatus(data)),
        Events.on("config:update", (_, data) => setConfig(data)),
        Events.on("stream:start", (_, data) => { /* ... */ }),
        // ... more events
    ];
    return () => unsubs.forEach(u => u());
}, [dependencies]);
```

### 3.2 Pattern: Null Guards

Since config/status load asynchronously, the render path must handle `null`:

```typescript
if (!status || !config) {
    return <LoadingSpinner />;
}
// From here, status and config are guaranteed non-null
```

### 3.3 Pattern: Optimistic UI

For actions like "send message", the frontend immediately adds the user message to the chat (optimistic), then the backend confirms via events:

```typescript
const send = async () => {
    // 1. Optimistic: add user message immediately
    applyMessages(prev => [...prev, userMsg]);
    setBusy(true);
    setStreaming(true);

    // 2. Send to backend (may fail)
    try {
        await Send(text);
    } catch (err) {
        // 3. Rollback on error
        setErrorState(true);
        setLastError(err.message);
        setBusy(false);
        setStreaming(false);
    }
};
```

### 3.4 Pattern: Tab-Based State

Chat state is per-tab, stored in `tabs: ChatTab[]`. Each tab has its own:
- `messages: ChatMessage[]`
- `tools: ToolActivity[]`
- `streaming: { ... } | null`
- `sessionId: string | null`

The active tab is identified by `activeTabId`. All stream/tool events update only the active tab.

## 4. Error Handling

| Error Type | Frontend Action |
|------------|-----------------|
| Network error | Show error banner, set `errorState=true` |
| Stream error | Show error in chat, set `errorState=true` |
| Config load error | Show loading error screen |
| Command failure | Show error toast/banner |

## 5. Lifecycle

```
1. App mounts
   → GetConfig() → set config
   → SessionList() → set sessions
   → Subscribe to Events

2. User sends message
   → Send(text) → backend processes
   → stream:start → busy=true
   → stream:delta → append to message
   → stream:end → finalize message

3. User opens settings
   → SettingsView mounts
   → GetConfig() → populate form
   → User edits → SaveConfig(partial)

4. User opens spend
   → SpendView mounts
   → GetSpendEntries(days) → populate entries

5. User closes app
   → Event subscriptions cleaned up
```

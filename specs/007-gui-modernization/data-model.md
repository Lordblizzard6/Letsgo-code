# Data Model: Modernización de la GUI — Entidades de UI

**Feature Branch**: `007-gui-modernization`

**Created**: 2026-08-27

## Entities

### ChatTab

Represents a chat tab in the tabs bar.

| Field | Type | Description |
|-------|------|-------------|
| id | string | Unique identifier |
| title | string | Tab title (first message or "New Chat") |
| messages | ChatMessage[] | Messages in this tab |
| tools | ToolActivity[] | Tool activity for this tab |
| streaming | StreamingState \| null | Active streaming state |
| saved | boolean | Whether linked to a saved session |
| sessionId | string \| null | Linked session ID |

**State Transitions**:
- Created → Active (when selected)
- Active → Closed (when tab closed)
- Empty → Has messages (when first message sent)

---

### ChatMessage

Represents a single message in the chat.

| Field | Type | Description |
|-------|------|-------------|
| id | number | Unique identifier |
| role | "user" \| "assistant" \| "tool" | Message role |
| content | string | Message content |
| usage | UsageInfo \| null | Token usage (assistant only) |
| duration_ms | number \| null | Duration in ms (assistant only) |
| model | string \| null | Model used (assistant only) |
| kind | "tool" \| undefined | Tool activity marker |
| toolStatus | "running" \| "done" \| "error" \| undefined | Tool status |

**Validation Rules**:
- `role = "user"`: content required, no usage/duration/model
- `role = "assistant"`: content required, usage/duration/model optional
- `role = "tool"`: content required, kind必须 = "tool", toolStatus required

**State Transitions**:
- Created → Sent (when added to messages)
- Tool Running → Done/Error (when tool completes)

---

### ToolActivity

Represents a tool execution in the activity pane.

| Field | Type | Description |
|-------|------|-------------|
| id | string | Unique identifier |
| name | string | Tool name (e.g., "grep_search") |
| arguments | string | Tool arguments (formatted) |
| status | "running" \| "done" \| "error" | Execution status |
| result | string \| undefined | Success result |
| error | string \| undefined | Error message |

**State Transitions**:
- Running → Done (success)
- Running → Error (failure)

**Validation Rules**:
- `status = "running"`: result/error must be undefined
- `status = "done"`: result required, error must be undefined
- `status = "error"`: error required, result must be undefined

---

### StreamingState

Represents active streaming from the assistant.

| Field | Type | Description |
|-------|------|-------------|
| msgId | number | Message ID being streamed |
| text | string | Accumulated text |
| usage | UsageInfo \| null | Final usage (when complete) |
| started | number | Start timestamp |

**State Transitions**:
- Started → Streaming (text accumulates)
- Streaming → Complete (usage received)
- Streaming → Cancelled (user stops)

---

### UsageInfo

Token usage information for a message or session.

| Field | Type | Description |
|-------|------|-------------|
| prompt_tokens | number | Input tokens |
| completion_tokens | number | Output tokens |
| total_tokens | number | Total tokens |

**Validation Rules**:
- All fields ≥0
- `total_tokens = prompt_tokens + completion_tokens`

---

### AppStatus

Application status displayed in the header.

| Field | Type | Description |
|-------|------|-------------|
| provider | string | Current provider (e.g., "openrouter") |
| model | string | Current model (e.g., "claude-3-5-sonnet") |
| language | string | UI language ("es" \| "en") |
| theme | string | Theme ("dark" \| "light") |
| active_account | string | Active account name |
| is_running | boolean | Whether agent is running |
| has_config | boolean | Whether config exists |
| project_dir | string | Current project directory |
| work_mode | string | Work mode (code/architecture/planning/research) |
| session_tokens | number | Tokens used in session |
| session_cost_usd | number | Cost in USD for session |
| account_cumulative_tokens | number | Total tokens for account |

---

### ConfigView

Configuration view for the settings modal.

| Field | Type | Description |
|-------|------|-------------|
| active_account | string | Active account name |
| language | string | UI language |
| theme | string | Theme |
| agent_timeout | number | Agent timeout in seconds |
| trusted_paths | string[] | Trusted directories |
| project_dir | string | Current project |
| budget | BudgetView | Budget settings |
| safety_mode | string | Safety mode |
| search | SearchView | Search settings |
| ui_zoom | number | UI zoom percentage |
| accounts | AccountView[] | Available accounts |
| config_path | string | Config file path |
| has_config | boolean | Whether config exists |

---

### AccountView

Account information for the account selector.

| Field | Type | Description |
|-------|------|-------------|
| name | string | Account name |
| provider | string | Provider ID |
| model | string | Default model |
| has_key | boolean | Whether API key is set |
| key_masked | string | Masked API key |
| active | boolean | Whether this is the active account |
| tokens | number | Tokens used |

---

### GUISessionInfo

Session information for the sidebar.

| Field | Type | Description |
|-------|------|-------------|
| id | string | Session ID |
| kind | string \| undefined | Session kind ("test" \| "skills") |
| title | string | Session title |
| created_at | string | Creation timestamp (ISO) |
| updated_at | string | Last update timestamp (ISO) |
| message_count | number | Number of messages |

---

### RecentProject

Recent project for the sidebar.

| Field | Type | Description |
|-------|------|-------------|
| path | string | Project path |
| name | string | Project name |
| last_opened | string | Last opened timestamp (ISO) |

---

### FileEntry

File/directory entry for the explorer.

| Field | Type | Description |
|-------|------|-------------|
| name | string | File/directory name |
| path | string | Relative path |
| is_dir | boolean | Whether it's a directory |
| size | number | File size in bytes |

---

### SpendEntry

Spend entry for the spend modal.

| Field | Type | Description |
|-------|------|-------------|
| ts | number | Timestamp |
| category | SpendCategory | Category |
| model | string | Model used |
| tokens | number | Tokens used |
| cost_usd | number | Cost in USD |
| label | string \| undefined | Optional label |

---

### Attachment

File attachment in the composer.

| Field | Type | Description |
|-------|------|-------------|
| id | number | Unique identifier |
| name | string | File name |
| size | number | File size in bytes |
| content | string \| undefined | File content (if loaded) |

---

### WorkMode

Work mode for the mode selector.

| Field | Type | Description |
|-------|------|-------------|
| id | string | Mode ID |
| icon | Component | SVG icon component |
| i18n | string | i18n key for label |
| hint | string | i18n key for hint |

**Enum Values**:
- `code`: Code development mode
- `architecture`: Architecture design mode
- `planning`: Planning mode
- `research`: Research mode

---

### BrowserHistoryEntry

Browser history entry for the browser pane.

| Field | Type | Description |
|-------|------|-------------|
| query | string | Search query |
| at | number | Timestamp |

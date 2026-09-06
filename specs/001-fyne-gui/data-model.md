# Data Model: Fyne Desktop GUI for LetsGO Code

**Date**: 2026-08-02 | **Plan**: [plan.md](plan.md) | **Spec**: [spec.md](spec.md)

The feature's entities map onto the **existing** SQLite schema in `internal/db` (file `~/.letsGo/history.db`, `modernc.org/sqlite`). No new tables are required for v1; the GUI and the engine access storage exclusively through `internal/db` (FR-005, FR-012, FR-026).

## Entity map

| Spec entity (spec.md) | Storage / representation | Notes |
|-----------------------|--------------------------|-------|
| **Session** | `sessions` table; `db.Session{ID, Name, CreatedAt, UpdatedAt, ProjectPath, IsActive}` | One active session at a time (`is_active` flag). Created via `CreateSession`, resumed via `ResumeSession`/`SetActiveSession`, renamed `RenameSession`, deleted `DeleteSession`. |
| **Message** | `messages` table; `db.Message{ID, SessionID, Role, Content, Timestamp, TokensInput, TokensOutput, CostUSD}` | `content` is JSON (string or `api.ContentBlock` array — tool_use/tool_result blocks included, matching `internal/api` types). Token/cost columns feed the usage view (FR-009). |
| **Tool Call** | Serialized inside `messages.content` as `api.ContentBlock{ToolUse}` / `api.ContentBlock{ToolResult}` | Live lifecycle (pending-approval → approved/rejected → executed/failed) exists only in `internal/engine` memory during a run; the persisted record is the tool_use + tool_result pair. |
| **Agent Task** | Conversation-level; assigned to a session; tool calls persisted as above | Parallel sub-agents (`AgentTool`, `TaskCreateTool`…) execute within the same engine loop; no separate table in v1. |
| **MCP Configuration** | `internal/config` (viper) + `internal/mcp.Manager` runtime | Servers configured in config; tools discovered at runtime and merged into the tool registry (`ListMcpResourcesTool`, `McpCallTool`). |
| **Plugin** | `internal/plugins` registry (config-driven install/load) | Loaded on launch; commands/contributions surface in the GUI's command palette / menu. |
| **Provider Configuration** | `internal/config` (viper): provider, API key, model, auto-approve categories | One active provider at a time; settings screen reads/writes the same keys the CLI uses, so both interfaces see the same config. |
| **Usage Record** | Aggregated from `messages.tokens_input`, `messages.tokens_output`, `messages.cost_usd` (+ `internal/tools.CostTracker` for the current run) | Usage view sums per provider/period. |

## Validation rules (from spec FRs)

- FR-005/FR-012: a session's last **complete** message is always flushed by the engine before a run is interrupted; partial stream content is never persisted.
- FR-007: API key stored via the existing config mechanism (same as CLI); GUI never stores keys elsewhere.
- FR-014/FR-016: a rejected tool call must never produce a persisted `tool_result` with success semantics — the engine records rejection in the tool_result (`is_error`-style) or omits it; the approval decision itself is not persisted in v1.
- Sessions without an active provider: GUI blocks sending until `Provider Configuration` is valid (SC-003), matching CLI behavior.

## State transitions

```text
Session:  created → active ⇄ (resumed|new) → deleted
Tool Call (live, in engine):  requested → pending-approval → approved → executed
                                                    → rejected  → (no execution)
                                                    → (timeout) → treated as rejected
Stream:   sending → streaming → done | cancelled | error
```

## Scale assumptions

- Single-user desktop; sessions are unbounded but the chat view renders incrementally (windowed history load for large sessions — research #9).
- Message rows are small; full-history search already exists (`SearchMessages`) and is reused by the GUI's session search.

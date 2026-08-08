# Contract: Engine Event/Command Protocol (`internal/engine`)

**Feature**: Fyne Desktop GUI for LetsGO Code | **Date**: 2026-08-02

The shared contract between presentations (GUI and TUI) and the conversation/agent loop. `internal/engine` is the ONLY owner of the loop; presentations translate user input into **commands** and render **events**. This is what makes FR-021/FR-026 (shared core, no duplication) hold.

## Commands (presentation → engine, over a buffered channel)

| Command | Payload | Semantics |
|---------|---------|-----------|
| `SendMessage` | `{text, sessionID}` | Start a user turn; persists user message first |
| `SendTask` | `{prompt, sessionID}` | Start agent-mode turn (tools enabled, approval flow active) |
| `Cancel` | — | Abort in-flight stream/tool cycle; keep last complete message (FR-004, FR-012) |
| `ApproveTool` | `{toolCallID}` | Execute the pending tool call |
| `RejectTool` | `{toolCallID}` | Never execute; feed rejection to model; continue |
| `SetAutoApprove` | `{category, enabled}` | Per-category auto-approve (file-edit, bash, web) persisted via `internal/config` |
| `SwitchSession` | `{sessionID}` | Flush current state; resume target session |
| `Compact` | — | Trigger context compaction (existing `db.SaveCompactHistory` flow) |
| `Stop` | — | Graceful shutdown of the engine loop |

## Events (engine → presentation, over a buffered channel)

| Event | Payload | When |
|-------|---------|------|
| `UserMessageAppended` | `{message}` | User message persisted |
| `StreamStart` | `{sessionID, requestID}` | Provider stream opened |
| `StreamDelta` | `{text}` | Chunk of assistant text (batched at ≤100 ms by engine) |
| `StreamDone` | `{messageID}` | Assistant turn complete (persisted) |
| `StreamCancelled` | — | User or error cancellation |
| `ToolRequested` | `{toolCallID, name, input}` | Tool wants execution; approval pending (unless auto-approved) |
| `ToolExecuting` | `{toolCallID}` | Approved, executing |
| `ToolResult` | `{toolCallID, name, content, isError}` | Execution finished |
| `ToolRejected` | `{toolCallID, name}` | User rejected |
| `AgentTaskUpdate` | `{taskID, status, summary}` | Parallel sub-agent progress (FR-018) |
| `UsageUpdate` | `{inputTokens, outputTokens, costUSD}` | After each persisted message |
| `Error` | `{message, recoverable}` | Provider/network/MCP failures (FR-010) |
| `Idle` | — | Loop has no work; presentation can re-enable input |

## Permission driver (presentation implements)

```go
type PermissionDriver interface {
    // Prompt is called when a tool call needs approval. Implementations MUST
    // be synchronous from the engine's perspective (block until user decides).
    Prompt(toolName string, input any) (allow bool, err error)
    // AutoApprove reports whether category is currently auto-approved.
    AutoApprove(category string) bool
}
```

- TUI: existing stdin prompt flow (behavior preserved byte-for-byte).
- GUI: keyboard-driven confirm dialog (Enter=approve, Esc=reject) — FR-025.
- Timeout: engine treats a prompt older than the configured approval timeout as rejected (edge case: no response → timeout, with notice event).

## Sequencing guarantees

1. Events for one turn are strictly ordered; `ToolResult` for call N is emitted before `ToolRequested` for call N+1.
2. `Idle` is emitted after every turn (success, cancel, or error) — presentations re-enable input only on `Idle` or `Error`.
3. `Cancel` is accepted at any point; the engine guarantees no tool executes after a `Cancel` (except one already mid-execution, which reports via `ToolResult`).
4. Zero silent modifications: no `Execute` call without a prior approve path (auto-approve counts as approved) — SC-007.

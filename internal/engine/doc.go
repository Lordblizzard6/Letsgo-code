// Package engine hosts the LetsGO conversation core: it owns the send/cancel
// loop, plan mode, tools, sessions and usage accounting, and exposes its state
// to presentations through typed events.
//
// # Contract boundary
//
// Every interface of the product (TUI in internal/tui, Wails GUI in
// cmd/wails) consumes this package — and only this package, internal/db and
// internal/config — through the surface defined in
// specs/005-bubbletea-tui-wails/contracts/frontend-contract.md (FR-001,
// FR-002). No interface duplicates conversation logic, and none reaches into
// the core outside the contract:
//
//   - Commands (§1): Engine.Send with the Command types of commands.go
//     (send, cancel, setModel, setAutoApprove, remember, resetMemory, slash,
//     switchSession, compact, ...).
//   - Events (§2): the typed events of events.go consumed via Engine.Events().
//   - Sessions/config (§3): SessionCreate/SessionList/ResumeSession and
//     GetHistory/GetMessages (internal/db — SessionCreate, ListSessions,
//     ResumeSession, GetHistory, GetMessages), GetConfig/SaveConfig
//     (internal/config — AppConfig, LoadConfig, SaveConfig) and
//     GetUsageToday (internal/analytics — GetUsageToday).
//
// Conformance: the headless suite in contract_test.go (C-001..C-006) is the
// single source of truth — a failing test there breaks integration of every
// interface (frontend-contract.md §4).
//
// # Contract §2 ↔ engine events
//
// The contract rows map 1:1 onto events.go:
//
//	stream:start     -> StreamStart
//	stream:delta     -> StreamDelta
//	stream:end       -> StreamDone
//	stream:cancelled -> StreamCancelled
//	stream:error     -> ErrorEvent
//	tool:start       -> ToolExecuting
//	tool:end         -> ToolResult
//	usage:update     -> UsageUpdate
//	session:list     -> SessionList
//	session:loaded   -> SessionLoaded
//	config:changed   -> ConfigChanged
//
// The last three are adapter events emitted by presentations-facing helpers
// (cmd/wails services, internal/tui) when the underlying operations in
// internal/db and internal/config complete; the engine itself streams the
// first eight during a turn.
package engine
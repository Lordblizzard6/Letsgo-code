# Post-change Review (slash commands)

## Scope reviewed
- Commit `0a188c3` (`internal/tui/slash_commands.go`, `internal/tui/ui.go`).

## Potential risks / missing pieces

1. **Global mutable session id (`currentSlashSessionID`) can become stale or race-prone**
   - Slash command handlers depend on a package-level variable.
   - In multi-session or concurrent flows this can point to the wrong session.
   - Suggestion: bind session id to the `model` instance and pass explicitly to handlers.

2. **`/clear` returns success even if auxiliary table cleanup fails**
   - `context_files` and `compact_history` deletes ignore returned errors.
   - This can leave partial state while showing a success message.
   - Suggestion: capture and report those errors (or wrap in transaction).

3. **`/compact` rewrites history without transactional safety**
   - Flow is clear history -> rewrite last 20 messages.
   - If a write fails mid-loop, session can be left truncated.
   - Suggestion: perform compact operation in a single DB transaction.

4. **`/compact` can lose message ordering guarantees if DB retrieval order changes**
   - Assumes `db.GetHistory()` returns ordered history.
   - If query ordering changes in future, retained "last 20" may be incorrect.
   - Suggestion: enforce ordering in `GetHistory` query or assert order before compacting.

5. **Mixed language UX and low observability for user-facing errors**
   - Several responses are in Spanish while other UI text is in English.
   - Debug logs go to stderr; users may not see actionable reason in UI.
   - Suggestion: standardize locale strategy and include concise actionable remediation in TUI messages.

## Validation commands run
- `go test ./...` (fails in `cmd` package due to `go vet`-style `fmt.Println` redundant newline warnings; unrelated to slash changes).

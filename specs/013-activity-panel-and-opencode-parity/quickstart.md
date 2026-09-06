# Quickstart & Validation Guide: Opencode Parity & Antigravity Activity Panel

## Prerequisites
- Windows environment with Go 1.26+ and Node.js 18+ installed.
- Repository cloned at `d:\projects\Letsgo-code`.

## Setup & Build
1. Build frontend bundle:
   ```bash
   cd cmd/wails/frontend
   npm run build
   ```
2. Build executable:
   ```bash
   cd d:\projects\Letsgo-code
   go build -o letsgo.exe .
   ```

## Validation Scenarios

### Scenario 1: Overview Panel & Changed Files Metrics
1. Launch `.\letsgo.exe gui .`.
2. Press `Ctrl+B` or click the Activity Panel toggle button on the header.
3. Verify the **Overview** tab is displayed:
   - Check uncommitted files counter badge.
   - Check changed files list displays filename, `+X` (green), `-Y` (red) badges, and the directory in grey (`App.tsx  cmd/wails/frontend/src`).
   - Check background tasks and skills sections are visible.
4. Click the Maximize button (`🗖`) on the inspector header:
   - Verify the panel expands across the main workspace.
   - Click Restore (`🗗`) to return to standard width.

### Scenario 2: Git Review & Center Diff Viewer
1. Click the **Git Review** tab (or type `/diff` in the chat composer).
2. Verify the panel switches directly to Git Review.
3. Select a modified file from the list:
   - Verify the center diff viewer renders line additions (`+`), deletions (`-`), hunk headers (`@@`), and line numbers.
4. Click the Stage (`+`) button next to a file.
   - Verify the file moves to the "Staged Changes" list.
5. Enter a commit message and verify commit succeeds.

### Scenario 3: Executed Commands & Embedded Diff
1. In the chat, issue a request that runs a tool or shell command (or type a bash command).
2. Click the **Commands & Tools** tab on the Activity Inspector.
3. Verify the execution appears with status, duration, and arguments.
4. If files were modified, expand the command item and verify the embedded git diff is displayed.

### Scenario 4: OpenCode Parity Slash Commands
1. In the composer textarea, type `/`.
2. Verify autocomplete popup displays:
   - `/init`, `/undo`, `/redo`, `/diff`, `/review`, `/compact`, `/cost`, `/skills`, `/tasks`, `/terminal`, `/share`.
3. Select `/cost` and hit Enter -> Verify the Spend modal opens.
4. Select `/share` and hit Enter -> Verify toast notification appears confirming chat copied to clipboard.

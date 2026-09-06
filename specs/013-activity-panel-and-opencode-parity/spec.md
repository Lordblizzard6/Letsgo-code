# Feature Specification: Opencode Command Parity & Antigravity CLI Activity Panel

**Feature Branch**: `013-activity-panel-and-opencode-parity`

**Created**: 2026-09-04

**Status**: Draft

**Input**: User description: "crea un Spec para añadir mas comandos para parity con opencode, y que la barra de actividades muestre Skills, cambios a archivos(con nombre del archivo Lineas+ Lineas-), toolcalls, mcp llamados, y sea desplegable como la de antigravity cli, tambien tenga una pestaña de comandos que se han ejecutado(y que alli mismo se meta el github diff) que sea un diseño como Antigravity cli donde se tiene: Overview(el panel que te dije donde se muestran files changed(uncomitted/committed) abajo nombres de archivo y al lado en gris la carpeta donde estan, las background tasks, las skills usadas y las terminales; otro icono que tambien de github de review de los cambios y lo staged tambien como antigravity; en el centro un file view (tambien replicado de antigravity cli) y un maximize, y un Toggle(para cerrar el panel); en fin una spec para acercar el funcionamiento del panel de actividades y el github diff a algo 1:1 con el de Antigravity Cli"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Antigravity CLI Activity Inspector Panel with Overview, File Changes & System Metrics (Priority: P1)

Developers interacting with LetsGo require deep visibility into what the assistant and tools are doing in real time. Rather than a static, plain right-side list, the application provides an Antigravity CLI-style collapsible Activity Inspector Panel with an Overview tab displaying:
- Header controls: Tab navigation icons (`Overview`, `Git Review`, `Command/Tool Executions`), active filename/context in the center, a Maximize toggle (`🗖` / `🗗`) to expand the panel for full-screen analysis, and a Close/Toggle button (`✕` / `⇲`).
- Files Changed counter distinguishing `uncommitted` and `committed` changes.
- Changed file items showing the file name in bold, line change metrics (`+X` in green, `-Y` in red), and directory path in muted grey (`App.tsx  cmd/wails/frontend/src`).
- Dedicated status sections for Background Tasks (with status dot and cancel/view actions), Skills Used (skills loaded or executed in session), and Terminals / Shell executions.

**Why this priority**: Directly implements the primary visual parity with Antigravity CLI and gives developers full context of session impact at a glance.

**Independent Test**:
Open LetsGo GUI. Toggle the activity panel (via button or `Ctrl+B`/`Ctrl+D`). Verify the Overview tab displays files changed with `+` and `-` line counts and grey directory paths, active background tasks, skills used, and terminal sessions. Click Maximize to expand, and click it again to restore.

**Acceptance Scenarios**:
1. **Given** the user is in a session with file changes, **When** the Activity panel is opened on the Overview tab, **Then** it renders the uncommitted and committed file count badges and a list of files showing `filename.ext`, `+Lines` / `-Lines` badges, and relative parent directories in grey text.
2. **Given** background tasks are running or completed, **When** viewing the Overview tab, **Then** the Background Tasks section displays each task ID, command, and status indicator.
3. **Given** skills are active or used in the session, **When** viewing the Overview tab, **Then** the Skills section lists active skills and their status.
4. **Given** the user clicks the Maximize button on the inspector header, **Then** the panel expands to occupy the primary viewport area; clicking it again restores the side-docked layout.
5. **Given** the user clicks the Toggle/Close button or presses `Ctrl+B`, **Then** the activity panel smoothly closes.

---

### User Story 2 - Git Review Tab with Staged/Unstaged Files and Center Diff Viewer (Priority: P1)

Developers need to inspect, stage, and commit code changes without leaving the conversation interface. The Git Review tab replicates Antigravity CLI's review workflow:
- Separate collapsible groupings for Unstaged Changes and Staged Changes.
- Quick stage (`+`) and unstage (`-`) actions per file or in bulk.
- Commit box with message input and one-click commit button.
- Center / Integrated Diff Viewer displaying syntax-highlighted hunks (`+`, `-`, `@@`), line numbers, and clean scrolling.

**Why this priority**: Enables complete git lifecycle control directly inside the activity panel, matching Antigravity CLI's code inspection and review flow.

**Independent Test**:
Make an edit to a tracked file. Switch to the Git Review tab. Verify unstaged file is listed, click stage to stage it, verify diff renders in the center viewer with colored additions/deletions, and commit changes using the commit box.

**Acceptance Scenarios**:
1. **Given** unstaged modifications in the workspace, **When** the user clicks the Git Review tab, **Then** unstaged files are listed with status badges (`M`, `A`, `D`, `?`).
2. **Given** a file in the list, **When** clicked, **Then** its diff is displayed in the center file diff viewer with additions, deletions, line numbers, and hunk headers.
3. **Given** unstaged files, **When** the user stages a file and enters a commit message, **Then** `GitService.Commit` executes and the tree cleans up automatically.

---

### User Story 3 - Executed Commands & Tool Calls Tab with Embedded Git Diff (Priority: P2)

When the assistant or user executes terminal commands, MCP calls, or editing tools, users must be able to inspect the history of executions and see the exact git diff resulting from each modifying command.
- Tab listing all executed commands, shell runs, tool calls, and MCP server invocations.
- For each entry: execution timestamp, duration, exit code / status, tool parameters, and console output.
- If a command or tool edited files (e.g., `bash`, `edit`, `write`), an embedded "View Diff" accordion appears directly within the execution card showing the git diff generated by that execution.

**Why this priority**: Eliminates guesswork regarding what a command did to the workspace, providing instant visual verification of command side-effects.

**Independent Test**:
Ask the AI to edit a file or run a terminal command. Navigate to the Executed Commands tab. Verify the tool call card displays status, execution time, and an expandable embedded diff viewer showing the file modifications.

**Acceptance Scenarios**:
1. **Given** tool calls or shell commands have run in the session, **When** opening the Commands & Executions tab, **Then** all calls (including MCP tools) are listed with status, duration, and arguments.
2. **Given** a command modified files in the repo, **When** expanding the command item, **Then** an embedded diff component renders the exact line changes produced.

---

### User Story 4 - OpenCode Slash Command Parity (Priority: P2)

To ensure seamless parity with the OpenCode ecosystem, LetsGo must provide the full set of essential OpenCode slash commands in both the GUI command palette/autocomplete and the CLI/TUI:
- `/init`: Initializes workspace configuration, `.specify`, or project rules (`AGENTS.md`).
- `/undo`: Rolls back the last assistant turn, restoring file states if edits were made.
- `/redo`: Re-applies the previously undone action.
- `/diff`: Toggles or jumps directly to the Git Review tab in the activity inspector.
- `/review`: Dispatches an automated code review request for all current repository changes.
- `/compact`: Truncates and summarizes the conversational context to save token window space.
- `/cost`: Opens or displays the token spend and cumulative cost modal.
- `/skills`: Lists loaded skills, active tools, and enables/disables skill packs.
- `/tasks`: Displays all running and background tasks.
- `/terminal`: Opens the terminal inspector tab or starts a command.
- `/share`: Copies a clean markdown summary of the conversation to the clipboard.

**Why this priority**: Delivers command interoperability and muscle-memory compatibility for developers transitioning between OpenCode and LetsGo.

**Independent Test**:
Type `/` in the composer. Verify autocomplete suggests `/init`, `/undo`, `/redo`, `/diff`, `/review`, `/compact`, `/cost`, `/skills`, `/tasks`, `/terminal`, and `/share`. Execute `/diff` and confirm it opens the Git Review inspector.

**Acceptance Scenarios**:
1. **Given** the composer is focused, **When** the user types `/diff`, **Then** the activity panel opens directly to the Git Review tab.
2. **Given** `/undo` is typed or selected, **Then** the last assistant message and its associated changes are rolled back.
3. **Given** `/cost` is executed, **Then** the Spend modal opens with the session token statistics.
4. **Given** `/share` is executed, **Then** the conversation transcript is sanitized and copied to clipboard with a toast notification.

---

### Edge Cases

- **Non-Git Workspace**: When opened in a folder that is not a git repository, the Overview and Git Review tabs clearly indicate "No Git repository detected" and provide an "Initialize Git" button without crashing.
- **Large Diffs**: Diffs exceeding 5,000 lines are truncated with a notice and an option to view specific files individually to preserve UI responsiveness.
- **Empty State**: In a brand-new session with no commands run and a clean git tree, helpful placeholder cards explain what each section tracks.
- **Rapid Mode Switching & Resizing**: Toggling maximize and collapsing the panel while streaming responses does not interrupt the active chat stream or cause layout shifts in the message feed.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide an integrated Activity Inspector Panel on the right side of the GUI that can be expanded, collapsed, or maximized to full screen.
- **FR-002**: The Activity Inspector header MUST feature view switchers for:
  - `Overview` (System metrics, changed files, background tasks, skills, terminals)
  - `Git Review` (Unstaged/staged review, file diff viewer, commit)
  - `Commands & Tools` (Executed commands, tool calls, MCP calls with embedded diffs)
- **FR-003**: The Overview tab MUST display:
  - Uncommitted and committed file counters
  - Changed files list showing file basename, addition/deletion line badges (`+N` / `-N`), and relative parent directory in muted grey text
  - Active background tasks list with status badges
  - Skills used section with skill identifiers
  - Terminal executions section
- **FR-004**: The Git Review tab MUST support:
  - Staging and unstaging individual files and all files
  - Interactive file selection with a center diff viewer
  - Unified diff view rendering additions (`+`), deletions (`-`), hunk headers (`@@`), and line numbers
  - Commit message input with commit execution via `GitService.Commit`
- **FR-005**: The Commands & Tools tab MUST log every executed command, tool call, and MCP invocation with duration, status, arguments, and an embedded git diff for modifying actions.
- **FR-006**: The Activity Inspector MUST provide a Maximize button (`🗖`) to toggle between docked panel mode and expanded viewport mode, and a Close button (`✕`) to hide the panel.
- **FR-007**: Keyboard shortcuts `Ctrl+B` (or `Cmd+B`) and `Ctrl+D` (or `Cmd+D`) MUST toggle the activity inspector panel.
- **FR-008**: System MUST support OpenCode parity slash commands in both autocomplete and execution:
  - `/init`, `/undo`, `/redo`, `/diff`, `/review`, `/compact`, `/cost`, `/skills`, `/tasks`, `/terminal`, `/share`, `/clear`, `/theme`, `/model`, `/help`.
- **FR-009**: The `/diff` slash command MUST immediately open the activity panel to the Git Review tab.
- **FR-010**: The `/cost` slash command MUST open the Spend modal.
- **FR-011**: The `/share` slash command MUST copy a markdown representation of the current chat to the system clipboard.

### Key Entities

- **ActivityPanelState**: Tracks active tab (`overview`, `git`, `commands`), collapsed state (`boolean`), maximized state (`boolean`), and currently inspected file (`string | null`).
- **FileDiffItem**: Represents a modified file with `path`, `filename`, `directory`, `status` (`M`, `A`, `D`, `U`), `additions` (`number`), `deletions` (`number`), and `staged` (`boolean`).
- **CommandExecutionItem**: Represents an executed command/tool with `id`, `name`, `type` (`tool` | `command` | `mcp`), `args`, `output`, `status`, `startTime`, `durationMs`, and optional `diff`.
- **BackgroundTaskItem**: Represents an active background process with `id`, `command`, `status`, `startTime`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of the requested Antigravity CLI activity panel elements (Overview with files changed `+`/`-` line counts & grey paths, background tasks, skills, terminals; Git Review with staged/unstaged & center diff; Executed Commands with embedded diffs; Maximize and Close toggles) are visible and interactive in the GUI.
- **SC-002**: Panel maximize toggle expands seamlessly to 100% width and restores without viewport jitter or message layout re-renders (< 50ms transition).
- **SC-003**: All 11 OpenCode parity slash commands (`/init`, `/undo`, `/redo`, `/diff`, `/review`, `/compact`, `/cost`, `/skills`, `/tasks`, `/terminal`, `/share`) appear in composer autocomplete and execute their expected actions.
- **SC-004**: Frontend build (`npm run build`) and backend tests (`go test ./cmd/...`) pass cleanly with zero regressions.

## Assumptions

- Git operations use the existing `GitService` Go backend via Wails bindings, extended where needed for line count stats and staging.
- Command execution history is captured from the existing tool events and shell executions in `App.tsx` and `internal/engine`.
- The activity inspector replaces both the separate `#git-diff-panel` and `#right-pane` with a single, coherent, modern dockable drawer.

# Research: CLI Path Argument and Project Scope Synchronization

## Problem Diagnosis

### Root Cause Analysis of "AI listing files from program directory"
When asking the assistant to list files (`ls`), the tool executes:
```go
entries, err := os.ReadDir(path) // path defaults to "."
```
In Go, relative file system operations and subprocess executions (`exec.Command` in `bash.go`) execute against the operating system process working directory (`os.Getwd()`).

When the user launched `letsgo.exe gui`:
1. The process working directory was inherited from the launch directory (the application binary directory).
2. While `CreateSession` set `os.Chdir(projectPath)`, `SessionsService.Open(id)` **never called `os.Chdir(sess.ProjectPath)`**.
3. In the React UI (`App.tsx`), clicking a project or opening a saved session only modified React state (`setStatus(s => ({ ...s, project_dir: ... }))`), which only updated the top header label cosmetically. The Go backend process was never instructed to change its directory.
4. Furthermore, `cmd/gui.go` completely ignored CLI arguments, so `letsgo gui <dir>` had no effect.

Therefore, both issues are fundamentally connected: the entire system lacked a unified working directory lifecycle spanning CLI launch, session switching, and frontend project selection.

---

## Architectural Decisions

### Decision 1: CLI Positional Path Syntax (`letsgo gui [path]`)
- **Decision**: Accept an optional positional argument `[path]` in Cobra (`cmd/gui.go`): `letsgo gui [path]`.
- **Rationale**: Familiar to all developers who use tools like `code .`, `subl .`, or `idea .`.
- **Validation**: If `path` is passed, resolve via `filepath.Abs(path)` and verify `os.Stat(abs)`. If it does not exist or is not a directory, output an informative error to `os.Stderr` and exit with code 1.
- **Alternatives Considered**: Using a flag like `--project-dir` or `-p`. Rejected because positional argument `[path]` is cleaner, more intuitive, and standard across modern editors.

### Decision 2: Working Directory Lifecycle Management (`Hub` & `SessionsService`)
- **Decision**: Centralize current project directory state in `Hub` (`Hub.SetCurrentProject(path)` and `Hub.GetCurrentProject()`), exposed through `SessionsService.SetCurrentProject(path)` and `SessionsService.GetCurrentProject()`.
- **Rationale**:
  - `Hub` is shared by all services (`ChatService`, `GitService`, `SessionsService`).
  - Calling `os.Chdir(path)` inside `SetCurrentProject` immediately updates all Go relative path resolutions (`os.ReadDir`, `os.ReadFile`, `exec.Command`).
  - Emits `project:changed` over Wails event bus so the frontend reactively synchronizes.
- **Alternatives Considered**: Passing `project_path` into every individual tool call. Rejected because Go tools (`tools.LsTool`, `tools.BashTool`, `tools.WriteTool`) and third-party commands in subshells all expect the OS process working directory to be accurate.

### Decision 3: Automatic Directory Switch on Session Open
- **Decision**: In `SessionsService.Open(id string)`, query the session record from SQLite. If `sess.ProjectPath != ""` and the directory exists on disk, invoke `s.SetCurrentProject(sess.ProjectPath)` automatically.
- **Rationale**: Ensures that whenever the user switches between conversations belonging to different projects, the underlying filesystem context immediately switches with zero manual steps.

### Decision 4: Unscoped / Empty Project State
- **Decision**: When `letsgo gui` is launched without arguments, `initialProjectPath` is `""`. The application starts with `project_dir = ""` and displays a dedicated empty-project card/prompt in the chat and header, guiding the user to open a folder before creating project-scoped chats.
- **Rationale**: Prevents accidental operations in the program folder, cleanly separating general-purpose chats from project workspaces.

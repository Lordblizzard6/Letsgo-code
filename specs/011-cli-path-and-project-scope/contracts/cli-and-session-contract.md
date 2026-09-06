# Interface Contract: CLI and Session Project Scope

## 1. CLI Contract (`cmd/gui.go`)

### Syntax
```bash
letsgo gui [path]
```

### Behaviors
1. `letsgo gui`:
   - No positional arguments.
   - Starts Wails GUI with initial project directory = `""`.
2. `letsgo gui <path>`:
   - Exactly 1 positional argument.
   - Resolves `<path>` to absolute path via `filepath.Abs`.
   - Validates existence and directory check (`os.Stat`).
   - If invalid: outputs error to `os.Stderr` and exits with status 1.
   - If valid: sets process working directory (`os.Chdir`) and launches GUI with that path as initial project directory.

---

## 2. Go Services Contract (`SessionsService`)

### Methods
- `GetCurrentProject() string`:
  Returns the active project directory path (or `""` if unscoped).
- `SetCurrentProject(path string) error`:
  - If `path != ""`: validates directory exists, runs `os.Chdir(path)`, updates `Hub`, and emits `project:changed`.
  - If `path == ""`: updates `Hub` to empty, emits `project:changed`.
- `Open(id string) ([]db.Message, error)`:
  - Resumes session.
  - Switches engine session ID.
  - Queries session row: if `session.ProjectPath != ""` and exists on disk, calls `SetCurrentProject(session.ProjectPath)`.
  - Emits `session:loaded` containing `session_id`, `messages`, and `project_path`.
  - Returns message history.

---

## 3. Frontend Contract (`App.tsx`)

### Events Listened
- `project:changed`: `{project_path: string, name: string}`
  Updates `status.project_dir` in React state, adds to `recentProjects`, and re-renders project badge.
- `session:loaded`: `{session_id: string, messages: any[], project_path?: string}`
  If `project_path` is present, updates `status.project_dir`.

### User Interactions
- Clicking a recent project or project folder: calls `SessionsService.SetCurrentProject(path)`.
- Opening a session: calls `SessionsService.Open(id)`.
- When `status.project_dir === ""`: chat displays an empty-project action bar: "No hay proyecto activo. Abre una carpeta para empezar." with a button calling `SessionsService.SelectProjectFolder()`.

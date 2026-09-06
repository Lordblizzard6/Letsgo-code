# Feature Specification: CLI Path Argument and Project Scope Synchronization

**Feature Branch**: `011-cli-path-and-project-scope`

**Created**: 2026-09-04

**Status**: Draft

**Input**: User description: "al ejecutar el comando letsgo gui, no deberia tenerse el path donde esta la aplicacion directamente, asi que me gustaria añadir a la sintaxis del comando el path, ejemplo 'letsgo.exe gui .' o 'letsgo.exe gui \"C:/hola mundo\"' asi cuando se use sin la sintaxis se empieza sin ninguna carpeta apuntando(y se le pida al usuario que se meta en una para empezar, y de una vez solucionar un error, de que por ejemplo tengo una carpeta de proyecto, doy a ese proyecto, y eligo una de las conversasiones en ese proyecto, pero le digo a la IA que liste los archivos y esta en la carpeta del programa(lets-go) eso diagnosticalo y si esta relacionado con la primera informame, pero has un spec para reparar ambos problemas"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Launch GUI with Project Path from CLI (Priority: P1)

Developers frequently invoke command-line tools from the terminal where their codebase is located. When running `letsgo gui .` or `letsgo gui "/path/to/project"`, the GUI starts directly focused on that project directory. The process working directory is automatically set to that project path, the header displays the folder name, and any tool executions (such as listing files or reading code) execute inside that project folder.

**Why this priority**: Directly solves the primary user request and standardizes CLI invocation matching modern developer editors (e.g. `code .`).

**Independent Test**:
Run `letsgo.exe gui .` from any test directory; verify the GUI opens with that directory active, and asking the assistant to list files returns the files in that directory.

**Acceptance Scenarios**:
1. **Given** a directory `D:/test-project` exists, **When** the user runs `letsgo gui "D:/test-project"`, **Then** the application opens with `D:/test-project` as the active project, the working directory is set to that path, and tool calls like `ls` list files inside `D:/test-project`.
2. **Given** the current terminal working directory is `D:/my-app`, **When** the user runs `letsgo gui .`, **Then** `.` resolves to the absolute path `D:/my-app` and opens as the active project.
3. **Given** the user passes an invalid or non-existent path, **When** running `letsgo gui "/path/does/not/exist"`, **Then** the CLI writes a clear error message to stderr and exits with a non-zero status code without crashing.

---

### User Story 2 - Launch Without Directory / Unscoped Empty State (Priority: P1)

When running `letsgo gui` without arguments, the application MUST NOT default to or expose the directory where the binary resides (the application installation directory). Instead, it starts in an unscoped state ("Sin proyecto activo"). The UI clearly prompts the user to open a project folder or choose one from recent projects to begin working.

**Why this priority**: Prevents AI tools from modifying or inspecting the application's internal installation files by default.

**Independent Test**:
Run `letsgo gui` with no arguments; verify the application opens showing "Sin proyecto activo" with a prominent action to select a folder, and does not point to the binary's directory.

**Acceptance Scenarios**:
1. **Given** no path arguments are passed to `letsgo gui`, **When** the application starts, **Then** `project_dir` is empty (`""`), the header displays "Sin proyecto activo", and an empty-state banner invites the user to open a project folder.
2. **Given** the application is in an unscoped state, **When** the user clicks "Seleccionar carpeta de proyecto" and picks a directory, **Then** that directory becomes the active project, the process working directory changes to it, and a new session can be started in that scope.

---

### User Story 3 - Synchronize Working Directory on Project & Session Selection (Priority: P1)

When a user selects a project from the sidebar, opens a conversation nested under a project, or switches between conversations belonging to different projects, the backend process working directory MUST immediately switch to that project's directory (`os.Chdir`). When the assistant executes tools (`ls`, `cat`, `write_file`, `edit`, `bash`), they execute in the currently active project directory.

**Why this priority**: Resolves the severe bug where asking the AI to list files in a project conversation was reading the binary's folder instead of the project directory.

**Independent Test**:
Create or open Session A with Project A (`/dir-a`), and Session B with Project B (`/dir-b`). Switch to Session A and ask `ls`: it lists `/dir-a`. Switch to Session B and ask `ls`: it lists `/dir-b`.

**Acceptance Scenarios**:
1. **Given** an existing conversation associated with `D:/my-project`, **When** the user clicks to open that conversation in the sidebar, **Then** `SessionsService.Open` sets the backend process working directory to `D:/my-project`, and assistant file tools run inside `D:/my-project`.
2. **Given** recent projects listed in the sidebar, **When** the user clicks a recent project, **Then** the working directory updates to that project and any new chat created under that project runs in that directory.
3. **Given** a session associated with a project whose directory was deleted on disk, **When** the user opens that session, **Then** the application notifies the user with a warning and does not crash.

---

## Edge Cases

- **Path with spaces or quotes**: CLI arguments like `letsgo gui "C:\My Documents\App"` must be handled properly without trimming quotes into the path.
- **Relative paths**: Inputs like `.` or `../other` must be cleanly resolved to absolute paths before passing to the backend.
- **Deleted or unreachable directory**: If a session's recorded project path no longer exists on disk, `os.Chdir` would fail; the system must gracefully fall back to empty project scope and alert the user.
- **Concurrent sessions in different projects**: Switching active tabs between different project sessions updates the process working directory to match the active tab's project.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI command `letsgo gui [path]` MUST accept an optional positional directory argument.
- **FR-002**: When a path argument is provided, the CLI MUST validate that it exists and is a directory; if invalid, it MUST output an error and exit with status 1.
- **FR-003**: When a valid path argument is provided, the application MUST resolve it to an absolute path, set it as the active project directory, and switch the process working directory to it.
- **FR-004**: When no path argument is provided, the application MUST start in an unscoped state (`project_dir = ""`), displaying an empty-project prompt rather than defaulting to the application binary folder.
- **FR-005**: The backend `SessionsService` and `Hub` MUST provide synchronized methods (`GetCurrentProject`, `SetCurrentProject`) to manage and query the active working directory.
- **FR-006**: Opening an existing session via `SessionsService.Open(id)` MUST automatically change the process working directory (`os.Chdir`) to the session's `project_path` if specified and existent.
- **FR-007**: Selecting a project in the UI (sidebar recent project, folder picker, or project menu) MUST notify the backend to change the process working directory via `SetCurrentProject(path)`.
- **FR-008**: All tool executions (`ls`, `cat`, `write_file`, `edit`, `bash`), `GitService`, and file explorer MUST execute against the synchronized active working directory.
- **FR-009**: When no project is active (`project_dir = ""`), the chat interface MUST show a visual prompt indicating that no project folder is open, guiding the user to select one.

### Key Entities

- **ActiveProject**: Represents the current project scope (`path: string`, `name: string`, `exists: bool`).
- **Session**: Retains association with `project_path`.
- **CLI Options**: `initialProjectPath: string`.

## Success Criteria *(mandatory)*

- **SC-001**: Running `letsgo gui <dir>` opens the GUI with `<dir>` active in under 2 seconds.
- **SC-002**: 100% of assistant tool calls (`ls`, `cat`, `bash`) execute inside the selected project directory instead of the program root.
- **SC-003**: Launching `letsgo gui` without arguments starts with 0 active project folders, prompting the user cleanly.
- **SC-004**: Switching sessions between different projects switches the underlying tool execution context reliably without restarting the application.

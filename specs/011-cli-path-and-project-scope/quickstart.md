# Quickstart Validation Guide: CLI Path Argument and Project Scope

## Validation Scenarios

### Scenario 1: Launch GUI with CLI path argument
1. In PowerShell:
   ```powershell
   go run . gui .
   ```
2. Check that the GUI opens.
3. Observe header shows current directory name (e.g. `Letsgo-code`).
4. Type in chat: `lista los archivos con ls`.
5. Expected outcome: Assistant invokes `ls` and returns files of `Letsgo-code` (e.g. `cmd/`, `internal/`, `go.mod`, etc.).

### Scenario 2: Launch GUI without arguments (unscoped)
1. Run:
   ```powershell
   go run . gui
   ```
2. Check that the GUI opens.
3. Observe header shows "Sin proyecto activo" (or localized equivalent).
4. Observe empty project prompt encouraging the user to select or open a folder.
5. Click "Seleccionar carpeta de proyecto" and pick a folder.
6. Observe the app switches to that project.

### Scenario 3: Switching sessions with different projects
1. Have two sessions in database:
   - Session A with `project_path = "D:/PathA"`
   - Session B with `project_path = "D:/PathB"`
2. Click Session A in sidebar: verify process switches working directory to `D:/PathA`.
3. Click Session B in sidebar: verify process switches working directory to `D:/PathB`.
4. Ask `ls`: verify files from `D:/PathB` are listed.

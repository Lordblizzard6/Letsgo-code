# Implementation Plan: CLI Path Argument and Project Scope Synchronization

**Branch**: `011-cli-path-and-project-scope` | **Date**: 2026-09-04 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/011-cli-path-and-project-scope/spec.md`

## Summary

Enable `letsgo gui [path]` positional argument to launch the GUI with a specified project directory, start unscoped without pointing to the program files when omitted, and synchronize the process working directory across session switching, sidebar project selection, and tool executions so that AI tools (like `ls`) operate within the selected project.

## Technical Context

**Language/Version**: Go 1.26, TypeScript / React 18
**Primary Dependencies**: Wails v3 (`github.com/wailsapp/wails/v3`), Cobra (`github.com/spf13/cobra`), Lucide React
**Storage**: SQLite (`internal/db/database.go`)
**Testing**: `go test ./...`, Vitest (`npm test`)
**Target Platform**: Windows (desktop), cross-platform compatible
**Project Type**: Desktop GUI / CLI hybrid
**Performance Goals**: Instant project switch (<100ms), no UI freeze
**Constraints**: No GPL3 libraries, no drive-by refactoring, test-last principle

## Constitution Check

- **I. Test-Last Verification**: Automated tests for `SessionsService.SetCurrentProject`, `Open`, and `cmd/gui.go` will be executed and verified at the end of implementation.
- **II. Contract-Only Core Access**: GUI communicates through `SessionsService` and `Hub` events, preserving the contract boundary.
- **III. One Source of Truth**: Active working directory and project state are managed in `Hub`/`SessionsService` and backed by `os.Chdir`.
- **IV. Secure Rendering**: All directory names and paths rendered in the UI continue to be sanitized.
- **V. Simplicity & YAGNI**: Direct, minimal changes to `cmd/gui.go`, `cmd/wails/gui/gui.go`, `cmd/wails/services/sessions_service.go`, `cmd/wails/services/hub.go`, and `cmd/wails/frontend/src/App.tsx`.

## Project Structure

### Documentation (this feature)

```text
specs/011-cli-path-and-project-scope/
├── spec.md
├── checklists/
│   └── requirements.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
└── contracts/
    └── cli-and-session-contract.md
```

### Source Code Changes

```text
cmd/
└── gui.go                                 # Add optional [path] argument, validate and pass to wailsgui.RunWithProject
cmd/wails/
├── gui/
│   └── gui.go                             # Add RunWithProject(assets, initialPath) and wire initial project
├── services/
│   ├── hub.go                             # Add currentProject, SetCurrentProject, GetCurrentProject
│   ├── sessions_service.go                # Expose GetCurrentProject/SetCurrentProject, update Open(id) to os.Chdir
│   └── sessions_service_test.go           # Unit tests for project path synchronization
cmd/wails/frontend/src/
└── App.tsx                                # Synchronize project on mount/click, show empty state prompt when unscoped
```

## Planned Phases

- **Phase 0: Research & Problem Diagnosis** (Completed in `research.md`)
- **Phase 1: Contracts & Data Model** (Completed in `contracts/` and `data-model.md`)
- **Phase 2: Backend CLI and Services Implementation**
  - Update `cmd/gui.go` to accept `[path]`.
  - Update `cmd/wails/gui/gui.go` with `RunWithProject(assets, initialPath)`.
  - Update `cmd/wails/services/hub.go` and `sessions_service.go` with project working directory tracking and `os.Chdir`.
  - Update `SessionsService.Open(id)` to switch working directory when resuming a project session.
- **Phase 3: Frontend Synchronization & Empty Project UX**
  - On startup: fetch initial project from `SessionsService.GetCurrentProject()`.
  - Listen for `project:changed` event.
  - When switching saved session: update active project.
  - When clicking recent projects: call `SessionsService.SetCurrentProject(path)`.
  - When `project_dir === ""`: render a clean, friendly banner prompting the user to open a project folder.
- **Phase 4: Verification & Test-Last Gate**
  - Run `go test ./...`
  - Run `npm test`
  - Verify build with `go build -o letsgo.exe ./cmd/wails`

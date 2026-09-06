# Tasks: CLI Path Argument and Project Scope Synchronization

**Feature Branch**: `011-cli-path-and-project-scope` | **Date**: 2026-09-04 | **Spec**: [spec.md](spec.md) | **Plan**: [plan.md](plan.md)

## Phase 1: Setup
- [X] T001 Setup working directory tracking and helper methods in `cmd/wails/services/hub.go`

## Phase 2: Foundational
- [X] T002 Implement `RunWithProject(assets fs.FS, initialProjectDir string)` in `cmd/wails/gui/gui.go`

## Phase 3: User Story 1 - Launch GUI with Project Path from CLI (P1)
- [X] T003 [US1] Update `cmd/gui.go` to accept optional `[path]`, validate directory existence, and pass to `RunWithProject`

## Phase 4: User Story 2 - Launch Without Directory / Unscoped Empty State (P1)
- [X] T004 [US2] Expose `GetCurrentProject` and `SetCurrentProject` on `SessionsService` in `cmd/wails/services/sessions_service.go`
- [X] T005 [US2] Add empty-project prompt banner and onboarding state in `cmd/wails/frontend/src/App.tsx` when `project_dir === ""`

## Phase 5: User Story 3 - Synchronize Working Directory on Project & Session Selection (P1)
- [X] T006 [US3] Update `SessionsService.Open(id)` and `SelectProjectFolder` to synchronize `os.Chdir` and active project in `cmd/wails/services/sessions_service.go`
- [X] T007 [US3] Update frontend session and project selection to invoke `SetCurrentProject` and react to `project:changed` in `cmd/wails/frontend/src/App.tsx`

## Phase 6: Polish & Verification (Test-Last per Constitution Principle I)
- [X] T008 Add unit tests for project path synchronization in `cmd/wails/services/sessions_service_test.go`
- [X] T009 Run full test suite (`go test ./...` and `npm test`)
- [X] T010 Build binary and verify end-to-end (`go build -o letsgo.exe ./cmd/wails`)

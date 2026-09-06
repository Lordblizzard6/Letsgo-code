# AGENTS.md

Operating instructions and project rules for coding agents in this repository.

## 0. Non-negotiables
1. Working code only. Finish the job. Plausibility is not correctness.
2. Direct, concise communication. No flattery, no filler.
3. Surgical changes: Touch only what you must. No drive-by refactoring or unnecessary format changes.
4. Goal-driven: Verify your changes with tests or build commands before reporting completion.

## 1. Project Context
- **Stack**: Go
- **Install**: go mod download
- **Build**: go build ./...
- **Test**: go test ./...
- **Lint**: golangci-lint run

## 2. Conventions & Style
- Follow existing patterns in the codebase.
- Error handling: return errors to callers, do not silently swallow them.
- Simplicity first: write the minimum code required to solve the task.

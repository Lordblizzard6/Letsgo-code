# Feature Specification: Modernización y Nuevas Funciones de la TUI

**Feature Branch**: `014-tui-modernization-and-features`
**Created**: 2026-09-05
**Status**: Draft
**Input**: User description: "crea un Spec para mejorar las funciones y modernizar completamente la TUI"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Modern Interactive Layout & Real-time Status Bar (Priority: P1) 🎯 MVP

As a developer using LetsGO Code in the terminal,
I want an aesthetically modern, responsive TUI layout with styled headers, real-time status bars, and rich markdown rendering,
So that my terminal experience feels polished, clean, and shows vital session context (active mode, model, git branch, tokens, and cost) at a glance.

**Why this priority**:
The core user interface is the daily driver for terminal users. Having clear visual hierarchy, dynamic responsive widths, and live telemetry immediately elevates the application quality.

**Independent Test**:
Launch `letsgo` in terminal. Verify that the header renders project path, session ID, git status, and the bottom status bar displays colored mode badges (BUILD/PLAN/RESEARCH), active model, token counters, and cost without layout glitches or overflow.

**Acceptance Scenarios**:
1. **Given** an active repository, **When** launching the TUI, **Then** the header displays the repository name, branch, clean/dirty status, and session ID.
2. **Given** the chat view, **When** pressing `Tab` while input is empty, **Then** the active mode cycles between `BUILD`, `PLAN`, and `RESEARCH`, updating the status badge color dynamically and notifying the engine.
3. **Given** model responses, **When** text streams or finishes, **Then** markdown and code blocks render with syntax styling via Glamour, with accurate line wrapping.

---

### User Story 2 - Dynamic Provider & Model Switcher with Discovery (Priority: P1)

As a developer,
I want to browse, search, and switch AI providers and models dynamically inside the TUI (via `Ctrl+S`, `/model`, or `/models`),
So that I am not limited to a hardcoded list and can use newly configured models (e.g. Gemini, Anthropic, OpenRouter, Ollama, DeepSeek, Groq) directly without restarting.

**Why this priority**:
Currently `internal/tui/ui.go` relies on hardcoded arrays of model names and counts, which contradicts the core dynamic discovery engine (`internal/api/discovery.go`) and prevents users from selecting custom or local models.

**Independent Test**:
Press `Ctrl+S` or run `/model`. Verify that configured providers and dynamically discovered models (including Ollama local models and OpenRouter models) appear in an interactive selector with search/filtering, and selecting one immediately updates `config.AppConfig.Model` and engine client.

**Acceptance Scenarios**:
1. **Given** the chat screen, **When** pressing `Ctrl+S`, **Then** an interactive model palette opens, displaying available providers and their available models fetched dynamically.
2. **Given** the model list, **When** navigating with arrows or typing a filter query, **Then** the list filters in real-time.
3. **Given** a selected model, **When** pressing `Enter`, **Then** the model is persisted to configuration, the active engine client is refreshed, and a status message confirms the change in chat.

---

### User Story 3 - Interactive Slash Command & Mention Autocomplete Overlay (Priority: P2)

As a developer typing prompts in the TUI,
I want an interactive autocomplete popup list when typing `/` or `@`,
So that I can discover and select commands (`/help`, `/diff`, `/review`, `/cost`, `/tasks`, `/terminal`, etc.) and context mentions (`@file`, `@skill`, etc.) with keyboard navigation.

**Why this priority**:
Terminal usability is significantly enhanced when commands don't have to be memorized, providing complete parity with the GUI composer autocomplete experience.

**Independent Test**:
Type `/` in the TUI textarea. Verify an autocomplete popup appears above the input listing matching slash commands with descriptions, navigable with Up/Down/Tab/Enter and dismissible with Esc.

**Acceptance Scenarios**:
1. **Given** the input textarea, **When** typing `/`, **Then** a floating/styled suggestions box appears listing all slash commands with their icons and descriptions.
2. **Given** the suggestions box, **When** typing additional characters (e.g. `/di`), **Then** the list dynamically filters down (e.g. `/diff`).
3. **Given** a highlighted suggestion, **When** pressing `Tab` or `Enter`, **Then** the command completes in the input box and the suggestion popup closes.
4. **Given** the input textarea, **When** typing `@`, **Then** file and mention suggestions appear.

---

### User Story 4 - TUI Git Review, Staging & Diff Overlay (Priority: P2)

As a developer modifying code with the assistant in the terminal,
I want a dedicated interactive Git Review view (`Ctrl+D` or `/diff` / `/review`),
So that I can inspect modified files, stage/unstage changes, view unified colored diffs, and create commits without leaving the TUI.

**Why this priority**:
Parity with Antigravity CLI and the GUI's Git Review tab allows full revision workflow directly from the command line interface.

**Independent Test**:
Execute `/diff` or press `Ctrl+D`. Verify a dual-pane or modal Git Review screen opens showing unstaged and staged files with numstat (+Lines/-Lines), diff preview pane, and stage/unstage/commit actions.

**Acceptance Scenarios**:
1. **Given** uncommitted repository changes, **When** pressing `Ctrl+D` or executing `/diff`, **Then** the Git Review view opens displaying staged and unstaged files with addition/deletion counts.
2. **Given** an unstaged file, **When** pressing `+` or `s`, **Then** the file is staged in Git index and the view refreshes.
3. **Given** a staged file, **When** pressing `-` or `u`, **Then** the file is unstaged.
4. **Given** a selected file, **When** pressing `Enter` or `v`, **Then** the unified diff is displayed with line numbers, hunk headers, and syntax colors.
5. **Given** changes ready to commit, **When** pressing `c`, **Then** a commit prompt allows writing a message and committing immediately.

---

### User Story 5 - Activity & Tool Execution Inspector (Priority: P2)

As a developer monitoring assistant actions,
I want an Activity Inspector view (`Ctrl+B` or `/tasks` / `/terminal`),
So that I can see the history of executed tool calls, command outputs, background tasks, and duration.

**Why this priority**:
Complex agent runs execute numerous commands and tools. Being able to inspect recent tool calls, timings, and stdout/stderr in a dedicated overlay prevents terminal clutter.

**Independent Test**:
Run tools in chat and press `Ctrl+B`. Verify that executed tools are listed with elapsed duration, parameters, exit status, and expandable output previews.

**Acceptance Scenarios**:
1. **Given** executed tools in a session, **When** pressing `Ctrl+B`, **Then** the Activity view opens showing a timeline of tool calls.
2. **Given** a file modification tool call, **When** selecting it, **Then** an embedded diff preview shows the lines changed.
3. **Given** active background tasks, **When** inspecting activities, **Then** their status and IDs are visible.

---

### User Story 6 - Session & Project Switcher Overlay (Priority: P3)

As a developer managing multiple tasks,
I want an interactive session and project switcher (`/sessions`, `/open`, `/new`, `/project`),
So that I can quickly switch between projects or resume previous chat sessions without restarting `letsgo`.

**Why this priority**:
Improves workflow continuity and multi-tasking across different directories and historical sessions.

**Independent Test**:
Type `/sessions`. Verify an interactive list of saved sessions opens with message counts and timestamps, and selecting one restores its conversation and working directory.

**Acceptance Scenarios**:
1. **Given** saved sessions in SQLite, **When** executing `/sessions`, **Then** an interactive table opens showing session ID, name, date, message count, and project path.
2. **Given** the session table, **When** selecting a session and pressing `Enter`, **Then** the TUI switches to that session and restores its conversation history.

---

## Requirements

### Functional Requirements

- **FR-001**: Modernize the top header with workspace project name, current git branch, repository status (clean/dirty), session short ID, and theme accent.
- **FR-002**: Modernize the bottom status bar with a 2-tier layout: upper tier for Mode indicator (`BUILD`, `PLAN`, `RESEARCH`) with Tab-toggle, active model name, token counters, and cost; lower tier for keyboard shortcut hints.
- **FR-003**: Replace hardcoded models in `internal/tui/ui.go` with dynamic discovery integrating `internal/api/discovery.go` and `internal/config/config.go`.
- **FR-004**: Implement an interactive model and provider selection overlay (`Ctrl+S` / `/model`) supporting search/filtering across all active providers (Anthropic, Gemini, OpenAI, Groq, OpenRouter, DeepSeek, Ollama).
- **FR-005**: Implement an interactive popup suggestions menu for slash commands and context mentions (`/` and `@`) with arrow navigation, fuzzy filtering, and Tab auto-completion.
- **FR-006**: Implement an interactive Git Review view (`Ctrl+D` / `/diff` / `/review`) with unstaged/staged sections, stage/unstage hotkeys, unified diff viewer, and commit composer.
- **FR-007**: Implement an Activity & Tool Inspector (`Ctrl+B` / `/tasks` / `/terminal`) showing execution history, elapsed time, status badges, and diffs.
- **FR-008**: Implement an interactive Session & Project browser (`/sessions`, `/open`, `/new`) with keyboard navigation and instant resume.
- **FR-009**: Ensure TUI supports terminal window resizing gracefully with adaptive widths and proper word wrapping.
- **FR-010**: Support theme variants (dark, light, dracula, nord) matching user configuration.

### Non-Functional Requirements

- **NFR-001**: Zero new third-party heavy dependencies; use established Charmbracelet ecosystem libraries (`bubbletea`, `lipgloss`, `bubbles`, `glamour`).
- **NFR-002**: High responsiveness: modal/overlay transitions and filter filtering must render in < 16ms (60 FPS feel in terminal).
- **NFR-003**: Graceful degradation in small terminals (fallback to single column when terminal width < 80 columns).
- **NFR-004**: Clean separation of presentation state from core domain models per Constitution Principle II & III.

---

## Success Criteria

1. Terminal users can inspect and change models across all enabled providers without hardcoded limitations.
2. Full command parity between TUI and GUI for slash commands and workflow navigation.
3. Git review and staging operations can be performed entirely within the TUI without opening external git commands.
4. Window resizing adapts smoothly without clipping text or corrupting terminal ANSI escape codes.
5. All existing and new automated tests pass (`go test ./internal/tui/...`).

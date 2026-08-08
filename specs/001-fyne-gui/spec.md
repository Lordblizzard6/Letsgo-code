# Feature Specification: Fyne Desktop GUI for LetsGO Code

**Feature Branch**: `Lets-go-Fyne`

**Created**: 2026-08-02

**Status**: Draft

**Input**: User description: "quiero crear una app a partir de esta, pero con una interfaz hecha en Fyne, ya que este es un proyecto go, crea una ramificación del proyecto llamada 'Lets-go-Fyne' para hacer una gui en fyne"

## Clarifications

### Session 2026-08-02

- Q: Alcance de v1 — ¿el bucle de agente tipo Codex entra en v1? → A: v1 = paridad 1:1 completa (chat, agente con herramientas y aprobaciones, git, agentes paralelos, MCP, plugins)
- Q: Resemblanza a Codex — ¿qué alcance tiene la semejanza? → A: Estética de terminal oscura (tema oscuro, tipografía monoespaciada) + patrones de interacción de Codex (flujos de aprobación, paneles de actividad de herramientas, atajos de teclado)
- Q: Arquitectura modular y escalable — ¿qué estructura prefieres? → A: Núcleo compartido modular (paquetes de dominio: providers, tools, sesiones, agente) consumido por igual por la GUI y el CLI

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Chat with the AI assistant in a desktop window (Priority: P1)

The user launches the application and sees a desktop window with a conversation area and a text input. They write a message, press send, and the AI's response appears in the conversation thread. The response is displayed as it arrives, and the user can keep the conversation going with follow-up messages. If they make a mistake, they can cancel an in-progress response.

**Why this priority**: Chatting with the AI is the core purpose of the application. Without a working chat conversation, the GUI delivers no value to the user. Every other feature (sessions, settings) exists to support this activity.

**Independent Test**: Can be fully tested by launching the app, typing a message, and confirming an AI response is displayed in the window. This delivers the primary value: using the AI assistant without the terminal.

**Acceptance Scenarios**:

1. **Given** the application is running with a valid provider configuration, **When** the user types a message and presses Send, **Then** the message appears in the conversation thread and an AI response is displayed within a reasonable time
2. **Given** a response is being generated, **When** the user clicks Cancel, **Then** the response stops and the input becomes editable again
3. **Given** the user has sent multiple messages, **When** they review the window, **Then** all prior messages and responses remain visible in chronological order
4. **Given** an error occurs (no network, invalid credentials), **When** the user sends a message, **Then** a clear, friendly error message is shown in the conversation and the app remains usable

---

### User Story 2 - Assign coding tasks to the agent with tool approvals (Priority: P1)

The user assigns a coding task ("fix this bug", "add a feature") and the agent works autonomously: it proposes file edits, runs commands in the project, and searches the web, requesting approval for each sensitive action. The user can approve, reject, or let the agent run in auto-approve mode. All activity (tool calls, results, diffs) is visible in the conversation as it happens.

**Why this priority**: Autonomous task execution with tools is the defining capability of Codex and the reason the goal is 1:1 parity. Without it, the GUI is a chat toy, not a coding assistant.

**Independent Test**: Can be fully tested by assigning a task that requires a file edit and a command run; the agent performs both with user approval visible in the GUI, delivering the core Codex-like value.

**Acceptance Scenarios**:

1. **Given** the user assigns a coding task, **When** the agent needs to edit a file, **Then** the proposed change (with diff) is shown and the user must approve or reject it before it is applied
2. **Given** the agent requests a tool action, **When** the user rejects it, **Then** the agent receives the rejection, adjusts, and continues without applying the action
3. **Given** the user enables auto-approve mode, **When** the agent requests actions, **Then** they are executed without per-action prompts, and the user can see each executed action in the activity log
4. **Given** the agent runs a command, **When** the command completes, **Then** its output is shown inline in the conversation

---

### User Story 3 - Manage conversations as sessions (Priority: P2)

The user can start a new conversation, see a list of past conversations, and reopen any of them to continue where they left off. Conversations persist across application restarts.

**Why this priority**: Long-lived conversations are essential for real coding work — users switch tasks and return to previous contexts. It can be delivered after basic chat works, as an independently valuable slice.

**Independent Test**: Can be fully tested by starting a conversation, closing the app, relaunching, and confirming the conversation appears in the list and can be reopened with its full history.

**Acceptance Scenarios**:

1. **Given** the user is in a conversation, **When** they start a new conversation, **Then** an empty conversation is created and the previous one remains in the session list
2. **Given** the app was closed with an active conversation, **When** the user relaunches, **Then** the conversation history is still available and can be resumed
3. **Given** a session list with several conversations, **When** the user selects one, **Then** its full message history is displayed in the conversation area

---

### User Story 4 - Use advanced capabilities: git, parallel agents, MCP, plugins (Priority: P3)

The user can perform git operations (branch, commit, diff, review) directly from the GUI, spawn parallel sub-agents for independent subtasks, connect external tools through MCP servers, and manage plugins. Each capability is a separately discoverable view or action within the app.

**Why this priority**: These capabilities complete the 1:1 parity goal with Codex, but they build on the agent and chat foundation, so they can be delivered after the core is stable.

**Independent Test**: Can be fully tested by connecting an MCP server and confirming its tools are usable in a conversation, or by running a git commit flow entirely from the GUI.

**Acceptance Scenarios**:

1. **Given** the user is in a project with git history, **When** they use the git view, **Then** they can create a branch, commit changes, view diffs, and trigger a review without leaving the app
2. **Given** a large task, **When** the user spawns parallel agents, **Then** each agent works independently and its results are reported in the conversation
3. **Given** an MCP server is configured, **When** the user asks for a capability it provides, **Then** the MCP tools are invoked as normal tool calls with the same approval flow
4. **Given** a plugin is installed, **When** the app restarts, **Then** the plugin is loaded and its commands/contributions are available

---

### User Story 5 - Configure providers and monitor usage (Priority: P3)

The user opens a settings area where they can select the AI provider, enter or update their API key, and choose a model. A usage/cost view shows how much of their allowance has been consumed.

**Why this priority**: Configuration is a prerequisite for using the app, but for users who already have credentials the app can offer reasonable defaults. Usage monitoring is a supporting feature that adds confidence without blocking core usage.

**Independent Test**: Can be fully tested by entering a new provider configuration, restarting the app, and confirming the configuration is retained and a chat works with it.

**Acceptance Scenarios**:

1. **Given** the settings area is open, **When** the user selects a provider and enters a valid API key, **Then** the configuration is saved and used for subsequent conversations
2. **Given** a saved configuration exists, **When** the user relaunches the app, **Then** the configuration is still active
3. **Given** conversations have taken place, **When** the user opens the usage view, **Then** cost and token usage are displayed for the current period

---

### Edge Cases

- What happens when the application is launched with no provider configured? The app should open to a configuration prompt instead of a broken chat.
- What happens when the network connection is lost mid-response? The response stops gracefully with an error notice, and the partial content remains visible.
- What happens when the API key is invalid or the provider is unreachable? A clear error message with guidance is shown; the app does not crash.
- What happens when the user closes the window while a response is generating? The conversation is saved up to the last complete message; no corruption of session data.
- What happens with an extremely long conversation? The window remains responsive; the full history stays accessible.
- What happens when a tool action awaits approval and the user never responds? The request times out after a defined period, with a clear notice, and the agent can be resumed manually
- What happens when the user rejects a critical tool action? The agent receives the rejection and must not apply the action; the conversation continues with the agent adjusting its plan
- What happens when an MCP server is unreachable or fails mid-call? A friendly error is shown, the failure is isolated to that tool, and the rest of the conversation keeps working
- What happens when two parallel agents try to modify the same file? Conflicts are surfaced to the user with clear messaging and no silent overwrite
- What happens when a git operation fails (conflict, dirty working tree)? The GUI shows the underlying error and recovery options without crashing
- What happens when the user wants to use the app without a mouse? All primary actions remain reachable through keyboard shortcuts and focus navigation

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The application MUST provide a desktop window with a conversation area and a message input field
- **FR-002**: Users MUST be able to send a message by pressing a Send button or keyboard shortcut
- **FR-003**: The system MUST display AI responses in the conversation thread, progressively as they are generated
- **FR-004**: Users MUST be able to cancel an in-progress response
- **FR-005**: The system MUST persist conversations so they survive application restarts
- **FR-006**: Users MUST be able to create a new conversation and return to any previous conversation from a visible list
- **FR-007**: Users MUST be able to select an AI provider, enter an API key, and choose a model from a settings screen
- **FR-008**: The system MUST retain provider configuration between application runs
- **FR-009**: The system MUST display usage and cost information for the current billing period
- **FR-010**: The system MUST show user-friendly error messages for network failures, invalid credentials, and provider errors without terminating the application
- **FR-011**: The application MUST launch to a usable state even when no configuration exists, guiding the user to configure it
- **FR-012**: The system MUST save the conversation up to the last complete message when the application closes during an active response
- **FR-013**: The system MUST support an agent mode in which the user assigns a coding task and the agent proposes file edits, runs commands, and searches the web
- **FR-014**: Every sensitive agent action (file modification, command execution) MUST require explicit user approval before execution, unless the user has enabled auto-approve mode for that action category
- **FR-015**: The system MUST show proposed changes (diffs) and command output inline in the conversation before and after execution
- **FR-016**: Users MUST be able to reject or cancel any pending agent action; rejected actions MUST NOT be applied
- **FR-017**: The system MUST provide git operations (branch, commit, diff, review) from within the application
- **FR-018**: Users MUST be able to spawn parallel sub-agents for independent subtasks, with results reported back into the conversation
- **FR-019**: The system MUST support connecting external tools via MCP servers, with MCP tools flowing through the same approval mechanism as built-in tools
- **FR-020**: Users MUST be able to install, list, and manage plugins from the application, with plugins loaded on subsequent launches
- **FR-021**: The GUI MUST expose every capability available in the existing command-line application (feature parity, no functional regression)
- **FR-022**: Provider access MUST be abstracted so no capability depends on a single AI provider; switching providers MUST NOT change the available feature set
- **FR-023**: The application MUST present a dark, terminal-inspired visual style with monospaced typography as the default appearance
- **FR-024**: Agent activity (tool calls, diffs, command output, approvals) MUST be presented in dedicated inline panels in the conversation, not as plain unstructured text
- **FR-025**: Users MUST be able to perform all primary actions (send message, approve/reject tool actions, switch sessions, cancel tasks) via keyboard shortcuts without requiring a mouse
- **FR-026**: The GUI and the existing command-line interface MUST consume the same shared core modules, so that any capability available in one is available in the other without duplication

### Key Entities *(include if feature involves data)*

- **Session**: A named conversation containing an ordered list of messages; persists across restarts; created, listed, reopened, renamed by the user
- **Message**: A single exchange within a session (user input, assistant response, or tool activity); carries a role, content, and timestamp
- **Tool Call**: A proposed agent action (file edit, command, web search, MCP tool) carrying a status (pending-approval, approved, rejected, executed, failed) and its result
- **Agent Task**: A unit of autonomous work assigned by the user (or a spawned sub-agent); tracks its assigned session, tool calls, and completion status
- **MCP Configuration**: The set of connected external MCP servers, their availability status, and the tools each exposes
- **Plugin**: An installable extension contributing commands or capabilities; managed from a plugin view
- **Provider Configuration**: The user's selected AI provider, API key, and model; one active configuration at a time; persisted locally
- **Usage Record**: Token counts and cost per provider for the current billing period; aggregated for display in the usage view

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can send a message and see the AI's response begin within 5 seconds on a typical broadband connection
- **SC-002**: 100% of conversations created in the GUI are available for resuming after an application restart
- **SC-003**: 90% of first-time users can configure the provider and complete their first conversation without external documentation
- **SC-004**: The application window appears and is interactive within 3 seconds of launch on typical desktop hardware
- **SC-005**: 100% of provider/network error cases result in a visible friendly message rather than a crash or silent failure
- **SC-006**: 100% of capabilities present in the existing command-line application are reachable from the GUI (verified against a parity checklist)
- **SC-007**: Every file modification performed by the agent is preceded by an explicit user approval (unless auto-approve is enabled), with zero silent or unapproved modifications
- **SC-008**: Users can switch AI providers without losing access to any feature; the feature set is identical across all supported providers
- **SC-009**: 100% of primary actions can be completed using only the keyboard, verified by a keyboard-only walkthrough of the main flows

## Assumptions

- The graphical interface will be built with the Fyne GUI toolkit, as explicitly requested by the user; this is a stated requirement, not an implementation choice
- The GUI is a new presentation layer for the existing application: the command-line interface continues to work and is not modified in this feature
- Target platforms are desktop operating systems (Windows, macOS, Linux)
- Version 1 delivers full feature parity with the reference coding assistant (Codex): chat, agent mode with tool approvals, sessions, git operations, parallel agents, MCP integration, and plugin management
- The application is not tied to a single AI provider: all features work with any supported provider, and provider access is isolated behind a common abstraction so the set of providers can grow without architectural change
- The architecture is a shared modular core: domain packages (providers, tools, sessions, agent) are consumed equally by the GUI and the existing CLI, so adding a provider or capability does not require touching other modules
- Existing persistent storage and core assistant logic from the current application will be reused rather than recreated
- Users have an account or API key with at least one supported AI provider before first use
- The existing multi-provider support (Anthropic, OpenAI, Groq, OpenRouter, Ollama) is preserved in the GUI

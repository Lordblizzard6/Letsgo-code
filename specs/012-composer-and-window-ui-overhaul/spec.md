# Feature Specification: Composer Layout and Window Sizing Antigravity Parity

**Feature Branch**: `012-composer-and-window-ui-overhaul`

**Created**: 2026-09-04

**Status**: Draft

**Input**: User description: "ahora crea un nuevo spec para hacer unos cambios a la aperiencia del colector de texto, el tamaño inicial de la ventana, y el boton de nueva conversacion moverlo a la navbar lateral, queiro que el tamaño de la ventana del gui tenga el mismo tamaño que antigravity, y el mismo tamaño de texto y disposicion, un visor de comandos al mismo estilo que antigravity tambien, y que el modo Plan/Codigo/Arch y research esten en la esquina superior del colector y que no sean un boton sino que se cambien con tab(obvio que cuando se este escribiendo un comando / o un @mention no se pueda alternar a los modos con tab, y que el selector de modelos y el boton de enviar y el de adjuntar esten en la misma disposicion que el antigrativy gui: / (posible contenido de texto / /boton adjuntar / Selector de modelo / --------------------------------------- Boton de enviar y en general hacer mas funcional y appealing el colector, lo mas parecido a antigravity cli pero con el flujo de trabajo de Opencode y su loop"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Antigravity Window Sizing, Typography Scale & Sidebar New Chat Button (Priority: P1)

Users opening the desktop application expect an immersive, comfortable workspace matching modern developer AI environments like Antigravity IDE. The window opens with optimal default dimensions (1280x850 with min 960x600). The lateral navbar / sidebar features a prominent "Nueva conversación" button at the top of the conversation list. The overall font scale and line-heights are aligned with Antigravity's clean typographic hierarchy.

**Why this priority**: Sets the foundational viewport and primary navigation action before refining the composer.

**Independent Test**:
Launch `letsgo gui`. Verify window opens with 1280x850 dimensions, sidebar contains a visible "+ Nueva conversación" button, and clicking it creates an active new chat tab.

**Acceptance Scenarios**:
1. **Given** the application is launched, **When** the desktop window initializes, **Then** its default size is 1280x850 pixels with a minimum constraint of 960x600.
2. **Given** the sidebar is open, **When** the user views the sidebar, **Then** a prominent "+ Nueva conversación" button is located at the top of the conversations section.
3. **Given** the user clicks "+ Nueva conversación", **Then** a new tab is created and focused immediately.

---

### User Story 2 - Work Modes Tabs in Composer Header with Tab-Key Navigation (Priority: P1)

In the composer box, the work modes (Plan, Código, Arch, Research) are located in the top corner as modern segmented tabs or pills. Instead of opening a dropdown menu, pressing the `Tab` key cycles through the modes (Código → Arch → Plan → Research → Código, and `Shift+Tab` in reverse), updating the active mode immediately with visual feedback and soundless keyboard fluidness.

**Why this priority**: Enhances speed of interaction during development, letting users switch context without touching the mouse.

**Independent Test**:
Focus the composer textarea and press `Tab`. Verify the active work mode shifts to the next mode without losing focus.

**Acceptance Scenarios**:
1. **Given** the user is focused in the composer textarea with no autocomplete open, **When** the user presses `Tab`, **Then** the active mode advances to the next work mode (`code` → `architecture` → `planning` → `research`), and focus remains in the textarea.
2. **Given** the user presses `Shift+Tab`, **Then** the active mode cycles backwards.
3. **Given** the user clicks directly on any mode tab in the top corner, **Then** that mode becomes active.
4. **Given** the user is typing a `/` command or `@` mention and the command suggestion popup is visible, **When** the user presses `Tab`, **Then** `Tab` selects or autocompletes the highlighted command/mention, and DOES NOT cycle the work mode.

---

### User Story 3 - Antigravity Composer Controls Layout (Priority: P1)

The composer controls are rearranged to match the Antigravity GUI layout:
- Top: Work mode tabs (`[Plan] [Código] [Arch] [Research]`) with a small `Tab ⇥` key shortcut indicator.
- Center: Multi-line message input area with clean placeholder and auto-expanding height.
- Bottom Bar:
  - Left: Attach file button (`[📎]`) followed immediately by the Model Selector dropdown (`[🤖 Model Name ▾]`).
  - Middle: Flexible spacer / divider.
  - Right: Send button (`[↑ Enviar]` / arrow icon) styled with accent colors.

**Why this priority**: Delivers the exact user requested layout, improving ergonomic flow and grouping file inputs and model configuration together while isolating the send action.

**Independent Test**:
Inspect the composer box. Verify attach and model selector are on the bottom-left, send button is on the bottom-right, and work modes are at the top.

**Acceptance Scenarios**:
1. **Given** the composer is rendered, **When** inspecting the bottom toolbar, **Then** the attach button and model selector dropdown are positioned on the bottom-left, and the send button is positioned on the bottom-right.
2. **Given** the user clicks the model selector dropdown, **Then** a popover lists available models for the provider, allowing one-click selection that updates the active model.
3. **Given** files are attached, **Then** attachment chips display above the text area with size and remove actions.

---

### User Story 4 - Antigravity Style Command Palette & Autocomplete Viewer (Priority: P2)

When the user types `/` in the composer, an elevated command viewer appears above the composer. It displays commands formatted with icons, formatted command names, description text, and keyboard navigation cues (`↑`, `↓`, `Tab`, `Enter`).

**Why this priority**: Replaces basic command lists with a modern, high-polish command experience matching Antigravity CLI and Opencode.

**Independent Test**:
Type `/` in the composer textarea. Verify the popover opens with command icons, descriptions, and keyboard navigation.

**Acceptance Scenarios**:
1. **Given** the composer textarea contains `/`, **When** typing characters, **Then** the command palette filters matching commands with title, icon, and description.
2. **Given** the palette is visible, **When** using Arrow Up/Down, **Then** the highlighted command changes; pressing `Enter` or `Tab` inserts the command into the input.

---

## Edge Cases

- **Tab key conflict**: If the command palette is open, `Tab` must complete the suggestion, not switch modes. If the textarea contains a tab intent (e.g. user selected text or in code block), mode switching must only trigger when autocomplete is idle.
- **Small screens or sidebar collapse**: The composer toolbar must wrap gracefully or maintain minimum touch targets if the window is shrunk to minimum width (960px).
- **Long model names**: Model names must truncate cleanly in the model selector button with an ellipsis (`...`) so they never push the send button off-screen.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The initial GUI window size MUST be set to 1280x850 with a minimum width of 960 and minimum height of 600 in `cmd/wails/gui/gui.go`.
- **FR-002**: A prominent "Nueva conversación" action button MUST be rendered at the top of the sidebar conversation list in `App.tsx`.
- **FR-003**: The work mode selector MUST be located in the top corner of the composer container, displaying Plan, Código, Arch, and Research.
- **FR-004**: Pressing `Tab` inside the composer textarea MUST cycle to the next work mode, provided the command/mention palette is NOT open.
- **FR-005**: Pressing `Shift+Tab` inside the composer textarea MUST cycle to the previous work mode when autocomplete is not active.
- **FR-006**: When autocomplete palette is open, `Tab` MUST apply the highlighted suggestion and NOT change the work mode.
- **FR-007**: The bottom toolbar of the composer MUST place the attach file button and model selector on the left, and the send button on the right.
- **FR-008**: The command palette viewer (`/`) MUST display each command with an icon, name, description, and keyboard shortcut cue.
- **FR-009**: The composer container MUST use rounded corners (12-16px), elevated background tokens, and smooth focus transition states matching the Antigravity aesthetic.

## Success Criteria *(mandatory)*

- **SC-001**: Window opens at 1280x850 on desktop startup.
- **SC-002**: Pressing `Tab` switches work mode in under 50ms with zero cursor jumping.
- **SC-003**: 100% of command suggestions in the `/` popup show icon, title, and description.
- **SC-004**: The bottom controls adhere strictly to `[Attach] [Model Selector] ... [Send Button]`.

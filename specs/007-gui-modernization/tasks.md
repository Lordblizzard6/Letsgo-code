# Tasks: Modernización de la GUI — Layout de3 columnas

**Feature Branch**: `007-gui-modernization`

**Created**: 2026-08-27

**Reference**: `specs/007-gui-modernization/plan.md`

## Organization

Tasks are organized by workstream phases (R1..R10). Each phase follows:
**implementación → tests → gate**.

- **[X]** = completed
- **[ ]** = pending

---

## Phase 1: R1 — Design Tokens + Layout Base

- [ ] T001 [R1] Reescribir `cmd/wails/frontend/src/app.css` con los design tokens del gui-shell
  (dark: `--bg: #0a0e15`, `--accent: #7c5cff`; light: `--bg: #f4f6f9`, `--accent: #6a4bef`);
  mantener alias legacy durante la transición (R1.1).

- [ ] T002 [R1] Crear `cmd/wails/frontend/src/components/header.svelte` — header de50px con:
  botón menú (toggle sidebar), logo + marca, selector de proyecto, selector de cuenta,
  selector de modelo, chip de uso, botón explorador, botón settings, botón palette (R1.2).

- [ ] T003 [R1] Crear `cmd/wails/frontend/src/components/sidebar.svelte` — sidebar de260px colapsable
  a0px con animación220ms; secciones: "Recientes" (proyectos) y "Sesiones" (lista con badges)
  (R1.3).

- [ ] T004 [R1] Crear `cmd/wails/frontend/src/components/activity-pane.svelte` — panel derecho
  de320px con: explorador de archivos (explorer-panel) + panel de actividad (activity-panel)
  con tool cards (R1.4).

- [ ] T005 [R1] Reescribir `cmd/wails/frontend/src/App.svelte` con layout de3 columnas:
  `#app-header` + `#app-main` (`#sidebar` + `#chat-pane` + `#right-pane`) (R1.5).

- [ ] T006 [R1] Test (Vitest) — verificar que el layout renderiza las3 columnas y que
  vitest pre-existente sigue verde; svelte-check 0 errores (R1.6).

**Checkpoint1**: Layout de3 columnas visible, header funcional, sidebar colapsable.

---

## Phase 2: R2 — Header + Selectors

- [ ] T007 [R2] Crear `cmd/wails/frontend/src/components/account-selector.svelte` — dropdown
  de cuentas con: trigger (nombre + caret), lista de cuentas (nombre + provider/model),
  active state con accent-soft (R2.1).

- [ ] T008 [R2] Crear `cmd/wails/frontend/src/components/model-selector.svelte` — dropdown
  de modelos con: trigger (provider + model + caret), lista agrupados por cuenta,
  active state con accent-soft (R2.2).

- [ ] T009 [R2] Crear `cmd/wails/frontend/src/components/project-selector.svelte` — dropdown
  de proyectos con: trigger (folder icon + nombre), proyectos recientes, botón open/create (R2.3).

- [ ] T010 [R2] Crear `cmd/wails/frontend/src/components/usage-chip.svelte` — chip con
  tokens + costo, click abre SpendView (R2.4).

- [ ] T011 [R2] Integrar selectors en `header.svelte` — account-selector, model-selector,
  project-selector, usage-chip (R2.5).

- [ ] T012 [R2] Test (Vitest) — selectors abren/cierran, selección actualiza store,
  vitest pre-existente sigue verde (R2.6).

**Gate**: Header con todos los selectors operativos.

---

## Phase 3: R3 — Tabs + Chat Viewport

- [ ] T013 [R3] Crear `cmd/wails/frontend/src/components/tabs-bar.svelte` — barra de pestañas
  con: tabs existentes, botón add (+), botón close (✕), active state con accent-soft,
  overflow-x auto (R3.1).

- [ ] T014 [R3] Crear `cmd/wails/frontend/src/components/chat-viewport.svelte` — viewport de
  mensajes con: scroll, stick-to-bottom, jump-to-bottom button (R3.2).

- [ ] T015 [R3] Reescribir `cmd/wails/frontend/src/components/message-row.svelte` — fila de
  mensaje con: user (burbuja derecha, accent-soft bg), assistant (full-width, avatar gradiente,
  nombre bold), tool (línea inline, icono estado, nombre mono) (R3.3).

- [ ] T016 [R3] Integrar markdown rendering con highlight.js en message-row — code blocks con
  header (lenguaje + copy) y syntax highlighting (R3.4).

- [ ] T017 [R3] Test (Vitest) — tabs funcionan, mensajes se renderizan correctamente,
  vitest pre-existente sigue verde (R3.5).

**Gate**: Chat con múltiples pestañas y mensajes correctos.

---

## Phase 4: R4 — Composer + Mode Selector + Cmd Palette

- [ ] T018 [R4] Reescribir `cmd/wails/frontend/src/components/composer.svelte` — composer con:
  attach button (paperclip), textarea auto-expansible (field-sizing: content), send button
  (gradiente púrpura), stop button (gradiente rojo) (R4.1).

- [ ] T019 [R4] Crear `cmd/wails/frontend/src/components/mode-selector.svelte` — dropdown con
 4 modos (code/architecture/planning/research) con iconos SVG; active state con accent-soft (R4.2).

- [ ] T020 [R4] Crear `cmd/wails/frontend/src/components/cmd-palette.svelte` — paleta de
  comandos con: input de búsqueda, items filtrados, ↑/↓ navigation, Enter ejecuta,
  Esc cierra (R4.3).

- [ ] T021 [R4] Crear `cmd/wails/frontend/src/components/attachment-chip.svelte` — chip de
  archivo adjunto con: nombre, tamaño, botón remove (✕) (R4.4).

- [ ] T022 [R4] Integrar composer en chat pane — composer.svelte + cmd-palette + attachments (R4.5).

- [ ] T023 [R4] Test (Vitest) — composer funciona, cmd palette navegable, attachments se
  muestran, vitest pre-existente sigue verde (R4.6).

**Gate**: Composer con attach, send/stop, mode selector y cmd palette.

---

## Phase 5: R5 — Global Palette Ctrl+K

- [ ] T024 [R5] Crear `cmd/wails/frontend/src/components/global-palette.svelte` — overlay +
  input + items agrupados (Acciones, Proyectos, Pestañas); focus trap; keyboard navigation (R5.1).

- [ ] T025 [R5] Integrar Ctrl+K handler en `App.svelte` — window keydown listener, abre
  global-palette (R5.2).

- [ ] T026 [R5] Implementar focus trap en global-palette — Tab cycling, Esc close, focus
  restoration al composer (R5.3).

- [ ] T027 [R5] Test (Vitest) — Ctrl+K abre, ↑/↓ navega, Enter ejecuta, Esc cierra,
  focus vuelve al composer (R5.4).

**Gate**: Paleta global funcional con keyboard navigation.

---

## Phase 6: R6 — Activity Pane + Tool Cards

- [ ] T028 [R6] Crear `cmd/wails/frontend/src/components/tool-card.svelte` — tarjeta con:
  head (icono estado + nombre mono), args (mono, bg), result (mono, success/error color) (R6.1).

- [ ] T029 [R6] Crear `cmd/wails/frontend/src/components/explorer-panel.svelte` — panel de
  explorador con: título, botón up, lista de archivos/directorios (R6.2).

- [ ] T030 [R6] Integrar activity pane en right-pane — explorer-panel + activity-panel
  con tool cards (R6.3).

- [ ] T031 [R6] Test (Vitest) — tool cards muestran estados running/done/error,
  vitest pre-existente sigue verde (R6.4).

**Gate**: Activity pane con tool cards y explorer funcional.

---

## Phase 7: R7 — Modales Settings + Spend

- [ ] T032 [R7] Crear `cmd/wails/frontend/src/components/settings-modal.svelte` — overlay +
  modal centrado (680px) con header, tabs (accounts/general/budget/safety/search/paths),
  body scrollable (R7.1).

- [ ] T033 [R7] Crear `cmd/wails/frontend/src/components/spend-modal.svelte` — overlay +
  modal centrado con gráfico de barras por día, rango selector (7/30/90), tabla de entries (R7.2).

- [ ] T034 [R7] Implementar focus trap y Esc close en ambos modales — aria-modal, tabindex,
  Tab cycling, Esc handler (R7.3).

- [ ] T035 [R7] Test (Vitest) — modales abren/cierran, focus trap funciona, vitest
  pre-existente sigue verde (R7.4).

**Gate**: Modales settings y spend funcionales.

---

## Phase 8: R8 — Estados Vacíos + Loading + Error

- [ ] T036 [R8] Crear `cmd/wails/frontend/src/components/empty-state.svelte` — icono SVG +
  título + texto + CTA button (R8.1).

- [ ] T037 [R8] Crear `cmd/wails/frontend/src/components/skeleton.svelte` — shimmer animation
  ≤150ms, off bajo prefers-reduced-motion (R8.2).

- [ ] T038 [R8] Crear `cmd/wails/frontend/src/components/error-banner.svelte` — icono +
  orientación + retry button (R8.3).

- [ ] T039 [R8] Crear `cmd/wails/frontend/src/components/thinking-dots.svelte` — 3 puntos
  parpadeantes (R8.4).

- [ ] T040 [R8] Aplicar estados a las10 superficies: chat, sessions, git, tasks, mcp,
  plugins, settings, usage, help, account — empty + loading + error (R8.5).

- [ ] T041 [R8] Test (Vitest) — empty/loading/error en7+ superficies, thinking dots
  aparecen cuando corresponde (R8.6).

**Gate**: Estados de UI consistentes en todas las superficies.

---

## Phase 9: R9 — Sidebar Proyectos + Sesiones

- [ ] T042 [R9] Crear `cmd/wails/frontend/src/components/recent-projects.svelte` — lista de
  proyectos recientes con: folder icon, nombre, delete button (R9.1).

- [ ] T043 [R9] Crear `cmd/wails/frontend/src/components/sessions-list.svelte` — lista de
  sesiones con: título, badge (test/skills), fecha, delete button (R9.2).

- [ ] T044 [R9] Implementar empty states accionables en sidebar — "No hay proyectos recientes"
  con CTA, "No hay sesiones" con CTA (R9.3).

- [ ] T045 [R9] Implementar colapsado del sidebar con animación220ms — width transition,
  border-right none cuando colapsado (R9.4).

- [ ] T046 [R9] Test (Vitest) — sidebar muestra proyectos y sesiones, colapso funciona,
  vitest pre-existente sigue verde (R9.5).

**Gate**: Sidebar con proyectos y sesiones funcionales.

---

## Phase 10: R10 — Regresión y Polish

- [ ] T047 [R10] Ejecutar `npm test -- --run` — vitest completo, tests pre-existentes
  pasan sin cambios + tests nuevos de cada fase (R10.1).

- [ ] T048 [R10] Ejecutar `npm run check` — svelte-check 0 errores (R10.2).

- [ ] T049 [R10] Ejecutar `npm run build` — build de producción exitoso (R10.3).

- [ ] T050 [R10] Ejecutar `npx playwright test` — e2e onboarding intacto (strings
  "Welcome to LetsGO", "Save & start chatting", "API Key", placeholder intactos) (R10.4).

- [ ] T051 [R10] Verificar gates G1-G10 — todos los gates del spec pasando (R10.5).

- [ ] T052 [R10] Aplicar polish visual — sombras, bordes, transiciones, gradientes del
  gui-shell (R10.6).

**Checkpoint5**: Todos los tests pasan, gates verificados, polish aplicado.

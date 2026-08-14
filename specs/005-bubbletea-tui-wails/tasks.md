# Tasks: Doble interfaz — TUI Bubble Tea + GUI Wails

**Input**: Design documents from `/specs/005-bubbletea-tui-wails/`

**Prerequisites**: plan.md (required), spec.md (user stories), research.md, data-model.md, `contracts/frontend-contract.md`, `contracts/gui-contract.md`

**Tests**: SI — el SPEC lo requiere explícitamente (conformidad del contrato FR-002/C-001..C-006, modelo TUI, Vitest+jsdom con bindings mockeados, Playwright contra el dev server). Según la constitución v1.0.0 (**Principio I, Test-Last, NON-NEGOTIABLE**), los tests se escriben y ejecutan **DESPUÉS** de la implementación, en batch al final de cada user story; NUNCA preceden ni bloquean la construcción (adoptado en `/speckit.constitution`, 2026-08-13; supersede la convención previa TDD/C-1 del repo).

**Organization**: Tareas agrupadas por user story para implementación y validación independiente de cada story.

## Format: `[ID] [P?] [Story] Descripción`

- **[P]**: Puede correr en paralelo (archivos distintos, sin dependencias)
- **[Story]**: US1..US4 (spec.md). Las fases Setup/Foundational/Retirada/Polish no llevan label.
- Rutas exactas de fichero en cada descripción. Referencias a IDs de contrato C-00x, FR-xxx y SC-xxx donde aplican.

## Build / Test del repo (Windows)

- Tests Go headless: `go test ./internal/... ./cmd/...` · Compilar: `go build ./...` → `letsgo.exe` · TUI: `letsgo chat`
- GUI: se ejecuta por el **MISMO ejecutable** que Fyne (`letsgo.exe gui` → cobra `cmd/gui.go` → `cmd/wails/gui.Run()`, plan.md Structure Decision); `wails3 dev` solo para hot-reload de desarrollo del frontend Svelte-TS; `wails3 build -nsis` queda como distribución opcional (D4).
- Frontend: `npm ci && npm run check && npm run build` y `npm run test` (vitest) y `npx playwright test` (contra el dev server) en `cmd/wails/frontend/`

---

## Phase 1: Setup (Infraestructura compartida)

**Proposito**: linea base verificada, andamiaje Wails v3 con pin exacto y las tres capas de testing (Go, vitest, playwright, teatest) listas.

- [x] T001 Verificar la linea base: `go vet ./...` y `go build ./...`; guardar el resultado de la suite headless actual (`go test ./internal/... ./cmd/...`) como referencia verde para las puertas de regresión (Gate B, T079) y la retirada de Fyne (G-2, Phase 3).
- [x] T002 Crear el andamiaje Wails v3 con `wails3 init` y adaptarlo a la estructura del repo: `cmd/wails/main.go`, `cmd/wails/services/`, `cmd/wails/frontend/` (plantilla Svelte + TypeScript + Vite con `src/components/`, `src/bindings/` generado por wails3 y `src/tests/`); el paquete reutilizable `cmd/wails/gui` se extrae en Phase 3 T012 (plan.md Project Structure, D2).
- [x] T003 Fijar Wails v3 en `go.mod`: `go get github.com/wailsapp/wails/v3@v3.0.0-beta.7` + `go mod tidy`; verificar que el pin exacto `v3.0.0-beta.7` queda en go.mod (D1, C-4).
- [x] T004 [P] Añadir dependencias del frontend en `cmd/wails/frontend/package.json`: runtime `@wailsio/runtime`, `streaming-markdown`, `dompurify` y dev `vitest`, `jsdom`, `@testing-library/svelte`, `@playwright/test`; `npm install` y build de humo (`npm run build`) (D3, D5).
- [x] T005 [P] Configurar los tests del frontend: config de vitest (jsdom) en `cmd/wails/frontend/vite.config.ts` + smoke test; `src/tests/setup.ts` y `playwright.config.ts` se crean con la infra e2e (T063-T065) (D5).
- [x] T006 [P] Añadir la dependencia de testing TUI: `go get github.com/charmbracelet/x/exp/teatest` en `go.mod` (D5).
- [x] T007 [P] Extender `.github/workflows/ci.yml`: subir `go-version` a `1.26` (go.mod pide 1.26.1) y añadir el job `gui`: setup-go 1.26 + setup-node → generate bindings → `npm ci && npm run test` (vitest) en `cmd/wails/frontend` → `npm run build` → `go build ./cmd/wails/`; Playwright contra el dev server se añade con la infra e2e (T063-T065) (D5).

**Checkpoint**: linea base guardada, andamiaje Wails compilando con pin v3.0.0-beta.7, deps de test (teatest/vitest/playwright) instaladas y CI con job GUI preparado.

---

## Phase 2: Foundational (Pre-requisitos bloqueantes)

**Proposito**: publicar el contrato de frontend como frontera oficial core↔interfaces y dejar listos los tipos de eventos y el harness de conformidad que usan todos los stories. Sin cambios de esquema (data-model.md: sin tablas nuevas).

- [x] T008 Publicar el contrato de frontend como frontera oficial: confirmar `specs/005-bubbletea-tui-wails/contracts/frontend-contract.md` como doc único de la superficie (comandos §1, eventos §2, sesiones/config §3, conformidad C-001..C-006 §4) y añadir `internal/engine/doc.go` que lo enlace y declare que toda interfaz consume el core solo vía esta superficie (FR-001, FR-002).
- [x] T009 [P] Tipos adaptadores del bus de eventos en `internal/engine/events.go`: tipos tipados por evento del contrato §2 (StreamStart, StreamDelta, StreamEnd, StreamCancelled, StreamError, ToolStart, ToolEnd, UsageUpdate, SessionList, SessionLoaded, ConfigChanged) consumibles por la TUI y la GUI; sin cambios de comportamiento del core (D3, FR-002).
- [x] T010 [P] Crear el harness de conformidad headless en `internal/engine/contract_test.go`: helpers de consumidor fake (envía comandos del contrato al motor real y recolecta los eventos tipados de T009); los tests C-001..C-006 se añaden en US1 sobre este harness (FR-002).
- [x] T011 [P] Verificar invariantes del core: `internal/db/database.go` no crea tablas nuevas (data-model.md) y la suite existente pasa sin cambios (`go test ./internal/engine/... ./internal/db/... ./internal/config/...`) (C-2).

**Checkpoint**: contrato publicado, tipos de eventos y harness de conformidad listos — los user stories pueden empezar.

---

## Phase 3: Retirada INMEDIATA de Fyne y ejecución por el mismo ejecutable (PRIMERA tarea del feature)

**Proposito**: cumplir la directiva del usuario (2026-08-13, plan.md/research D6/gui-contract §7): **eliminar Fyne YA** — `internal/gui`, `fyne.io/fyne` y el árbol de assets — como **primera tarea** del feature, y que la GUI Wails arranque por el **mismo método que Fyne hoy** (`letsgo.exe gui` → `cmd/gui.go` → `cmd/wails/gui.Run()`), SIN binario Wails separado. La paridad de flujos ya no es prerrequisito de la retirada: se construye en US3/US4 y su regresión (G-1) se cierra al final de cada historia.

- [x] T012 Extraer el paquete reutilizable `cmd/wails/gui/gui.go` (package `gui`, func `Run(assets fs.FS) error`): mover el wiring de `cmd/wails/main.go` (embed `frontend/dist`, `application.New` + services bound T040-T049, Event bus de Wails, `OnShutdown` con cancelación de stream) al paquete `gui`; `cmd/wails/main.go` queda como `main()` delgado que llama `gui.Run()` (permite `wails3 dev`/`wails3 build`) (plan.md Project Structure, D4, FR-014).
- [x] T013 [P] Reescribir `cmd/gui.go` (comando cobra `letsgo gui`): quitar el import de `internal/gui` y llamar `gui.Run()` de `cmd/wails/gui` — MISMO método de ejecución que lanzaba Fyne, sin binario separado (FR-018, gui-contract §7). Ambos embeds de `frontend/dist` resuelven index.html vía búsqueda recursiva de Wails.
- [x] T014 [P] Eliminar `internal/gui/` completo: código Fyne, tests (`rail_test.go`, `settings_test.go`, `theme_test.go`, `usage_test.go`, `empty_test.go`, `tooltips_test.go`, `icons_test.go`, `account_test.go`, `navigation_test.go`, `composer_test.go`, `message_test.go`, `shortcuts_test.go`, `app_shell_test.go`, `shell_smoke_test.go`) y assets (`CascadiaMono.ttf`); borrar toda referencia a fyne en `cmd/` e `internal/` (`rg -n "fyne|Fyne|internal/gui" cmd/ internal/ --glob '*.go'`) (FR-018). Actualizados `build.bat`, README §GUI y comentarios de services; queda `contract_c003_c006_test.go` (test C-006, intencional).
- [x] T015 `go mod tidy` y retirar el árbol de dependencias fyne: verificado — `go mod why fyne.io/fyne/v2` y `fyne.io/systray` devuelven "main module does not need"; `go.mod` limpio de fyne/glfw/OpenGL (G-3, SC-012).
- [x] T016 [P] Migrar assets de la GUI a Wails: el frontend referencia la fuente por nombre con fallback (`--mono: "Cascadia Mono", Consolas, monospace`) — no requiere embebido; no hay `FyneApp.toml`; `rg -n -i "fyne|Fyne|internal/gui" --glob '!specs/**' .` solo devuelve referencias intencionales (nota de `build.bat`, comentario de `cmd/gui.go`, test C-006, bindings autogenerados) (G-3).
- [x] T017 Gate de retirada (G-2): `go build ./...` y `go vet ./...` verdes **sin** importar fyne; `letsgo.exe` construido y `letsgo gui` arranca ventana Wails (vive >6s sin crash, AssetServer handler OK); `letsgo chat` (/--help) funcional (FR-018, SC-010, SC-012). Nota: `go test ./...` muestra 3 fallos preexistentes de fases previas (TestToolCycleWithAutoApprove/TestPlanGateHoldsToolsUntilApproved de 004 en engine, TestSessionsServiceLifecycle de US4 en services) — NON-gating por decisión del usuario: el flujo real del programa aún no está fijado y esos tests pueden dejar de ser válidos.

**Checkpoint**: Fyne fuera del proyecto (código + go.mod), el proyecto compila y las suites pasan sin la dependencia, y la GUI Wails se ejecuta por `letsgo.exe gui`. **No existe referencia a fyne fuera de `specs/`.**

---

## Phase 4: User Story 1 — Core compartido: contrato de frontend y conformidad (Prioridad P1) 🎯 MVP

**Goal**: el contrato de frontend (comandos, eventos de stream, sesiones y configuración) queda documentado y verificado por la suite de conformidad C-001..C-006, garantizando "una sola fuente de verdad" (motor `internal/engine`, sqlite `internal/db`, configuración `internal/config`) para la TUI y la GUI Wails sin cambiar el comportamiento del core.

**Independent Test**: `go test ./internal/engine/... ./internal/db/... ./internal/config/...` verde sin pantalla; además, iniciar una conversación desde la TUI y comprobar (vía harness del contrato) que la misma sesión es legible con historial completo (J1, SC-002).

### Implementación para User Story 1

- [x] T018 [P] [US1] Formalizar la API de comandos del contrato §1 en `internal/engine/commands.go` (`send`, `cancel`, `setModel`, `setAutoApprove`, `resetMemory`, `remember`, dispatch slash) como envoltorios estables documentados con referencia a C-001/C-002; sin cambios de comportamiento (FR-001).
- [x] T019 [US1] Formalizar las funciones de sesión y configuración del contrato §3 (`SessionCreate`, `SessionList`, `SessionOpen`, `GetMessages`, `GetConfig`, `SaveConfig`, `GetUsageToday`) donde ya viven (`internal/engine/engine.go`, `internal/db/database.go`, `internal/config/config.go`) con firmas estables y doc vinculada a C-003/C-004 (FR-001, FR-002).
- [x] T020 [US1] Hacer pasar la suite de conformidad completa: ejecutar `go test ./internal/engine/... ./internal/db/... ./internal/config/...` y resolver cualquier brecha entre la superficie contratada y el core sin cambiar su comportamiento (C-001..C-006, FR-002, J1).

### Tests para User Story 1 (Test-Last, constitución I — escritos/ejecutados al final del story) ⚠️

- [x] T021 [P] [US1] Test de conformidad C-001 en `internal/engine/contract_test.go`: `send` emite `stream:start` → `stream:delta…` → `stream:end` con mensaje persistido (contrato §1/§2, FR-003).
- [x] T022 [P] [US1] Test de conformidad C-002 en `internal/engine/contract_test.go`: `cancel` a mitad de stream no persiste el texto parcial y vuelve a estado usable (contrato §1, FR-007, SC-008).
- [x] T023 [P] [US1] Test de conformidad C-003 en `internal/engine/contract_test.go`: una sesión creada por un consumidor (harness TUI) es retomable por otro (harness GUI) con historial completo vía `internal/db` (FR-006/FR-009/FR-012, SC-002).
- [x] T024 [P] [US1] Test de conformidad C-004 en `internal/engine/contract_test.go`: `setModel`/`SaveConfig` emite `config:changed` y refresca el motor sin reinicializar la interfaz (FR-016).
- [x] T025 [P] [US1] Test de conformidad C-005 en `internal/engine/contract_test.go`: un `stream:error` deja el historial intacto (FR-013, SC-011).
- [x] T026 [P] [US1] Test de conformidad C-006 en `internal/engine/contract_test.go`: las interfaces solo consumen la superficie del contrato — falla si `internal/tui` (y `cmd/wails/services` cuando exista) accede al core fuera de los comandos/eventos/funciones contratados (FR-001, contrato §4).
- [ ] T027 [US1] Validación independiente manual (J1): iniciar una conversación con `letsgo chat`, cerrarla y verificar que un consumidor distinto del contrato (harness GUI) lee el historial completo de esa sesión (SC-002).

**Checkpoint**: US1 completo y testeable aislado — MVP del feature (el contrato verifica "una sola fuente de verdad" sin requerir ninguna de las dos interfaces nuevas).

---

## Phase 5: User Story 2 — TUI Bubble Tea formalizada sobre el contrato (Prioridad P2)

**Goal**: la TUI existente (`internal/tui`, bubbletea v1.3.10 + bubbles + glamour + lipgloss) queda formalizada sobre el contrato conservando todas sus capacidades: streaming+markdown, slash commands con autocompletado, selector de modelo Ctrl+S, retomar sesiones, Ctrl+C limpio, referencias `@archivo` y barra de estado; testeada con modelo puro (`Update` sintético) + teatest.

**Independent Test**: `letsgo chat` en un terminal 80x24 y completar el ciclo: enviar mensaje → streaming → interrumpir con Ctrl+C → cambiar modelo con Ctrl+S → `/help`, `/clear`, `/cost` → retomar una sesión anterior (J2, FR-003..FR-007).

### Implementación para User Story 2

- [x] T028 [US2] Formalizar el modelo TUI sobre el contrato en `internal/tui/ui.go`: enviar/cancelar vía los comandos del contrato (T018) y consumir los eventos tipados de T009 para streaming/estado; sin cambios de UX (FR-001, FR-003).
- [x] T029 [P] [US2] Despachar los slash commands en `internal/tui/slash_commands.go` vía el comando slash del contrato, manteniendo el autocompletado (FR-004).
- [x] T030 [P] [US2] Selector de modelo Ctrl+S en `internal/tui/commands.go`: persiste provider/model vía `SaveConfig` del contrato y refresca el motor in-place (FR-005, FR-016).
- [x] T031 [US2] Control de sesiones (resumen, retomar, crear) en `internal/tui/commands.go` sobre `SessionCreate`/`SessionList`/`SessionOpen` del contrato, compartiendo la misma DB (FR-006, C-003).
- [x] T032 [US2] Cancelación limpia: el manejador de Ctrl+C en `internal/tui/ui.go` llama a `cancel()` del contrato y responde a `stream:cancelled` sin persistir el parcial (FR-007, SC-008).
- [x] T033 [US2] Re-layout al redimensionar durante un stream activo en `internal/tui/ui.go`: el viewport se re-laya sin perder contenido acumulado ni interrumpir la respuesta (edge case; SC-008, SC-009).

### Tests para User Story 2 (Test-Last, constitución I — modelo puro + teatest; plan/research D5) ⚠️

- [x] T034 [P] [US2] TestUpdateStreamLifecycle en `internal/tui/model_test.go`: `Update()` sintético con `stream:start`/`stream:delta`/`stream:end` → el viewport acumula el markdown y el estado vuelve a "listo" (FR-003, SC-005).
- [x] T035 [P] [US2] TestUpdateCancelClean en `internal/tui/model_test.go`: mensaje de cancelación (equivalente a Ctrl+C) a mitad de stream → prompt usable y sin mensaje parcial (FR-007, SC-008, C-002).
- [x] T036 [P] [US2] TestSlashCommandsAndSuggestions en `internal/tui/slash_commands_test.go`: `/help`, `/clear`, `/model`, `/cost`, `/tokens`, `/compact`, `/quit` responden y aparecen sugerencias de autocompletado al escribir "/" (FR-004, SC-004).
- [x] T037 [P] [US2] TestModelPickerCtrlS en `internal/tui/model_test.go`: Ctrl+S abre el selector; al confirmar, la selección persiste en `internal/config` y el motor se refresca sin reiniciar (FR-005, FR-016, C-004).
- [x] T038 [P] [US2] TestSessionResume en `internal/tui/model_test.go`: retomar una sesión previa y crear una nueva contra la misma `internal/db` (FR-006, C-003).
- [x] T039 [P] [US2] TestTeaStreamingProgram en `internal/tui/teatest_test.go`: programa Bubble Tea completo headless con teatest que reproduce send → streaming → end sobre el motor real (D5, FR-003).

**Checkpoint**: US2 independiente — TUI completa y testeada con modelo puro + teatest (J2); US1 y US2 funcionan juntos.

---

## Phase 6: User Story 3 — GUI Wails: paridad completa sobre el mismo ejecutable (Prioridad P3)

**Goal**: la GUI Wails (`cmd/wails/`) alcanza paridad funcional completa con la GUI Fyne retirada (FR-017): railway de 10 destinos, chat con streaming, historial de sesiones, paleta Ctrl+K, atajos Alt+1..8, overlays, temas y estados vacíos/de carga. La retirada de Fyne ya ocurrió en **Phase 3** (primera tarea del feature, FR-018); los gates G-2/G-3 quedaron verificados allí y **G-1** (regresión de cada flujo portado) se cierra aquí y en US4 con tests escritos al final de cada historia (constitución I). La GUI se ejecuta por `letsgo.exe gui` (T013).

**Independent Test**: compilar y lanzar la GUI (`letsgo.exe gui`), completar una conversación (streaming, cancelación, herramienta), retomar una sesión creada desde la TUI, redimensionar la ventana y recorrer los 10 destinos con Alt+1..8 y Ctrl+K en ambos temas (J3).

### Implementación — servicios bound (capa fina sin lógica de dominio, solo el contrato)

- [x] T040 [P] [US3] ChatService en `cmd/wails/services/chat_service.go`: `send`/`cancel` vía el contrato; re-emite el stream como eventos `chat:delta` batcheados ~50ms por el Event bus de Wails (D3, FR-008, SC-005).
- [x] T041 [P] [US3] SessionsService en `cmd/wails/services/sessions_service.go`: `SessionCreate`/`List`/`Open`/`GetMessages` del contrato (FR-009, SC-002).
- [x] T042 [P] [US3] GitService en `cmd/wails/services/git_service.go`: branches, status, commit/diff/review/tags/hooks/undo/redo, PR+push (gui-contract §4, FR-017).
- [x] T043 [P] [US3] TasksService en `cmd/wails/services/tasks_service.go`: lista de sub-agentes con estado (FR-017).
- [x] T044 [P] [US3] MCPService en `cmd/wails/services/mcp_service.go`: lista de servidores + detalle (estado, tools, start/stop) (FR-017).
- [x] T045 [P] [US3] PluginsService en `cmd/wails/services/plugins_service.go`: lista + instalar/cargar/habilitar/deshabilitar (FR-017).
- [x] T046 [P] [US3] SettingsService en `cmd/wails/services/settings_service.go`: `GetConfig`/`SaveConfig` del contrato (FR-010, C-004).
- [x] T047 [P] [US3] UsageService en `cmd/wails/services/usage_service.go`: `GetUsageToday` del contrato (coste, requests, tokens) (FR-017).
- [x] T048 [P] [US3] ThemeService en `cmd/wails/services/theme_service.go`: get/set del tema persistido en `internal/config` (FR-010, FR-016).
- [x] T049 [P] [US3] AccountService en `cmd/wails/services/account_service.go`: estado online/ocupado y datos de cuenta (gui-contract §1).
- [x] T050 [US3] `cmd/wails/main.go` (package main) arranca la app Wails con WebView2 (`-webview2 embed`, D4) y registra los services bound de T040-T049 y cierre limpio sin procesos colgados con stream en curso (FR-014); el wiring completo se consolidó en el paquete `cmd/wails/gui` (Phase 3 T012), quedando `main()` como wrapper (FR-014).

### Implementación — frontend Svelte-TS (paridad)

- [x] T051 [US3] Shell de la app en `cmd/wails/frontend/src/App.svelte`: layout rail + pane + chat responsive (chat, panel de sesiones y composer legibles y operativos al redimensionar) (FR-015, SC-009).
- [x] T052 [P] [US3] Rail de 10 destinos en `cmd/wails/frontend/src/components/Rail.svelte`: 3 zonas con separador "HERRAMIENTAS", colapso con tooltip "Label (Alt+N)", marca activa única y avatar con dot de estado (gui-contract §1).
- [x] T053 [P] [US3] Chat y mensajes en `cmd/wails/frontend/src/components/Chat.svelte` y `Message.svelte`: suscripción a `chat:delta`, acumulado saneado con DOMPurify + streaming-markdown (doble sanitize sobre el acumulado, nunca innerHTML), indicador de estado, tarjetas borde 1px/radius 6px/padding 16px y columna ~720px (gui-contract §2, D3).
- [x] T054 [P] [US3] Composer en `cmd/wails/frontend/src/components/Composer.svelte`: botón Enviar↔Detener, placeholder "Escribe a LetsGO…", input siempre activo durante el stream (gui-contract §2, FR-008).
- [x] T055 [P] [US3] Actividad de herramientas en `cmd/wails/frontend/src/components/ToolActivity.svelte`: `tool:start`/`tool:end` (nombre y resultado éxito/error) mostrados dentro del chat sin bloquear (FR-008, contrato §2).
- [x] T056 [P] [US3] Historial de sesiones en `cmd/wails/frontend/src/components/SessionsPanel.svelte`: crear/listar/retomar con estados vacío ("No hay conversaciones todavía." + CTA "Nueva conversación") y de carga ("Refrescando…") (gui-contract §3, FR-009).
- [x] T057 [P] [US3] GitPanel.svelte en `cmd/wails/frontend/src/components/GitPanel.svelte`: branches, status, commit/diff/review/tags/hooks/undo/redo, PR+push (gui-contract §4, FR-017).
- [x] T058 [P] [US3] TasksPanel.svelte en `cmd/wails/frontend/src/components/TasksPanel.svelte`: lista de sub-agentes con estado; vacío "Todavía no hay tareas." (gui-contract §4, FR-017).
- [x] T059 [P] [US3] MCPPanel.svelte en `cmd/wails/frontend/src/components/MCPPanel.svelte`: lista + detalle (estado, tools, start/stop); vacío con CTA "Añadir servidor" (gui-contract §4, FR-017).
- [x] T060 [P] [US3] PluginsPanel.svelte en `cmd/wails/frontend/src/components/PluginsPanel.svelte`: lista + instalar/cargar/habilitar/deshabilitar; vacío con CTA "Instalar…" (gui-contract §4, FR-017).
- [x] T061 [P] [US3] SettingsOverlay.svelte en `cmd/wails/frontend/src/components/SettingsOverlay.svelte`: overlay con tabs Cuenta/Apariencia/Preferencias/Uso; guardar persiste vía SettingsService (gui-contract §4, FR-010).
- [x] T062 [P] [US3] UsagePanel.svelte en `cmd/wails/frontend/src/components/UsagePanel.svelte`: KPI hoy (coste, requests, tokens) + tabla por proveedor + presupuesto; vacío "Sin actividad registrada hoy." (gui-contract §4, FR-017).
- [x] T063 [P] [US3] HelpPanel.svelte en `cmd/wails/frontend/src/components/HelpPanel.svelte`: guía de atajos y comandos (gui-contract §4, FR-017).
- [x] T064 [P] [US3] ThemeSwitch.svelte en `cmd/wails/frontend/src/components/ThemeSwitch.svelte`: cambio dark/light aplicado al motor y persistido vía ThemeService (gui-contract §4, FR-016).
- [x] T065 [P] [US3] AccountFlyout.svelte en `cmd/wails/frontend/src/components/AccountFlyout.svelte`: flyout del avatar con datos de cuenta y acciones (claves, uso, salir), cierre con Esc (gui-contract §1, FR-017).
- [x] T066 [P] [US3] Palette.svelte en `cmd/wails/frontend/src/components/Palette.svelte`: paleta Ctrl+K (comando + salto a superficies), atajos Alt+1..8 y Esc con foco visible en ambos temas (gui-contract §5, SC-003).
- [x] T067 [US3] Estados vacíos/de carga en las 6+ superficies (chat, sesiones, git, tasks, mcp, plugins, usage) en `cmd/wails/frontend/src/components/` (gui-contract §5, FR-017).

### Tests para User Story 3 (Test-Last, constitución I — Vitest + bindings mockeados y Playwright contra el dev server; plan/research D5) ⚠️

> **NOTA**: tras la regeneración de bindings (10 services: Chat/Sessions/Git/Tasks/MCP/Plugins/Settings/Usage/Theme/Account), el smoke `src/app.test.ts` quedó obsoleto (importaba `AppService`); se corrige en T068 ANTES de correr el resto de la suite.

- [x] T068 [US3] Corregir el smoke test de bindings en `cmd/wails/frontend/src/app.test.ts`: reemplazar `AppService.AppInfo()` (ya no existe en los bindings regenerados) por una llamada real a un service bound (p.ej. `SettingsService.GetConfig()` o `ChatService`) con `Call.ByID` mockeado; `npm run test` (vitest) verde (D5).
- [ ] T069 [P] [US3] TestVitestRail en `cmd/wails/frontend/src/tests/rail.test.ts`: el rail renderiza 10 destinos en 3 zonas (work/tools/system) con marca activa única y atajos Alt+1..8 (gui-contract §1, FR-017).
- [x] T070 [P] [US3] TestVitestChatStreaming en `cmd/wails/frontend/src/tests/chat.test.ts`: con bindings mockeados, Chat.svelte acumula los eventos `chat:delta` y renderiza con streaming-markdown + DOMPurify sobre el acumulado (nunca innerHTML) (gui-contract §2, D3, FR-008).
- [x] T071 [P] [US3] TestVitestComposerSwap en `cmd/wails/frontend/src/tests/composer.test.ts`: el botón alterna Enviar↔Detener según el estado de stream y el input sigue activo durante el stream (gui-contract §2, FR-008).
- [ ] T072 [P] [US3] TestVitestSessionHistory en `cmd/wails/frontend/src/tests/sessions.test.ts`: crear/listar/retomar con bindings mockeados; estado vacío "No hay conversaciones todavía." + CTA "Nueva conversación"; "Refrescando…" conserva el contenido previo (gui-contract §3).
- [ ] T073 [P] [US3] TestVitestPaletteOverlays en `cmd/wails/frontend/src/tests/palette.test.ts`: Ctrl+K abre la paleta, Alt+1..8 navega, Esc cierra overlays/flyout y el foco vuelve al composer con focus visible (gui-contract §5).
- [ ] T074 [P] [US3] TestVitestThemeContrast en `cmd/wails/frontend/src/tests/theme.test.ts`: palette smoke dark/light — texto ≥4.5:1 y no-texto ≥3:1 (contraste AA, regresión quickstart, SC-003).
- [ ] T075 [P] [US3] TestPlaywrightChatE2E en `cmd/wails/frontend/src/tests/e2e/chat.spec.ts`: contra el dev server — enviar mensaje → streaming con indicador → cancelar → conversación persistida (FR-008, FR-014, SC-005).
- [ ] T076 [P] [US3] TestPlaywrightSessionsE2E en `cmd/wails/frontend/src/tests/e2e/sessions.spec.ts`: crear una sesión en la GUI, retomarla con historial completo y redimensionar la ventana sin romper el layout (FR-009, FR-015, SC-002).
- [ ] T077 [P] [US3] TestPlaywrightRailE2E en `cmd/wails/frontend/src/tests/e2e/rail.spec.ts`: recorrer los 10 destinos del rail con Alt+1..8 y la paleta Ctrl+K en ambos temas (FR-017, SC-010).

**Checkpoint (paridad)**: la GUI Wails cubre el rail completo que tenía Fyne (ya retirada en Phase 3) y supera Vitest + Playwright (G-1 cerrado para los flujos de US3). J3 completo vía `letsgo.exe gui`.

---

## Phase 7: User Story 4 — Onboarding y configuración compartida en la GUI Wails (Prioridad P3)

**Goal**: el primer arranque sin API key guía al usuario al flujo de configuración de proveedor/key antes de poder chatear; la superficie de Configuración persiste en `internal/config` y se aplica al motor sin reiniciar; los errores de red/API muestran orientación accionable con el historial intacto (FR-010, FR-011, FR-013).

**Independent Test**: borrar la configuración de API keys, abrir la GUI Wails (`letsgo.exe gui`) y completar el flujo de primer arranque (configurar → chatear); después cambiar el modelo en Configuración, guardar y verificar que la TUI y la configuración persistida reflejan el cambio (J4, SC-007).

### Implementación para User Story 4

- [x] T078 [US4] Gate de primer arranque en `cmd/wails/gui/gui.go` (o main wrapper) y `cmd/wails/frontend/src/App.svelte`: al arrancar se consulta `GetConfig` del contrato; sin key → renderizar solo el flujo de configuración y bloquear el chat hasta tener key válida (FR-011, gui-contract §6).
- [x] T079 [P] [US4] ProviderSetup.svelte en `cmd/wails/frontend/src/components/ProviderSetup.svelte`: formulario proveedor + API key con validación; al guardar → `SaveConfig` + refresco del motor → habilitar el chat (FR-011, SC-007).
- [x] T080 [US4] Aplicar cambios de Configuración al motor sin reinicio: SettingsOverlay + SettingsService emiten `config:changed` y refrescan el cliente del motor; verificación cruzada: la TUI (`letsgo chat`) lee la misma config persistida (FR-010, FR-016, C-004).
- [x] T081 [P] [US4] Manejo de errores de red/API en `cmd/wails/frontend/src/components/Chat.svelte`: banner claro con orientación y botón de reintentar; historial intacto (FR-013, SC-011).

### Tests para User Story 4 (Test-Last, constitución I — Vitest + Playwright) ⚠️

- [x] T082 [P] [US4] TestVitestOnboarding en `cmd/wails/frontend/src/tests/onboarding.test.ts`: con `GetConfig` mockeado sin key, la única superficie es el flujo de proveedor/key; el chat se habilita solo tras guardar una key válida (FR-011, gui-contract §6).
- [x] T083 [P] [US4] TestSettingsServiceRefresh en `cmd/wails/services/settings_service_test.go`: `SaveConfig` persiste en `internal/config` y emite `config:changed` que refresca el motor sin reiniciar (FR-010/FR-016, C-004).
- [x] T084 [P] [US4] TestVitestErrorStates en `cmd/wails/frontend/src/tests/errors.test.ts`: `stream:error` → mensaje orientativo (revisar key/conexión/modelo) y la conversación previa permanece intacta (FR-013, SC-011, C-005).
- [x] T085 [P] [US4] TestPlaywrightOnboardingE2E en `cmd/wails/frontend/src/tests/e2e/onboarding.spec.ts`: arrancar con la config de keys borrada (como J4) → flujo key→modelo→chat en ≤4 acciones (FR-011, SC-007).

**Checkpoint**: US4 independiente — la GUI v1 es autosuficiente: onboarding, configuración compartida y errores orientativos (J4). G-1 cerrado para US3+US4 (cada flujo del gui-contract tiene su test).

---

## Phase 8: User Story 5 — Restyle & UX tipo Codex/Qwen (Prioridad P4)

**Goal**: estilizar y mejorar la UI de la GUI Wails con la filosofía de diseño convergente Codex/Qwen (plan.md §US5 + research.md D8/D9 + gui-contract §8): chrome mínimo, dark-first, chat como canvas, keyboard-first, tokens semánticos solo en `app.css`. **Restyle SOLO frontend presentacional** — cero diff de bindings/services/backend (gate UX-14); atajos Alt+1..9 y rutas hash intactos; strings de test intactos (UX-12/UX-13).

**Independent Test**: `letsgo.exe gui` y recorrido J5 de quickstart.md — composer siempre visible, Retry reproduce el último turno, paleta navegable con teclado, Esc cierra overlays/system, tool activity 1 fila por llamada, affordance de stream, contrastes AA en ambos temas (J5, G2, UX-01..UX-09).

### Implementación — Fase U5-A (P0, Test-Last) ⚠️

- [x] T090 [US5] U5-A/I-1 Composer always-on: re-diseñar el layout de `cmd/wails/frontend/src/components/chat.svelte` (y `app.css`) — el área de mensajes/stream hace scroll dentro del pane y el composer queda fijado/físicamente siempre visible al fondo, aunque el scroll esté arriba (UX-02, plan §US5 punto 5).
- [x] T091 [US5] U5-A/I-2 Retry reproduce último turno: en `cmd/wails/frontend/src/components/chat.svelte` `handleRetry` debe reenviar **Start + Send del último turno de usuario** (helper `useChat.lastUserText()` en `lib/store.svelte.ts`) en lugar de solo `Start()` (UX-01).
- [x] T092 [US5] U5-A/I-3 Paleta ↑/↓+Enter: en `cmd/wails/frontend/src/components/palette.svelte` — ArrowUp/ArrowDown mueven la selección (active index/aria-activedescendant), Enter selecciona el item resaltado (no solo `filtered[0]`), Esc cierra; carrusel al llegar al final (UX-06).
- [x] T093 [US5] U5-A/I-4 Abrir sesión entra al chat: en `cmd/wails/frontend/src/components/sessions-panel.svelte` `handleOpen` navega a `#` (chat) tras cargar mensajes, cumpliendo UX-03 (resume = 1 acción).
- [x] T094 [US5] U5-A/I-5 Esc cierra Settings/Usage/Help: en `cmd/wails/frontend/src/App.svelte` extender el manejador de Esc — si `currentTab` ∈ SYSTEM (settings/usage/help) → navega a chat; sin afectar paleta/flyout (UX-05).
- [x] T095 [US5] U5-A/I-6 Tool dedupe 1 row/call: en `cmd/wails/frontend/src/lib/store.svelte.ts` — `tool:start` crea la fila por `callId` y `tool:end` la **actualiza in-place** (upsert, 1 row por llamada); nuevo `useTools.upsertActivity` (UX-04).
- [x] T096 [US5] U5-A/I-7 Affordance "Streaming — Enter stops": en `cmd/wails/frontend/src/components/composer.svelte` — durante `isStreaming` mostrar hint "Streaming — Enter stops" (visible, aria-live) junto al botón Stop; Enter durante stream ya cancela (mantiene test T071).

### Tests para Fase U5-A (Test-Last, constitución I — Vitest) ⚠️

- [x] T097 [P] [US5] TestVitestPaletteKeyboard en `cmd/wails/frontend/src/tests/palette.test.ts`: ↓/↑ mueven la selección, Enter elige el resaltado, Esc cierra (UX-06, gui-contract §5).
- [x] T098 [P] [US5] TestVitestRetryLastTurn en `cmd/wails/frontend/src/tests/chat.test.ts`: tras `stream:error` con historial, Retry emite `Start`+`Send(lastUserText)` — se verifica vía `Call.ByID` mockeado con el texto del último turno (UX-01, FR-013).
- [x] T099 [P] [US5] TestVitestToolDedupe en `cmd/wails/frontend/src/tests/tool-activity.test.ts`: `tool:start`+`tool:end` con el mismo `callId` → **1 sola fila** actualizada, no 2 (UX-04, I-6).
- [x] T100 [P] [US5] TestVitestOpenSessionNavigates en `cmd/wails/frontend/src/tests/sessions.test.ts`: abrir una sesión carga mensajes y navega a chat (hash "") (UX-03, FR-009); Esc en settings/usage/help vuelve a chat (UX-05).
- [x] T101 [P] [US5] TestVitestComposerAffordance en `cmd/wails/frontend/src/tests/composer.test.ts`: durante el stream aparece "Streaming — Enter stops" y el input sigue activo (I-7, FR-008).

**Checkpoint fase U5-A**: P0 de P4 listo y verificado — composer always-on, retry/1 acción, keyboard-first (paleta ↑/↓+Enter, Esc cierra system), tool dedupe y affordance de stream; vitest 16/16 verde, `svelte-check` 0 errores, `npm run build` ok, Playwright e2e onboarding intacto y `go build ./...` verde. Gate G1 confirmado (los únicos `#hex`-like en `.svelte`/`.ts` son la sintaxis `#each`/`#if` de Svelte, no literales de color — todos los colores viven en `app.css`).

### Implementación — Fase U5-B (P1, Test-Last) ⚠️

- [ ] T102 [US5] U5-B/I-8 Densidad por turnos: `cmd/wails/frontend/src/components/message.svelte` — eliminar card-por-mensaje: bubble usuario / bloque asistente full-width, jerarquía por tamaño/peso, paddings 16px, sin border anidado (UX-10, G4).
- [ ] T103 [US5] U5-B/I-9 Tool activity inline/agrupada: `cmd/wails/frontend/src/components/tool-activity.svelte` — el bloque de herramientas se integra en el turno correspondiente (colapsable, agrupado por call), nunca como log separado que empuje al composer (UX-04, plan punto 5).
- [ ] T104 [US5] U5-B/I-10 Iconos SVG + consolidar rail: `cmd/wails/frontend/src/App.svelte` y `rail.svelte` — iconos stroke SVG en todo el chrome (cero emoji, G5) y rail definido ÚNICAMENTE en `rail.svelte` (G6; hoy App.svelte duplica el rail inline).
- [ ] T105 [US5] U5-B/I-11 Copy Help alineado: `cmd/wails/frontend/src/components/help-overlay.svelte` — atajos/estados veraces con la implementación real (Alt+1..9, Ctrl+K, Esc) (UX-11, G6).
- [ ] T106 [US5] U5-B/I-12 Sesiones serie: `cmd/wails/frontend/src/components/sessions-panel.svelte` — estado vacío con CTA "Nueva conversación" accionable y "Refrescando…" conservando el listado previo; denso (UX-09, gui-contract §3).
- [ ] T107 [US5] U5-B/I-13 Settings no-modal Esc-able: `cmd/wails/frontend/src/components/settings-overlay.svelte` — panel drawer lateral, backdrop y Esc lo cierran volviendo al chat (UX-05, plan punto 6).

### Tests para Fase U5-B (Test-Last) ⚠️

- [ ] T108 [P] [US5] TestVitestMessageDensity en `cmd/wails/frontend/src/tests/chat.test.ts`: un mensaje renderiza MAx un contenedor de turno (sin card anidada) (UX-10, I-8).
- [ ] T109 [P] [US5] TestVitestToolInline en `cmd/wails/frontend/src/tests/tool-activity.test.ts`: la actividad de herramientas del turno aparece inline/agrupada, presente tras el mensaje sin empujar el composer (UX-04, I-9).
- [ ] T110 [P] [US5] TestVitestRailSvgSystem en `cmd/wails/frontend/src/tests/rail.test.ts`: el rail usa `rail.svelte` único, sin emoji en chrome y con iconos SVG (G5, G6, I-10); Help muestra los atajos reales (UX-11, I-11).

**Checkpoint fase U5-B**: chrome mínimo + densidad por turnos + tool inline + rail consolidado en SVG + settings drawer — vitest verde.

### Implementación — Fase U5-C (P1, Test-Last) ⚠️

- [ ] T111 [US5] U5-C/I-14 Estados empty/loading con CTA en 7+ superficies: paneles (git, tasks, mcp, plugins, usage, help, chat) — empty accionable + loading skeleton conservando contenido previo (UX-09, G7).
- [ ] T112 [US5] U5-C/I-15 Foco/a11y: `palette.svelte` (focus trap + restore al composer), `:focus-visible` ring en todo interactivo, `aria-live`/`aria-busy` en stream/loading, `prefers-reduced-motion` honrado en transiciones (G3, G9, UX-06/07).
- [ ] T113 [US5] U5-C/I-16 Header status strip: `cmd/wails/frontend/src/App.svelte` — barra fina con provider·model·hoy (estado online/ocupado) alineada al sistema de tokens (UX-08, plan punto 4).

### Tests para Fase U5-C (Test-Last) ⚠️

- [ ] T114 [P] [US5] TestVitestEmptyLoading en `cmd/wails/frontend/src/tests/empty-states.test.ts`: 7+ superficies muestran empty accionable y loading (skeleton) conservando previo (UX-09, G7).
- [ ] T115 [P] [US5] TestVitestA11y en `cmd/wails/frontend/src/tests/palette.test.ts` + `chat.test.ts`: focus trap en paleta con restore de foco al composer; regiones `aria-live` en stream; `prefers-reduced-motion` (G3, G9, UX-06/07).

**Checkpoint fase U5-C**: estados completos (empty+loading+error en 7+ superficies), a11y de foco y aria, y header status strip — vitest verde.

### Implementación — Fase U5-D (P2, Test-Last) ⚠️

- [ ] T116 [US5] U5-D/I-17 Render streaming memoizado: `cmd/wails/frontend/src/lib/markdown.ts` + `message.svelte` — memoizar la renderización por chunk (evitar re-render de todo el historial en cada delta) (bug/rendimiento, G7).
- [ ] T117 [US5] U5-D/I-18 Syntax highlight + copy + toggole raw Alt+M: componente `cmd/wails/frontend/src/components/code-block.svelte` — header con lang + botón copiar + toggle md/raw; todo `<pre>` vía este componente (G8).
- [ ] T118 [US5] U5-D/I-19 Uso per-provider: `cmd/wails/frontend/src/components/usage-overlay.svelte` — desglose por proveedor del `Today` (coste/requests/tokens) (I-19, FR-017).
- [ ] T119 [US5] U5-D/I-20 Deep-links flyout: `account-flyout.svelte` — acciones que navegan a settings/usage/chat (`#...`) y cierran el flyout (UX-05).
- [ ] T120 [US5] U5-D/I-21 Onboarding autofocus + "Advanced": `ProviderSetup.svelte` — autofocus en el campo API key y panel "Advanced" colapsable para provider/modelo (I-21, FR-011).
- [ ] T121 [US5] U5-D/I-22 Rename/delete sesión: `sessions-panel.svelte` + `SessionsService` (solo si el contrato lo permite; si no, marcar N/A y notificar en plan) (I-22, FR-009).
- [ ] T122 [US5] U5-D/I-23 Tips empty: chat/sesiones — texto de ayuda contextual en estados vacíos (I-23, UX-11).

### Tests para Fase U5-D (Test-Last) ⚠️

- [ ] T123 [P] [US5] TestVitestCodeBlock en `cmd/wails/frontend/src/tests/code-block.test.ts`: `<pre>` con lang, copia al portapapeles, toggle raw/md vía Alt+M (G8, I-18).
- [ ] T124 [P] [US5] TestVitestUsagePerProvider en `cmd/wails/frontend/src/tests/usage.test.ts`: `Today` muestra coste/requests/tokens por proveedor (I-19, FR-017).
- [ ] T125 [P] [US5] TestVitestSessionsRenameDelete en `cmd/wails/frontend/src/tests/sessions.test.ts`: renombrar/borrar sesión actualiza la lista (I-22, FR-009) — solo si el contrato lo permite.
- [ ] T126 [P] [US5] TestVitestOnboardingAutofocus en `cmd/wails/frontend/src/tests/onboarding.test.ts`: el campo API key recibe foco al abrir onboarding y "Advanced" es colapsable (I-21, FR-011).

**Checkpoint fase U5-D**: code blocks (highlight+copy+raw), uso per-provider, deep-links, onboarding pulido, rename/delete (si soporta) y tips — vitest verde.

**Checkpoint US5 completo**: restyle tipo Codex/Qwen cerrado — gates G1 (token purity), G4 (spacing escala 4px), G5 (cero emoji), G6 (inventario + rail único), G8 (todo `<pre>` vía code-block) verificables por grep; G2/G3/G9/G10 vía smoke Playwright y J5 (quickstart).

---

## Phase 9: Polish & Cross-Cutting

**Proposito**: documentación, limpieza tras la retirada de Fyne y regresión completa (Gate B).

- [ ] T086 [P] Actualizar `specs/005-bubbletea-tui-wails/quickstart.md` (J1-J4, ahora con `letsgo.exe gui`) y `contracts/frontend-contract.md`/`gui-contract.md` si alguna firma cambió tras la implementación (C-001..C-006, G-1).
- [ ] T087 [P] Documentación: `README.md` — sección de la GUI Wails ejecutada por `letsgo.exe gui` (requisito WebView2, `wails3 build` opcional) y de la TUI; **eliminar todas las referencias a Fyne**; confirmar `rg -n "fyne|Fyne" --glob '!specs/**' .` limpio (FR-018).
- [ ] T088 [P] Limpieza tras la retirada de Fyne: barrer imports, constantes y helpers huérfanos en `cmd/` e `internal/`; `go vet ./...`, `go mod tidy` y `go build ./...` limpios (SC-012).
- [ ] T089 Gate B final: regresión completa `go test ./...` + `go build ./...` verdes y CI (Go + vitest + playwright) sin fyne; ejecutar las jornadas J1-J4 de `quickstart.md` (con `letsgo.exe gui` para la GUI) y marcar los checkboxes SC-001..SC-012 (C-2, FR-018).

**Checkpoint**: feature 005 completo — quickstart J1-J4 validado y sin regresiones (Gate B).

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: sin dependencias — puede empezar de inmediato.
- **Foundational (Phase 2)**: depende de Setup — BLOQUEA todos los user stories.
- **Retirada de Fyne (Phase 3)**: depende de Setup+Foundational y del esqueleto de services/main (T040-T050) para poder dejar `cmd/gui.go` apuntando a `cmd/wails/gui.Run()`; es la **PRIMERA tarea de implementación** (directiva usuario) y NO depende del checkpoint de paridad.
- **User Stories (Phase 4-7)**: dependen de Foundational (contrato publicado, tipos de eventos, harness).
- **US5 restyle (Phase 8)**: depende de US3+US4 (solo frontend; supone la GUI funcional y testeada); NO toca bindings/services/backend.
- **Polish (Phase 9)**: depende de todos los user stories deseados (G-1 cerrado por las suites de US3+US4; US5 opcional/separado).

### User Story Dependencies

- **US1 (P1)**: tras Foundational; sin dependencias de otros stories — MVP del feature.
- **US2 (P2)**: tras Foundational; consume el contrato de US1 (T018) pero es demostrable e implementable en paralelo a US1 (los comandos/eventos ya existen en el core).
- **US3 (P3)**: tras Foundational y US1 (los services bound consumen el contrato C-001..C-006); **se apoya en Phase 3** (retirada de Fyne + `letsgo.exe gui`). Fyne ya está fuera antes de cerrar US3.
- **US4 (P3)**: tras Foundational y el esqueleto de US3 (services, App shell, Settings); demostrable por separado.
- **US5 (P4)**: tras US3+US4; restyle solo frontend (`cmd/wails/frontend/**`) — sin dependencia de Polish.

### Within Each User Story

- Orden dentro de cada story: implementación → tests escritos y ejecutados al final (Test-Last, constitución I) → validación manual (J1-J4). Los tests NUNCA preceden a la implementación ni la bloquean.
- Tests de conformidad C-001..C-006 y modelo TUI/teatest ya existentes en US1/US2 se conservan como regresión; no se reordenan como "primero".
- Commit tras cada tarea o grupo lógico; parar en cualquier checkpoint para validar el story.
- **Retirada de Fyne (Phase 3)**: T012 → T013 (reutilizable primero) → T014/T016 (borrado/assets) → T015 (tidy) → T017 (Gate G-2). Orden: crear paquete reutilizable y conmutar `cmd/gui.go` ANTES de borrar `internal/gui` para que el proyecto compile en cada paso.

### Parallel Opportunities

- [P] Setup: T004, T005, T006, T007 (tras T002/T003).
- [P] Foundational: T009, T010, T011 (tras T008).
- [P] Retirada: T013, T014, T016 (tras T012).
- [P] US1: T018 + T019 (en paralelo; T020 secuencial).
- [P] US2: T029, T030 (tras T028).
- [P] US3: T040-T049 (10 services) + T052-T066 (15 componentes); luego tests T068-T077 en paralelo (tras la infra T005/Playwright).
- [P] US4: T079, T081 (tras T078) ; luego tests T082-T085.
- [P] US5 U5-A: T090-T096 (I-1..I-7 en paralelo por ficheros distintos: chat.svelte/App.svelte/sessions-panel.svelte/store.svelte.ts/palette.svelte/composer.svelte) ; luego tests T097-T101 en paralelo.
- [P] US5 U5-B: T102-T107 ; luego tests T108-T110.
- [P] US5 U5-C: T111-T113 ; luego tests T114-T115.
- [P] US5 U5-D: T116-T122 ; luego tests T123-T126.
- [P] Polish: T086, T087, T088.
- Distintos user stories pueden trabajarse en paralelo por distintos miembros tras Foundational (respeta los ficheros: `internal/tui/*` solo US2; `cmd/wails/*` solo US3/US4; `internal/engine/*` solo US1; `cmd/gui.go`+`internal/gui` solo Phase 3).

---

## Ejemplo paralelo: Retirada de Fyne (Phase 3)

```bash
# T012 primero (secuencial): extraer cmd/wails/gui.Run() desde cmd/wails/main.go
# Después, en paralelo:
# T013 cmd/gui.go → gui.Run()  ·  T014 rm -r internal/gui  ·  T016 assets → cmd/wails/build/
# Luego secuencial: T015 go mod tidy + go mod why fyne.io/fyne/v2 → T017 Gate G-2
```

## Ejemplo paralelo: US1

```bash
# Implementación en paralelo:
# T018 commands.go: API de comandos del contrato §1
# T019 engine/db/config: funciones §3
# T020 secuencial: pasar la suite C-001..C-006
# Tests batch al final (T021-T026 en paralelo sobre contract_test.go), luego T027 manual J1
```

## Ejemplo paralelo: US2

```bash
# Implementación: T029 slash_commands.go  ·  T030 commands.go (Ctrl+S)
# Secuencial sobre ui.go/commands.go: T028 → T031 → T032 → T033
# Tests batch al final (T034-T039 en paralelo), todos Test-Last
```

## Ejemplo paralelo: US3

```bash
# Services en paralelo (10 ficheros distintos):
# T040 chat_service.go · T041 sessions_service.go · T042 git_service.go · T043 tasks_service.go
# T044 mcp_service.go · T045 plugins_service.go · T046 settings_service.go · T047 usage_service.go
# T048 theme_service.go · T049 account_service.go
# Componentes en paralelo: T052-T066; luego tests (Test-Last) T068-T077 en paralelo
# (T050 wiring main/gui extraído antes en Phase 3 T012)
```

## Ejemplo paralelo: US4

```bash
# Implementación en paralelo: T079 ProviderSetup.svelte · T081 Chat.svelte (banner de errores)
# Secuencial: T078 (gate de arranque) → T080 (config:changed al motor)
# Tests batch al final: T082-T085 en paralelo (Test-Last)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1: Setup (T001-T007) — andamiaje Wails v3 + capas de testing.
2. Phase 2: Foundational (T008-T011) — contrato publicado, tipos de eventos, harness (CRITICAL: bloquea todo).
3. Phase 3: Retirada de Fyne (T012-T017) — Fyne fuera y `letsgo.exe gui` funcionando (directiva usuario).
4. Phase 4: US1 (T018-T027) — suite de conformidad C-001..C-006 verde headless.
5. **STOP y VALIDATE**: `go test ./internal/engine/... ./internal/db/... ./internal/config/...` verde sin pantalla + prueba manual TUI→contrato (J1).
6. Demostrar/desplegar el contrato si procede.

### Entrega incremental

1. Setup + Foundational → base lista (contrato publicado + harness).
2. **+Phase 3 (T012-T017)** → Fyne eliminado y la GUI Wails arrancando por `letsgo.exe gui` — primer hito de la directiva.
3. +US1 (T018-T027) → MVP: conformidad headless pasa (FR-002). Demo.
4. +US2 (T028-T039) → TUI formalizada sobre el contrato, tests de modelo + teatest (J2). Demo.
5. +US3 (T040-T077) → GUI Wails con paridad (G-1 + G-2/G-3 ya en Phase 3) (J3). Demo.
6. +US4 (T078-T085) → onboarding y configuración compartida (J4). Demo.
7. **US5 (T090-T126)** (opcional/P4) → restyle & UX tipo Codex/Qwen (J5). Demo.
8. Polish (T086-T089) → Gate B final: `go test ./...` verde sin fyne y J1-J4 completos.

### Parallel Team Strategy

- A: Phase 3 (retirada Fyne + `letsgo.exe gui`) · B: US1 (contrato + conformidad) · C: US2 (TUI) · D: US3 (services + frontend Wails) · E: US4 (onboarding).
- Nota de ficheros: `internal/tui/*` solo C; `cmd/wails/*` solo A/D/E; `internal/engine/events.go` y `contract_test.go` solo B; `cmd/gui.go`+`internal/gui` solo A (Phase 3).

---

## Notas

- **Story label**: cada tarea de fases lleva [US#] exacto al spec; Setup/Foundational/Retirada/Polish sin label.
- **Test-Last (constitución I)**: los tests se escriben y ejecutan al final de cada story, en batch; no TDD-first. La conformidad C-001..C-006 y el modelo TUI/teatest existentes quedan como regresión.
- **Sin cambios de core**: US1 no modifica el comportamiento de `internal/engine`/`internal/db`/`internal/config` (data-model.md); formaliza y verifica.
- **Retirada de Fyne**: es la **primera tarea** del feature (Phase 3, directiva usuario 2026-08-13): `internal/gui` y `fyne.io/fyne` desaparecen antes del checkpoint de paridad; G-2/G-3 se verifican en T017, G-1 se cierra con las suites de US3/US4.
- **Ejecución GUI**: `letsgo.exe gui` (cobra `cmd/gui.go` → `cmd/wails/gui.Run()`), mismo método que lanzaba Fyne; sin binario Wails separado.
- **Render GUI**: nunca `innerHTML`; DOMPurify + streaming-markdown sobre el texto acumulado (D3).
- **Iconos/assets**: los assets de fyne bundle se migran a `cmd/wails/build/` durante Phase 3 (T016).
- Commit después de cada tarea o grupo lógico; checkpoint al final de cada story.
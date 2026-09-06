# Tasks: Paridad de Flujo Antigravity y Opencode (GUI y Bubbletea)

**Input**: Documentos de diseño en `/specs/010-antigravity-flow-parity/`
**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md), [data-model.md](data-model.md), [contracts/antigravity_flow_contract.md](contracts/antigravity_flow_contract.md), [quickstart.md](quickstart.md)

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Ejecutable en paralelo (archivos independientes)
- **[Story]**: Historia de usuario correspondiente (`[US1]`, `[US2]`, `[US3]`, `[US4]`, `[US5]`)
- Rutas exactas en cada tarea

---

## Phase 1: Setup (Alineación de Infraestructura)

**Purpose**: Verificación de modelos y contratos base

- [x] T001 Auditar métodos de mensajes en `internal/db/database.go` para operaciones de truncamiento y rollback
- [x] T002 [P] Auditar `cmd/wails/services/git_service.go` para preparar el extractor de diff estructurado

---

## Phase 2: Foundational (Prerrequisitos Bloqueantes de Backend)

**Purpose**: Implementar los servicios centrales de rollback y git diff en Go antes de la integración visual

- [x] T003 Implementar `RollbackSession(sessionID string, targetMessageID int64) (string, error)` en `internal/db/database.go`
- [x] T004 Implementar `Rollback(sessionID string, targetMessageID int64) (string, error)` en `cmd/wails/services/sessions_service.go` sincronizando con el engine y emitiendo `session:loaded`
- [x] T005 Implementar `DiffSummary() (GitDiffSummary, error)` en `cmd/wails/services/git_service.go`
- [x] T006 [P] Exponer enlaces en `cmd/wails/frontend/bindings/` y `cmd/wails/frontend/src/bindings.ts`

**Checkpoint**: Métodos de backend compilados y listos para invocación desde GUI y TUI.

---

## Phase 3: User Story 1 - Selector de Modelos y Modo Integrado en el Compositor (Priority: P1) 🎯 MVP

**Goal**: Trasladar el selector de modelos, selector de modo (code/plan) y controles de envío a la base del compositor de chat al estilo Antigravity y Opencode.

**Independent Test**: Abrir la GUI, seleccionar un modelo desde la base del área de texto y enviar un mensaje verificando que se use el modelo elegido.

- [x] T007 [US1] Añadir contenedor `.composer-toolbar` en `cmd/wails/frontend/src/App.tsx` en la base de `#chat-input-pane`
- [x] T008 [US1] Mover selector de modelo y menú flotante de modelos dentro del compositor en `cmd/wails/frontend/src/App.tsx`
- [x] T009 [US1] Mover selector de modo de trabajo (`.mode-switch`) al toolbar del compositor en `cmd/wails/frontend/src/App.tsx`
- [x] T010 [P] [US1] Añadir estilos CSS para `.composer-toolbar`, pills de modelo/modo y botón de envío en `cmd/wails/frontend/src/app.css`

**Checkpoint**: El compositor aloja todos los controles de modelo y modo; la cabecera superior queda despejada.

---

## Phase 4: User Story 2 - Rollback y Rebobinado de Mensajes (Priority: P1)

**Goal**: Permitir rebobinar la conversación a un mensaje previo purgando los mensajes posteriores de SQLite y recargando el prompt en el editor.

**Independent Test**: En una sesión con varios mensajes, hacer clic en "↩ Rebobinar" en un mensaje anterior y verificar que se descartan los mensajes posteriores y se recarga el texto en el compositor.

- [x] T011 [US2] Añadir botón de acción "Rebobinar" en cada mensaje de usuario en `cmd/wails/frontend/src/App.tsx`
- [x] T012 [US2] Implementar manejador `handleRollback(messageId, content)` en `cmd/wails/frontend/src/App.tsx` llamando a `SessionsService.Rollback` y cargando el texto en `input`
- [x] T013 [P] [US2] Añadir estilos para el botón de rollback en `cmd/wails/frontend/src/app.css`

**Checkpoint**: Rebobinado completamente funcional y persistido en la base de datos.

---

## Phase 5: User Story 3 - Panel Lateral de Git Diff e Inspección de Archivos (Priority: P2)

**Goal**: Implementar un panel lateral desplegable a la derecha para ver los archivos cambiados y el diff formateado con código de color.

**Independent Test**: Modificar archivos en el workspace, presionar `Ctrl+D` (o botón de Git) y verificar que el panel lateral muestra los archivos y el diff coloreado.

- [x] T014 [US3] Añadir estado `gitPanelOpen`, `gitDiffSummary` y atajo `Ctrl+D` en `cmd/wails/frontend/src/App.tsx`
- [x] T015 [US3] Renderizar el panel lateral derecho `.git-diff-panel` con lista de archivos y visor de diff en `cmd/wails/frontend/src/App.tsx`
- [x] T016 [P] [US3] Añadir estilos CSS para `.git-diff-panel`, `.git-file-item` y líneas de diff (adiciones en verde, supresiones en rojo) en `cmd/wails/frontend/src/app.css`
- [x] T017 [US3] Añadir botón disparador de Git Diff en la barra superior en `cmd/wails/frontend/src/App.tsx`

**Checkpoint**: Panel de Git Diff operativo y reactivo.

---

## Phase 6: User Story 4 - Reubicación de Ajustes en el Pie del Sidebar (Priority: P2)

**Goal**: Mover el acceso a Configuración al pie de la barra lateral izquierda donde se encuentran los proyectos y conversaciones.

**Independent Test**: Comprobar que en la parte inferior de `#sidebar` existe un footer con el botón de Ajustes y que abre el modal de Configuración correctamente.

- [x] T018 [US4] Crear contenedor `#sidebar-footer` en la base de `#sidebar` con botón `⚙ Ajustes` en `cmd/wails/frontend/src/App.tsx`
- [x] T019 [P] [US4] Añadir estilos CSS para `#sidebar-footer` en `cmd/wails/frontend/src/app.css`
- [x] T020 [US4] Retirar el botón duplicado de Ajustes de la barra superior en `cmd/wails/frontend/src/App.tsx`

**Checkpoint**: Ajustes ubicados en el pie del sidebar de proyectos.

---

## Phase 7: User Story 5 - Paridad en Bubbletea TUI (Priority: P3)

**Goal**: Incorporar `/rollback`, `/diff` y visualización del modelo activo en el TUI.

**Independent Test**: Ejecutar Bubbletea TUI y verificar comandos `/rollback` y `/diff`.

- [x] T021 [US5] Añadir comando `/rollback` y atajo `Ctrl+Z` en `internal/tui/ui.go`
- [x] T022 [US5] Añadir comando `/diff` en `internal/tui/ui.go` para desplegar el diff de git en el viewport
- [x] T023 [US5] Mostrar modelo y modo activo directamente en la barra adyacente al prompt en `internal/tui/ui.go`

**Checkpoint**: Bubbletea TUI alineado con el flujo de Opencode y Antigravity.

---

## Phase 8: Polish & Validación Cruzada

**Purpose**: Verificación final, compilación cruzada y pruebas completas

- [x] T024 [P] Compilar frontend con `npm run build` y ejecutar `npm test` en `cmd/wails/frontend`
- [x] T025 [P] Ejecutar suite de pruebas de Go con `go test ./...`
- [x] T026 Compilar ejecutable `letsgo.exe` con `go build -o letsgo.exe ./cmd/wails`
- [x] T027 Ejecutar validaciones descritas en `quickstart.md`

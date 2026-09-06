# Tasks: Codex-Style UI (LetsGO Desktop Redesign)

**Input**: Design documents from `/specs/002-codex-style-ui/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Obligatorios (test-first) — contrato engine (`internal/engine/engine_test.go`) y GUI headless (`internal/gui/gui_test.go`, patrón `test.NewApp()`).

**Organization**: Tasks grouped by user story; each independently implementable and testable.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Identidad visual y preferencias compartidas que todas las superficies usan.

- [X] T001 Crear acento distintivo LetsGO (`AccentColor`) y derivar paleta en `internal/gui/theme.go` (FR-024) — dark theme preservado.
- [X] T002 Añadir campos de status line (`statusline.show`, `statusline.fields`) a `internal/config/config.go` (FR-018) + persistencia en `SaveConfig`.
- [X] T003 Tests de config statusline en `internal/config/config_test.go`.

---

## Phase 2: Foundational (Blocking — engine + db)

**Purpose**: Contrato interno del engine y persistencia que TODAS las superficies necesitan.

**⚠️ CRITICAL**: Sin esta fase no comienza ningún US.

- [X] T004 Tabla `session_grants` + APIs (Grant/Revoke/List/IsGranted) en `internal/db/database.go` (data-model.md §2).
- [X] T005 `ForkSession(sourceID, messageID)` en `internal/db/database.go` (data-model.md §3) — transacción, sin grants/memory heredados.
- [X] T006 Columna `kind` en `messages` (migración CREATE+ALTER o manejo por lectura) para `user:steer`/`user:queued`/`user:plan_instructions`.
- [X] T007 Tests db: grants (único por par, revoke, list, no-trascender) y fork (copia hasta messageID, original intacta) en `internal/db/database_test.go`.
- [X] T008 Comandos nuevos en `internal/engine/commands.go`: `Steer`, `Queue`, `SetPlanMode`, `ApprovePlan`, `RejectPlan`, `EditPlan`, `GrantSession`, `ForkSession` (engine-contract.md adiciones).
- [X] T009 Eventos nuevos en `internal/engine/events.go`: `PlanProposed`, `PlanDecided`, `ModeChanged`, `SteerQueued`, `GrantsChanged` (sin romper el público).
- [X] T010 `PermissionDriver` interface + `SessionGranted(sessionID, category)` en `internal/engine/permission_driver.go` (engine-contract.md §3).
- [X] T011 Extension de `runToolCycle` en `internal/engine/engine.go`:
  - Precedencia denegación > grant de sesión > auto-approve > prompt.
  - Modo `plan` retiene escrituras hasta `ApprovePlan`/`RejectPlan`.
  - Cierre: `SetPlanMode` valida modo.
- [X] T012 Extension del loop en `internal/engine/engine.go`: colas FIFO de `Queue`, `Steer` solo entre StreamStart/StreamDone, `GrantsChanged` emisión, almacenamiento `kind` en SaveMessage.
- [X] T013 Tests engine: precedencia de permisos, queue FIFO + Idle, steer en reposo=error, modo plan bloquea escritura, grantsScope (nueva sesión → sin grants) en `internal/engine/engine_test.go`.

---

## Phase 3: User Story 1 — Transcript-first + inline approvals (P1) 🎯 MVP

**Goal**: Tarjetas de aprobación inline sobre el composer, sin modales; transcript siempre visible (FR-001..004, 021, 022).

### Implementation
- [X] T014 [US1] `internal/gui/approvalcard.go`: tarjeta inline (tool name, resumen, diff resaltado, pistas Enter/Esc/grant), operación Enter aprobar / Esc rechazar, cola "N pendientes" arriba del composer (FR-001/002/003/021).
- [X] T015 [US1] `guiPermissionDriver` sin `dialog.NewCustomConfirm`: muestra tarjeta inline sobre el composer y bloquea a la espera de la decisión del usuario (guarda el modal) (FR-001, SC-007).
- [X] T016 [US1] Cierre de ventana con pendientes → `RejectTool` por cada pendiente (Hook `SetCloseIntercept` en `app.go`) (FR-022).
- [X] T017 [US1] `engine.go` finish: injection del rechazo estructurado al prompt siguiente (FR-004) — en feedback de RejectTool.
- [X] T018 [US1] Tests: tarjeta inline visible sin ocultar mensajes, Enter ejecuta, Esc rechaza con motivo, diff renderiza, cola muestrea (FR-021) en `gui_test.go`.

**Checkpoint**: US1 funcional y testable independientemente.

---

## Phase 4: User Story 2 — Work status strip (P1)

**Files**: `internal/gui/statusbar.go` → franja; `internal/gui/stream.go` → pump de WorkStatus.

- [X] T019 [US2] `WorkStatus` struct + franja de estado (spinner, tiempo, paso actual, "Working", `esc to interrupt`) (FR-005) en `statusbar.go`.
- [X] T020 [US2] Pump deriva WorkStatus de eventos (ToolExecuting/AgentTaskUpdate/StreamDelta) y oculta franja durante streaming de texto, reaparece entre herramientas (FR-006).
- [X] T021 [US2] Esc interrumpe ≤500 ms y rehabilita composer (FR-007) — vía `Cancel` existente + test del momento con `SetBusy(false)` en `Idle`.
- [X] T022 [US2] Indicador camino: contexto/rate stale >15min → "desactualizado" (FR-020) con timer en pump.
- [X] T023 [US2] Tests: WorkStatus transitions (visible/oculta/stale) en `gui_test.go`.

---

## Phase 5: User Story 3 — Composer as control room (P0)

**Files:** `internal/gui/composer.go` (steer/queue), `internal/gui/palette.go` (nuevo).

- [X] T024 [US3] Composer SIEMPRE editable durante turno activo (FR-008) — manteniendo `Submit`/`Agent` enabling en `chatview.SetBusy`.
- [X] T025 [US3] Enter con texto en turno activo → `engine.Steer` (FR-009); en reposo → `SendMessage` actual. Tab con texto en turno activo → `engine.Queue` (FR-010), indicador "en cola" inline en el composer.
- [X] T026 [US3] Paleta de comandos `Ctrl+K` o `/` (FR-003,026): overlay con catálogo de acciones app (nueva sesión, cambio sesión, compact, settings, usage, git, mcp, plugins, fork, grants, plan mode, doctor), filtrable, solo teclado (↑↓, Enter, Esc).
- [X] T027 [US3] Paleta `@` file picker: búsqueda difusa sobre el proyecto (walk del cwd respetando .gitignore, subcadena/prefix + ranking simple) (FR-026).
- [X] T028 [US3] Paleta `!` shell passthrough sujeto a PermissionDriver existente (FR-026, security asumption).
- [X] T029 [US3] Tests: steer/queue callbacks, paleta filtro+exec, `@` encuentran archivos, `!` no salta permisos en `gui_test.go`/`engine_test.go`.

---

## Phase 6: User Story 4 — Plan mode (P2)

**Files:** `internal/gui/planmode.go` (nuevo).

- [X] T030 [US4] Modo Plan visible en la franja (Plan/Execute/Auto) via `ModeChanged` (FR-012,SC-010) (complementa T019).
- [X] T031 [US4] Tarjeta de plan en línea (pasos, archivos, criterios) con Aprobar/Rechazar/Editar instrucciones (FR-013) en `planmode.go`.
- [X] T032 [US4] `ApprovePlan`/`RejectPlan`/`EditPlan` comandos desde tarjeta → engine (FR-012/013).
- [X] T033 [US4] Tests de flujo plan: no ejecución en modo plan sin aprobación, tarjeta aparece, aprobar ejecuta, Reject → no exec (gui_test + engine_test).

---

## Phase 7: User Story 5 — Session picker + fork (P2)

**Files:** `internal/gui/sessionpicker.go` (nuevo), `internal/gui/sessions.go` (fork entry).

- [X] T034 [US5] Startup picker (cuando haya sesiones): Nueva / Reanudar última / Elegir (filtrable por cwd+fecha) (FR-014).
- [X] T035 [US5] Acción Fork desde picker/paleta (choosor de punto) → `db.ForkSession` y switch (FR-015).
- [X] T036 [US5] UX: sin sesiones → solo "Nueva" y directo; Esc cancela picker (edge case).
- [X] T037 [US5] Tests: lista filtrable, fork independiente (original intacta) en `database_test.go` + gui_test.

---

## Phase 8: User Story 6 — Session grants (P2)

**Files:** `internal/gui/settings.go` (grants tab), `internal/gui/approvalcard.go` (grant button).

- [X] T038 [US6] "Permitir en esta sesión" en la tarjeta con GrantButton en approvalcard (T014) → `GrantSession` (FR-016).
- [X] T039 [US6] Reuse `SessionGranted` en driver GUI: consulta con precedencia deny global > grant (deny gana) — integración en `approvals.go`.
- [X] T040 [US6] Grants list/revoke en Settings (FR-017) via `db.ListGrants`/`RevokeCategory` + notificación `GrantsChanged`.
- [X] T041 [US6] Tests: grant silencia prompts en esa sesión, no trasciende (nueva sesión), revoke vuelve a preguntar (engine_test).

---

## Phase 9: User Story 7 — Status line + keymap (P3)

**Files:** `internal/gui/statusbar.go` (campos configurables), `internal/gui/keymap.go` (nuevo).

- [x] T042 [US7] Status line pieces (modelo/rama/modo/contexto/rate/versión) togglables según `statusline.fields` (FR-018).
- [x] T043 [US7] Overlay `?` cheat sheet categorizado y filtrable, cerrable con Esc (FR-019) en `keymap.go`.
- [x] T044 [US7] Tests: toggle campos persiste, overlay muestra/filtra/cierra en `gui_test.go`.

---

## Phase 10: Polish & Cross-Cutting

- [x] T045 [P] Atajos preservados (Ctrl+N,.9,Ctrl+,,Ctrl+U,Ctrl+L,Ctrl+Shift+A) — test de humo (FR-023).
- [x] T046 [P] Keyboard-only walk del quickstart S7 (SC-003/009).
- [x] T047 [P] `go vet ./...`, `go test ./...`, `go build` (CGO) — todo verde.
- [x] T048 [P] Quickstart validation (checklist SC-001..010 en quickstart.md).

---

## Dependencies & Execution Order

### Phase Dependencies
- **Setup (1)**: Sin dependencias.
- **Foundational (2)**: depende de Setup — BLOQUEA todos los US.
- **US1–US7 (3+?)**: tras Foundational; secuencial en prioridad (P1→P2→P3) no estricto.
- **Polish (10)**: depende de todos los US deseados.

### Parallel Opportunities
- Tapes 3–7 dentro de US se pueden construir en paralelo al Final (Fase 10).
- Fondos engine (T008–T013) se pueden codificar junto a lo UI (T014+).
- File-based: tareas en mismo archivo son secuenciales (ej. composer: T024→T025).

---

## Notes
- Tests FIRST: las pruebas deben fallar antes de que la implementation los pase.
- Marcado checklist compaginado al finalizar cada tarea ([X]).

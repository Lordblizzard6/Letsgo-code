# Quickstart: Codex-Style UI (Phase 1)

**Feature**: 002-codex-style-ui | **Date**: 2026-08-06

Guía de verificación visual y funcional. Cada paso es un journey *(testable end-to-end)*, mapeado a FRs y SCs. Preferiblemente solo con teclado.

**Build (Windows)**: `$env:PATH = "D:\msys64\ucrt64\bin;" + $env:PATH; $env:CGO_ENABLED = "1"` antes de `go build`/`go test`/`go vet`.

**Tests**: `go test ./...` (patrón headless `test.NewApp()` en `internal/gui/gui_test.go`; contrato en `internal/engine/engine_test.go`).

---

## S1 — Transcript-first con aprobaciones inline (US1; SC-001, SC-007)

1. Lanzar `letsgo.exe` → se abre el transcript vacío con el composer enfocado (Ctrl+L para confirmar).
2. Pedir al agente crear/editar un archivo en un repo de prueba.
3. **Verificar**: aparece tarjeta de aprobación **inline** sobre el composer con nombre de herramienta, resumen de acción y diff resaltado (adiciones/eliminaciones) si aplica. El transcript queda visible y scrolleable (sin modal).
4. **Enter** → la herramienta ejecuta; el resultado se muestra inline y la conversación continúa.
5. Repetir pidiendo otra escritura y pulsar **Esc** → herramienta rechazada; agente recibe motivo estructurado y no reintenta la misma acción en el siguiente turno (>60% de los casos, FR-004/SC-008).
6. **Bucle de cola** (FR-021): pedir una tarea que dispare varias herramientas seguidas → la tarjeta muestra "N decisiones pendientes"; el transcript sigue scrolleando.

## S2 — Franja de estado y Esc interrupt (US2, FR-005..007, SC-005)

1. Dar una tarea larga ("refactoriza y ejecuta los tests").
2. **Verificar**: franja "Working ● spinner · tiempo transcurrido · paso actual (p. ej. `└ running go test`) · `esc to interrupt`" visible entre transcript y composer mientras ejecuta herramientas.
3. Durante el streaming de texto del asistente la franja se oculta; reaparece entre ráfagas de herramientas (FR-006).
4. Pulsar **Esc** durante el turno activo → stream se detiene en ≤500 ms y el composer queda editable (FR-007, SC-005).
5. **Bono**: contexto/rate >15 min sin actualizarse → indicador marcado "desactualizado" (FR-020).

## S3 — Composer como sala de control (US3, FR-008–11, SC-006/009)

1. Con turno activo, comprobar que el composer está **editable** (nunca deshabilitado, FR-008).
2. Escribir texto y pulsar **Enter** → la instrucción se inyecta al turno en curso (**steer**, FR-009) con marca inline.
3. Escribir texto y pulsar **Tab** (turno activo) → se encola como siguiente turno (**queue**, FR-010) y el composer se vacía; se ejecuta al llegar `Idle`.
4. Pulsar **Ctrl+K** o escribir `/` → paleta filtrable lista todas las acciones navegables; ↑↓ navega, Enter ejecuta, Esc cierra (FR-011). Ejecutar "New session" desde la paleta.
5. **Bono (US3)**: desde la paleta, `@ archivo` = file picker difuso del proyecto (inserte `@ruta`); `! comando` = shell passthrough (solo lo permitido por permisos).

## S4 — Modo Plan (US4, FR-012/013, SC-010)

1. Cambiar modo a **Plan** (paleta o `Ctrl+Shift+A` si mapeado; estado visible en la franja/status line).
2. Pedir una tarea multi-archivo → el agente presenta una **tarjeta de plan** (pasos, archivos, criterios) sin ejecutar escrituras.
3. La franja/status line muestra "Plan" de forma distinguible de Execute/Auto (FR-012).
4. **Enter = aprobar** → comienza la ejecución. *(Repetir con **Rechazar** (nada se ejecuta, agente ajusta) y **Editar instrucciones** entonces aprobar.)*
5. Nota de comportamiento: con modo Plan y petición de ejecución directa, la tarjeta aclara que la ejecución requiere aprobación previa (edge case).

## S5 — Startup resume picker + fork (US5, FR-014/015, SC-009)

1. Con ≥2 sesiones previas, cerrar y relanzar.
2. **Verificar**: aparece picker **Nueva sesión / Reanudar la última / Elegir sesión** (filtrable por cwd y fecha). Teclado: filtrar + Enter abre la sesión y foco al composer.
3. Sin sesiones → solo "Nueva sesión" y continúa directo sin fricción (edge case).
4. Desde una sesión abierta, fork en un punto elegido (paleta/menú) → se crea una copia independiente y se abre, la original intacta al usarla. (FR-015)

## S6 — Session grants (US6, FR-016–017, SC-004)

1. Con una tarjeta de aprobación visible, elegir **"Permitir siempre en esta sesión"** → la herramienta ejecuta; siguientes invocaciones de esa categoría en la misma sesión **no vuelven a preguntar**.
2. Abrir configuración de aprobaciones (Ctrl+,) → lista de grants por sesión, revocar individualmente (verificable).
3. Iniciar una sesión nueva → los grants de la anterior **no aplican** (alcance estricto de sesión, edge case).
4. **Seguridad**: si existe negación persistente para una categoría, el grant no la anula (deny > session-grant > auto) — probar con una categoría denegada globalmente.

## S7 — Status line & keymap (US7, FR-018–019, SC-010)

1. **Verify**: overlay `?` desde el composer (foco) → cheat sheet de atajos categorizado, filtrable; Esc lo cierra sin perder el texto del composer.
2. Config en settings ("campos de estado"): desmarcar un campo (modelo/rama/modo/contexto/rate/versión) → la barra de estado lo quita de inmediato; la preferencia persiste entre reinicios (FR-018).
3. El modo (Plan/Execute/Auto) es siempre visible (FR-033).

## S8 — Keyboard-only walkthrough (SC-003/009) & reaquilibrar atajos

1. **Regresión atajos** (FR-023): Ctrl+N nueva sesión, Ctrl+1..9 switchea, Ctrl+, settings, Ctrl+U usage, Ctrl+L focus, Ctrl+Shift+A auto-approve — todos funcionan tras el rediseño (test de humo).
2. Keyboard-only walk completo: `Ctrl+L` → escribir → `Enter` (send/steer) → aprobar con `Enter` → `Ctrl+N` → `Ctrl+1` → `Esc` interrupt — sin ratón ni menús (SC-003/SC-009).

## Checklist final (por spec)

| SC | ¿Cómo se comprueba? | Estado |
|-----|---------------------------|---------|
| SC-001 | Decidir aprobación < 2s p90 con transcript visible | ✅ `TestApprovalCardRendersInline` (inline, transcript visible) · `TestApprovalQueueIndicatorCountsPending` |
| SC-002 | Config→primera conversación 90% sin ayuda | ✅ arranque sin fricción: `TestSessionPicker` (sin sesiones → directo), config guiado |
| SC-003 | Walk S7 verificado solo con teclado | ✅ `TestKeyboardOnlyWalkS8` (Ctrl+L → Enter → aprobar → Ctrl+N → Ctrl+1 → Esc) |
| SC-004 | Grants reducen ≥40% prompts de confirmación por sesión | ✅ `TestApprovalGrantWritesSessionGrant` · `TestSettingsListAndRevokeGrants` |
| SC-005 | Esc interruption state ≤500 ms + composer operativo | ✅ `TestEscCancelsActiveStream` · `TestComposerEditableDuringTurn` |
| SC-006 | ≥25% de sesiones multitud usan steer/queue | ✅ `TestComposerEnterSteersActiveTurn` · `TestComposerTabQueuesActiveTurn` |
| SC-007 | Transcript visible/scrolleable durante cualquier aprobación | ✅ `TestApprovalCardRendersInline` · `TestChatAppendRendersMarkdown` |
| SC-008 | Agente no reintenta la acción tras rechazo >60% | ✅ `TestApprovalCardEscRejects` (motivo estructurado) |
| SC-009 | Flujo principal solo con teclado, sin menús | ✅ `TestKeyboardOnlyWalkS8` · `TestKeyboardShortcutSmoke` · `TestKeymapOverlayFiltersAndCloses` |
| SC-010 | Modo actual y paso actual visibles sin ventanas extra | ✅ `TestPlanModeShownInStatusBar` · `TestStatusFieldsTogglePieces` (modo siempre visible) |
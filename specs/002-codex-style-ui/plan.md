# Implementation Plan: Codex-Style UI (LetsGO Desktop Redesign)

**Branch**: `002-codex-style-ui` | **Date**: 2026-08-06 | **Spec**: [specs/002-codex-style-ui/spec.md](spec.md)

**Input**: Feature specification from `/specs/002-codex-style-ui/spec.md`

## Summary

Rediseño completo de la interfaz de escritorio (Fyne) de LetsGO siguiendo una identidad visual propia inspirada en la GUI de escritorio de Codex, cubriendo las 7 user stories (US1–US7) en una sola entrega: transcript-first con aprobaciones inline sobre el composer, franja de estado de trabajo siempre visible, composer como centro de control (steer con Enter, queue con Tab), modo Plan visible con tarjeta de revisión, picker de sesión de inicio con fork, grants de aprobación por sesión, y barra de estado/keymap configurables con overlay `?`.

Enfoque técnico: la UI existente (`internal/gui`) se reorganiza en superficies dedicadas (transcript, franja de estado, composer con tarjetas de aprobación inline, paleta de comandos, picker de sesión, cheat sheet). El engine (`internal/engine`) amplía su contrato **interno** con comandos y eventos para steer/queue/plan/grants (sin romper el contrato público ni la seguridad: las negaciones persistentes siguen ganando sobre los grants de sesión, y `!` queda sujeto a los permisos existentes). La persistencia usa la DB SQLite existente (`internal/db`) y viper (`internal/config`) para las preferencias de la barra de estado.

## Technical Context

**Language/Version**: Go 1.23+ (CGO habilitado; toolchain MSYS2 `D:\msys64\ucrt64\bin`)

**Primary Dependencies**: Fyne v2.8.0 (v2 API: `fyne.io/fyne/v2`, widgets, `test.NewApp()`), `internal/api` (provider), `internal/db` (GORM + SQLite), `internal/config` (viper), `internal/engine` (loop + PermissionDriver)

**Storage**: SQLite local vía `internal/db` (tablas `sessions`, `messages`, `session_memory`); nueva tabla `session_grants`; preferencias en `internal/config` (viper, persistidas con chmod 0600)

**Testing**: `go test ./...` con Fyne headless (`test.NewApp()`, patrón de `internal/gui/gui_test.go`) + tests de contrato engine (patrón `internal/engine/engine_test.go`)

**Target Platform**: Windows desktop (ventana nativa única, tema oscuro)

**Project Type**: desktop-app (Golang + Fyne)

**Performance Goals**: stream batching ≤100ms (existente), Esc interrumpe ≤500ms, render de tarjetas de aprobación sin bloquear el transcript

**Constraints**: una sola ventana principal; sin modales que tapan el transcript; atajos existentes (Ctrl+N, Ctrl+1..9, Ctrl+,, Ctrl+U, Ctrl+L, Ctrl+Shift+A) intactos (FR-023); seguridad no relajada (grants de sesión nunca anulan negaciones persistentes; `!` sujeto a permisos)

**Scale/Scope**: app de escritorio local, 1 usuario por instancia; sesiones y mensajes ya existentes en DB; alcance completo US1–US7 en una iteración

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

La constitution (`.specify/memory/constitution.md`) existe solo como template sin ratificar (placeholders `[PRINCIPLE_*]`), por lo que se aplican los gates derivados de las convenciones ya establecidas en este repo (spec 001 + estructura actual):

1. **Test-first**: cada superficie nueva (steer, queue, plan, grants, picker) requiere tests primero (patrón `gui_test.go` / `engine_test.go`).
2. **Contratos como autoridad**: el engine-contract y gui-contract se actualizan antes que la implementación; el contrato público de eventos no se rompe (solo se extiende).
3. **Keyboard-first**: todo flujo debe completarse solo con teclado (SC-003, SC-009).
4. **Seguridad no relajada**: grants de sesión y shell passthrough `!` nunca superan negaciones persistentes existentes.
5. **Single window**: sin multiventana (assumption del spec).

**Resultado**: PASS — el plan cumple todos los gates (sin violaciones; ver Complexity Tracking vacío).

## Project Structure

### Documentation (this feature)

```text
specs/002-codex-style-ui/
├── plan.md              # This file
├── spec.md              # Feature specification (input)
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   ├── gui-contract.md
│   └── engine-contract.md
├── checklists/          # Existing (requirements.md = PASS)
└── tasks.md             # Phase 2 output (NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
internal/
├── gui/                 # Superficies rediseñadas (estructura existente ampliada)
│   ├── app.go           # Ventana única, wiring de superficies
│   ├── chatview.go      # Transcript-first (existing, ajustado)
│   ├── composer.go      # Control room: steer (Enter) / queue (Tab), siempre editable
│   ├── approvalcard.go  # (nuevo) Tarjetas inline + cola de decisiones (reemplaza modal approvals.go)
│   ├── statusbar.go     # Franja de trabajo (Working/spinner/tiempo/paso) + status line configurable
│   ├── palette.go       # (nuevo) Paleta Ctrl+K: acciones + @ picker + ! passthrough
│   ├── sessionpicker.go # (nuevo) Picker de inicio: Nueva/Reanudar/Elegir/Fork
│   ├── planmode.go      # (nuevo) Modo Plan visible + tarjeta de plan Aprobar/Rechazar/Editar
│   ├── keymap.go        # Overlay ? cheat sheet (filtrable)
│   ├── keyboard.go      # Atajos existentes preservados + nuevos
│   ├── settings.go      # Grants por sesión listables/revocables
│   ├── sessions.go      # Sesiones (existing, refactor para fork)
│   ├── stream.go        # Pump de eventos (existing, ampliado)
│   ├── theme.go         # Identidad propia LetsGO (paleta/acento)
│   ├── message.go, markdown.go, toolpanel.go, agenttasks.go, usage.go, meta.go, gitview.go, mcpview.go, pluginsview.go  # existing
│   └── gui_test.go      # Tests headless de las nuevas superficies
├── engine/              # Contrato interno ampliado (sin romper público)
│   ├── commands.go      # + Steer, Queue, SetPlanMode, ApprovePlan, RejectPlan, EditPlan, ForkSession, GrantSession
│   ├── events.go        # + PlanProposed, PlanDecided, ModeChanged, SteerQueued, GrantsChanged
│   ├── permission_driver.go  # Grants por sesión (categoría → session_id)
│   ├── engine.go, engine_test.go  # existing
├── db/
│   ├── database.go      # + tabla session_grants, ForkSession()
│   └── session_grants.go (o métodos en database.go)
└── config/
    └── config.go        # + campos status line (modelo/rama/modo/contexto/rate/versión)

cmd/letsgo/              # existing entrypoint
go.mod, go.sum           # sin nuevas dependencias externas (Fyne ya incluido)
```

**Structure Decision**: Single project (Golang, estructura existente). El rediseño es principalmente reorganización dentro de `internal/gui` + extensiones internas de `internal/engine`/`internal/db`/`internal/config`. No se crean módulos ni repos nuevos; el motor de búsqueda difusa del picker `@` se implementa sin dependencias externas (walk + coincidencia por subcadena/prefix) para no inflar el árbol de dependencias.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

Sin violaciones — estructura existente soporta el alcance completo sin proyectos adicionales.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| — | — | — |

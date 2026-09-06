# Contract: Engine Event/Command Protocol v2 (`internal/engine`)

**Feature**: 002-codex-style-ui | **Date**: 2026-08-06 | **Extends**: contracts/engine-contract.md de 001 (sin romper el contrato público)

`internal/engine` sigue siendo el único dueño del loop. Este documento **extiende** el contrato de 001 con los comandos y eventos internos nuevos para steer/queue/plan/grants/fork. Todos los comandos y eventos de 001 permanecen exactamente como estaban (backward compatible — los consumidores TUI/GUI v1 no se rompen).

## Commands (presentation → engine) — adiciones

| Command | Payload | Semantics |
|---------|---------|-----------|
| `Steer` | `{text, sessionID}` | Inyecta instrucción al turno EN CURSO; mensaje persistido con `kind='user:steer'`; el stream sigue (FR-009). No válido en reposo (debe usarse `SendMessage`) |
| `Queue` | `{text, sessionID}` | Persiste con `kind='user:queued'`; se ejecuta como turno al llegar `Idle` (FR-010) |
| `SetPlanMode` | `{mode}` | `plan` / `execute` / `auto`. En `plan`, el engine NO ejecuta herramientas de escritura hasta `ApprovePlan` (FR-012) |
| `ApprovePlan` | `{planID}` | Habilita la ejecución del plan aprobado |
| `RejectPlan` | `{planID, reason}` | Plan denegado; retroalimentación al modelo |
| `EditPlan` | `{planID, instructions}` | Actualiza instrucciones del plan; nueva iteración de propuesta |
| `GrantSession` | `{category, sessionID, allow}` | Persiste/revoca grant en `session_grants` (FR-016/017) |
| `ForkSession` | `{sourceID, messageID}` | Nueva sesión: metadata copiada + mensajes hasta `messageID`; original intacta; SIN grants/memory (FR-015) |

## Events (engine → presentation) — adiciones

| Event | Payload | When |
|-------|---------|------|
| `PlanProposed` | `{planID, steps, files, criteria}` | Agente presenta plan en modo Plan (FR-013) |
| `PlanDecided` | `{planID, decision}` | `approved` / `rejected` / `edited` |
| `ModeChanged` | `{mode}` | Cambio de modo Plan/Execute/Auto (FR-012) |
| `SteerQueued` | `{text}` | Confirmación de que el steer se inyectó en el turno actual (FR-009) |
| `QueueRuns` | `{text}` | Un mensaje encolado comenzó su turno (FR-010); la UI decrementa el indicador "en cola" |
| `GrantsChanged` | `{sessionID, category, allow}` | Grant creado o revocado; la UI refresca settings y cards (FR-016/017) |

## Permission driver (presentation implementa) — reglas ampliadas

```go
type PermissionDriver interface {
    Prompt(toolName string, input any) (allow bool, err error)
    AutoApprove(category string) bool
    SessionGranted(sessionID, category string) bool // NUEVO
}
```

- **Precedencia (non-negotiable, assumption de seguridad)**: negación persistente (deny) > `SessionGranted` > `AutoApprove` global > `Prompt`.
- `Prompt` en GUI v2: el bloqueo sigue siendo síncrono desde la perspectiva del engine (la goroutine de aprobación se bloquea), pero la UI ya NO abre modal: la tarjeta inline aparece sobre el composer y el transcript sigue renderizando (FR-001, SC-007). El `Idle` del transcript no depende de la decisión de la tarjeta — solo la tarjeta en sí.
- Cierre de ventana con pendientes: la GUI emite `RejectTool` por cada `ToolRequested` pendiente antes de `Stop` (FR-022) — el timeout existente del engine actúa como red de seguridad adicional.
- Timeout existente: prompt sin respuesta dentro del timeout configurado → `ToolTimedOut` (tratado como rechazo) — se mantiene.

## Sequencing guarantees (001 preservadas + adiciones)

1–4. Igual que 001 (orden estricto por turno, `Idle` tras cada turno, `Cancel` siempre aceptado, cero modificaciones silenciosas).
5. `Steer` solo es válido entre `StreamStart` y `StreamDone`/`StreamCancelled` del turno activo; si llega en reposo, el engine lo ignora y emite `ErrorEvent{recoverable:true}`.
6. `Queue` válido en cualquier momento; encolado FIFO; al llegar `Idle` el primero de la cola se ejecuta como turno normal.
7. En modo `plan`, cualquier `ToolRequested` de categoría de escritura se retiene en el engine (no llega al driver) hasta `ApprovePlan` o `RejectPlan`.
8. `GrantSession{allow:false}` sobre categoría negada persistentemente no tiene efecto visible (deny global gana siempre).
9. `ForkSession` emite `SwitchSession` interno a la nueva sesión; la UI muestra el picker/transcript de la copia.

# Data Model: Codex-Style UI (Phase 1)

**Feature**: 002-codex-style-ui | **Date**: 2026-08-06

Fuente: entidades de `spec.md` (§ Key Entities). Persistencia existente: SQLite (GORM) en `internal/db` + config viper en `internal/config`. Regla: **sin nuevas dependencias**; solo una tabla nueva + columnas/config nuevas.

## 1. Entidades de la spec → ubicación

| Entidad | Naturaleza | Dónde vive |
|---|---|---|
| `ApprovalDecision` | En memoria (cola UI) + efímera | `internal/gui/approvalcard.go` (struct GUI), alimentada por `engine.ToolRequested` |
| `SessionGrant` | Persistente (DB) | Tabla nueva `session_grants` (`internal/db`) |
| `WorkStatus` | En memoria, derivada de eventos | `internal/gui/statusbar.go` (struct GUI, sin persistencia) |
| `PlanCard` | En memoria (ciclo de vida de un turno) | `internal/gui/planmode.go` + estado del engine |
| `SessionPickerEntry` | Proyección de `sessions` | `internal/db.ListSessions` (campos ya existen) |
| `StatusField` | Persistente (config viper) | `internal/config` |
| `CommandPaletteEntry` | En memoria (catálogo estático + picker) | `internal/gui/palette.go` |

Solo `SessionGrant` y `StatusField` requieren persistencia nueva. `Steer`/`Queue` requieren un flag en `messages`.

## 2. Tabla nueva: `session_grants`

```text
session_grants
├── id            INTEGER PRIMARY KEY AUTOINCREMENT
├── session_id    TEXT NOT NULL          -- FK lógica a sessions.id
├── category      TEXT NOT NULL          -- categoría de herramienta (p.ej. "write", "shell")
├── created_at    DATETIME NOT NULL
└── UNIQUE(session_id, category)         -- upsert = grant; delete = revoke
```

- Indices: `idx_session_grants_session` en `session_id` (lookup por sesión), `idx_session_grants_cat` en `category`.
- API en `internal/db`:
  - `GrantCategory(sessionID, category string) error` — upsert.
  - `RevokeCategory(sessionID, category string) error` — delete.
  - `ListGrants(sessionID string) ([]SessionGrant, error)` — para settings y lookup del driver.
  - `IsGranted(sessionID, category string) (bool, error)` — lookup del PermissionDriver.
- Precaución: nunca borrar grants al eliminar la sesión original por fork — el fork es una sesión nueva sin grants heredados (los grants no trascienden: FR-017).

## 3. Cambios en tablas existentes

### `messages` — columna `kind`

```text
kind  TEXT NOT NULL DEFAULT 'user'
```

Valores: `user` (existente), `user:steer` (FR-009), `user:queued` (FR-010), `user:plan_instructions` (FR-013). El transcript los muestra con marca sutil ("steer", "en cola", "instrucciones"). Backward compatible: migración de schema existente (GORM `AutoMigrate` añade columna con default).

Sin cambios en `sessions`: el fork (FR-015) reusa la fila de sesión con `name` copiado + prefijo "Fork de ..." y copia `messages` hasta un `message_id` dado (transacción: nueva sesión + INSERT SELECT de mensajes).

### `session_memory` — sin cambios

### Config viper (`internal/config`) — status line (FR-018)

```text
statusline.show       bool     (default true)
statusline.fields     []string (default ["model","mode","context","rate","version"])
statusline.order      []string (opcional, orden de campos)
```

Mecanismo de guardado existente (`AppConfig.Save()` → `SaveConfig` con chmod 0600).

## 4. Modelos en memoria (GUI)

### ApprovalDecision (cola de aprobaciones, FR-001..003/021)

```go
type ApprovalDecision struct {
    ToolCallID string
    Name       string
    Input      map[string]interface{}
    Diff       string          // renderer markdown existente
    Decision   string          // "", "approved", "rejected"
    RejectReason string        // motivo estructurado si rechazo
    Position   int             // 1-based en cola
}
```

### WorkStatus (franja de estado, FR-005/006/020)

```go
type WorkStatus struct {
    Active        bool    // turno en curso
    StreamingText bool    // StreamDelta activo → ocultar franja
    Phase         string  // "working" | "tool" | "agent"
    CurrentStep   string  // herramienta/agente actual
    Elapsed       time.Duration
    ContextStale  bool    // >15 min sin actualizar (FR-020)
    RateStale     bool
}
```

### PlanCard (FR-012/013)

```go
type PlanCard struct {
    PlanID   string
    Steps    []PlanStep  // {Action, File, Criteria}
    Files    []string
    Status   string      // pending | editing | approved | rejected
}
```

## 5. Reglas de integridad (seguridad, non-negotiable)

1. **Precedencia de permisos** (assumption spec): `negación persistente (deny global) > grant de sesión > auto-approve global > preguntar`.
2. **Grants nunca globales**: `GrantCategory` siempre lleva `session_id`; no existe columna opcional NULL-global.
3. **`!` passthrough**: ejecuta solo lo que el PermissionDriver permite; los grants no aplican a categorías negadas persistentemente.
4. **Fork aislado**: mensajes copiados, grants NO copiados, memory NO copiada.

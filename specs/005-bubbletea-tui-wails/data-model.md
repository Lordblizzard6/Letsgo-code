# Data Model: Doble interfaz — TUI Bubble Tea + GUI Wails

**Branch**: `005-bubbletea-tui-wails` | **Fecha**: 2026-08-11 | **Spec**: [spec.md](spec.md)

El feature no crea tablas nuevas: **formaliza y reutiliza** el modelo persistente
existente (`internal/db`) y la configuración (`internal/config`) como la única
fuente de verdad compartida por TUI y GUI. Este documento describe las entidades,
sus relaciones y las transiciones de estado relevantes para ambas interfaces.

## Entidades

### Sesión (Conversación)

Representa una conversación persistida, creada o retomada indistintamente desde la
TUI o la GUI.

| Campo | Tipo | Notas |
|-------|------|-------|
| id | string (PK) | Identificador único generado al crear la sesión |
| name | string | Nombre legible de la sesión |
| created_at / updated_at | datetime | Fechas de creación y última actividad |
| project_path | string | Ruta del proyecto asociado (referencias `@archivo`) |
| is_active | boolean | Marca de sesión en curso |

**Reglas**:
- Cualquier interfaz puede crear, listar y retomar sesiones (FR-006, FR-009, FR-012).
- Una sesión creada en la TUI es retomable en la GUI y viceversa (SC-002).

### Mensaje

Pertenece a una sesión; el contenido del chat.

| Campo | Tipo | Notas |
|-------|------|-------|
| id | integer (PK, autoincrement) | |
| session_id | FK → sessions.id | |
| role | enum | `user`, `assistant`, `tool` |
| content | text | Markdown crudo (fuente única: TUI renderiza con glamour; GUI con streaming-markdown) |
| timestamp | datetime | Orden cronológico (índice por session+timestamp desc) |
| tokens_input / tokens_output | integer | Métrica por mensaje |
| cost_usd | real | Coste por mensaje (acumulado para el panel Uso) |

**Reglas**:
- Un stream interrumpido (Ctrl+C / cancelar) **no** persiste el texto parcial como
  respuesta final (FR-007; edge: "interrupción").
- Los mensajes `tool` exponen actividad/resultado de herramientas a ambas
  interfaces (FR-008).

### ContextFile

Archivos referenciados con `@` dentro de una sesión.

| Campo | Tipo | Notas |
|-------|------|-------|
| id | integer (PK) | |
| session_id | FK | |
| file_path | string | Ruta relativa al project_path |
| content | text | Contenido leído para contexto |
| added_at | datetime | |

### SessionMemory

Memoria persistente clave-valor del sistema (recordatorios del modelo).

| Campo | Tipo | Notas |
|-------|------|-------|
| id | integer (PK) | |
| key | string (UNIQUE) | |
| value | string | |
| created_at / updated_at | datetime | |

### SessionGrant

Aprobaciones de categoría concedidas por sesión (modelo de permisos heredado).

| Campo | Tipo | Notas |
|-------|------|-------|
| id | integer (PK) | |
| session_id | FK | |
| category | string | |
| UNIQUE (session_id, category) | | |

### CompactHistory

Registro de compactaciones (para el comando `/compact` y depuración).

| Campo | Tipo | Notas |
|-------|------|-------|
| id | integer (PK) | |
| session_id | FK | |
| original_messages / compacted_messages | integer | |
| summary | text | Resumen generado |
| compacted_at | datetime | |

### Configuración de usuario

Persistida por `internal/config` (provider/modelo/API key/preferencias), consumida
por ambas interfaces (FR-010, FR-016).

| Aspecto | Notas |
|---------|-------|
| proveedores + API keys | cifradas/guardadas según convención actual |
| modelo activo | el que usa el motor en ambas interfaces |
| preferencias | auto-aprobación, tema (dark/light), rail colapsado, etc. |

**Reglas**: cambiar el modelo/preferencias desde cualquier interfaz refresca el
motor sin reiniciar (FR-016); primer arranque sin key → onboarding (FR-011).

## Relaciones

```text
Configuración (internal/config) ── usada por ──► Motor ──► TUI
        ▲                                        │        GUI
        └──────── persistida / compartida ───────┴────────┘

Sesión 1─∞ Mensaje, ContextFile, CompactHistory, SessionGrant
SessionMemory: global (sin FK a sesión)
```

## Transiciones de estado relevantes

### Stream de conversación (visible en ambas interfaces)

```text
idle ──send──► streaming ──delta──► streaming ──complete──► idle
                 │
                 ├──cancel (Ctrl+C / botón Detener / cierre)──► idle (sin mensaje parcial persistido)
                 └──error (red/API)───────────────────────────► idle + error mostrado, historial intacto
```

### Sesión

```text
creada ─► activa (is_active) ─► retomada (updated_at, is_active)
                              └► cerrada (is_active = 0)
```

### Configuración

```text
sin key ──onboarding──► configurada ──cambio modelo/preferencias──► refresco del motor (sin reinicio)
```

## Validaciones derivadas de los requisitos

| Regla | Origen |
|-------|--------|
| El texto parcial de un stream cancelado no se persiste como mensaje final | FR-007, edge interrupción |
| Ambas interfaces leen/escriben la misma base de datos y configuración | FR-012 |
| Cambios de configuración se aplican al motor sin reiniciar la interfaz | FR-016 |
| La GUI Wails debe exponer el mismo historial que la TUI (sin export/import) | SC-002 |
| Transcripts ≥500 mensajes sin degradación perceptible | SC-009 |

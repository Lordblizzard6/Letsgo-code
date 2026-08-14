# Contrato de Frontend (Core Contract)

**Branch**: `005-bubbletea-tui-wails` | **Fecha**: 2026-08-11 | **Spec**: [spec.md](../spec.md)

**Propósito** (FR-001, FR-002): toda interfaz del producto (TUI, GUI Wails) consume
el mismo core (motor, base de datos, configuración) a través de esta superficie
estable. Ninguna interfaz duplica lógica de conversación.

## Alcance

- Proyecta la superficie del core **tal como existe hoy** en `internal/engine`,
  `internal/db` y `internal/config` y que ya consumen la TUI (`internal/tui`) y la
  GUI Fyne (`internal/gui`). Este contrato **no** cambia el comportamiento del
  core; lo documenta y lo verifica.
- Los adaptadores específicos (loop Bubble Tea, servicios bound de Wails) quedan
  fuera del contrato y se describen en [gui-contract.md](gui-contract.md).

## 1. Comandos de conversación (core → motor)

| Comando | Entrada | Salida/Eventos |
|---------|---------|----------------|
| `send(message)` | texto del usuario (markdown crudo) | `stream:start`, `stream:delta…`, `stream:end` o `stream:error` |
| `cancel()` | — | detiene el stream; `stream:cancelled` (no persiste parcial) |
| `setModel(provider, model)` | ids | `config:changed` + refresh del motor |
| `setAutoApprove(bool)` | bool | `config:changed` |
| `resetMemory()`, `remember(key, value)` | | persistencia en session_memory |
| `slash:{help,clear,compact,cost,tokens,model,quit,…}` | args opcionales | respuesta textual o acción + eventos de estado |

## 2. Eventos de streaming/estado (motor → interfaz)

| Evento | Payload | Consumidores |
|--------|---------|--------------|
| `stream:start` | sessionId, model | TUI (spinner/estado), GUI (indicador + botón Detener) |
| `stream:delta` | fragmento de markdown (batch ~50 ms) | TUI (viewport + glamour), GUI (streaming-markdown + DOMPurify) |
| `stream:end` | mensaje persistido (id, tokens, cost) | ambas: actualizan historial + estado "listo" |
| `stream:cancelled` | — | ambas: estado "listo" sin mensaje final parcial |
| `stream:error` | mensaje de error orientativo | ambas: historial intacto + acción sugerida (key/conexión/modelo) |
| `tool:start` / `tool:end` | nombre herramienta, resultado éxito/error | GUI: actividad en el chat; TUI: log de herramientas |
| `usage:update` | coste hoy, requests, tokens | paneles Uso de ambas |
| `session:list` / `session:loaded` | sesiones / sesión retomada | ambas: panel de sesiones |
| `config:changed` | config aplicada | ambas: refresco de modelo/tema sin reiniciar |

## 3. Funciones de sesión y configuración

| Función | Firma lógica | Notas |
|---------|--------------|-------|
| `SessionCreate(name?, projectPath?)` | → sessionId | |
| `SessionList()` | → []sesión (id, name, updated_at) | |
| `SessionOpen(id)` | → []mensaje (historial completo) | |
| `GetMessages(sessionId, limit?, offset?)` | → []mensaje paginado | para transcripts largos (SC-009) |
| `GetConfig()` / `SaveConfig(partial)` | Config | proveedores, modelo, preferencias |
| `GetUsageToday()` | → {cost, requests, tokens} | alimenta KPI Uso |

## 4. Tests de conformidad (obligatorios, headless)

Toda interfaz del proyecto debe pasar la suite de conformidad **sin pantalla**
(fuente única: # de fallos rompe la integración):

- C-001: `send` emite `stream:start` → deltas → `stream:end` con mensaje persistido.
- C-002: `cancel` a mitad de stream no persiste texto parcial y vuelve a estado usable.
- C-003: la misma sesión creada por una interfaz es retomable por la otra con historial completo.
- C-004: `setModel`/`SaveConfig` refresca el motor sin reinicializar la interfaz.
- C-005: un `stream:error` deja el historial intacto.
- C-006: la GUI solo usa la superficie de este contrato (no accede al core por fuera).

## 5. Suposiciones

- Los mensajes se persisten en **markdown crudo**; el rendering es responsabilidad
  de cada interfaz (glamour en TUI, streaming-markdown + DOMPurify en GUI).
- La suite actual de tests headless de `internal/gui` (fyne test) se porta 1:1 a
  tests del motor/servicios al migrar (D6 de [research.md](../research.md)).
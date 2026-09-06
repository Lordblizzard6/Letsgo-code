# Contract: Antigravity Flow Parity

Este contrato especifica las funciones RPC, eventos de Wails y atajos de teclado requeridos para la paridad con el flujo de Antigravity y Opencode.

---

## 1. RPC Backend Surfaces (`cmd/wails/services`)

### `SessionsService`
```go
// Rollback trunca el historial de la sesión a partir de targetMessageID (inclusive),
// resiembra el motor y retorna el texto del prompt revertido.
func (s *SessionsService) Rollback(sessionID string, targetMessageID int64) (string, error)
```

- **Llamada frontend**: `SessionsService.Rollback(sessionID, messageID): Promise<string>`
- **Efectos colaterales**:
  - Detiene cualquier generación en progreso (`Cancel`).
  - Purga de la tabla `messages` en SQLite.
  - Actualiza el motor de IA (`s.hub.Engine.SeedMessages`).
  - Emite `session:loaded` con el historial truncado.

### `GitService`
```go
// DiffSummary retorna el estado estructurado del repositorio: archivos modificados y diff.
func (s *GitService) DiffSummary() (GitDiffSummary, error)

// DiffFile retorna el diff específico de un archivo dado.
func (s *GitService) DiffFile(filePath string) (string, error)
```

---

## 2. Eventos Wails (`@wailsio/runtime`)

| Evento | Origen | Destino | Payload | Descripción |
| :--- | :--- | :--- | :--- | :--- |
| `session:rollback` | Backend | Frontend | `{ session_id: string, target_id: number }` | Notifica que la sesión ha sido rebobinada. |
| `git:changed` | Backend | Frontend | `{ branch: string, clean: boolean }` | Notifica cambios en el árbol de git tras ejecuciones de herramientas. |

---

## 3. Atajos de Teclado y Comandos TUI

| Comando / Atajo | Entorno | Acción |
| :--- | :--- | :--- |
| `/rollback` | TUI | Rebobina el último turno de conversación y recarga el prompt en el `textarea`. |
| `Ctrl+Z` / `Alt+U` | TUI | Atajo rápido para `/rollback`. |
| `/diff` | TUI | Abre el visor interactivo de Git Diff en el viewport. |
| `Ctrl+D` / `Ctrl+Shift+G` | GUI | Abre o cierra el panel lateral de Git Diff. |
| `Ctrl+Enter` / `Enter` | GUI | Envía el prompt redactado en el compositor. |
| `Shift+Enter` | GUI | Inserta un salto de línea sin enviar en el compositor. |

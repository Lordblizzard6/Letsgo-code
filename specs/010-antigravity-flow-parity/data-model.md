# Data Model: Paridad de Flujo Antigravity y Opencode

Este documento define las entidades de datos, extensiones al esquema de base de datos y modelos de interfaz para el soporte de rollback, panel de git diff y controles integrados del compositor.

---

## 1. Extensiones al Esquema SQLite (`internal/db`)

### Tabla `messages` (Existente)
```sql
CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    tool_calls TEXT,
    tool_call_id TEXT,
    tokens INTEGER DEFAULT 0,
    cost REAL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES sessions(id)
);
```

### Operación de Rollback
- **Consulta**:
  ```sql
  -- Obtener el mensaje seleccionado
  SELECT role, content FROM messages WHERE session_id = ? AND id = ?;

  -- Eliminar todos los mensajes a partir de ese ID inclusive
  DELETE FROM messages WHERE session_id = ? AND id >= ?;
  ```
- **Invariante**: Al ejecutar rollback sobre un mensaje de usuario de ID `N`, todos los mensajes con `id >= N` se purgan, dejando la sesión en el estado exacto anterior a la emisión de esa solicitud.

---

## 2. Entidades de Backend (Go)

### `RollbackResult`
Representa el resultado de una operación de reversión de mensajes.
```go
type RollbackResult struct {
    SessionID      string `json:"session_id"`
    TargetID       int64  `json:"target_id"`
    MessagesPruned int    `json:"messages_pruned"`
    RestoredText   string `json:"restored_text"`
}
```

### `GitDiffSummary`
Estructura representativa del estado de control de versiones para el panel lateral.
```go
type GitDiffFile struct {
    Path      string `json:"path"`
    Status    string `json:"status"` // "M", "A", "D", "R", "??"
    Staged    bool   `json:"staged"`
    Additions int    `json:"additions"`
    Deletions int    `json:"deletions"`
}

type GitDiffSummary struct {
    IsRepo    bool          `json:"is_repo"`
    Branch    string        `json:"branch"`
    Clean     bool          `json:"clean"`
    Files     []GitDiffFile `json:"files"`
    RawDiff   string        `json:"raw_diff"`
}
```

---

## 3. Entidades de Frontend (TypeScript)

### `ComposerState`
Representa el estado reactivo dentro de la barra de herramientas del compositor:
```typescript
export interface ComposerState {
    model: string;
    provider: string;
    mode: "code" | "plan" | "ask";
    isStreaming: boolean;
    hasAttachments: boolean;
    characterCount: number;
}
```

### `RollbackAction`
Payload emitido al rebobinar un mensaje desde la GUI:
```typescript
export interface RollbackAction {
    sessionId: string;
    messageId: number | string;
    content: string;
}
```

### `GitPanelState`
Estado del panel lateral de Git Diff:
```typescript
export interface GitPanelState {
    open: boolean;
    selectedFile: string | null;
    summary: GitDiffSummary | null;
    loading: boolean;
}
```

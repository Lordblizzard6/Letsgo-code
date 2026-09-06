# Data Model: CLI Path Argument and Project Scope Synchronization

## Entities

### 1. ActiveProjectState (In-Memory / Hub)
Tracks the active project working directory across services and the presentation layer.

| Field | Type | Description |
|-------|------|-------------|
| `Path` | `string` | Absolute path to the active project folder, or `""` if unscoped |
| `Name` | `string` | Display name (basename of the path), or `"Sin proyecto"` |
| `Exists` | `bool` | Whether the path currently exists as a directory on disk |

### 2. Session (SQLite / `internal/db`)
Existing schema in `sessions` table (retained, no schema migration required):

| Column | Type | Description |
|--------|------|-------------|
| `id` | `TEXT PRIMARY KEY` | Unique session identifier |
| `name` | `TEXT` | Session display name |
| `project_path` | `TEXT` | Associated project directory path, or empty string |
| `created_at` | `DATETIME` | Creation timestamp |
| `updated_at` | `DATETIME` | Last update timestamp |
| `is_active` | `INTEGER` | Active flag |

### 3. Events

| Event Name | Payload | Emitted By | Handled By |
|------------|---------|------------|------------|
| `project:changed` | `{"project_path": string, "name": string}` | `Hub` / `SessionsService` | Frontend (`App.tsx`), `GitService` |
| `session:loaded` | `{"session_id": string, "messages": Message[], "project_path": string}` | `SessionsService.Open` | Frontend (`App.tsx`) |

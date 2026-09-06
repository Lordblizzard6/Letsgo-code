# Contract: Servicios de Sesión, Proyectos y Temas

## 1. SessionsService (Wails RPC Binding)

### `Create(name string, projectPath string) (db.Session, error)`
- **Descripción**: Crea una nueva sesión directamente en la base de datos SQLite y asocia la ruta del proyecto especificada.
- **Entrada**:
  - `name`: Nombre inicial de la sesión (ejemplo: `[2026-09-03] mi-proyecto`).
  - `projectPath`: Ruta absoluta de la carpeta de proyecto.
- **Salida**: Objeto `db.Session` con `id`, `name`, `project_path`, `created_at` y `updated_at`.
- **Efectos secundarios**: Emite el evento Wails `session:list` con la lista actualizada de sesiones.

### `List() ([]db.Session, error)`
- **Descripción**: Devuelve todas las sesiones almacenadas en SQLite ordenadas descendentemente por fecha de última actualización.
- **Salida**: Array de sesiones con metadatos y conteo de mensajes.

### `SelectProjectFolder() (string, error)`
- **Descripción**: Despliega el diálogo nativo de selección de directorios del sistema operativo.
- **Salida**: Ruta absoluta seleccionada por el usuario o cadena vacía si fue cancelado.
- **Efectos secundarios**: Actualiza el directorio de trabajo del proceso (`os.Chdir`) si no está vacío.

### `Rename(id string, newName string) error`
- **Descripción**: Modifica el título de una conversación existente.
- **Efectos secundarios**: Persiste el nuevo nombre en SQLite y emite `session:list`.

### `Delete(id string) error`
- **Descripción**: Elimina una conversación y todos sus mensajes asociados de la base de datos SQLite.
- **Efectos secundarios**: Emite `session:list`.

---

## 2. ThemeService (Wails RPC Binding)

### `Get() string`
- **Descripción**: Devuelve la variante de tema activa (`"goulm"`, `"dark"` o `"light"`).

### `Set(variant string) (string, error)`
- **Descripción**: Valida y persiste la variante de tema en `config.yaml`.
- **Entrada**: `variant` (`"goulm" | "dark" | "light"`).
- **Salida**: La variante guardada o error si no es válida.
- **Efectos secundarios**: Emite el evento Wails `theme:changed` con el payload `{"theme": variant}`.

---

## 3. Eventos Wails Emitidos

| Evento | Payload | Descripción |
|---|---|---|
| `session:list` | `db.Session[]` | Notifica cambios en la lista de conversaciones (creación, renombramiento, eliminación). |
| `session:loaded` | `{"session_id": string, "messages": []}` | Emite el historial de mensajes cuando una sesión es abierta. |
| `theme:changed` | `{"theme": string}` | Notifica a todas las ventanas el cambio de tema visual. |

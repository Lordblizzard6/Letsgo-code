# Data Model: Proyectos, Sesiones y Temas

## Entidades y Estructuras de Datos

### 1. Proyecto (`Project`)
Representa una carpeta de proyecto registrada por el usuario en el entorno local.

| Campo | Tipo | Requerido | Descripción |
|---|---|---|---|
| `name` | `string` | Sí | Nombre representativo (`basename` de la ruta). |
| `path` | `string` | Sí | Ruta absoluta única en el sistema de archivos del usuario. |
| `last_opened` | `string (ISO-8601)` | Sí | Marca temporal del último acceso o apertura del proyecto. |
| `is_pinned` | `boolean` | No | Indicador opcional de proyecto anclado en la vista. |

### 2. Sesión / Conversación (`Session`)
Hilo de diálogo e interacciones de herramientas persistido en SQLite (`internal/db`).

| Campo | Tipo | Requerido | Descripción |
|---|---|---|---|
| `id` | `string (UUID)` | Sí | Identificador único de la sesión. |
| `title` | `string` | Sí | Título descriptivo de la conversación (ej: `[2026-09-03] Letsgo-code`). |
| `project_path`| `string` | No | Ruta del proyecto al que pertenece (vacío para sesiones "General"). |
| `created_at` | `string (ISO-8601)` | Sí | Fecha y hora de creación de la sesión. |
| `updated_at` | `string (ISO-8601)` | Sí | Fecha y hora de la última interacción o mensaje. |
| `message_count`| `integer` | Sí | Número total de mensajes y ejecuciones en la sesión. |

### 3. Pestaña de Navegación en GUI (`ChatTab`)
Estado de presentación en el frontend de Wails para el espacio de pestañas activas.

| Campo | Tipo | Requerido | Descripción |
|---|---|---|---|
| `id` | `string` | Sí | Identificador interno de la pestaña (`tab-1`, `tab-2`, etc.). |
| `title` | `string` | Sí | Título visible en la pestaña. |
| `sessionId` | `string` | Sí | Clave foránea al `id` de la sesión en SQLite. **Nunca null** en pestañas operativas. |
| `messages` | `ChatMessage[]` | Sí | Lista de mensajes y llamadas de herramientas parseadas. |
| `project_dir` | `string` | No | Directorio de trabajo en el que opera la pestaña. |
| `saved` | `boolean` | Sí | Indicador de sincronización con la base de datos. |

### 4. Configuración de Apariencia (`ThemeConfig`)
Configuración de diseño persistida en `internal/config` (`config.yaml`).

| Campo | Tipo | Valores Válidos | Por Defecto | Descripción |
|---|---|---|---|---|
| `theme` | `string` | `"goulm"`, `"dark"`, `"light"` | `"goulm"` | Variante visual activa en la interfaz y en el motor. |

---

## Relaciones entre Entidades

```text
[Project] (1) ─────────── (0..N) [Session]
                                    │ (1)
                                    │
                                    │ (1)
                              [ChatTab (GUI)]
```

- Un `Project` puede tener 0 o más `Session`s asociadas.
- Si un `Project` tiene 0 `Session`s, el contenedor del proyecto sigue existiendo y mostrándose en el árbol lateral.
- Cada `Session` activa abierta en la interfaz corresponde a un `ChatTab` con `sessionId` válido y persistido.

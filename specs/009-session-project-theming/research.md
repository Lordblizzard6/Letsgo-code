# Research: Revisión de Proyectos, Sesiones y Temas

## 1. Alcance de Proyectos sin Conversaciones

### Decisión
Unificar la lista de proyectos en la barra lateral (`sidebar`) para que provenga de la lista de proyectos registrados / recientes (`RecentProjects` / rutas de proyecto en base de datos). El árbol de proyectos debe iterar sobre todos los proyectos conocidos:
- Si un proyecto tiene sesiones, muestra sus sesiones anidadas.
- Si un proyecto tiene 0 sesiones, se muestra igualmente con su cabecera, nombre, ruta y estado vacío ("Sin conversaciones"), acompañado de su botón de acción directa para crear la primera conversación.

### Razón
El usuario añade carpetas de proyecto para establecer un espacio de trabajo. Si un proyecto recién añadido desaparece porque aún no tiene sesiones, la interfaz resulta desconcertante. Mantener el contenedor visible refuerza el concepto de proyecto como ámbito de trabajo primario.

### Alternativas consideradas
- *Mostrar solo proyectos con sesiones activas*: Descartada por petición explícita del usuario ("el apartado donde estan los proyectos deben figurar aunque no tengan una conversasion añadida").
- *Separar proyectos recientes y árbol de sesiones en dos vistas distintas*: Genera fragmentación y requiere cambiar de menú; la integración directa en el árbol es más limpia.

---

## 2. Creación Contextual de Sesiones desde el Ítem de Proyecto

### Decisión
Incorporar un botón `+` (Nueva conversación) visible o accesible en la cabecera de cada carpeta de proyecto en la barra lateral.
Al hacer clic:
1. Se invoca `SessionsService.Create("[YYYY-MM-DD] <NombreProyecto>", projectDir)`.
2. Se obtiene de inmediato el ID persistido en SQLite.
3. Se actualiza el directorio de trabajo del sistema (`os.Chdir(projectDir)`).
4. Se abre una pestaña de chat vinculada a ese `sessionId` y `project_dir`.
5. Se refresca la lista de sesiones emitiendo `session:list`.

### Razón
Evita que las sesiones se creen en el vacío o en la raíz del ejecutable. Cada sesión nace con un proyecto y una carpeta de trabajo claros desde su creación.

---

## 3. Limpieza de Controles de Nueva Conversación en el Espacio de Chat

### Decisión
Eliminar cualquier botón redundante de "Nueva conversación" que se encuentre dentro del panel de chat, el área de bienvenida central o la barra de herramientas del compositor.
La creación de conversaciones queda centralizada en:
1. El botón `+` de cada proyecto en el árbol lateral.
2. El botón general `+` en la barra de pestañas (que hereda el proyecto actualmente activo o solicita uno).

### Razón
Tener botones flotantes o en el cuerpo del chat invitando a "Nueva conversación" sin contexto de proyecto provoca que se creen sesiones en el directorio raíz o sin asociar al proyecto deseado.

---

## 4. Confiabilidad en el Registro Inmediato de Sesiones

### Decisión
Eliminar la creación de "pestañas sin sesión" en la interfaz. Cuando el usuario solicita un nuevo chat:
- Se llama de forma sincrónica a `SessionsService.Create(...)` antes de renderizar la pestaña como lista para chatear.
- El objeto de pestaña (`ChatTab`) siempre cuenta con un `sessionId` válido desde el momento cero.
- En el backend Go (`internal/db`), la sesión se inserta de inmediato en SQLite con `created_at`, `updated_at` y `project_path`.
- Cualquier mensaje posterior guarda sus turnos asociados a esa clave foránea existente.

### Razón
Anteriormente, al pulsar un botón de nueva pestaña, solo se creaba un estado local en React (`sessionId: null`). Si el usuario no enviaba un mensaje o la creación tardía fallaba, la sesión no quedaba registrada en la base de datos. Al registrarla en SQLite desde el primer instante, no se pierde ninguna sesión.

---

## 5. Persistencia y Reactividad Total de Temas

### Decisión
1. **En Go (`internal/config` y `ThemeService`)**:
   - `AppConfig.ThemeVariant` soporta `"goulm"`, `"dark"` y `"light"`.
   - `viper.Set("theme", variant)` escribe en `config.yaml`.
   - `GetConfig()` expone `theme` a la GUI.
2. **En Frontend (`App.tsx` y `SettingsView.tsx`)**:
   - Lectura inicial síncrona de `localStorage.getItem("app_theme") || "goulm"` antes del render de React.
   - Inyección inmediata en `document.documentElement.dataset.theme`.
   - En `SettingsView.tsx`, todos los componentes (fondo, tarjetas, bordes, pestañas, textos, botones) usan variables CSS (`var(--bg-panel)`, `var(--bg-elevated)`, `var(--border)`, `var(--text)`, etc.) en lugar de colores fijos.
   - Al seleccionar un tema, se aplica `document.documentElement.dataset.theme`, se guarda en `localStorage` y se invoca `ThemeService.Set` para persistirlo en `config.yaml`.

# Feature Specification: Revisión de Proyectos, Sesiones y Temas

**Feature Branch**: `009-session-project-theming`

**Created**: 2026-09-03

**Status**: Draft

**Input**: User description: "crea un spec para revisionar la UI y el comportamiento de las conversasiones, nueva conversacion, y el scope de carpetas, y revisar el sistema que registra las nuevas sesiones, ya que inicie una y no se registro, ademas el apartado donde estan los proyectos deben figurar aunque no tengan una conversasion añadida(se supone que para eso se añaden) y que cada item de proyecto tenga su boton para una nueva conversación, y quitar el boton de nueva conversasion del espacio de chat, y revisar el comportamiento del aplicar los temas que no tiene persistencia aun, y la ventana de settings no cambia de color con los temas"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Proyectos Visibles sin Sesiones y Creación Contextual (Priority: P1)

Como usuario del asistente, deseo que los proyectos que he añadido aparezcan en la barra lateral aunque aún no contengan ninguna conversación, y que cada proyecto disponga de un botón dedicado para iniciar una nueva conversación en su contexto, para organizar mi trabajo por repositorios de forma clara e intuitiva.

**Why this priority**: Los usuarios configuran proyectos para asociar su trabajo a carpetas específicas. Ocultar los proyectos sin conversaciones impide utilizarlos como punto de partida y confunde al usuario sobre si el proyecto fue guardado.

**Independent Test**: Añadir un nuevo proyecto desde el explorador; verificar que aparece de inmediato en la lista de proyectos de la barra lateral con 0 conversaciones, y pulsar su botón "+" para abrir una conversación vinculada a esa carpeta.

**Acceptance Scenarios**:

1. **Given** que el usuario añade una carpeta de proyecto sin conversaciones previas, **When** la carpeta es seleccionada en el diálogo de proyecto, **Then** el proyecto aparece visible en la sección de proyectos de la barra lateral mostrando su nombre y ruta.
2. **Given** un proyecto listado en la barra lateral (con o sin conversaciones), **When** el usuario hace clic en el botón "+" de ese proyecto, **Then** se abre una nueva pestaña de conversación configurada en la ruta de ese proyecto.
3. **Given** que un proyecto contiene varias conversaciones, **When** el usuario colapsa o expande la carpeta, **Then** el botón de nueva conversación permanece accesible y operativo.

---

### User Story 2 - Registro Confiable e Inmediato de Nuevas Sesiones (Priority: P1)

Como usuario, deseo que cada nueva conversación que inicio quede registrada de forma inmediata y persistente en el sistema, para no perder el contexto de mis consultas ni sufrir sesiones invisibles tras reiniciar la aplicación.

**Why this priority**: La pérdida de sesiones es un problema crítico de confiabilidad que provoca frustración y pérdida de tiempo al usuario.

**Independent Test**: Crear una nueva conversación, escribir una consulta, verificar que aparece listada en la barra lateral con su título y fecha relativa, cerrar la aplicación y volver a abrirla comprobando que la conversación sigue presente y accesible.

**Acceptance Scenarios**:

1. **Given** que se solicita una nueva conversación para un proyecto, **When** la sesión es generada, **Then** se registra inmediatamente en la lista de sesiones activas del proyecto.
2. **Given** una sesión recién iniciada, **When** el usuario envía el primer mensaje, **Then** el historial y los metadatos de la sesión se guardan de forma permanente sin requerir acciones manuales de guardado.
3. **Given** una sesión registrada previamente, **When** se reinicia la aplicación, **Then** la sesión aparece en el árbol de proyectos con su marca de tiempo relativo de último acceso intacta.

---

### User Story 3 - Eliminación de Controles Redundantes en el Espacio de Chat (Priority: P2)

Como usuario, deseo una interfaz de chat despejada donde la creación de conversaciones se centralice en los proyectos de la barra lateral, eliminando botones confusos de nueva conversación dentro del área de mensajes.

**Why this priority**: Tener botones de nueva conversación dentro del espacio de chat genera ambigüedad sobre en qué proyecto o carpeta se creará la sesión, además de saturar la vista de lectura.

**Independent Test**: Navegar por la vista de chat y compositor; confirmar visualmente que no existen botones secundarios de "nueva conversación" en el panel de conversación y que la acción principal se realiza desde las pestañas o la barra lateral.

**Acceptance Scenarios**:

1. **Given** que el usuario está interactuando en el espacio de chat, **When** revisa el encabezado de chat y la barra de herramientas del compositor, **Then** no se visualizan botones redundantes de "nueva conversación" dentro de esa área.
2. **Given** la barra de pestañas superior, **When** el usuario desea abrir una pestaña rápida o crear una sesión, **Then** el flujo es coherente con el proyecto activo.

---

### User Story 4 - Persistencia Absoluta de Temas y Adaptación de la Ventana de Configuración (Priority: P2)

Como usuario, deseo que la selección de temas (Goulm, Oscuro, Claro) sea respetada en todas las vistas de la aplicación —incluyendo el modal de Configuración— y se mantenga activa de forma persistente tras cerrar y reiniciar el programa.

**Why this priority**: La inconsistencia visual en la ventana de ajustes y el restablecimiento del tema al reiniciar rompen la experiencia de usuario y causan fatiga visual.

**Independent Test**: Abrir Ajustes, cambiar a tema Claro o Goulm, comprobar que el modal de Ajustes adapta sus fondos, bordes y textos al nuevo tema al instante; reiniciar la aplicación y verificar que inicia con el tema seleccionado.

**Acceptance Scenarios**:

1. **Given** la ventana de Configuración abierta, **When** el usuario selecciona un tema distinto, **Then** el fondo, las tarjetas, los textos y los controles de la ventana de configuración cambian de color inmediatamente reflejando el tema elegido.
2. **Given** un tema seleccionado por el usuario, **When** la aplicación se cierra y se vuelve a iniciar, **Then** la interfaz carga directamente con dicho tema sin volver al predeterminado.
3. **Given** el comando `/theme <nombre>`, **When** se ejecuta en el chat, **Then** el tema cambia en toda la aplicación y se guarda para futuros inicios.

---

### Edge Cases

- **Proyecto sin acceso al sistema de archivos**: Si una carpeta de proyecto registrada ya no existe en el disco al abrir la aplicación, el proyecto se muestra en la lista con una advertencia visual sin bloquear la interfaz.
- **Eliminación de la última conversación de un proyecto**: Si el usuario borra todas las conversaciones de un proyecto, el contenedor del proyecto debe permanecer en la barra lateral con su botón para crear nuevas conversaciones.
- **Cierre repentino tras crear sesión**: Si la aplicación se cierra inmediatamente después de crear una sesión pero antes del primer mensaje del asistente, la sesión debe figurar registrada en la base de datos con su título inicial.
- **Alternancia rápida de temas**: Si el usuario hace clic sucesivamente entre distintos temas, no deben quedar estilos remanentes ni parpadeos de color en la ventana de configuración.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema DEBE listar todos los proyectos registrados en la barra lateral, independientemente de si poseen conversaciones asociadas.
- **FR-002**: Cada ítem de proyecto en la barra lateral DEBE contar con un botón interactivo para iniciar una nueva conversación vinculada a su carpeta.
- **FR-003**: El sistema DEBE persistir inmediatamente cada nueva conversación en el almacenamiento permanente para evitar sesiones no registradas.
- **FR-004**: El área de chat y compositor DEBE eliminar los botones redundantes de nueva conversación, concentrando dicha acción en la gestión de proyectos y pestañas.
- **FR-005**: La selección del tema visual DEBE persistir de manera fiable entre reinicios de la aplicación.
- **FR-006**: La ventana de Configuración DEBE aplicar dinámicamente las variables de diseño del tema activo a todos sus elementos (fondos, bordes, pestañas, campos de entrada y botones).
- **FR-007**: El sistema DEBE calcular y mostrar el tiempo transcurrido desde el último acceso a cada conversación en formato conciso `(Nm)`, `(Nh)` o `(Nd)`.

### Key Entities

- **Proyecto**: Representa un espacio de trabajo en disco. Atributos: nombre identificativo, ruta absoluta en el sistema de archivos, fecha de última apertura y colección de conversaciones asociadas.
- **Sesión / Conversación**: Representa un hilo de interacción con el asistente. Atributos: identificador único, título, fecha de creación, fecha de última actualización, ruta del proyecto al que pertenece y secuencia de mensajes e intervenciones de herramientas.
- **Tema Visual**: Configuración de apariencia de la interfaz. Valores admitidos: `goulm`, `dark`, `light`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El 100% de los proyectos abiertos por el usuario permanecen visibles en la lista de proyectos, incluso cuando contienen 0 conversaciones.
- **SC-002**: Al presionar el botón de nueva conversación en cualquier proyecto, la nueva sesión queda registrada y lista para interactuar en menos de 300 ms.
- **SC-003**: El 100% de las nuevas sesiones creadas se conservan en la base de datos y se restauran íntegramente tras reiniciar la aplicación.
- **SC-004**: El tema visual seleccionado se conserva en el 100% de los reinicios de la aplicación sin intervención del usuario.
- **SC-005**: El modal de Configuración refleja el cambio de tema en tiempo real sin requerir recarga ni dejar zonas con paletas de color incompatibles.

## Assumptions

- El usuario dispone de permisos de lectura y escritura en las rutas de proyecto seleccionadas.
- La persistencia de configuración utiliza el almacenamiento local del sistema operativo y las preferencias del usuario.
- Las conversaciones creadas desde un proyecto heredan automáticamente la carpeta de trabajo de dicho proyecto para la ejecución de herramientas.

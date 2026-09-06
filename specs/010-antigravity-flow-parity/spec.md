# Feature Specification: Paridad de Flujo Antigravity y Opencode (GUI y Bubbletea)

**Feature Branch**: `010-antigravity-flow-parity`  
**Created**: 2026-09-04  
**Status**: Draft  
**Input**: "crea una SPEC para acercar el proyecto(tanto en bubbletea como en GUI) a Antigravity GUI al flujo de Opencode, quiero que el selector de modelos y la ventana de chat y funciones sea Similar a la de antigravity, con capacidad para roll back de mensajes, el selector de modelos en el mismo espacio donde esta el text que se envia a la IA, un panel lateral para ver el git diff y archivos, y que las settings este en el lado inferior en la barra donde estan los proyectos, y demas"

---

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Selector de Modelos y Controles Integrados en el Compositor (Priority: P1) 🎯 MVP

Como desarrollador, deseo que el selector de modelos y controles de ejecución residan directamente dentro del contenedor del compositor de texto (en la base del cuadro de entrada, al estilo Antigravity y Opencode), para poder alternar modelos, proveedores y modos de trabajo sin desviar la mirada hacia barras superiores o menús externos.

**Why this priority**: Es la seña de identidad ergonómica de Antigravity y Opencode. Reduce la dispersión visual, manteniendo la atención focalizada en la redacción del prompt y en el modelo que lo procesará.

**Independent Test**:
Abrir la GUI, enfocar el compositor de texto y comprobar que en la base inferior del cuadro de entrada se encuentra el selector de modelo activo (`[🤖 modelo ▾]`), selector de modo (`[🛠️ code/plan ▾]`) y el botón de enviar/cancelar. Cambiar de modelo directamente desde allí y enviar un mensaje, confirmando que la IA responde utilizando el modelo elegido.

**Acceptance Scenarios**:
1. **Given** la ventana de chat en la GUI, **When** el usuario observa el compositor, **Then** visualiza un toolbar integrado dentro del contenedor del compositor con el pill selector de modelo, modo de ejecución y botón de acción.
2. **Given** el selector de modelo del compositor, **When** el usuario hace clic, **Then** se despliega una lista emergente filtrable con los modelos disponibles del proveedor activo y opción de alternar proveedor.
3. **Given** la terminal Bubbletea TUI, **When** el usuario interactúa con la línea de entrada, **Then** una barra inferior contextual muestra el modelo y modo activo con comandos rápidos (`/model`) para alternarlo sin salir del flujo.

---

### User Story 2 - Rollback y Rebobinado de Mensajes (Priority: P1)

Como desarrollador interactuando con el asistente, deseo poder rebobinar (roll back) la conversación a un turno previo cuando una respuesta de la IA no sea satisfactoria o se desvíe del objetivo, de modo que se descarten de SQLite y del motor los mensajes posteriores y el texto del prompt original vuelva al compositor para reintentar.

**Why this priority**: Previene la degradación del contexto y evita tener que borrar toda la sesión y empezar de cero. Proporciona control total sobre la historia de la conversación.

**Independent Test**:
Generar al menos 2 intercambios en el chat. En el mensaje de usuario que se desea rebobinar, hacer clic en la acción "↩ Rebobinar" (o escribir `/rollback` en TUI). Verificar que todos los mensajes generados posteriormente se purgan en SQLite y en la UI, y el texto original del mensaje se carga en el compositor listo para editar.

**Acceptance Scenarios**:
1. **Given** una conversación con varios mensajes en la GUI, **When** el usuario pulsa el botón "Rebobinar" en un mensaje anterior, **Then** se eliminan de la base de datos SQLite todos los mensajes posteriores a dicho punto, se resincroniza el motor y el texto del mensaje vuelve al área de edición.
2. **Given** Bubbletea TUI, **When** el usuario escribe `/rollback` o pulsa `Ctrl+Z`, **Then** se descarta el último turno de respuesta y se recupera el prompt anterior en el `textarea`.
3. **Given** un stream en progreso, **When** el usuario solicita rollback, **Then** se cancela de inmediato el streaming activo antes de efectuar el truncamiento en la base de datos.

---

### User Story 3 - Panel Lateral de Git Diff e Inspección de Archivos (Priority: P2)

Como desarrollador, deseo disponer de un panel lateral desplegable para visualizar en tiempo real los cambios del repositorio (`git diff`) y la lista de archivos modificados, para auditar los cambios producidos por las herramientas de la IA de manera inmediata.

**Why this priority**: En Antigravity y Opencode, tener a la vista las diferencias de código en el lateral del chat agiliza la revisión de código sin recurrir a herramientas externas.

**Independent Test**:
Realizar cambios en archivos del proyecto activo (o dejar que el asistente ejecute herramientas de edición). Abrir el panel de Git Diff en la GUI; comprobar que muestra la lista de archivos modificados y el diff unificado con código de color (verde para inserciones, rojo para eliminaciones). En Bubbletea, ejecutar `/diff` para ver el diff en el visor.

**Acceptance Scenarios**:
1. **Given** un proyecto versionado con Git en la GUI, **When** el usuario abre el panel lateral de Git (botón en la barra superior o atajo `Ctrl+D`), **Then** se expande un panel a la derecha mostrando el árbol de archivos cambiados y el visor de diff coloreado.
2. **Given** el panel de Git Diff abierto, **When** el usuario hace clic en un archivo de la lista, **Then** el visor se posiciona en el bloque de cambios correspondiente.
3. **Given** Bubbletea TUI, **When** el usuario escribe `/diff`, **Then** el viewport entra en modo visualizador de git diff interactivo y permite salir con `Esc` o `q`.

---

### User Story 4 - Reubicación de Ajustes en la Barra Inferior de Proyectos (Priority: P2)

Como usuario de la aplicación, deseo que el acceso a Configuración (`Settings`) se ubique en la parte inferior de la barra lateral izquierda (donde están los proyectos y conversaciones), despejando la cabecera superior y alineándose con la arquitectura de navegación de Antigravity / Opencode.

**Why this priority**: Optimiza el espacio de trabajo, concentrando la gestión de proyectos, conversaciones y configuración global en la barra lateral primaria.

**Independent Test**:
Abrir la GUI y verificar que la barra lateral izquierda contiene un pie (`sidebar-footer`) con el botón de Ajustes (`⚙ Ajustes`), y que al hacer clic se abre el modal de Configuración conservando los estilos cromáticos del tema activo.

**Acceptance Scenarios**:
1. **Given** la barra lateral izquierda `#sidebar`, **When** se observa el pie del panel, **Then** contiene el botón de acceso directo a Ajustes (`⚙ Configuración`).
2. **Given** la cabecera superior, **When** se verifica su distribución, **Then** queda más limpia y enfocada en el título del proyecto y paneles de trabajo.

---

### User Story 5 - Tarjetas de Ejecución de Herramientas Estilo Antigravity (Priority: P3)

Como usuario, deseo que las llamadas a herramientas en el historial de conversación se presenten en tarjetas estilizadas y colapsables con estados claros (ejecutando, completado, error), tiempo transcurrido y vista previa de diff o salida.

**Why this priority**: Mejora radicalmente la legibilidad del historial cuando el agente encadena múltiples llamadas a herramientas de análisis o edición.

**Independent Test**:
Pedir al asistente una tarea que invoque herramientas (`view_file`, `write_to_file`, etc.). Verificar que cada herramienta aparece como una tarjeta compacta con icono, argumentos esenciales colapsables y badge de estado.

**Acceptance Scenarios**:
1. **Given** una respuesta del asistente que ejecuta una herramienta, **When** se muestra en el chat, **Then** se renderiza como una tarjeta colapsable con badge de estado e información clara de la operación.

---

## Edge Cases

- **Rollback de la primera pregunta**: Si se rebobina el mensaje inicial de la sesión, la sesión se preserva vacía en SQLite y el texto vuelve al compositor.
- **Rollback durante streaming activo**: Se cancela el stream mediante `ChatService.Cancel()` antes de alterar la base de datos.
- **Proyecto sin repositorio Git**: El panel lateral de Git Diff detecta la ausencia de `.git` y muestra un mensaje informativo en lugar de romperse.
- **Árbol de trabajo sin modificaciones**: El panel lateral de Git Diff muestra *"Árbol de trabajo limpio"* con opción de refrescar.

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema MUST integrar el selector de modelos y selector de modos de trabajo en la barra de herramientas interna del compositor de chat en la GUI.
- **FR-002**: El selector del compositor MUST desplegar los modelos configurados con filtrado rápido y permitir cambiar de modelo activo inmediatamente.
- **FR-003**: El backend MUST exponer un método `Rollback(sessionID string, messageID int64) error` que elimine todos los mensajes posteriores o iguales al ID seleccionado y resincronice la memoria del motor.
- **FR-004**: La GUI MUST mostrar una acción "Rebobinar" en los mensajes del chat que ejecute el rollback y cargue el texto del mensaje en el compositor.
- **FR-005**: Bubbletea TUI MUST proveer el comando `/rollback` para descartar el último turno de conversación y restaurar el prompt en el `textarea`.
- **FR-006**: El backend MUST proveer datos de diff y estado de archivos mediante `GitService.Diff()` y `GitService.Status()`.
- **FR-007**: La GUI MUST implementar un panel lateral desplegable de Git Diff con lista de archivos modificados y visor de diff unificado coloreado.
- **FR-008**: Bubbletea TUI MUST proveer el comando `/diff` para inspeccionar el diff en el visor de terminal.
- **FR-009**: El botón de Configuración en la GUI MUST reubicarse en el pie de la barra lateral izquierda (`#sidebar-footer`).
- **FR-010**: Las ejecuciones de herramientas en el historial de chat MUST renderizarse como tarjetas colapsables con estados visuales claros.

---

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El cambio de modelo en la GUI se realiza directamente desde el compositor en 2 clics o menos.
- **SC-002**: El rollback de mensajes elimina los registros en SQLite en menos de 100ms y repuebla el compositor al instante.
- **SC-003**: El panel de Git Diff se abre y presenta los cambios del workspace en menos de 300ms.
- **SC-004**: El 100% de las suites de prueba existentes en Go (`go test ./...`) y frontend (`npm test`) se mantienen en estado aprobado.

---

## Assumptions

- El motor `internal/engine` permite resembrar o recargar el historial de mensajes de la sesión tras una operación de rollback.
- Las utilidades de git del sistema operativo están disponibles para la lectura de diffs cuando el workspace es un repositorio.
- Todos los componentes visuales nuevos heredan y respetan las variables CSS de los 3 temas admitidos (`goulm`, `dark`, `light`).

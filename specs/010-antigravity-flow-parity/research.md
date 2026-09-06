# Research: Paridad de Flujo Antigravity y Opencode (GUI y Bubbletea)

## 1. Integración del Selector de Modelos en el Compositor

### Decisión
Trasladar el selector de modelos y el selector de modo de trabajo (Code / Plan) desde la barra de cabecera superior (`#app-header`) directamente a la barra inferior interna del compositor de chat (`#chat-input-pane`).
- La tarjeta del compositor contendrá:
  - En la parte superior/central: el área de texto de redacción multilínea (`textarea`).
  - En la barra de herramientas inferior del compositor:
    - A la izquierda: El botón pill del modelo activo con icono de proveedor (`[🤖 modelo ▾]`) que al pulsar despliega un menú emergente con filtro y lista de modelos.
    - Junto al modelo: El selector de modo (`[🛠️ Code ▾]`).
    - A la derecha: Botón de adjuntos (`+` o clip), contador de tokens / contexto y botón de acción principal (Enviar `↑` o Detener `■` si está en streaming).
- En Bubbletea TUI:
  - Mostrar en una barra de estado directamente adyacente a la caja de texto el modelo activo (`[openrouter/auto-beta]`), modo (`[build/plan]`) y atajo `/model`.

### Razón
En Antigravity y Opencode, el compositor es el centro de gravedad de la experiencia. La elección del modelo y modo forma parte del contexto de formulación de la pregunta, no de la configuración global de la ventana.

### Alternativas consideradas
- *Mantener selector en cabecera y duplicar en compositor*: Causa confusión sobre cuál es la fuente de verdad.
- *Usar solo comandos de barra `/model`*: Es rápido para usuarios avanzados pero menos intuitivo para usuarios visuales en la GUI.

---

## 2. Mecanismo de Rollback y Rebobinado de Mensajes

### Decisión
Implementar la capacidad de rebobinar la conversación a cualquier mensaje previo del usuario o turno del asistente:
1. **En la base de datos (`internal/db/database.go`)**:
   - Crear la función `RollbackSession(sessionID string, targetMessageID int64) (string, error)`:
     - Consulta el contenido del mensaje `targetMessageID`.
     - Ejecuta: `DELETE FROM messages WHERE session_id = ? AND id >= ?`.
     - Si el mensaje eliminado tenía llamadas a herramientas de edición de archivos registradas en `context_files` o metadatos, se limpian las referencias huérfanas.
     - Actualiza el `updated_at` de la sesión.
     - Retorna el contenido de texto del mensaje rebobinado para cargarlo de nuevo en el editor.
2. **En el backend de servicios (`cmd/wails/services/sessions_service.go`)**:
   - Exponer `Rollback(sessionID string, messageID int64) (string, error)`:
     - Cancela cualquier stream activo en el motor (`s.hub.Engine.Send(engine.Cancel{})`).
     - Invoca `db.RollbackSession`.
     - Recarga la lista de mensajes en el motor mediante `db.GetHistory(sessionID)` y `s.hub.Engine.SeedMessages(...)`.
     - Emite el evento `session:loaded` para que todas las vistas actualicen su estado al instante.
3. **En la GUI (`cmd/wails/frontend/src/App.tsx`)**:
   - Añadir en cada mensaje de usuario un botón discreto de acción "↩ Rebobinar" (o icono de undo).
   - Al pulsar: llama a `SessionsService.Rollback(activeTab.sessionId, message.id)`, carga el texto del mensaje en `input` y enfoca el compositor.
4. **En Bubbletea TUI (`internal/tui/ui.go`)**:
   - Añadir comando `/rollback` y soporte para atajo de teclado:
     - Obtiene el último mensaje del usuario del historial de la sesión.
     - Ejecuta rollback en SQLite y el motor.
     - Carga el contenido de texto en el `textarea` del TUI.

### Razón
Proporciona el flujo iterativo no destructivo típico de Antigravity: si una respuesta no es adecuada, el desarrollador hace clic en rebobinar, ajusta el prompt y vuelve a intentarlo sin ensuciar la ventana de contexto del LLM.

---

## 3. Panel Lateral de Git Diff e Inspección de Archivos

### Decisión
Implementar un panel lateral desplegable a la derecha en la GUI dedicado a cambios de Git y archivos modificados:
- En la barra de herramientas superior se agrega un botón selector "Git Changes" (icono de Git / branch) o atajo `Ctrl+D`.
- Al activarse, la ventana adopta un layout de 3 columnas o panel superpuesto:
  - Lista de archivos modificados (con badges de `M` modificado, `A` añadido, `D` eliminado).
  - Visor de diff unificado con numeración de líneas y coloreado de sintaxis (verde para líneas añadidas, rojo para líneas borradas).
- En el backend:
  - `GitService.Diff()` y `GitService.Status()` proveen los datos requeridos.
- En Bubbletea TUI:
  - Comando `/diff` que alterna la visualización del `git diff` coloreado en el `viewport` principal, con salida rápida mediante `Esc` o `q`.

### Razón
En el flujo de desarrollo asistido por IA, los agentes leen, escriben y modifican archivos. Tener un panel de diff accesible permite revisar visualmente las modificaciones en caliente antes de continuar con el siguiente paso.

---

## 4. Reubicación de Ajustes al Pie de la Barra Lateral

### Decisión
Reubicar el acceso a Configuración (`SettingsView`) desde la cabecera superior al pie de la barra lateral izquierda (`#sidebar`):
- Crear un contenedor `#sidebar-footer` en la base del sidebar con:
  - Botón de Configuración (`⚙ Ajustes` / `Configuración`).
  - Información de versión o estado rápido.
- Al hacer clic, abre el modal de `SettingsView` ya existente con soporte multitema.

### Razón
En Antigravity, Opencode, VS Code y editores modernos, la cabecera superior se reserva para pestañas de trabajo o título de proyecto, mientras que las opciones de configuración y perfil se ubican en la esquina inferior izquierda.

---

## 5. Tarjetas de Ejecución de Herramientas Estilo Antigravity

### Decisión
Refactorizar la visualización de herramientas en el historial de mensajes de la GUI:
- Las llamadas a herramientas (`tool:start`, `tool:end`) se muestran como una tarjeta compacta:
  - Encabezado: Icono de herramienta (ej. terminal para bash, lápiz para edit_file, ojo para view_file), nombre de la herramienta, argumento principal (ej. ruta de archivo) y badge de estado (`running`, `success`, `error`).
  - Cuerpo colapsable: Entrada detallada y salida devuelta por la herramienta (con resaltado de diff para herramientas de edición).
- En TUI:
  - Renderizar cajas con bordes redondeados usando Lipgloss mostrando el estado y resumen de la herramienta.

### Razón
Reemplaza volcados planos de texto JSON por una interfaz estructurada y legible, idéntica a la experiencia de Antigravity.

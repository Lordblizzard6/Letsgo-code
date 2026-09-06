# Feature Specification: Codex-Style GUI Redesign

**Feature Branch**: `002-codex-style-ui`

**Created**: 2026-08-06

**Status**: Draft

**Input**: User description: "mejorar la interfaz para ser similar a la de Codex (OpenAI), basándose en el repo github.com/openai/codex y su GUI"

**Referencia de diseño**: El usuario se refiere tanto al TUI de Codex como a su **GUI de escritorio** (la app `codex app` / la visión de la interfaz de Codex). La dirección visual resultante es una *interpretación moderna propia* (paleta y acento propios de LetsGO), no una réplica literal del TUI ANSI.

## User Scenarios & Testing *(mandatory)*

Priorizadas como journeys independientemente implementables: cada una entrega valor por sí sola y se puede demostrar sin el resto.

### User Story 1 - Transcript-first chat with inline approvals (Priority: P1)

El usuario conversa con el agente en un único scrollback continuo: prompt, respuesta en streaming, llamadas a herramientas, diffs, aprobaciones, errores — todo en línea, nunca en ventanas modales que tapen la conversación. Cuando el agente pide permiso para una herramienta, aparece una tarjeta de aprobación *inline* sobre el composer (el transcript sigue visible y scrolleable encima) con: nombre de la herramienta, descripción de la acción, vista previa del diff cuando aplique, y pistas de teclado (`Enter` aprobar, `Esc` rechazar). El usuario decide en menos de 2 segundos sin perder contexto.

**Why this priority**: Es la diferencia más visible con el TUI de Codex (principio "el transcript es la interfaz") y elimina el mayor dolor actual: el diálogo modal bloqueante que oculta la conversación durante una decisión.

**Independent Test**: Se puede probar pidiendo al agente que cree/edite un archivo en un repo de prueba: aparece la tarjeta inline, el transcript permanece visible y scrolleable, Enter aprueba, Esc rechaza, y la tarjeta desaparece sin cerrar la ventana.

**Acceptance Scenarios**:

1. **Given** una conversación con el agente en curso, **When** el agente solicita ejecutar una herramienta, **Then** se muestra una tarjeta de aprobación inline sobre el composer, sin ocultar los mensajes previos y sin bloquear el scroll del transcript.
2. **Given** una tarjeta de aprobación visible, **When** el usuario pulsa Enter, **Then** la herramienta se ejecuta, la tarjeta muestra el resultado inline y la conversación continúa.
3. **Given** una tarjeta de aprobación visible, **When** el usuario pulsa Esc, **Then** la herramienta se rechaza, el agente recibe un motivo estructurado de denegación y continúa adaptándose (no reintenta la misma acción en el siguiente turno en >60% de los casos).
4. **Given** una herramienta con diff aplicable (p. ej. edición de archivo), **When** se muestra la tarjeta de aprobación, **Then** el diff se renderiza inline con resaltado de adiciones/eliminaciones antes de decidir.

---

### User Story 2 - Always-on work status strip (Priority: P1)

Mientras el agente trabaja, el usuario siempre sabe *qué está haciendo ahora mismo*: una franja entre el transcript y el composer muestra estado ("Working"), spinner animado, cronómetro transcurrido (47s, 4m 27s), el paso actual (p. ej. `└  running npm test`), y la pista `esc to interrupt`. La franja aparece al iniciar un turno y se oculta mientras el agente escribe texto (streaming), reapareciendo entre ráfagas de actividad. Al pulsar Esc durante un turno activo, el stream se detiene en ≤500 ms y el composer se rehabilita.

**Why this priority**: Es el punto de mayor confusión con agentes autónomos — "¿qué está haciendo?" — y Codex lo resuelve con una señal siempre presente (principios P-3/P-7).

**Independent Test**: Se puede probar con un turno largo (p. ej. "refactoriza y ejecuta los tests"): la franja aparece con spinner y cronómetro, muestra el paso actual cambiando, y Esc interrumpe en <500 ms.

**Acceptance Scenarios**:

1. **Given** un turno activo del agente, **When** el agente está ejecutando herramientas, **Then** una franja de estado muestra "Working", un spinner, el tiempo transcurrido y el paso actual, visible entre el transcript y el composer.
2. **Given** un turno activo, **When** el usuario pulsa Esc, **Then** el stream se detiene en ≤500 ms y el composer queda habilitado para escribir.
3. **Given** el agente escribiendo texto en streaming, **When** no hay actividad de herramientas, **Then** la franja de estado se oculta para no duplicar la señal.
4. **Given** una tarea con múltiples pasos (herramientas secuenciales), **When** cada paso inicia, **Then** la franja muestra el paso actualizado (comando, archivo o herramienta en curso).

---

### User Story 3 - Composer as the control room (Priority: P1)

El composer es el centro de control: con el turno activo, el input permanece editable y `Enter` con texto inyecta una nueva instrucción en el turno en curso (steer); `Tab` encola un seguimiento para el siguiente turno (queue); `Ctrl+K` (o `/`) abre una paleta de comandos filtrable que da acceso a cada acción de la app (nueva sesión, reanudar, fork, settings, usage, git, mcp, plugins, plan mode, compact, doctor, keymap). El usuario nunca necesita el ratón ni los menús para fluir.

**Why this priority**: Es el "composer es la sala de control" de Codex (P-2/P-9): elimina la fricción de esperar a que el agente termine para volver a escribir y reemplaza la navegación por menús con un solo atajo.

**Independent Test**: Se puede probar con un turno largo: escribir texto y pulsar Enter (steer) se entrega al turno activo sin esperar; Tab encola sin ejecutar; Ctrl+K muestra la paleta y filtrar + Enter ejecuta una acción.

**Acceptance Scenarios**:

1. **Given** un turno activo, **When** el usuario escribe y pulsa Enter, **Then** la instrucción se inyecta en el turno en curso y el agente la atiende sin esperar al fin del turno.
2. **Given** un turno activo, **When** el usuario pulsa Tab con un texto escrito, **Then** el texto se encola como siguiente turno y el composer se vacía sin ejecutarlo aún.
3. **Given** el composer con foco, **When** el usuario pulsa Ctrl+K o escribe `/`, **Then** se abre una paleta filtrable que lista todas las acciones navegables y ejecutables solo con teclado.
4. **Given** la paleta abierta, **When** el usuario teclea para filtrar y pulsa Enter sobre una acción, **Then** la acción se ejecuta y el foco vuelve al composer.
5. **Given** el turno activo, **When** el usuario intenta escribir, **Then** el input sigue editable (nunca deshabilitado durante un turno).

---

### User Story 4 - Plan mode as a visible mode with review card (Priority: P2)

El modo plan se convierte en un estado explícito y visible: la barra de estado indica el modo actual (Plan / Execute / Auto). En modo Plan, el agente propone una tarjeta de plan estructurada (pasos, archivos afectados, criterios) y el usuario la **aprueba, edita las instrucciones o la rechaza** antes de tocar archivos. Aprobar la estrategia, no cada microacción.

**Why this priority**: Cierra la brecha funcional más grande con Codex (P-6 "plan primero, ejecutar después") y reduce el riesgo percibido en tareas multi-archivo.

**Independent Test**: Se puede probar pidiendo una tarea multi-archivo en modo Plan: el agente produce la tarjeta de plan, el usuario la aprueba/rechaza con teclado, y solo tras aprobar se ejecutan herramientas.

**Acceptance Scenarios**:

1. **Given** el modo Plan activo, **When** el usuario pide una tarea de varios archivos, **Then** el agente presenta una tarjeta de plan (pasos, archivos, criterios) sin ejecutar herramientas de escritura.
2. **Given** una tarjeta de plan visible, **When** el usuario pulsa Enter (aprobar), **Then** comienza la ejecución; al pulsar Esc (rechazar), **Then** nada se ejecuta y el agente ajusta el plan.
3. **Given** una tarjeta de plan visible, **When** el usuario elige "editar instrucciones", **Then** puede añadir correcciones al plan antes de aprobarlo.
4. **Given** el modo Plan activo, **When** la barra de estado se renderiza, **Then** muestra "Plan" de forma distinguible del modo Execute/Auto.

---

### User Story 5 - Startup resume picker with fork (Priority: P2)

Al abrir la app (cuando existen sesiones), se muestra un selector rápido de sesión: **Nueva sesión / Reanudar la última / Elegir sesión** (filtrable por cwd y fecha). Desde cualquier sesión, el menú permite **fork** (duplicar la sesión en un punto elegido) sin tocar la original. El usuario retoma exactamente donde lo dejó.

**Why this priority**: Sesiones como ciudadanos de primera clase (P-8): reduce el costo de "¿en qué sesión trabajé?" y habilita exploración en paralelo sin riesgo.

**Independent Test**: Se puede probar con dos sesiones existentes: al lanzar, aparece el picker; reanudar la última carga la historia completa; fork crea una copia independiente que no altera la original al usarla.

**Acceptance Scenarios**:

1. **Given** la app con sesiones previas, **When** el usuario la inicia, **Then** aparece un picker con Nueva/Reanudar última/Elegir sesión (filtrable por cwd y fecha).
2. **Given** una sesión abierta, **When** el usuario elige "Fork", **Then** se crea una copia independiente en el punto elegido y se abre, sin modificar la original.
3. **Given** el picker abierto, **When** el usuario teclea para filtrar y pulsa Enter, **Then** la sesión seleccionada se abre y el foco va al composer.

---

### User Story 6 - Smarter approvals: session grants (Priority: P2)

La tarjeta de aprobación ofrece una tercera decisión además de Aprobar/Rechazar: **"Permitir siempre en esta sesión"** (accept-for-session). Las repeticiones del mismo tipo de herramienta en la misma sesión no vuelven a preguntar, y el usuario puede ver/limpiar los grants en la configuración. Las aprobaciones por sesión reducen la fatiga de confirmación sin comprometer la seguridad global.

**Why this priority**: La fatiga de aprobación repetida ("el agente vuelve a ejecutar `git status`") es la queja más frecuente; Codex la resuelve con grants por sesión (P-4/P-9).

**Independent Test**: Se puede probar con un turno que ejecuta la misma herramienta varias veces: tras "permitir en esta sesión", las siguientes invocaciones se ejecutan sin prompt y el grant aparece listado y removible en settings.

**Acceptance Scenarios**:

1. **Given** una tarjeta de aprobación, **When** el usuario elige "Permitir en esta sesión", **Then** la herramienta se ejecuta y las siguientes invocaciones de esa categoría en la sesión no piden confirmación.
2. **Given** grants activos, **When** el usuario abre la configuración de aprobaciones, **Then** ve la lista de grants por sesión y puede revocarlos individualmente.
3. **Given** una sesión con grants, **When** el usuario inicia una sesión nueva, **Then** los grants de la sesión anterior no aplican (alcance estricto de sesión).

---

### User Story 7 - Status line & keymap transparency (Priority: P3)

La barra de estado es configurable y honesta: muestra modelo, rama git, modo (Plan/Execute/Auto), uso de contexto (progreso usado/restante), límites de rate y versión; el usuario puede elegir qué campos ver desde un diálogo de "campos de estado" persistido. Un overlay `?` muestra el cheat sheet de atajos categorizado y filtrable, sin salir del teclado.

**Why this priority**: Transparencia anti-caja-negra (P-7) y descubribilidad de atajos (P-9) con esfuerzo acotado.

**Independent Test**: Se puede probar abriendo el overlay `?` desde el composer y verificando que lista los atajos por categoría; y en settings, desmarcando campos de la barra de estado y viendo el cambio inmediato tras guardar.

**Acceptance Scenarios**:

1. **Given** el composer con foco, **When** el usuario pulsa `?`, **Then** se muestra un cheat sheet de atajos categorizado y filtrable, cerrable con Esc.
2. **Given** la barra de estado visible, **When** el usuario configura los campos en settings, **Then** solo se muestran los campos elegidos y la preferencia persiste entre reinicios.
3. **Given** el modo de permisos activo, **When** la barra de estado se renderiza, **Then** el modo (Plan/Execute/Auto) es siempre visible.

---

### Edge Cases

- ¿Qué pasa si llega una aprobación mientras otra está visible? → Se encolan en orden; la tarjeta indica cola ("2 decisiones pendientes") y el transcript nunca se bloquea.
- ¿Qué pasa si el usuario cierra la ventana con una aprobación pendiente? → Se trata como rechazo (comportamiento actual, sin cambio).
- ¿Qué pasa si un steered message llega cuando el turno está terminando? → Se entrega al turno siguiente si el actual ya emitió Idle; se muestra una nota inline.
- ¿Qué pasa si el modo Plan está activo y el usuario pide ejecución directa? → La tarjeta de plan explica que la ejecución requiere aprobación previa; no se ejecuta nada.
- ¿Qué pasa si un grant de sesión choca con una regla persistente de negación? → Gana la regla más restrictiva (deny > session-grant > auto).
- ¿Qué pasa si el picker de sesión se abre sin sesiones previas? → Solo muestra "Nueva sesión" y continúa directo al chat sin fricción.
- ¿Qué pasa si el rate limit está stale (>15 min sin refresh)? → El indicador se marca como desactualizado en vez de mostrar datos falsos.
- ¿Qué pasa si Ctrl+K/`/` colisiona con el teclado del composer? → La paleta se abre solo en modo no-escritura; el texto de `/` se mantiene si el usuario cancela.
- ¿Qué pasa si el usuario pulsa Enter (steer) mientras el turno está a punto de terminar (race)? → El mensaje se encola y se muestra como cola pendiente.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Las aprobaciones de herramientas MUST mostrarse como tarjetas inline sobre el composer, dejando el transcript visible y scrolleable (sin modales a pantalla completa).
- **FR-002**: La tarjeta de aprobación MUST mostrar nombre de la herramienta, resumen de la acción, y diff inline resaltado cuando la acción tenga diff aplicable.
- **FR-003**: La tarjeta de aprobación MUST ser operable solo con teclado: Enter aprueba, Esc rechaza.
- **FR-004**: Al rechazar una herramienta, el agente MUST recibir un motivo estructurado de denegación y continuar sin reintentar la misma acción en el turno siguiente (salvo que cambie el contexto) — verificable en >60% de los casos.
- **FR-005**: Durante un turno activo, una franja de estado MUST mostrar "Working", spinner, tiempo transcurrido, paso actual y pista `esc to interrupt`.
- **FR-006**: La franja de estado MUST ocultarse durante el streaming de texto del asistente y reaparecer entre ráfagas de actividad de herramientas.
- **FR-007**: Esc durante un turno activo MUST detener el stream en ≤500 ms y rehabilitar el composer.
- **FR-008**: El composer MUST permanecer editable durante un turno activo (nunca deshabilitado).
- **FR-009**: Enter con texto durante un turno activo MUST inyectar la instrucción en el turno en curso (steer).
- **FR-010**: Tab con texto durante un turno activo MUST encolar el texto como siguiente turno sin ejecutarlo.
- **FR-011**: Una paleta de comandos (Ctrl+K o `/`) MUST listar todas las acciones navegables de la app, filtrable y ejecutable solo con teclado.
- **FR-012**: El modo Plan MUST ser un estado visible en la barra de estado (Plan/Execute/Auto) y MUST impedir escritura de archivos hasta que el usuario apruebe el plan.
- **FR-013**: En modo Plan, el agente MUST presentar una tarjeta de plan (pasos, archivos, criterios) con opciones Aprobar / Rechazar / Editar instrucciones.
- **FR-014**: Al iniciar la app con sesiones previas, un picker MUST ofrecer Nueva / Reanudar última / Elegir sesión, filtrable por cwd y fecha.
- **FR-015**: El usuario MUST poder hacer fork de una sesión en un punto elegido, creando una copia independiente sin modificar la original.
- **FR-016**: La tarjeta de aprobación MUST ofrecer "Permitir siempre en esta sesión", y las invocaciones posteriores de esa categoría en la misma sesión MUST no preguntar.
- **FR-017**: Los grants de sesión MUST ser listados y revocables desde la configuración de aprobaciones, y MUST no trascender a otras sesiones.
- **FR-018**: La barra de estado MUST ser configurable (modelo, rama, modo, uso de contexto, rate limit, versión), con preferencias persistidas.
- **FR-019**: Un overlay `?` MUST mostrar el cheat sheet de atajos categorizado y filtrable, cerrable con Esc.
- **FR-020**: Si el uso de contexto o rate limit supera 15 min sin actualizarse, el indicador MUST marcarse como desactualizado.
- **FR-021**: Las decisiones de aprobación en cola MUST mostrarse como cola ("N decisiones pendientes") sin bloquear el transcript.
- **FR-022**: Cerrar la ventana con aprobación pendiente MUST tratarse como rechazo.
- **FR-023**: Los atajos existentes (Ctrl+N, Ctrl+1..9, Ctrl+,, Ctrl+U, Ctrl+L, Ctrl+Shift+A) MUST seguir funcionando tras el rediseño.

*Ejemplo de marcado de requisitos poco claros:*

- **FR-024**: El rediseño visual MUST seguir una dirección propia de LetsGO inspirada en la GUI de escritorio de Codex: paleta y acento distintivo definidos en la guía de diseño (no una réplica literal del TUI ANSI magenta/terminal), manteniendo el tema oscuro y aplicando las buenas prácticas de legibilidad, jerarquía y densidad de la referencia. *(Decidido: Q1=C — identidad propia con acento distintivo, informado por la GUI desktop de Codex.)*
- **FR-025**: El alcance del rediseño MUST cubrir todas las capacidades definidas en este documento (US1–US7): transcript-first con aprobaciones inline, franja de estado, composer como centro de control, modo Plan visible, picker de sesión con fork, grants por sesión, y barra de estado/keymap configurables. *(Decidido: Q2=C — alcance completo en una sola entrega.)*
- **FR-026**: La paleta de comandos MUST incluir acciones internas de la app, picker de archivos (`@`, búsqueda difusa sobre el proyecto) y shell passthrough (`!`), replicando la paridad de Codex. *(Decidido: Q3=C — alcance completo; el shell passthrough queda sujeto a las restricciones de seguridad existentes.)*

### Key Entities *(include if feature involves data)*

- **ApprovalDecision**: Representa una decisión del usuario sobre una herramienta (aprobar / rechazar / permitir-en-sesión), con herramienta, resumen, diff aplicable y motivo de rechazo estructurado; alimenta al motor y a la UI.
- **SessionGrant**: Regla de aprobación persistente por sesión (categoría → permitido), con visibilidad y revocación individual.
- **WorkStatus**: Estado en vivo del turno activo (fase, paso actual, tiempo transcurrido, interrumpible) derivado de los eventos del motor.
- **PlanCard**: Representación del plan propuesto (pasos, archivos, criterios, estado: pendiente/aprobado/editando/rechazado).
- **SessionPickerEntry**: Sesión candidata en el picker de inicio (nombre, cwd, última actividad) para Nueva/Reanudar/Fork.
- **StatusField**: Campo configurable de la barra de estado (modelo, rama, modo, uso de contexto, rate limit, versión) con preferencia persistida por usuario.
- **CommandPaletteEntry**: Acción navegable de la app (etiqueta, categoría, atajo asociado) para la paleta de comandos.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El usuario puede decidir una aprobación de herramienta en menos de 2 segundos (p90) desde que la tarjeta aparece, con el transcript visible en todo momento.
- **SC-002**: Al menos el 90% de los usuarios nuevos completan configuración → primera conversación sin ayuda externa ni documentación.
- **SC-003**: El 100% del walkthrough de teclado (S7: Ctrl+L → enviar → aprobar con Enter → Ctrl+N → Ctrl+1 → Esc) funciona sin ratón tras el rediseño.
- **SC-004**: Las aprobaciones por sesión reducen el número de prompts de confirmación por sesión en al menos un 40% respecto al estado actual.
- **SC-005**: Esc interrumpe el stream en ≤500 ms y el composer queda operativo (medible en el escenario S2 del quickstart).
- **SC-006**: Al menos el 25% de las sesiones de múltiples turnos usan steer (Enter) o queue (Tab) una vez implementada la paleta.
- **SC-007**: El transcript permanece visible y scrolleable durante cualquier decisión de aprobación (nunca tapado por un modal).
- **SC-008**: Tras un rechazo, el agente adapta su siguiente acción (no reintenta la misma herramienta) en más del 60% de los casos medidos.
- **SC-009**: El usuario puede completar el flujo principal de la app (nueva sesión, chat, aprobación, reanudar) solo con teclado, sin abrir menús.
- **SC-010**: El usuario puede identificar el modo actual (Plan/Execute/Auto) y el paso actual del agente sin abrir ninguna ventana adicional.

## Assumptions

- Los usuarios objetivo son desarrolladores que usan CLIs de IA (Codex, Claude Code) y valoran el teclado, la densidad de información y la transparencia del estado.
- La entrega cubre todo el alcance (US1–US7) en una sola iteración (decisión Q2=C).
- El rediseño se aplica solo al escritorio (ventana nativa); no hay versión móvil/web en v1.
- El motor (engine) existente ya emite eventos suficientes (stream, herramientas, aprobaciones, tareas) para alimentar las nuevas superficies; los eventos adicionales para steer/queue/plan se añadirán al contrato interno sin cambiar el contrato público.
- El tema oscuro por defecto se mantiene; el acento visual es una identidad propia de LetsGO inspirada en la GUI de escritorio de Codex (panes, jerarquía, densidad), no una réplica literal del TUI ANSI.
- La seguridad no se relaja: las negaciones persistentes (permissions) siempre ganan sobre los grants de sesión, y el shell passthrough (`!`) queda sujeto a las reglas de permisos existentes.
- Los menús existentes se mantienen como ruta secundaria; la paleta y los atajos son la ruta primaria.
- El overlay de aprobación usa el contrato de permisos existente (PermissionDriver) sin cambios de seguridad.
- Los datos de sesiones y mensajes ya existen en la base de datos local; el picker y fork no requieren datos nuevos.
- La experiencia multiventana no aplica: una sola ventana principal, como hoy.

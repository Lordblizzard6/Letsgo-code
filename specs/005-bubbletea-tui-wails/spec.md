# Feature Specification: Doble interfaz — TUI con Bubble Tea y GUI de escritorio con Wails

**Feature Branch**: `005-bubbletea-tui-wails`

**Created**: 2026-08-11

**Status**: Borrador

**Input**: User description: "crea un SPEC para crear una TUI con bubble tea y la GUI en Wails"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - El core compartido: una sola fuente de verdad para ambas interfaces (Priority: P1)

El producto ofrece dos formas de usar el mismo cliente de chat: una terminal (TUI) y una ventana de escritorio (GUI). El usuario alterna entre ellas sin perder nada: la conversación, el modelo elegido, la API key y el historial son **los mismos** en ambas. Detrás de eso hay un único núcleo compartido — el motor de conversación (con su bus de comandos/eventos), la base de datos sqlite de sesiones y la configuración persistente — y un **contrato de frontend** documentado y verificado por tests que la TUI y la GUI Wails consumen por igual. Ninguna interfaz duplica la lógica de conversación; ambas se limitan a enviar comandos, consumir eventos y persistir/leer sesiones y configuración.

**Why this priority**: Es la base de todo el feature. El motor (`internal/engine`), la base de datos (`internal/db`) y la configuración (`internal/config`) ya existen y ya sirven a la TUI y a la GUI Fyne actuales; sin un contrato formalizado y verificado, la nueva GUI Wails correría el riesgo de forkar la lógica o derivar en comportamientos divergentes entre interfaces. Es el primer slice porque tanto la TUI (P2) como la GUI Wails (P3) se construyen encima.

**Independent Test**: Suite de tests de conformidad del contrato (headless, sin pantalla) que ejercita comandos, eventos, sesiones y configuración contra el core compartido, más una prueba manual: iniciar una conversación desde la TUI, cerrarla y abrirla desde la GUI viendo el mismo historial. Entrega la garantía de "una sola fuente de verdad" sin necesidad de ninguna de las dos interfaces terminadas.

**Acceptance Scenarios**:

1. **Given** el contrato de frontend documentado, **When** la TUI y la GUI Wails envían comandos y consumen eventos del motor, **Then** ambas reciben exactamente los mismos tipos de evento (inicio/delta/fin de stream, ejecución de herramientas, errores, uso) y los traducen a su propia vista sin modificar el core.
2. **Given** una conversación iniciada desde la TUI y persistida en la base de datos, **When** el usuario la abre desde la GUI, **Then** ve el historial completo de esa misma sesión (mismo id, mismos mensajes).
3. **Given** una configuración guardada (API key, modelo, proveedor), **When** se inicia cualquiera de las dos interfaces, **Then** ambas usan esa misma configuración y el modelo activo se refleja en el motor (refresco de cliente sin reiniciar la interfaz).
4. **Given** el core compartido, **When** se ejecuta la suite de conformidad del contrato en CI, **Then** pasa sin errores y falla si una interfaz deja de respetar la superficie contratada (comandos, eventos, funciones de sesión/config).

---

### User Story 2 - TUI con Bubble Tea: chat completo en terminal (Priority: P2)

El usuario abre la terminal, ejecuta el chat y conversa con el asistente de principio a fin sin salir de ella: escribe una pregunta, ve la respuesta en **streaming** con markdown renderizado, usa **comandos slash** (ayuda, limpiar, coste, tokens, compactar, salir…), cambia de **proveedor/modelo** desde la propia TUI, retoma **sesiones anteriores**, referencia archivos con `@`, navega el historial de sus propios inputs con ↑/↓, y ve el estado (modelo, tokens, coste) en la barra de estado. La TUI ya existe sobre Bubble Tea y conserva este conjunto de capacidades, ahora formalizado sobre el contrato compartido; cualquier interrupción (Ctrl+C) o error de red se maneja de forma limpia sin perder la conversación.

**Why this priority**: La TUI es el frontend que ya existe y entrega valor por sí sola: es el slice de mayor retorno inmediato. P2 porque se apoya en el contrato del core (P1), pero es totalmente demostrable de forma independiente antes de que exista la GUI Wails.

**Independent Test**: Ejecutar `letsgo chat` en un terminal 80x24 y completar un ciclo real: enviar mensaje → ver streaming → interrumpir con Ctrl+C → cambiar modelo con Ctrl+S → ejecutar `/help`, `/clear`, `/cost` → retomar una sesión anterior. Entrega un cliente de terminal completo y utilizable sin tocar la GUI.

**Acceptance Scenarios**:

1. **Given** un terminal 80x24 con configuración válida, **When** el usuario escribe un mensaje y pulsa Enter, **Then** la respuesta aparece en streaming incremental con markdown renderizado (glamour), la conversación scrollea con el viewport y el estado vuelve a "listo" al terminar.
2. **Given** el input comenzando con `/`, **When** el usuario escribe o ejecuta un comando slash, **Then** ve sugerencias de autocompletado y los comandos existentes responden (p.ej. `/help`, `/model`, `/clear`, `/cost`, `/tokens`, `/compact`, `/quit`).
3. **Given** la TUI en pleno streaming, **When** el usuario pulsa Ctrl+C, **Then** la respuesta se cancela de forma limpia, la interfaz vuelve al prompt en <2s y la sesión queda consistente (sin mensajes parciales corruptos).
4. **Given** sesiones anteriores guardadas, **When** el usuario inicia la TUI y solicita el historial de sesiones, **Then** puede retomar una sesión existente o crear una nueva, compartiendo la misma base de datos que la GUI.
5. **Given** el selector de modelos (Ctrl+S), **When** el usuario navega entre proveedores y modelos y confirma, **Then** la selección se persiste en la configuración y el motor se refresca con el nuevo modelo sin reiniciar la TUI.

---

### User Story 3 - GUI de escritorio con Wails: reemplazo completo de la GUI Fyne (Priority: P3)

El usuario abre la aplicación de escritorio (ventana nativa construida con Wails) y chatea con el asistente con la misma naturalidad que en la terminal: escribe, ve el **streaming** con indicador de estado y botón de cancelar, ve la actividad de herramientas dentro del chat, y gestiona su **historial de sesiones** (crear, retomar, listar) en el panel lateral. La GUI Wails reemplaza por completo a la GUI Fyne actual: **todos** los flujos que hoy ofrece `internal/gui` (chat, historial de sesiones, paneles Git, Tareas, MCP, Plugins, Uso/Analytics, Configuración, Ayuda, Tema y Cuenta, junto con el rail de navegación, la paleta, los atajos y los estados vacíos/de carga) deben alcanzar paridad funcional en Wails, porque Fyne se **elimina del proyecto**: una vez alcanzada la paridad, `internal/gui` y la dependencia `fyne.io/fyne` se retiran (es más fácil de mantener una única GUI).

**Why this priority**: Wails aún no es una dependencia del proyecto; el esqueleto completo de escritorio (rail de 10 paneles, paleta, atajos, estados) ya existe en Fyne, por lo que el valor incremental está en portar el flujo central (chat + sesiones) primero y completar la paridad después, sin romper lo que ya funciona. P3 porque depende del contrato (P1) y se beneficia de la TUI como referencia de comportamiento (P2).

**Independent Test**: Compilar y lanzar la GUI Wails, completar una conversación completa (streaming, cancelación, ejecución de herramientas), retomar una sesión creada previamente desde la TUI, redimensionar la ventana y comparar el comportamiento de estos flujos contra la GUI Fyne (mismo resultado, sin regresión). Entrega un cliente de escritorio utilizable para el flujo principal.

**Acceptance Scenarios**:

1. **Given** la GUI Wails abierta con un proveedor configurado, **When** el usuario escribe y envía un mensaje, **Then** ve la respuesta en streaming con indicador de estado, puede cancelar en cualquier momento y la conversación queda persistida.
2. **Given** sesiones previas (creadas desde la TUI o la GUI), **When** el usuario abre el historial de sesiones, **Then** puede crear una sesión nueva, listar las existentes y retomar cualquiera con su historial completo.
3. **Given** la ejecución de una herramienta dentro del chat, **When** el motor notifica su actividad/resultado, **Then** la GUI muestra la ejecución y su resultado (éxito/error) sin bloquear la conversación.
4. **Given** un stream activo, **When** el usuario cierra la ventana o pulsa cancelar, **Then** el stream se detiene, no quedan procesos colgados y la próxima apertura recupera la sesión intacta.
5. **Given** la ventana redimensionada, **When** el usuario continúa la conversación, **Then** el layout se adapta (chat, panel de sesiones y composer siguen legibles y operativos).
6. **Given** la paridad funcional alcanzada en los paneles avanzados (Git, Tareas, MCP, Plugins, Uso, Ayuda, Tema, Cuenta, paleta y atajos), **When** el equipo retira la GUI Fyne, **Then** el proyecto compila y la suite de regresión de esos flujos pasa sin la dependencia Fyne.

---

### User Story 4 - Configuración y primer arranque en la GUI Wails (Priority: P3)

El usuario que abre la GUI Wails por primera vez sin ninguna API key configurada no se encuentra con una ventana vacía: la aplicación lo **guía para configurar proveedor y API key** antes de poder conversar. Una vez configurado, el usuario puede volver a **Configuración** desde la propia GUI para cambiar proveedor, modelo y preferencias, y los cambios persisten en la misma configuración que usa la TUI (al volver a la terminal, todo sigue igual).

**Why this priority**: sin configuración la GUI no puede chatear, así que el onboarding y los ajustes son la pieza que cierra el flujo v1 de la GUI (chat + sesiones + configuración). Depende del esqueleto de la US3, pero es demostrable por separado.

**Independent Test**: Borrar la configuración de API keys, abrir la GUI Wails y verificar el flujo de primer arranque (configurar → poder chatear); después cambiar el modelo en Configuración, guardar, y verificar que la TUI y la configuración persistida reflejan el cambio. Entrega la GUI v1 completa y autosuficiente.

**Acceptance Scenarios**:

1. **Given** un primer arranque sin ninguna API key configurada, **When** el usuario abre la GUI Wails, **Then** ve el flujo de configuración de proveedor/API key como pantalla inicial y solo tras guardar una key válida puede acceder al chat.
2. **Given** la configuración existente, **When** el usuario abre Configuración en la GUI y cambia modelo o preferencias y guarda, **Then** los cambios se persisten en la configuración compartida y se aplican al motor sin reiniciar la aplicación.
3. **Given** una API key inválida o sin conexión, **When** el usuario intenta conversar, **Then** ve un mensaje de error claro con orientación (revisar key, conexión, modelo) y la conversación previa permanece intacta.

---

### Edge Cases

- **Redimensionamiento del terminal durante un stream activo**: la TUI re-laya el viewport sin perder el contenido acumulado ni interrumpir la respuesta en curso.
- **Sin conexión o error de red/API**: ambas interfaces muestran un error legible con orientación (revisar key, conexión, modelo) y conservan el historial; el usuario puede reintentar sin perder contexto.
- **Transcripts muy largos**: una conversación de cientos de mensajes se desplaza (scroll) sin degradación perceptible ni bloqueo de la interfaz en ninguna de las dos.
- **Interrupción (Ctrl+C en TUI, cancelar/cerrar en GUI)**: la sesión queda consistente — el texto parcial no se guarda como respuesta final, la interfaz vuelve a un estado usable y no quedan procesos colgados.
- **Primer arranque sin configuración**: la TUI lo orienta con hints (`/settings`, `Ctrl+S`) y la GUI Wails lo lleva al flujo de configuración inicial; ninguna de las dos se queda en blanco.
- **Sesiones compartidas TUI ↔ GUI**: ambas interfaces leen y escriben la misma base de datos; una sesión creada en una es retomable en la otra. La ejecución simultánea de ambas interfaces no es un requisito de v1 (se asume en la sección Assumptions).
- **Retirada de Fyne sin pérdida de funcionalidad**: ningún panel o flujo de la GUI Fyne actual se pierde en la migración; Fyne (código, tests y assets) se elimina al inicio del feature (inmediato) y cada superficie portada a Wails queda cubierta por su test de regresión del motor/servicios + Vitest/Playwright.
- **Terminal sin soporte de color/UTF-8 o de ancho mínimo**: la TUI degrada con elegancia (sin glifos rotos ni layout roto en 80x24 y por debajo).
- **Cambio de modelo/configuración durante una sesión**: el motor se refresca con la nueva configuración sin reiniciar la interfaz y sin perder la conversación abierta.
- **Cierre de la ventana de la GUI con aprobaciones/tareas pendientes**: el cierre rechaza o descarta limpiamente lo pendiente antes de detener el motor (comportamiento heredado del baseline Fyne).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema MUST exponer un único core conversacional (motor, base de datos y configuración) compartido por la TUI y la GUI Wails; ninguna interfaz duplica la lógica de conversación (envío, streaming, persistencia).
- **FR-002**: El contrato de frontend del core (comandos, eventos, funciones de sesión y configuración) MUST estar documentado y cubierto por tests de conformidad que toda interfaz del producto debe pasar.
- **FR-003**: La TUI MUST permitir enviar mensajes y visualizar la respuesta en streaming incremental, con indicador de estado durante la generación y el renderizado markdown del contenido del asistente.
- **FR-004**: La TUI MUST soportar los comandos slash existentes (ayuda, salir, limpiar, modelo, coste, tokens, compactar, contexto, memoria, tareas, etc.) con sugerencias de autocompletado mientras el usuario escribe.
- **FR-005**: La TUI MUST permitir cambiar de proveedor/modelo desde la propia interfaz y persistir la selección en la configuración compartida.
- **FR-006**: La TUI MUST permitir retomar sesiones anteriores y crear sesiones nuevas usando la misma base de datos que la GUI (historial compartido).
- **FR-007**: La TUI MUST manejar la interrupción (Ctrl+C) cancelando el stream en curso y volviendo al prompt sin corromper la sesión persistida.
- **FR-008**: La GUI Wails MUST permitir iniciar y mantener una conversación completa: envío, streaming con indicador de estado, cancelación y visualización de la actividad/resultado de las herramientas dentro del chat.
- **FR-009**: La GUI Wails MUST ofrecer gestión de sesiones: crear, listar y retomar, con el mismo historial que la TUI.
- **FR-010**: La GUI Wails MUST ofrecer una superficie de Configuración (proveedor, API key, modelo y preferencias) que persista en la configuración compartida con la TUI y se aplique al motor sin reiniciar la aplicación.
- **FR-011**: En el primer arranque sin ninguna API key configurada, la GUI Wails MUST guiar al usuario a configurar un proveedor antes de permitirle conversar.
- **FR-012**: Ambas interfaces MUST usar la misma base de datos y la misma configuración: una conversación iniciada en una de ellas está disponible en la otra.
- **FR-013**: El sistema MUST manejar errores de red/API en ambas interfaces con un mensaje claro y orientativo, sin pérdida del historial y permitiendo reintentar.
- **FR-014**: La GUI Wails MUST arrancar como ventana nativa de escritorio y cerrarse limpiamente (sin procesos colgados) incluso con un stream o tarea en curso.
- **FR-015**: La GUI Wails MUST adaptar su layout al redimensionado de la ventana manteniendo el chat, el panel de sesiones y el composer legibles y operativos.
- **FR-016**: El cambio de modelo o configuración MUST reflejarse en el motor (refresco de cliente) sin reiniciar la interfaz, en ambas interfaces.
- **FR-017**: La GUI Wails MUST alcanzar paridad funcional completa con la GUI Fyne actual (chat, sesiones, configuración y todos los paneles del rail: Git, Tareas, MCP, Plugins, Uso/Analytics, Ayuda, Tema, Cuenta, incluidos paleta, atajos y estados vacíos/de carga), sin regresiones en esos flujos.
- **FR-018**: El sistema MUST eliminar la GUI Fyne del proyecto de forma inmediata al inicio del feature (retirada de `internal/gui` y de la dependencia `fyne.io/fyne`) de modo que el proyecto compile y las suites de regresión pasen sin ella; la GUI Wails se ejecuta por el mismo método que Fyne (`letsgo.exe gui`).

### Key Entities *(include if feature involves data)*

- **Sesión (Conversación)**: unidad de conversación persistida en sqlite; atributos como id, nombre, ruta de proyecto y fechas de creación/actualización. Es la entidad compartida: creada o retomada indistintamente desde la TUI o la GUI.
- **Mensaje**: pertenece a una sesión; tiene rol (usuario/asistente/herramienta) y contenido (texto plano o bloques de contenido, incluidos resultados de herramientas). Se persiste en la misma base de datos para ambas interfaces.
- **Configuración de usuario**: proveedores y API keys, modelo activo, preferencias (auto-aprobación, tema, etc.); persistida en la configuración compartida (`internal/config`) y consumida por ambas interfaces.
- **Contrato de frontend (core contract)**: superficie estable y documentada que toda interfaz debe implementar — comandos del motor, eventos de streaming/estado, funciones de sesión y acceso a configuración — acompañada de su suite de tests de conformidad. No es una entidad de datos, pero es un entregable de primer nivel del feature.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: El usuario puede iniciar una conversación desde cualquiera de las dos interfaces en ≤2 pasos (abrir la interfaz → enviar el mensaje).
- **SC-002**: El 100% de las sesiones creadas desde una interfaz es visible y retomable desde la otra, sin pasos adicionales de exportación/importación.
- **SC-003**: La TUI arranca y muestra el prompt en <1s en un terminal 80x24 con configuración existente.
- **SC-004**: Al menos el 90% de los comandos slash disponibles se completan sin ayuda externa, con sugerencias visibles mientras se escribe.
- **SC-005**: En un stream de prueba, el primer token de la respuesta aparece en <3s y el texto avanza de forma continua y fluida (percepción de generación en vivo) en ambas interfaces.
- **SC-006**: La ventana de la GUI Wails se abre y muestra el chat listo para usar en <5s tras lanzar el binario.
- **SC-007**: El usuario completa el flujo "configurar API key → elegir modelo → chatear" en la GUI Wails en ≤4 acciones.
- **SC-008**: Interrumpir un stream (Ctrl+C en la TUI; cancelar o cerrar en la GUI) devuelve la interfaz a un estado usable en ≤2s en el 100% de los casos, sin sesión corrupta ni procesos colgados.
- **SC-009**: Un transcript de ≥500 mensajes se desplaza (scroll) sin degradación perceptible en ambas interfaces.
- **SC-010**: La GUI Wails alcanza paridad funcional completa con la GUI Fyne actual (todos los paneles y flujos) sin regresiones detectadas en la suite de conformidad, y el proyecto compila con la dependencia Fyne ya retirada.
- **SC-011**: Un error de red/API en cualquier interfaz deja el historial intacto y al usuario con una orientación accionable (qué revisar) en el 100% de los casos probados.
- **SC-012**: La retirada de Fyne reduce el mantenimiento observable: `go.mod` sin dependencias fyne y un build de GUI único (una sola carpeta de GUI y un solo conjunto de tests de GUI en el proyecto).

## Assumptions

- Las tecnologías de ambas interfaces están **mandatadas por el usuario**: TUI con Bubble Tea (charmbracelet) y GUI de escritorio con Wails, **reemplazando por completo a Fyne**, que se elimina del proyecto por mantenibilidad (una sola GUI que mantener); no son decisiones de diseño del equipo y quedan fuera del debate de este SPEC.
- **La GUI Fyne actual (`internal/gui`) se retira YA**: no conviven dos GUIs. Queda como referencia de comportamiento solo para portar sus flujos, pero `internal/gui` (código, tests y assets) y la dependencia `fyne.io/fyne` se eliminan de forma **inmediata** al inicio del feature (directiva del usuario 2026-08-13), manteniendo el mismo método de ejecución (`letsgo.exe gui`). La GUI Wails debe cubrir el alcance completo del rail Fyne actual: chat, sesiones, Git, Tareas, MCP, Plugins, Uso/Analytics, Configuración, Ayuda, Tema, Cuenta, paleta, atajos de teclado y estados vacíos/de carga.
- La sincronización TUI ↔ GUI se resuelve compartiendo la misma base de datos y la misma configuración; la ejecución simultánea de ambas interfaces no es un requisito de v1, y la elección de proceso único vs. multi-proceso queda a criterio de la implementación.
- El contrato de frontend se formaliza sobre el motor, la base de datos y la configuración existentes, sin reescribir su comportamiento (tanto la TUI como la GUI Fyne ya consumen este core).
- La TUI conserva su comportamiento de permisos actual (las herramientas se ejecutan sin prompt interactivo, auto-aprobación) mientras no se decida lo contrario.
- La GUI Wails en Windows depende del runtime WebView2; se asume disponible o instalable como requisito de entorno para los usuarios finales.
- El frontend de la GUI Wails usa HTML/CSS/TypeScript estándar del andamiaje de Wails, sin framework obligatorio; la elección concreta queda para la implementación.
- La infraestructura de red avanzada (proxy corporativo, multi-cuenta simultánea) queda fuera del alcance de v1.

# Feature Specification: GUI Overhaul — Codex-style App Shell

**Feature Branch**: `003-gui-overhaul`

**Created**: 2026-08-06

**Status**: Draft

**Input**: User description: "Overhaul total a la GUI, ya que en este momento no se ve nada appealing a la vista. Necesito una apariencia similar a Codex gui pero original, una barra vertical con la configuración, cuentas y demás en la parte inferior de la misma, los demás elementos y todo bien estructurado."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Shell de navegación vertical (rail + panes) (Priority: P1)

El usuario abre LetsGO y encuentra una barra vertical al lado izquierdo que estructura toda la app: arriba el trabajo principal (conversación, sesiones, git, tareas, herramientas/MCP, plugins) y abajo, fijados, el control de la app (uso/cuota, cuenta/proveedores, tema, ajustes, ayuda). Cada vista vive como un panel del área principal — nunca una ventana modal — con la conversación siempre montada detrás. El usuario navega por la barra con el ratón o con el teclado (`Alt+n`, paleta `Ctrl+K` sigue funcionando), y `Esc`/`Alt+Left` lo devuelve a la conversación con el composer enfocado.

**Why this priority**: Es el pedido central ("una barra vertical... bien estructurado") y el cambio más visible. Reescribe la arquitectura de pantalla y la identidad visual; el resto de retoques dependen de él. Transversal, testable de forma aislada con solo invocar cada slot.

**Independent Test**: Abrir la app, seleccionar cada slot del rail desde el teclado y con el ratón, entrar y salir de vistas con `Esc`/`Alt+Left` sin perder la conversación. Entrega la estructura "tipo Codex" solicitada con navegación completa solo con teclado.

**Acceptance Scenarios**:

1. **Given** la app abierta, **When** el usuario selecciona cualquier elemento de la barra (ratón o `Alt+n`), **Then** el panel correspondiente se muestra en el área principal, el elemento queda marcado como activo, y la conversación permanece montada tras él.
2. **Given** una vista secundaria abierta, **When** el usuario pulsa `Esc` o `Alt+Left`, **Then** regresa a la conversación con el composer enfocado, sin abrir diálogos y sin perder el estado de la sesión.
3. **Given** la parte inferior de la barra, **When** el usuario busca cuentas/proveedores, ajustes, uso o tema, **Then** los encuentra anclados siempre en la parte inferior de la barra vertical, visibles sin viajar ni scroll de la barra.

---

### User Story 2 - Aspecto visual rediseñado (tema "Deep Goblue") (Priority: P2)

El usuario percibe una app por fin "appealing": paleta oscura escalonada por niveles (chrome, superficie, tarjetas, hover), acento LetsGO Goblue reservado a momentos firmados, tipografía dual (sans para interfaz, mono para telemetría/código), radios y separadores consistentes, y foco de teclado visible en toda la app. Las tarjetas de aprobación, plan, mensajes y composer se restilan en bloque.

**Why this priority**: El pedido explícito "no se ve nada appealing a la vista" apunta a estética; condiciona la percepción de calidad pero no bloquea funcionalidad. Depende de US1 (montar el nuevo shell) para aplicarse de forma coherente.

**Independent Test**: Comprobaciones visuales (screenshots comparados antes/después a resolución fija) y pruebas headless que verifiquen colores/importancia de componentes clave; confirman que fondo, tarjetas, acento, tipografía y foco radial aplican en toda la app a la vez.

**Acceptance Scenarios**:

1. **Given** la app renderizada, **When** el usuario explora cualquier panel, **Then** observa una paleta oscura con ≥3 niveles de profundidad diferenciados por luminosidad y líneas de separación de 1px, y un único acento LetsGO con colores semánticos solo para estados.
2. **Given** la conversación, **When** el usuario escribe y lee, **Then** los mensajes del usuario y el asistente se distinguen por posición y apariencia sin perder legibilidad, y el composer muestra un anillo de foco visible.
3. **Given** el modo oscuro activo, **When** el usuario cambia el tema claro/oscuro, **Then** toda la paleta se adapta sin romper legibilidad y la preferencia persiste entre reinicios.

---

### User Story 3 - Densidad, atajos y microcopy (P3)

Sobre la base nueva, el usuario accede a velocidad de teclado a todo: la barra responde a `Alt+1..n`, la paleta (Ctrl+K) sigue siendo el nexo universal, el cheat sheet `?` lista los atajos nuevos de la barra, y cada elemento muestra su atajo en el tooltip. La app mantiene su voz original en español en microcopy y estados vacíos.

**Why this priority**: P3 aporta densidad de teclado (atado a SC-009) y refuerza la identidad original, pero el valor de la estructura y el tema se entrega con US1+US2.

**Independent Test**: Suite de smoke de teclado con la barra presente: ejecutar el walk del quickstart (incluido el rail nuevo) 100% con teclado, sin menús, y comprobar que Esc/cancel de stream, aprobación y paleta conservan sus atajos actuales (regresión FR-023).

**Acceptance Scenarios**:

1. **Given** la app con la barra visible, **When** el usuario pulsa `Alt+1..9`, **Then** cada slot de la barra activa su panel y el anillo de foco se mueve con la barra sin atascos de foco.
2. **Given** el cheat sheet `?`, **When** el usuario lo abre, **Then** aparecen listados los atajos de la barra con su categoría, y `Esc` lo cierra.
3. **Given** una sesión con muchas vistas, **When** el usuario cambia de panel rápidamente, **Then** ninguna vista bloquea el scroll/teclado y la barra nunca scrollea (máximo 11 slots fijos).

---

### Edge Cases

- **Colisión de atajos**: `Alt+n` no pisa los atajos existentes (Ctrl+1..9 = sesiones, Ctrl+K paleta, Ctrl+U usage, Ctrl+,, Ctrl+L, Ctrl+Shift+A). Si hay colisión, la barra usa `Alt+` y el cheat sheet documenta la nueva marca.
- **Esc ambiguo**: Esc hace cosas según el contexto (interrumpir stream en composer, cerrar paleta/picker, volver de vista). Regla: quien tiene el foco decide; la vuelta de vista con `Esc` solo cuando el panel activo es una vista secundaria (índice != chat) y ni palette ni picker están abiertos.
- **Tema del SO**: el modo oscuro de la app no debe alternar por la detección del sistema; ambos variantes deben renderizar de forma estable con el tema propio.
- **Vistas ocultas**: al seleccionar una vista, las demás quedan montadas pero ocultas; marca activa solo en el panel visible. Al volver, se muestra la misma.
- **Primer arranque sin sesiones**: el shell se ve igual; el picker de arranque sigue en el stack sin romper la barra.
- **Aprobaciones pendientes**: abrir una vista secundaria no descarta las tarjetas de aprobación pendientes ni bloquea su entrada; al regresar, el transcript muestra el mismo estado.
- **Ventana estrecha**: la barra colapsable (ícono-only 44–56px) mantiene la navegación y el transcript colapsa con gracia sin ocultar el composer ni la franja de estado.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: La app DEBE mostrar una barra vertical (rail) de navegación fija a la izquierda, con dos grupos separados por una línea: arriba el trabajo (Conversación, Sesiones, Git, Tareas, Herramientas/MCP, Plugins) y abajo, anclados (Uso/cuota, Cuenta/Proveedores, Tema, Ajustes, Ayuda).
- **FR-002**: La barra DEBE colapsar a modo ícono (≥44px) y expandirse a modo ícono+etiqueta, con la preferencia persistida por usuario entre reinicios.
- **FR-003**: Cada elemento de la barra DEBE ser seleccionable con el teclado mediante un atajo dedicado (`Alt+1..9`) además del ratón, y DEBE mostrar su atajo en el cheat sheet `?` y en el tooltip.
- **FR-004**: Las vistas (Uso, Ajustes, Git, MCP, Plugins, Tareas) DEBEN mostrarse como paneles dentro del área principal (no modales) de manera que el usuario pueda volver a la conversación con una única acción (`Esc` o `Alt+Left`) con el composer enfocado.
- **FR-005**: El elemento de cuenta/proveedores DEBE estar anclado en el borde inferior de la barra, presentar un avatar/identificador del proveedor activo con estado, y abrir en-el-panel la configuración de cuenta y proveedores.
- **FR-006**: La app DEBE aplicar un tema oscuro de niveles múltiples (fondo, superficie, tarjeta elevada) con acento LetsGO distintivo, colores semánticos (éxito/aviso/error) solo para estados, y separaciones de 1px — en toda la app.
- **FR-007**: La app DEBE ofrecer tema claro y oscuro, con cambio inmediato y persistente; DEBE renderizar de forma estable en ambos modos de tema del sistema sin pestañeo.
- **FR-008**: La tipografía DEBE diferenciar UI (sans) y contenido técnico (mono: status line, código, payloads de herramientas, composer), manteniendo la legibilidad a 11–14px.
- **FR-009**: Las tarjetas de aprobación DEBEN mostrar un borde/barra en acento, nombre de herramienta mono, payload colapsable y botones Aprobar/Rechazar visibles (Enter/Esc) sin volverse modales.
- **FR-010**: Los mensajes del usuario y del asistente DEBEN distinguirse visualmente (posición y/o destacado) y el composer DEBE mostrar foco de teclado visible; la franja de estado permanece como línea de 1px con campos configurables existentes (p. ej., modo siempre visible).
- **FR-011**: La paleta (Ctrl+K) DEBE seguir funcionando desde cualquier panel y ser el acceso universal; el cheat sheet `?` DEBE incluir los atajos de la barra.
- **FR-012**: La app DEBE preservar sin regresión los flujos existentes: aprobación inline ≤2s p90, interrupt con Esc ≤500ms con composer operativo, grants por sesión, steers/queue, fork, picker de arranque, y la precedencia de permisos (deny > session-grant > auto-approve > prompt).

### Key Entities

- **Rail (barra de navegación)**: colección ordenada de slots; cada slot referencia una vista, estado (activa/inactiva), atajo asignado, grupo (superior/inferior) y visibilidad.
- **Vista (panel)**: superficie intercambiable en el área principal (Conversación, Sesiones, Git, Tareas, Herramientas/MCP, Plugins, Uso, Cuenta, Ajustes, Ayuda), con una sola activa a la vez y una pila de navegación (historia).
- **Cuenta/Proveedor**: identidad mostrada en el borde inferior de la barra (proveedor, estado online/idle/error, acceso a configuración de claves y proveedores).
- **Preferencias de apariencia**: tema (claro/oscuro), colapso de la barra, visibilidad; persisten entre reinicios.
- **Vista activa**: índice del panel visible; usada por navegación (rail, Alt+Left, Esc) y por el marcador activo.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: >90% de los elementos clave de la navegación (esp. cuenta/ajustes/uso) son localizables en la barra inferior en <5s de la primera visita sin asistencia (test de descubrimiento).
- **SC-002**: Config → model → tema → vuelta al chat en ≤3 acciones y <20s, sin cerrar la ventana ni perder la conversación (sin regresión de SC-002 previo).
- **SC-003**: El walk oficial de la app se completa 100% solo con teclado (incluida la barra nueva: `Alt+digit`, `Esc`/`Alt+Left`), cero uso de menús o ratón (mantiene SC-009 previo).
- **SC-004**: Media de tiempo para decidir una aprobación ≤2s p90 con transcript visible (sin regresión de SC-001 previo).
- **SC-005**: El interrupt con `Esc` mantiene ≤500ms con el composer operativo tras el rediseño (sin regresión de SC-005 previo).
- **SC-006**: ≥80% de usuarios en un test de descubrimiento de 10 sujetos enumeran la ubicación de cuentas/ajustes sin ayuda tras la primera vista del shell.
- **SC-007**: El transcript es visible/scrolleable durante cualquier aprobación o vista alterna (sin regresión de SC-007 previo).
- **SC-008**: La barra NO scrollea, no supera 11 slots fijos, y el colapso ícono-only (44px) mantiene toda la navegación accesible.
- **SC-009**: La suite de regresión del quickstart previo (SC-002, 003, 004, 005, 006, 007, 009, 010) sigue en verde tras el overhaul; sin tests rotos.

## Assumptions

- La app conserva su identidad LetsGO (acento Goblue `#00ADD8`, dark-preserving); "similar a Codex pero original" = se toma la *arquitectura* de shell (barra vertical, densidad, teclado) y NO los colores/íconos/marca de OpenAI/Codex.
- El modo oscuro se mantiene por defecto y como identidad; el claro es una opción accesible pero no un rediseño separado.
- Las seis vistas secundarias actuales (Settings, Usage, Git, MCP, Plugins, Tasks) se convierten de overlays a paneles intercambiables por el rail; las superficies transitorias (paleta, picker de sesión, cheat sheet `?`, fork) permanecen como overlays.
- La navegación por teclado usa `Alt+1..9` para el rail para no colisionar con `Ctrl+1..9` (sesiones) existentes; `Ctrl+K` sigue siendo el acceso universal.
- El colapso de la barra es global (no por vista) y su preferencia se guarda en configuración.
- El microcopy y los estados vacíos siguen en español (voz original), con labels técnicos en inglés (branch, commit, model) preservados.
- El usuario usa una pantalla de desktop ≥1100px; la ventana por defecto puede ensancharse a ~1280px para acomodar la barra + sesiones + chat.
- Los flujos verificados en 002 (aprobaciones inline, steers/queues, plans, fork, grants, picker) quedan funcionalmente intactos; solo cambia su presentación y ubicación.

**Note on scope**: "Overhaul total" = rediseño estructural + visual del shell (US1+US2) y refuerzo teclado/microcopy (US3). Fuera de alcance: reescritura del motor de conversación, cambios de API/proveedores, nuevas features de agente, análisis de uso.
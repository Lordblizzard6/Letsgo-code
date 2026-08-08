# Feature Specification: UI/UX Refinement — Consistency, Legibility & Menu Architecture

**Feature Branch**: `004-ui-redesign`

**Created**: 2026-08-07

**Status**: Draft

**Input**: User description: "UI/UX holística — la interfaz no se ve consistente ni atractiva, se ve antigua y con elementos dispersos. Se pide confianza, legibilidad, funcionalidad y elegancia. Tiene una barra lateral izquierda vertical con settings y accesos; los menús deben organizarse según como el @UX Researcher sugiere organizarlos."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Rail coherente: zonas, nomenclatura y un destino por slot (Priority: P1)

El usuario abre LetsGO y encuentra la barra lateral izquierda organizada en **tres zonas separadas por hairline**: abajo (anclada al pie) la zona de sistema/identidad (Cuenta·proveedores, Tema, Colapsar), en medio la zona de herramientas con etiqueta de sección (Git · Tareas · MCP · Plugins), y arriba el trabajo principal (Chat). Cada slot apunta a un destino único, con un nombre corto en español, y el estado activo (accent bar) se muestra de forma unívoca — ningún slot es redundante ni comparte panel con otro. En modo colapsado todos los iconos muestran tooltip con nombre y acceso directo (p.ej. "Git (Alt+2)").

**Why this priority**: Es la causa raíz del "desorden": el rail actual mezcla propósitos (un slot "Sessions" apunta al mismo panel que "Chat", impidiendo el estado activo correcto), y tres sistemas de navegación nombran los mismos destinos de forma distinta (menu bar, rail, paleta). Corregir la arquitectura de información devuelve confianza y es la base sobre la que se apoyan los demás retoques.

**Independent Test**: Abrir la app, recorrer cada zona del rail con ratón y con `Alt+1..8` (correctamente re-mapeados), verificar que cada slot resalta como único activo, que no hay duplicados, que el tooltip aparece en modo colapsado y que cada destino se alcanza en ≤2 clics. Entrega una navegación coherente y con nomenclatura completa sin necesidad de retocar el resto de paneles.

**Acceptance Scenarios**:

1. **Given** la app abierta, **When** el usuario recorre el rail, **Then** ve tres zonas separadas por hairline (trabajo / herramientas con etiqueta "HERRAMIENTAS" / sistema anclado abajo) y ningún slot duplica el destino de otro.
2. **Given** el rail colapsado, **When** el usuario pasa el cursor por cualquier icono, **Then** ve un tooltip con el nombre del destino y su acceso directo (p.ej. "Git (Alt+2)").
3. **Given** la app, **When** el usuario navega con teclado por el rail (flechas + Enter), **Then** el foco del rail es siempre visible (focus ring) y cada activación recupera el panel correcto.

---

### User Story 2 - Profundidad correcta: Settings agrupado, Cuenta en flyout, Uso compacto (Priority: P2)

El usuario necesita configurar sin perder de vista la conversación y sin perderse en una sola página extensa. "Configuración" (accesible vía `Ctrl+,` y rail) se organiza en **secciones con selector** (Cuenta · Apariencia · Preferencias · Uso) en lugar de un formulario lineal; la cuenta/proveedores vive en un **flyout del avatar** (estado online/idle/error, provider+model activo, resumen de uso de hoy y acciones: "API keys & providers…", "Uso…", "Cerrar sesión"); "Uso" deja de ser una pared de texto monoespaciado y pasa a ser un **panel de tarjetas KPI** (coste de hoy, tokens, requests; tabla por provider/model; barra de presupuesto). El tema queda como quick-toggle en el pie del rail además de un control en Apariencia. Al cerrar Configuración o el flyout, el foco vuelve al composer.

**Why this priority**: una sola página extensa, un avatar que abre la misma configuración y un panel de uso ilegible son los síntomas "antiguo/disperso" que menciona el usuario; dar profundidad correcta (diálogo con pestañas, popover, dashboard) moderniza el resto sin reescribir funcionalidad. Depende de US1 (misma navegación y nomenclatura).

**Independent Test**: Abrir Settings, recorrer sus 4 secciones, editar valores y Guardar/Cancelar; abrir el flyout de cuenta desde el avatar; comprobar el panel Uso compacto; comprobar que la conversación sigue montada tras el diálogo y que al cerrar cualquier superficie el foco vuelve al composer. Entregable aislado y medible por estado de cada superficie.

**Acceptance Scenarios**:

1. **Given** la app abierta, **When** el usuario pulsa `Ctrl+,` o el slot Configuración, **Then** se abre un diálogo no modal con las secciones Cuenta · Apariencia · Preferencias · Uso, la conversación permanece visible tras él, y `Esc` lo cierra devolviendo el foco al composer.
2. **Given** el avatar de usuario (abajo a la izquierda), **When** el usuario hace clic, **Then** se abre un flyout con estado [online/idle/error], provider+model activos, resumen de uso de hoy y accesos a "API keys & providers…", "Uso…", "Cerrar sesión" — sin reutilizar la superficie de Configuración.
3. **Given** un estado de uso con datos, **When** el usuario abre el panel "Uso", **Then** ve tarjetas KPI (coste de hoy / tokens / requests), tabla por provider/modelo y barra de presupuesto — nunca un volcado monoespacial.

---

### User Story 3 - Legibilidad y elegancia: medida de mensaje, tipografía y consistencia de tarjetas (Priority: P2)

La conversación se lee con comodidad: los mensajes no se estiran a todo el ancho del panel (columna de ~720px centrada), las tarjetas (mensaje, aprobación, plan, panel de herramientas) comparten el mismo marco (border 1px, radius 6px, padding 16px), el espaciado sigue una rejilla base de 4px, y la tipografía usa **una rampa** (12/14/16/20/24 sans + mono solo para código/telemetría). El composer es prominente, con focus ring accent de 2px, y el botón principal intercambia "Enviar" ↔ "Detener" en el mismo lugar durante el streaming. El microcopy se armoniza: chrome en español con términos técnicos en inglés; los placeholders describen la acción (composer: "Pregunta a LetsGO… (Enter enviar · Tab encolar · Esc detener)"; sesiones: "Filtrar conversaciones…").

**Why this priority**: la legibilidad es un pilar explícito del requerimiento y se percibe en cada pantalla; armonizar tarjetas, métricas y texto es lo que convierte "antigua/dispersa" en "elegante y confiable". Depende de US1 (superficies) y de US2 (diálogo de configuración).

**Independent Test**: Inspección visual y checks headless: columna de mensajes ≤720px; mismo marco (border/radius/padding) entre mensajes, tarjetas de aprobación y paneles de herramientas; tokens de la rampa tipográfica usados en todas las superficies; ring de focus en composer; estado hover visible. Medible sin funcionalidad extra.

**Acceptance Scenarios**:

1. **Given** una conversación con mensajes largos, **When** el usuario la visualiza en una ventana amplia, **Then** la prosa no supera ~720px de ancho, los bloques de mensaje se separan con ritmo consistente (24px) y el código permite ancho completo con wrap opcional.
2. **Given** el composer, **When** el usuario lo enfoca, **Then** muestra un ring de focus accent de 2px; durante el streaming el botón principal cambia de "Enviar" a "Detener" en el mismo lugar.
3. **Given** cualquier placeholder o subetiqueta, **When** el usuario la lee, **Then** el chrome está en español, los términos técnicos (branch, commit, model, API key) en inglés, y los placeholders describen la acción correcta (el filtro de sesiones dice "Filtrar conversaciones…", no "Buscar mensajes…").

---

### User Story 4 - Accesibilidad y consistencia técnica: contraste AA, foco e iconos semánticos (Priority: P3)

El tema aplica en ambas variantes (oscuro/claro) sin huecos: ningún color se hardcodea en widgets que también corren en claro (composer ring, avatar), `lightPlaceholder` se oscurece para cumplir WCAG AA 4.5:1, los separadores mantienen ≥3:1, y cada elemento focuseable (rail, entries, listas, tarjeta de aprobación) exhibe un focus ring visible accent 2px en ambas variantes. Los iconos genéricos que engañan (Chat=MailCompose, Git=ViewRefresh) se sustituyen por glifos semánticos propios (burbuja de chat, branch, checklist, servidor, puzzle, gear, sun/moon, gauge) como recursos SVG, y todos los iconográficos tienen tooltip. Las listas vacías (conversaciones, MCP, plugins, tareas, git, uso) muestran mensaje + CTA en lugar de vacío; los refrescos asíncronos mantienen el contenido previo con un indicador "Refrescando…".

**Why this priority**: son las cualidades de madurez visual; no desbloquean funciones nuevas (por eso P3) pero evitan fallas de auditoría de accesibilidad y estados que parecen bugs (lista vacía). Depende de US1-US3 (base de tokens y tarjetas).

**Independent Test**: suite automatizada de contraste (cada token de texto ≥4.5:1 y separadores/iconos ≥3:1 en ambos temas); un test de estado vacío por superficie; un recorrido de todos los focuseables verificando focus visible; verificación de tooltips en modos colapsados. Todo verificable en CI sin pantalla.

**Acceptance Scenarios**:

1. **Given** el tema claro activo, **When** el usuario inspecciona texto secundario, **Then** el contraste es ≥4.5:1 (el placeholder claro pasa del actual ~2.7:1) y los separadores ≥3:1 en ambos temas.
2. **Given** cualquier lista vacía (conversaciones, MCP, plugins, tareas, git, uso), **When** el usuario la abre, **Then** ve un mensaje de estado legible (p.ej. "No hay conversaciones todavía.") con una CTA (p.ej. "Nueva conversación"), no un cuadro en blanco.
3. **Given** un icono en modo colapsado, **When** el usuario recorre con teclado, **Then** el foco es visible (ring accent) y el glifo es semánticamente correspondiente al destino.

---

### Edge Cases

- **¿Qué pasa con el acceso a "Sesiones"?** No se elimina funcionalidad: la lista de conversaciones sigue embebida en el borde izquierdo de la vista Chat (búsqueda, nueva, renombrar, borrar). Solo se retira del rail el slot "Sessions" que compartía el panel 0 con Chat. Si en el futuro se quiere gestión global, se añade un panel "Sesiones" con pane propio (fuera de alcance).
- **¿Cómo se comporta la Configuración durante un stream activo?** El diálogo no desmonta el shell: la conversación permanece montada detrás; al cerrarlo el foco vuelve al composer y el stream continúa.
- **¿Qué ocurre con los atajos al re-mapear el rail?** Se reasignan los `Alt+1..8` a la nueva composición (Chat, Git, Tareas, MCP, Plugins, Configuración, Uso, Ayuda) y el keymap se actualiza en paralelo; paleta `Ctrl+K`, `?`, `Ctrl+,` y `Esc`/`Alt+Left` quedan invariantes.
- **¿Y si un widget hardcodea colores oscuros en tema claro?** Ningún token de color se hardcodea en widgets compartidos; la suite de contraste CI falla si algún token queda bajo umbral en cualquier variante.
- **¿Qué pasa durante la carga asíncrona de listas?** El refresco nunca vacía el panel: mantiene el contenido previo con un indicador sutil "Refrescando…"; el estado vacío solo aparece cuando el resultado definitivo es vacío.
- **¿Y si el modo colapsado pierde contexto del icono?** Los tooltips con nombre + atajo son obligatorios en colapsado; los iconos SVG semánticos admiten fallback al icono Fyne genérico mientras no exista el recurso.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El rail MUST mostrar tres zonas separadas por hairline (trabajo · herramientas con etiqueta de sección · sistema anclado abajo), con nomenclatura en español y un destino único por slot (ningún slot reutiliza el panel de otro).
- **FR-002**: La vista de conversaciones permanece como lista embebida en la vista Chat; el rail no contiene un slot separado hacia ese mismo panel (el slot `sessions` queda fuera del rail).
- **FR-003**: En modo colapsado, cada icono del rail MUST mostrar tooltip con nombre + acceso directo; la navegación por teclado del rail MUST exponer foco visible en ambos modos.
- **FR-004**: "Configuración" MUST abrirse como diálogo con selector/secciones por pestañas (Cuenta · Apariencia · Preferencias · Uso), accesible desde rail y `Ctrl+,`; la conversación NO se desmonta al abrir; al cerrar, el foco vuelve al composer.
- **FR-005**: El avatar/cuenta MUST abrir un flyout con provider/model activo, resumen de uso de hoy y accesos a "API keys & providers…", "Uso…", "Cerrar sesión" — sin reutilizar la superficie de Configuración como destino de cuenta.
- **FR-006**: "Uso" MUST mostrarse como panel de tarjetas KPI (coste de hoy, tokens, requests) con desglose por provider/model y barra de presupuesto cuando aplique; la vista de texto monoespaciado plano queda prohibida como superficie principal del panel.
- **FR-007**: El cambio de tema MUST exponerse como quick-toggle en el pie del rail (icono sol/luna, persistente) y como pestaña Apariencia en Configuración.
- **FR-008**: Microcopy de chrome en español con términos técnicos en inglés; placeholders que describen la acción (composer "Pregunta a LetsGO… (Enter enviar · Tab encolar · Esc detener)"; sesiones "Filtrar conversaciones…").
- **FR-009**: Los mensajes de conversación en prosa MUST limitarse a ~720px de ancho máximo con la columna centrada; el código permite ancho completo con wrap opcional; tarjetas (usuario/asistente, aprobación, plan, herramientas) comparten marco: border 1px, radius 6px, padding 16px.
- **FR-010**: El composer enfocado MUST mostrar ring accent de 2px; el botón principal MUST intercambiar "Enviar"↔"Detener" en el mismo sitio durante el streaming; estados hover, focus y active deben ser distinguibles.
- **FR-011**: Cada lista sin datos (conversaciones, MCP, plugins, tareas, git, uso) MUST mostrar estado vacío con mensaje y CTA; los refrescos asíncronos NO vacían el panel (indicador sutil "Refrescando…").
- **FR-012**: MUST sustituir iconos engañosos por glifos semánticos SVG (Chat burbuja, Git branch, Tareas checklist, MCP servidor, Plugins puzzle, Settings gear, Tema sun/moon, Uso gauge).
- **FR-013**: Todo par de colores MUST cumplir ≥4.5:1 (texto) y ≥3:1 (no-texto/separadores) en AMBAS variantes de tema; ningún widget hardcodea un token de color que lo haga ilegible en la otra variante (composer ring, avatar).

### Key Entities *(include if feature involves data)*

- **RailSlot**: entrada de navegación con id estable (compatible con keymap y paleta), icono semántico, label ES, shortcut y destino/acción; organizable por zona y modo expandido/colapsado. Sin cambios de datos persistidos.
- **SettingsCatalog**: opciones de configuración agrupadas en 4 secciones (Cuenta · Apariencia · Preferencias · Uso) reutilizando la persistencia existente (config.go); separación de estado login vs. apariencia, sin nuevos objetos de datos.
- **UsageSummary**: presentación agregada de la telemetría existente (coste de hoy, tokens, requests, desglose por provider) con representación compacta (KPI + tabla + barra), sin nuevos tipos de datos subyacentes.
- **EmptyState / LoadingState**: plantillas documentadas de estado sin datos (mensaje + CTA) y de refresco conservador (contenido previo + indicador) aplicadas a las listas existentes.

## Success Criteria *(mandatory)*

- **SC-001**: ≥4/5 usuarios novatos localizan un destino del rail en ≤2 clics a la primera (top-5: Chat, Git, Configuración, Uso, Cuenta) sin confundirse con el slot duplicado "Sesiones".
- **SC-002**: El camino de configuración "ajustar modelo → cambiar tema → volver" se completa en ≤3 acciones desde cualquier vista (rail/`Ctrl+,` → pestaña → control → cerrar).
- **SC-003**: El recorrido completo de la nueva interfaz (rail, paneles, paleta `Ctrl+K`, `?`) es 100% realizable con teclado, con foco visible en cada paso y en ambos temas.
- **SC-004**: El panel de uso responde "¿cuánto he gastado hoy?" en ≤2 segundos: tarjetas KPI visibles a primer vistazo, alcanzables en un clic desde avatar o rail.
- **SC-005**: Contraste AA automatizable en CI: todos los pares de texto ≥4.5:1 y no-texto (separadores, iconos, foco) ≥3:1 en ambos temas (palette smoke test).
- **SC-006**: Al menos 6 superficies de lista (conversaciones, MCP, plugins, tareas, git, uso) presentan estado vacío con mensaje + CTA; ninguna se vacía en blanco durante un refresco.
- **SC-007**: Decisiones de aprobación/rechazo en ≤2 s (p90) con el transcript siempre visible (heredado del contrato 003, sin regresión).

## Assumptions

- Los usuarios usan español como lengua principal del chrome de interfaz y aceptan términos técnicos en inglés (branch, commit, model, API key).
- El slot "Sessions" se retira del rail sin pérdida de funcionalidad porque la lista de conversaciones ya está embebida a la izquierda de la vista Chat (coherente con el 003).
- Los iconos SVG semánticos admiten fallback a iconos Fyne genéricos mientras no exista el recurso; el tooltip es obligatorio en colapsado.
- No hay cambios al modelo de datos ni ruptura de contrato con la telemetría/uso existente; cambia la presentación, no el contenido.
- El feature 004 se construye sobre el shell heredado de 003 (rail + tema Deep Goblue, panes no modales), extendiendo su arquitectura sin reescribir el motor de streaming/aprobación.
- Queda fuera de alcance la gestión global de sesiones como panel independiente (se documenta en quickstart como extensión futura).

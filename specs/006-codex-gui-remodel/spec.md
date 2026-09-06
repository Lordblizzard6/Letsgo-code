# Feature Specification: Remodelado visual de la GUI Wails — apariencia CODEX GUI

**Feature Branch**: `006-codex-gui-remodel`

**Created**: 2026-08-14

**Status**: Draft

**Input**: User description: "crea otro spec, usando @UI Designer para remodelar la gui para que se vea como CODEX GUI, no se ve igual en lo absoluto actualmente; revisalo"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - La GUI se ve y se siente como CODEX GUI (Priority: P1)

El usuario abre `letsgo gui` y, a primera vista, la ventana le recuerda a la GUI
desktop de **OpenAI Codex**: mismo lenguaje visual (chrome casi monocromo, superficies
que escalan por luminancia, tipografía de lectura, acento sobrio y único), misma
anatomía de chat (transcript como lienzo con burbuja de usuario a la derecha y
bloques de asistente full-width sin borde; bloques de código como objetos elevados
con header + botón copiar + resaltado de sintaxis), mismo rail lateral fino con
iconografía lineal SVG y marcador activo único, y el composer como "command shell"
anclado abajo que sigue editable durante el streaming (Send ⇄ Stop). El usuario
actual percibe que la app actual (GitHub-style: tarjetas por mensaje, acento azul
`#2f81f7`, iconos emoji, pills flotantes) **no se parece en absoluto a Codex**, y
este spec exige un rediseño visual 1:1 de la filosofía Codex manteniendo toda la
funcionalidad ya implementada (chat streaming, sesiones, herramientas, configuración,
paleta, atajos, onboarding).

**Why this priority**: es el único objetivo del spec. El usuario ya declaró que la
iteración previa no alcanzó el parecido ("no mms no se ve igual en lo absoluto
actualmente"), así que el rediseño visual es la prioridad única y el resto son
guardarraíles de regresión.

**Independent Test**: Abrir `letsgo gui` y comparar contra una captura de la GUI de
Codex (referencia visual): a 8ft de distancia la composición general (rail fino,
status line, transcript sin tarjetas, composer docked) es la misma; no hay emoji en
el chrome; los mensajes del asistente no tienen borde de tarjeta; los bloques de
código tienen header con lenguaje y botón copiar; dark-first con acento único.
Cobertura automatizada: smoke visual Playwright + gates G1–G10 (gui-contract §8)
pasando.

**Acceptance Scenarios**:

1. **Given** la GUI Wails abierta con una conversación, **When** el usuario mira la
   ventana, **Then** el transcript es el lienzo: mensajes de usuario como burbuja
   compacta y mensajes de asistente como bloques full-width **sin borde de tarjeta**,
   separados por ritmo tipográfico, no por cajas.
2. **Given** un mensaje con código, **When** se renderiza, **Then** el bloque de
   código es un objeto elevado con header (lenguaje + botón copiar) y resaltado de
   sintaxis, sin `<pre>` desnudo.
3. **Given** el rail lateral, **When** el usuario navega, **Then** los iconos son SVG
   lineales monocromos (cero emoji), el slot activo tiene un único marcador de acento
   y el rail colapsa a iconos con tooltip `Label Alt+n`.
4. **Given** el streaming activo, **When** el usuario sigue escribiendo, **Then** el
   composer permanece anclado y editable, con el botón en la misma ranura cambiando
   a "Stop" y un indicador de cursor de escritura en el bloque en curso.
5. **Given** la ventana en ambos temas, **When** se alterna dark/light, **Then** ambos
   cumplen WCAG AA (texto ≥4.5:1, no-texto ≥3:1) usando el mismo juego de tokens.
6. **Given** el estado actual de la funcionalidad, **When** se aplica el rediseño,
   **Then** ninguna funcionalidad se pierde: chat, sesiones, herramientas, paleta
   Ctrl+K, atajos Alt+1..9, overlays y onboarding siguen operativos (regresión
   UX-12..UX-14).

---

### User Story 2 - Consistencia de estados, foco y feedback (Priority: P2)

El usuario, al interactuar con la GUI rediseñada, encuentra el mismo lenguaje de
estados en todas las superficies: skeletons de carga (no spinners genéricos),
estados vacíos con CTA, errores con orientación y retry, foco visible en ambos temas,
motion sutil (≤150ms) y `prefers-reduced-motion` respetado. La paleta Ctrl+K es
navegable por teclado (↑/↓ + Enter) con fila activa resaltada y restauración de foco
al composer.

**Why this priority**: la filosofía Codex/Qwen es *keyboard-first* y *calm states*;
sin esto el parecido es solo superficial. P2 porque complementa al P1 sin bloquearlo.

**Independent Test**: Recorrer el 100% de los controles interactivos solo por
teclado (Tab / ↑↓ / Enter / Esc / Alt+n / Ctrl+K) con foco visible en dark y light;
verificar skeletons, empty-states y error-banner en las 7+ superficies de datos;
verificar `aria-live`/`aria-busy` en streaming y ausencia de animaciones bajo
`prefers-reduced-motion` (Playwright con emulación).

**Acceptance Scenarios**:

1. **Given** cualquier superficie de datos, **When** se carga sin contenido previo,
   **Then** se muestra un skeleton de carga (nunca un spinner gigante) y un estado
   vacío con CTA cuando aplica.
2. **Given** un error de red/API, **When** ocurre, **Then** aparece un banner con
   orientación accionable y el botón Retry reproduce el último turno de usuario; el
   historial permanece intacto (FR-013).
3. **Given** la paleta abierta, **When** el usuario navega con ↑/↓ y pulsa Enter,
   **Then** la fila activa se resalta con acento y la selección cierra la paleta
   devolviendo el foco al composer.
4. **Given** cualquier overlay (Settings/Usage/Help), **When** se pulsa Esc o el
   backdrop, **Then** el overlay se cierra sin oscurecer el chat subyacente y el foco
   vuelve al composer.

---

### Edge Cases

- **Redimensionado extremo (ventana 800×600)**: el rail colapsa a iconos, el
  transcript mantiene columna legible y el composer no se sale del viewport.
- **Transcript ≥500 mensajes**: el render sigue fluido (memoización del último
  bloque; sin re-parseo O(n²) del markdown acumulado), SC-009.
- **Código con líneas largas**: overflow horizontal dentro del bloque (nunca scroll
  de página); botón copiar disponible.
- **Stream interrumpido/error**: cursor de escritura desaparece, banner con Retry que
  reproduce el último turno, historial intacto.
- **prefers-reduced-motion activo**: sin pulse/shimmer/cursor blink.
- **Emoji solo en contenido del mensaje**: permitido en el markdown del usuario/
  asistente, prohibido en el chrome (G5).
- **Acento: contraste AA** en ambas themes (incluido texto sobre botones de acento).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: La GUI MUST verse como la GUI desktop de OpenAI Codex: chrome casi
  monocromo, superficies por luminancia, tipografía de lectura 13–14px con jerarquía
  por tamaño/peso, un solo acento restringido y dark-first.
- **FR-002**: El transcript MUST usar burbuja de usuario (derecha) + bloque de
  asistente full-width **sin borde de tarjeta**; separación por ritmo tipográfico.
- **FR-003**: Los bloques de código MUST ser objetos elevados con header (lenguaje +
  botón copiar on hover/focus) y resaltado de sintaxis; ningún `<pre>` desnudo.
- **FR-004**: El rail MUST usar iconografía SVG lineal monocroma (cero emoji),
  marcador activo único, y colapsar 48↔56px con tooltip `Label Alt+n`.
- **FR-005**: El composer MUST ser un "command shell" docked y anclado, editable
  durante el streaming, con Send ⇄ Stop en la misma ranura y affordance
  "Streaming — Enter stops".
- **FR-006**: La actividad de herramientas MUST mostrarse inline en el turno como
  cards colapsables agrupadas (5+ → grupo compacto), nunca como log colgante que
  empuje al composer.
- **FR-007**: Debe haber una status line inferior persistente (proveedor · modelo ·
  uso hoy · dot busy) que absorba el estado "Ready/Streaming" y elimine las pills
  flotantes.
- **FR-008**: La paleta Ctrl+K MUST ser navegable por ↑/↓ + Enter con fila activa
  resaltada, agrupada, y restaurar foco al composer al cerrar.
- **FR-009**: Los overlays Settings/Usage/Help MUST no oscurecer el chat (drawer/
  panel) y cerrarse con Esc/backdrop devolviendo el foco al composer.
- **FR-010**: Las 7+ superficies de datos MUST definir empty + loading (skeleton) +
  error con CTA; el chat define streaming/error/empty/idle.
- **FR-011**: Ambos temas MUST cumplir WCAG AA con el mismo juego de tokens; focus
  visible en todo interactivo; motion ≤150ms y `prefers-reduced-motion` honrado.
- **FR-012**: El rediseño MUST ser **solo frontend** (Svelte + app.css + store para
  dedupe de tool rows): cero cambios en bindings, servicios bound, contrato de
  eventos, rutas hash, atajos y motor (UX-14).
- **FR-013**: La regresión MUST preservar la funcionalidad existente: tests vitest
  pre-existentes pasan sin cambios, e2e `onboarding.spec.ts` intacto, y la demo
  funcional (chat, sesiones, herramientas, configuración, onboarding) no regresiona.

### Key Entities

- **Tokens de diseño (CSS custom properties)**: la única fuente de color/tipografía/
  espaciado/radio/sombra; viven en `app.css` bajo `html[data-theme="dark"|"light"]`;
  los componentes consumen `var(--token)` y nunca hex literal (G1).
- **Componentes primitivos**: button, card, input, code-block, tool-card, empty-state,
  skeleton, chip, badge, kpi, status-dot, palette, overlay, flyout — el inventario
  común del que se construye toda superficie (G6).
- **Estado de la UI (presentación)**: derivado de los eventos del contrato
  (`chat:*`, `session:*`, `tool:*`, `theme:changed`, `config:changed`); no se añade
  estado de dominio al frontend (constitución III).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Una captura de la GUI con una conversación de ejemplo, comparada con la
  referencia Codex, coincide en composición general y en los 10 puntos del checklist
  G10 (smoke visual Playwright) sin fallos de token.
- **SC-002**: `rg "#[0-9a-fA-F]{3,8}" src/ --glob '*.svelte' --glob '*.ts'` devuelve
  0 fuera de `app.css` (G1 token purity).
- **SC-003**: Todo interactivo es operable y focuseable por teclado en dark+light
  (G3); flujo Ctrl+K ⇄ palette ⇄ Enter/arrows ⇄ Esc ⇄ composer completo con foco
  visible.
- **SC-004**: 10/10 superficies del rail definen estados empty + loading + error; el
  chat define streaming/error/empty/idle (G7); tool rows sin duplicados (1/call).
- **SC-005**: Contraste AA verificado en ambos temas (texto ≥4.5:1, no-texto ≥3:1) y
  focus ring ≥3:1 (G2); smoke Playwright a 100/125/150% de escala.
- **SC-006**: Todos los `<pre>` se renderizan vía CodeBlock (header+lang+copy+highlight)
  (G8); cero emoji en chrome (G5); spacing en escala 4px y chat ≤760px (G4).
- **SC-007**: Regresión estabilizada: los tests vitest pre-existentes pasan sin
  cambios, `e2e/onboarding.spec.ts` intacto y `git diff` (fuera de `specs/`) solo
  toca `cmd/wails/frontend/**` (UX-12..UX-14).
- **SC-008**: La demo manual (J6 en quickstart) completa: key → modelo → chat (≤4
  acciones), tool inline, Ctrl+K, Esc overlays, sesión de la TUI retomada en 1 clic,
  tema alterno con AA.

## Assumptions

- El rediseño es **solo presentacional** (frontend Wails/Svelte + `app.css`); el
  motor, los servicios bound, el contrato de eventos, las rutas hash y los atajos
  Alt+1..9 no cambian (guardarraíl UX-14).
- La funcionalidad completa de la GUI actual (US3/US4 del feature 005) está
  implementada y es la base sobre la que se aplica el rediseño; este spec NO agrega
  superficies nuevas (excepto la status line y el CodeBlock como componentes visuales).
- Se conservan los strings que bloquean los tests existentes ("Welcome to LetsGO",
  "Save & start chatting", "API Key", placeholder del composer) salvo cambio
  deliberado documentado en el mismo commit.
- La referencia visual de Codex GUI es el documento de diseño canónico (research.md
  D10); se usa con fines de estilo (OpenAI no es propietaria de los tokens concretos
  aquí elegidos).
- El acento elegido es único e indigo-violeta (familia Qwen/Codex, research D8);
  intercambiable solo en tokens si el equipo lo decide después.
- El proyecto conserva el patrón de testing existente: Vitest + jsdom + Playwright
  contra el dev server, con mock del runtime Wails vía `page.route` (sin backend ni
  red).

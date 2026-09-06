# Feature Specification: Modernización de la GUI — Layout de3 columnas basado en gui-shell

**Feature Branch**: `007-gui-modernization`

**Created**: 2026-08-27

**Status**: Draft

**Input**: User description: "crea un spec para modernizar la interfaz en base a la que deje en la carpeta ejemplo de gui con el @UI Designer"

**Reference Design**: `ejemplo de gui/gui-shell/` (React19 + Vite7, cáscara visual autocontenida)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Layout de3 columnas con sidebar, chat y activity pane (Priority: P1)

El usuario abre `letsgo gui` y encuentra una interfaz de escritorio clásica de3 columnas:
- **Sidebar izquierda** (260px, colapsable): proyectos recientes + lista de sesiones con badges
- **Chat central** (flex): barra de pestañas + viewport de mensajes + composer con attach/send/stop
- **Activity pane derecho** (320px): explorador de archivos + panel de actividad de herramientas

La interfaz actual (rail de48px + status line inferior) se reemplaza por este layout de3 columnas
inspirado en el diseño de referencia `gui-shell`. El header contiene: logo/marca, selector de proyecto,
selector de cuenta, selector de modelo, chip de uso, botón explorador, botón settings, botón palette.

**Why this priority**: es el objetivo principal del spec — modernizar el layout para parecerse al
diseño de referencia. Sin este cambio no hay modernización.

**Independent Test**: Abrir `letsgo gui` y verificar que el layout tiene3 columnas visibles (sidebar +
chat + activity pane), header con los controles, y que la interfaz es navegable por teclado.

**Acceptance Scenarios**:

1. **Given** la GUI abierta, **When** el usuario mira la ventana, **Then** hay3 columnas: sidebar
   (260px), chat (flex), activity pane (320px), con un header fijo de50px arriba.
2. **Given** el sidebar, **When** el usuario hace clic en el botón menú, **Then** el sidebar colapsa
   a0 px y el chat se expande; al hacer clic de nuevo, el sidebar se restaura a260px.
3. **Given** el sidebar, **When** el usuario ve la lista de sesiones, **Then** cada sesión tiene un
   badge (test/skills) y una fecha de actualización.
4. **Given** el chat central, **When** el usuario tiene múltiples conversaciones, **Then** puede
   alternar entre pestañas en la barra de tabs.
5. **Given** el activity pane, **When** hay herramientas en ejecución, **Then** se muestran como
   tarjetas con estado (running/done/error) y el nombre de la herramienta en mono.
6. **Given** el header, **When** el usuario hace clic en el selector de cuenta, **Then** se abre un
   dropdown con las cuentas disponibles; lo mismo para el selector de modelo.

---

### User Story 2 - Composer con attach, mode selector y command palette (Priority: P1)

El composer (área de entrada de texto) se moderniza para incluir:
- Botón de adjuntar archivos (paperclip icon)
- Textarea auto-expansible (field-sizing: content)
- Botón Send (gradiente púrpura) / Stop (gradiente rojo)
- Mode selector (code/architecture/planning/research) con dropdown
- Command palette (/commands) con filtrado por teclado
- Hint row con atajos y estado de adjuntos

**Why this priority**: el composer es el componente más interactivo; su diseño afecta directamente
la experiencia del usuario.

**Independent Test**: Verificar que el composer tiene attach, send/stop, mode selector, y que
el command palette se abre con "/" y se navega con ↑/↓.

**Acceptance Scenarios**:

1. **Given** el composer vacío, **When** el usuario escribe texto, **Then** el textarea crece
   automáticamente (field-sizing: content) hasta un max-height de160px.
2. **Given** el composer, **When** el usuario hace clic en el botón paperclip, **Then** se abre
   el selector de archivos y los archivos adjuntos aparecen como chips debajo del textarea.
3. **Given** el composer, **When** el usuario escribe "/" , **Then** aparece el command palette
   con los comandos disponibles filtrados; ↑/↓ navegan y Enter selecciona.
4. **Given** el streaming activo, **When** el usuario hace clic en Stop, **Then** el streaming
   se cancela y el botón vuelve a Send.
5. **Given** el mode selector, **When** el usuario hace clic, **Then** se abre un dropdown con
   los4 modos (code/architecture/planning/research) y al seleccionar uno se cierra el dropdown.

---

### User Story 3 - Header con selectors y usage chip (Priority: P1)

El header de50px contiene de izquierda a derecha:
- Botón menú (toggle sidebar)
- Logo + marca ("LetsGO")
- Selector de proyecto (folder icon + nombre)
- Selector de cuenta (dropdown)
- Selector de modelo (dropdown con provider + model)
- Chip de uso (tokens + costo)
- Botón explorador (folder icon)
- Botón settings (gear icon)
- Botón palette (Ctrl+K)

**Why this priority**: el header es el punto de entrada principal para configuración y navegación.

**Independent Test**: Verificar que cada botón del header funciona y que los dropdowns se abren/cierran
correctamente.

**Acceptance Scenarios**:

1. **Given** el header, **When** el usuario hace clic en el selector de cuenta, **Then** se abre
   un dropdown con las cuentas disponibles y al seleccionar una se cierra el dropdown.
2. **Given** el header, **When** el usuario hace clic en el selector de modelo, **Then** se abre
   un dropdown con los modelos disponibles agrupados por cuenta.
3. **Given** el header, **When** el usuario hace clic en el chip de uso, **Then** se abre la vista
   de gasto (SpendView) como modal.
4. **Given** el header, **When** el usuario hace clic en el botón settings, **Then** se abre el
   modal de configuración.
5. **Given** el header, **When** el usuario presiona Ctrl+K, **Then** se abre la paleta de
   comandos global.

---

### User Story 4 - Paleta de comandos global Ctrl+K (Priority: P2)

La paleta de comandos se abre con Ctrl+K y muestra:
- Input de búsqueda con placeholder
- Items agrupados por categoría (Acciones, Proyectos, Pestañas)
- Fila activa resaltada con accent-soft
- Navegación ↑/↓ + Enter para ejecutar + Esc para cerrar
- Foco restaurado al composer al cerrar

**Why this priority**: la paleta es la forma más rápida de acceder a funciones.

**Independent Test**: Verificar que Ctrl+K abre la paleta, ↑/↓ navegan, Enter ejecuta, Esc cierra.

**Acceptance Scenarios**:

1. **Given** la paleta cerrada, **When** el usuario presiona Ctrl+K, **Then** se abre la paleta
   con el input enfocado y los items listados.
2. **Given** la paleta abierta, **When** el usuario escribe texto, **Then** los items se filtran
   por coincidencia de subcadena.
3. **Given** la paleta abierta, **When** el usuario presiona ↑/↓, **Then** la fila activa se
   mueve y se resalta con accent-soft.
4. **Given** la paleta abierta, **When** el usuario presiona Enter, **Then** se ejecuta la acción
   seleccionada y se cierra la paleta.
5. **Given** la paleta abierta, **When** el usuario presiona Esc, **Then** se cierra la paleta
   y el foco vuelve al composer.

---

### User Story 5 - Mensajes y herramientas con diseño consistente (Priority: P2)

Los mensajes del chat siguen el diseño de referencia:
- User: burbuja derecha con accent-soft bg y accent-border
- Assistant: bloque full-width con panel bg, avatar gradiente púrpura, nombre en bold
- Tool: línea inline con icono de estado (spinner/check/cross) y nombre en mono
- Markdown: renderizado con code blocks elevados (header + copy + highlight)
- Streaming: cursor ▍ parpadeante al final del bloque en curso

**Why this priority**: la consistencia visual de los mensajes es clave para el parecido con el
diseño de referencia.

**Independent Test**: Verificar que los mensajes user/assistant/tool se renderizan correctamente
y que el streaming muestra el cursor.

**Acceptance Scenarios**:

1. **Given** un mensaje de usuario, **When** se renderiza, **Then** aparece como burbuja derecha
   con accent-soft bg y accent-border, sin avatar.
2. **Given** un mensaje del asistente, **When** se renderiza, **Then** aparece como bloque
   full-width con avatar gradiente púrpura, nombre "LetsGO" y contenido markdown.
3. **Given** una herramienta en ejecución, **When** se renderiza, **Then** aparece como línea
   con spinner animado y nombre en mono; al completar, el spinner se reemplaza por check/cross.
4. **Given** un bloque de código, **When** se renderiza, **Then** tiene header con lenguaje y
   botón copy, fondo code-bg y mono font.
5. **Given** streaming activo, **When** el asistente está escribiendo, **Then** aparece el cursor
   ▍ parpadeante al final del bloque.

---

### User Story 6 - Modales settings y spend (Priority: P2)

Los modales siguen el diseño de referencia:
- Settings: overlay oscuro + modal centrado (680px) con header, tabs, body scrollable
- Spend: overlay oscuro + modal centrado con gráfico de barras por día y tabla de entries
- Ambos se cierran con Esc o clic en el overlay
- Focus trap dentro del modal

**Why this priority**: los modales son parte integral de la experiencia.

**Independent Test**: Verificar que settings y spend se abren/cierran correctamente y que el
focus trap funciona.

**Acceptance Scenarios**:

1. **Given** settings abierto, **When** el usuario presiona Esc, **Then** se cierra el modal
   y el foco vuelve al botón que lo abrió.
2. **Given** spend abierto, **When** el usuario cambia el rango (7/30/90 días), **Then** el
   gráfico se actualiza con los datos filtrados.
3. **Given** cualquier modal abierto, **When** el usuario hace clic en el overlay, **Then**
   se cierra el modal.

---

### User Story 7 - Sidebar con proyectos recientes y sesiones (Priority: P2)

El sidebar contiene:
- Sección "Recientes" con proyectos recientes (folder icon + nombre + delete button)
- Sección "Sesiones" con lista de sesiones (título + badge + fecha + delete button)
- Empty states accionables cuando no hay contenido
- Colapso a0 px con animación de220ms

**Why this priority**: el sidebar es la navegación principal entre proyectos y sesiones.

**Independent Test**: Verificar que el sidebar muestra proyectos y sesiones, y que el colapso funciona.

**Acceptance Scenarios**:

1. **Given** el sidebar con proyectos, **When** el usuario hace clic en uno, **Then** se
   selecciona el proyecto y se cierra el menú de proyecto.
2. **Given** el sidebar con sesiones, **When** el usuario hace clic en una, **Then** se abre
   la sesión en una nueva pestaña.
3. **Given** el sidebar vacío, **When** no hay proyectos recientes, **Then** se muestra un
   empty state con texto "No hay proyectos recientes".
4. **Given** el sidebar expandido, **When** el usuario hace clic en el botón menú, **Then**
   el sidebar colapsa a0 px con animación de220ms.

---

### User Story 8 - Estados vacíos, loading y error (Priority: P3)

Cada superficie de datos tiene:
- Empty state con icono + título + CTA accionable
- Loading state con skeleton (shimmer ≤150ms, off bajo reduced-motion)
- Error state con banner + orientación + retry
- Thinking dots (3 puntos parpadeantes) cuando el asistente está procesando

**Why this priority**: los estados de UI son clave para la experiencia de usuario.

**Independent Test**: Verificar que cada superficie muestra empty/loading/error correctamente.

**Acceptance Scenarios**:

1. **Given** una superficie sin datos, **When** se carga, **Then** se muestra un empty state
   con icono SVG + título + CTA.
2. **Given** una superficie cargando, **When** se carga, **Then** se muestra un skeleton con
   shimmer que preserva el contenido previo si existe.
3. **Given** un error de red/API, **When** ocurre, **Then** aparece un banner con orientación
   y botón retry.
4. **Given** el asistente procesando, **When** no hay streaming ni herramientas, **Then** se
   muestran thinking dots (3 puntos parpadeantes).

---

## Cross-cutting Concerns

### G1 — Design Tokens (CSS Custom Properties)

Todos los colores, espaciados, tipografías y efectos se definen como CSS custom properties en
`app.css` con namespace `--`. Dark y light themes se implementan con `[data-theme="dark"]` y
`[data-theme="light"]`. Referencia: `ejemplo de gui/gui-shell/src/App.css:1-80`.

### G2 — Iconografía SVG

Todos los iconos son SVG inline con stroke="currentColor" strokeWidth="2". Cero emoji en chrome.
Referencia: `ejemplo de gui/gui-shell/src/App.tsx:128-290` (PaperclipIcon, SendIcon, StopIcon, etc.).

### G3 — Tipografía

Font family: Nunito (400,600,700) + fallback system. Mono: JetBrains Mono / Consolas.
Base size:14px. Referencia: `ejemplo de gui/gui-shell/src/style.css`.

### G4 — Accesibilidad

- Focus visible: `outline: 2px solid var(--accent); outline-offset: 2px`
- Aria labels en todos los botones interactivos
- Focus trap en modales
- Keyboard navigation: Tab, ↑/↓, Enter, Esc, Ctrl+K
- prefers-reduced-motion: desactivar animaciones

### G5 — Performance

- Transiciones ≤150ms
- Lazy loading de react-markdown y rehype-highlight
- Memoización de MessageRow
- Scroll-to-bottom automático con stick detection

### G6 — Multi-idioma

i18n con keys en formato `section.key`. Soporte para español e inglés.
Referencia: `ejemplo de gui/gui-shell/src/i18n.ts`.

## Technical Context

**Stack actual**:
- Frontend: Svelte5 (runes mode) + Vite + TypeScript
- Backend: Go (Wails v2)
- Tests: Vitest + Playwright
- Build: `npm run build` + `go build`

**Stack de referencia**:
- Frontend: React19 + Vite7 + TypeScript
- Markdown: react-markdown + rehype-highlight + remark-gfm
- i18n: custom (t() function)

**Decisión técnica**: mantener Svelte5 como framework (no migrar a React). Adaptar el diseño
visual y layout del gui-shell a componentes Svelte5 con runes mode ($props, $state, $derived, $effect).

**Dependencias**:
- highlight.js (ya instalado en Spec 006)
- Iconos SVG (ya definidos en `lib/icons.ts` del Spec 006)
- Design tokens (ya definidos en `app.css` del Spec 006)

## Constitution Check

No existe constitution.md. Se aplican las siguientes restricciones del Spec 006 como constitution:

1. **Cero cambios a bindings Go** — el contrato de eventos y rutas hash se mantiene intacto
2. **Cero emoji en chrome** — solo SVG inline para iconografía
3. **Dark-first** — el tema oscuro es el primario; light es el secundario
4. **WCAG AA** — texto ≥4.5:1, no-texto ≥3:1
5. **prefers-reduced-motion** — respetado en todas las animaciones
6. **Transiciones ≤150ms** — motion sutil, no intrusivo
7. **Test-Last** — tests después de implementación en cada fase
8. **Strings bloqueados** — "Welcome to LetsGO", "Save & start chatting", "API Key", placeholder "Type your message..."

## Gates

- **G1**: Todos los colores en CSS custom properties (cero hex fuera de app.css)
- **G2**: Cero emoji en chrome (solo SVG inline)
- **G3**: Focus visible en ambos temas
- **G4**: Transiciones ≤150ms
- **G5**: prefers-reduced-motion respetado
- **G6**: Vitest verde sin cambios en tests pre-existentes
- **G7**: svelte-check 0 errores
- **G8**: npm run build exitoso
- **G9**: Playwright e2e intacto
- **G10**: Todos los modales accesibles (aria-modal, focus trap, Esc close)

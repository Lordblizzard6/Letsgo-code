# Research: Modernización de la GUI — Decisiones de Diseño

**Feature Branch**: `007-gui-modernization`

**Created**: 2026-08-27

## Decision1: Framework — Mantener Svelte5 (no migrar a React)

**Decision**: Mantener Svelte5 (runes mode) como framework del frontend.

**Rationale**: El proyecto ya tiene un frontend Svelte5 funcional con Wails v2 backend.
Migrar a React requeriría reescribir todos los componentes, cambiar el build pipeline,
y potencialmente romper la integración con Wails. El diseño visual del gui-shell se puede
adaptar perfectamente a Svelte5.

**Alternatives considered**:
- Migrar a React19: más trabajo, riesgo de romper integración con Wails
- Migrar a Vue3: mismo problema que React
- Mantener Svelte4: no tiene runes mode, menos features

---

## Decision2: Layout — De Rail a3 Columnas

**Decision**: Cambiar de layout rail (48px) a layout de3 columnas (sidebar260px + chat flex + activity pane320px).

**Rationale**: El diseño de referencia (gui-shell) usa layout de3 columnas. Este layout es
más tradicional de desktop apps y permite mostrar más información simultáneamente (proyectos,
sesiones, actividad de herramientas). El rail de48px es demasiado compacto para mostrar
toda la información requerida.

**Alternatives considered**:
- Mantener rail + drawer: no coincide con el diseño de referencia
- Layout de2 columnas (sidebar + chat): pierde el activity pane
- Layout de4 columnas: demasiado complejo para el espacio disponible

---

## Decision3: Design Tokens — Migrar a CSS Custom Properties del gui-shell

**Decision**: Reemplazar los design tokens actuales por los del gui-shell (dark: `--bg: #0a0e15`, `--accent: #7c5cff`; light: `--bg: #f4f6f9`, `--accent: #6a4bef`).

**Rationale**: Los tokens del gui-shell definen la paleta de colores completa del diseño de
referencia. Usar los mismos tokens garantiza consistencia visual. Se mantienen alias legacy
durante la transición para no romper componentes existentes.

**Alternatives considered**:
- Mantener tokens actuales: no coincide con el diseño de referencia
- Crear tokens híbridos: complejo de mantener
- Usar tokens de Tailwind: dependencia adicional innecesaria

---

## Decision4: Tipografía — Nunito + JetBrains Mono

**Decision**: Usar Nunito (400,600,700) como font family principal y JetBrains Mono / Consolas como mono font.

**Rationale**: El gui-shell usa Nunito como font family principal. Es una font legible y
amigable que funciona bien en interfaces de escritorio. Para código, JetBrains Mono ofrece
buena legibilidad y soporte de ligatures.

**Alternatives considered**:
- Inter: más común en web apps, pero menos "amigable"
- SF Pro: solo disponible en macOS
- Fira Code: buena para código, pero Nunito ya incluye mono fallback

---

## Decision5: Iconografía — SVG Inline con Stroke

**Decision**: Todos los iconos son SVG inline con `stroke="currentColor"` y `strokeWidth="2"`. Cero emoji en chrome.

**Rationale**: El gui-shell define sus iconos como componentes React con SVG inline. Este
patrón se adapta fácilmente a Svelte5. Los SVG inline permiten controlar el color vía
CSS (currentColor) y son más ligeros que icon fonts.

**Alternatives considered**:
- Icon fonts (FontAwesome, Material Icons): dependencia adicional, menos control
- SVG sprites: más complejo de mantener
- Emoji: no accesible, inconsistente entre plataformas

---

## Decision6: Composer — Auto-expansible con field-sizing

**Decision**: El textarea del composer usa `field-sizing: content` para auto-expandirse
verticalmente hasta un max-height de160px.

**Rationale**: El gui-shell usa `field-sizing: content` (CSS Feature) que permite que el
textarea crezca automáticamente con el contenido. Esto mejora la experiencia del usuario
al escribir mensajes largos. El max-height previene que el textarea ocupe demasiado espacio.

**Alternatives considered**:
- JavaScript auto-resize: más complejo, requiere event listeners
- Fijo con scroll: peor experiencia de usuario
- contenteditable: más complejo, problemas de accesibilidad

---

## Decision7: Transiciones — ≤150ms con prefers-reduced-motion

**Decision**: Todas las transiciones CSS son ≤150ms y se respetan `prefers-reduced-motion`.

**Rationale**: El gui-shell usa transiciones sutiles (≤150ms) que mejoran la experiencia sin
ser intrusivas. `prefers-reduced-motion` es esencial para accesibilidad (usuarios con
 vértigo o mareos).

**Alternatives considered**:
- Transiciones más largas (300ms): perceptibles pero lentas
- Sin transiciones: interface se siente estática
- JavaScript animations: más complejo, peor performance

---

## Decision8: Modales — Focus Trap + aria-modal

**Decision**: Los modales (settings, spend) implementan focus trap, aria-modal, y Esc close.

**Rationale**: Los modales del gui-shell usan `position: fixed` + `inset: 0` + overlay.
Para accesibilidad, se requiere focus trap (Tab cycling dentro del modal), aria-modal, y
cierre con Esc. Esto sigue las WAI-ARIA Modal Dialog Pattern.

**Alternatives considered**:
- Drawers laterales: diferente patrón, no coincide con el diseño de referencia
- Inline panels: menos separación del contexto
- Toast notifications: no aplica para configuración/gasto

---

## Decision9: Paleta Ctrl+K — Overlay con Agrupación

**Decision**: La paleta global usa overlay + input + items agrupados por categoría.

**Rationale**: El gui-shell agrupa los items en "Acciones", "Proyectos", "Pestañas". Esta
agrupación ayuda al usuario a encontrar rápidamente lo que busca. El overlay oscurece el
fondo para enfocar la atención en la paleta.

**Alternatives considered**:
- Sidebar palette: menos discoverable
- Dropdown simple: no suficiente espacio para muchos items
- Búsqueda full-text: más complejo, innecesario para este caso

---

## Decision10: Tool Cards — Inline en Activity Pane

**Decision**: Las herramientas se muestran como tarjetas en el activity pane derecho, no
inline en el chat.

**Rationale**: El gui-shell muestra las herramientas en el panel derecho (activity pane).
Esto mantiene el chat limpio de "ruido" de herramientas y permite ver la actividad de
herramientas mientras se lee la conversación. Cada tool card muestra: icono de estado,
nombre, args, y resultado.

**Alternatives considered**:
- Inline en chat: más ruido visual, peor legibilidad del chat
- Drawer lateral: menos discoverable
- Toast notifications: se pierde el historial

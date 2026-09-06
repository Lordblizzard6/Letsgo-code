# Investigación: Remodelado visual de la GUI Wails — apariencia CODEX GUI

**Branch**: `006-codex-gui-remodel` | **Fecha**: 2026-08-14 | **Spec**: [spec.md](spec.md)

## Resumen

El feature 006 es el **ejecutor independiente** del lenguaje visual Codex/Qwen que
el feature 005 definió en su enmienda US5 (research D8/D9, gui-contract §8,
quickstart J5). 005 fue implementado por fases y su plan (T090-T101) ya aplicó
parte del restyle P0; sin embargo, el usuario declara que la GUI **"no se ve igual
en lo absoluto"** a Codex. Este documento audita el **estado real del código**
(`cmd/wails/frontend/src/`) contra la anatomía Codex y produce la lista de brechas
P0/P1/P2 que el plan.md resuelve. El documento canónico de referencia visual sigue
siendo la sección **D10** de este archivo; los tokens finales (una sola fuente)
viven en [contracts/gui-contract.md](contracts/gui-contract.md) §9.

Las decisiones D1..D9 (Wails v3, Svelte-TS, streaming, sistema de diseño, backlog
UX) se heredan de `specs/005-bubbletea-tui-wails/research.md` y **no se re-abren**;
006 solo añade D10 (auditoría del estado actual) y D11 (decisiones de remediación)
para el remodelado completo.

---

## D10: Auditoría del estado actual vs. Codex GUI

### 10.1 Anatomía Codex que vamos a imitar (referencia canónica)

La GUI desktop de OpenAI Codex (y su convergente Qwen Code) se describe con estos
rasgos, que son los que la audiencia percibe a 8ft de distancia:

1. **Chrome casi monocromo**: ventana con rail fino y barras de superficie neutra;
   el color de la superficie habla por luminancia (bg → surface → elevated), no por
   borde de tarjeta. Cero sombras en el contenido; las sombras solo aparecen en
   overlays/flotantes.
2. **El transcript es un lienzo**: mensajes de usuario como **burbuja compacta a la
   derecha** (tint de superficie activa, no borde completo); mensajes de asistente
   como **bloques full-width sin borde**, separados por ritmo tipográfico (header
   de mensaje discreto: rol + timestamp + modelo), no por cajas.
3. **Los bloques de código son objetos elevados**: contenedor con header
   (lenguaje en caps + botón copiar visible en hover/focus), superficie elevada,
   resaltado de sintaxis, scroll horizontal dentro del bloque. Nunca un `<pre>`
   desnudo pegado al texto.
4. **Rail fino** (~48px) con iconografía **SVG lineal monocroma** (cero emoji),
   marcador activo **único** (barra/tint de acento), secciones en caps y colapso a
   iconos con tooltip *Label Alt+n*.
5. **Composer como "command shell" docked**: caja anclada al fondo, siempre
   visible y **editable durante el streaming**; el botón en la misma ranura cambia
   Send ⇄ Stop; affordance de streaming textual; hints de teclado en 11px bajo el
   campo.
6. **Status line inferior persistente** (28px): proveedor · modelo · uso hoy · dot
   de estado busy/sync. El estado "Ready/Streaming" **no** vive en el transcript.
7. **Cero emoji en el chrome**: rail, botones, headers, status, paleta. El emoji
   solo puede aparecer en el contenido markdown del usuario/asistente.
8. **Dark-first**: dark es la navegación principal; light es traducción AA
   verificada. Acento único restringido (familia indigo-violeta Codex/Qwen).
9. **Calm states**: skeletons de carga (no spinners gigantes), empty states con
   CTA, errores con orientación + Retry, foco visible, motion ≤150ms y
   `prefers-reduced-motion` honrado.

### 10.2 Estado real verificado (archivos leídos)

Auditoría directa de `cmd/wails/frontend/src/` (2026-08-14):

| Archivo | Línea(s) | Estado verificado |
|---------|----------|-------------------|
| `app.css` | 11-38 | Tokens **GitHub**: dark `#0d1117` bg, acento `#2f81f7`, light `#f6f8fa`/`#0969da`; sin escala de superficies por luminancia |
| `app.css` | 192-198 | `.msg` = **tarjeta por mensaje** (borde 1px + radius 6px + padding 16px) |
| `app.css` | 229-242 | `.spinner` genérico (borde giratorio) para stream/saving |
| `app.css` | 276-283 | `.composer` simple (flex textarea+botón), sin shell ni controles laterales |
| `app.css` | 323 | `.overlay` con `rgba(0,0,0,0.45)` → **oscurece el chat** |
| `app.css` | 124-128 | `.rail` 64px (no 48↔56), sin colapso |
| `chat.svelte` | 41-49 | `.stream-indicator` con spinner + "Streaming..." / "Ready" **dentro del transcript** (pill flotante de estado) |
| `chat.svelte` | 52 | `<ToolActivity />` renderizado **después** de los mensajes → log colgante que empuja el layout |
| `composer.svelte` | 46-65 | Send⇄Stop ya en la misma ranura (bien) pero sin chip de modelo, sin controles izquierda, sin hint row 11px fija |
| `message.svelte` | 8-14 | Cada mensaje es `<article class="msg">` con label "You"/"Assistant"; sin burbuja usuario, sin header-per-message |
| `message.svelte` | 36-42 | `<pre>` desnudo del renderer default (sin header/lang/copy/highlight) |
| `rail.svelte` | 13-46 | **Rail duplicado**: componente standalone con **emoji** (💬📁🌿✅🔌🧩⚙️📊❓), 64px, sin collapse ni tooltip `Label Alt+n` |
| `App.svelte` | 43-53 | `ICONS` = **emoji**; rail inline duplicado en App.svelte 160-194 (G6: rail duplicado) |
| `App.svelte` | 218-240 | **Pills flotantes**: ThemeSwitch pill top-right + avatar top-right; sin status line inferior |
| `App.svelte` | 115-117 | Esc solo navega SYSTEM→chat; **sin restore de foco al composer** para overlays |
| `palette.svelte` | 96 | Icono default `"›"`, items con emoji; sin agrupación, sin footer hints; row activa usa `--bg-input` (no acento) |
| `settings-overlay.svelte` | 117-123 | Overlay **modal que oscurece** el chat (rgba 0.45), spinner en saving, botón "✕" |
| `theme-switch.svelte` | 13-19 | Pill flotante con **emoji** ☀/🌙, radius pill 20px |
| `sessions-panel.svelte` | 56-59 | Empty "No sessions yet...", sin "Refrescando…", sin agrupar por fecha |
| `tool-activity.svelte` | 5-14 | Lista **global plana** de rows (borde izquierdo), no inline, no colapsable, no agrupada |
| `lib/store.svelte.ts` | 249-257 | `upsertActivity` ya deduplica por `callId` (1 row/call OK en store), pero la UI lo muestra plano y suelto |
| `lib/markdown.ts` | 23-33 | `renderMarkdown` = streaming-markdown + DOMPurify; renderer default → `<pre>` sin componentes |

**Strings bloqueados por tests** (no se tocan salvo cambio deliberado documentado):
`"Welcome to LetsGO"`, `"Save & start chatting"`, `"API Key"`,
`"Type your message..."` (composer placeholder) — verificados en
`ProviderSetup.svelte`, `settings-overlay.svelte`, `composer.svelte` y los tests
`onboarding.test.ts` / `e2e/onboarding.spec.ts`.

### 10.3 Brechas priorizadas vs. Codex (accionable)

| ID | Prioridad | Brecha | Evidencia | Remedio (workstream) |
|----|-----------|--------|-----------|----------------------|
| P0-1 | **P0** | Token set es GitHub (acento azul `#2f81f7`/`#0969da`, bg `#0d1117`, radius 6px, 2 niveles de surface), no dark-first Codex indigo | `app.css:11-38` | R1 (tokens completos en app.css) |
| P0-2 | **P0** | Transcript = tarjeta por mensaje con borde; sin burbuja usuario derecha ni bloque asistente full-width | `message.svelte:8-14`, `app.css:192-198` | R2 (transcript) |
| P0-3 | **P0** | Rail con emoji, 64px, duplicado (rail.svelte + inline en App.svelte), sin collapse ni tooltip `Label Alt+n` | `rail.svelte:13-46`, `App.svelte:160-194` | R4 (rail SVG + collapse) |
| P0-4 | **P0** | No hay status line inferior; estado "Ready/Streaming" es pill flotante en el transcript; pills flotantes theme/avatar | `chat.svelte:41-49`, `App.svelte:218-240`, `theme-switch.svelte:13-19` | R7 (status line) + R10 |
| P0-5 | **P0** | Bloques de código = `<pre>` desnudo, sin header+lang+copy+highlight | `message.svelte:36-42`, `lib/markdown.ts:23-33` | R3 (CodeBlock) |
| P1-6 | P1 | Tool activity = log global plano que empuja el layout; debe ser inline por turno, colapsable, agrupada | `tool-activity.svelte:5-14`, `chat.svelte:52` | R6 (tool cards) |
| P1-7 | P1 | Overlays oscurecen el chat (backdrop 0.45); deben ser drawers/paneles que dejan el chat visible + foco al cerrar | `settings-overlay.svelte:117-123`, `app.css:323` | R8 (overlays) |
| P1-8 | P1 | Spinners genéricos en stream/saving en vez de skeletons/cursor de escritura | `app.css:229`, `chat.svelte:41-49`, `settings-overlay.svelte:228-243` | R9 (estados) |
| P1-9 | P1 | Estados empty/loading/error incompletos en superficies (sessions, git, tasks, mcp, plugins, usage, help) | `sessions-panel.svelte:56-59` y pares | R9 |
| P1-10 | P1 | Paleta sin agrupación, sin footer hints, fila activa sin acento, iconos emoji | `palette.svelte:96-104,163-166` | R10 |
| P1-11 | P1 | Foco: overlays no restauran foco al composer; Esc incompleto (solo SYSTEM→chat) | `App.svelte:109-122` | R8 + R10 |
| P2-12 | P2 | Composer no es "command shell": sin chip de modelo, sin hint row fija 11px, sin affordance streaming visual | `composer.svelte:46-65` | R5 |
| P2-13 | P2 | Theme-switch pill emoji vs. control de tema consistente con tokens | `theme-switch.svelte:13-19` | R10 |
| P2-14 | P2 | Sesiones sin agrupar por fecha, sin "Refrescando…", sin dot busy consistente | `sessions-panel.svelte:33,56-59` | R9/R10 |

### 10.4 Decisiones de remediación (D11)

**Decision**: Aplicar el remodelado como **reemplazo visual completo** (no parche):
reemplazar los tokens de `app.css` por el set Codex de 005 D8 (una sola fuente en
`contracts/gui-contract.md` §9), reconstruir transcript/rail/composer/overlays desde
las primitivas del sistema, y reutilizar toda la lógica existente (store, bindings,
eventos, rutas hash, atajos Alt+1..9). El único cambio de lógica permitido es la
dedupe/agrupación de tool rows en `lib/store.svelte.ts` (ya parcialmente presente
en `upsertActivity`) y el estado derivado para la status line; **cero cambios** en
bindings Go, servicios bound, contrato de eventos y motor (constitución II/III,
gate UX-14).

**Rationale**: la auditoría muestra que las piezas funcionales ya existen (Send⇄Stop
en composer, upsertActivity dedupe, Esc cierra SYSTEM, paleta con teclado); lo que
falta es el **lenguaje visual** completo y coherente. El reemplazo de tokens +
primitivas es la única forma de cumplir el objetivo del usuario ("no se ve igual en
lo absoluto") sin regrabar el backend. El formato de 005 (gates G1–G10, UX-01..UX-14,
tests vitest/Playwright, quickstart J5) se reutiliza tal cual para que la verificación
sea idéntica y comparable.

**Alternatives considered**:
- Parche progresivo por componente (mantener tokens GitHub): descartado — deja
  colores/radii/acento inconsistentes y el usuario ya rechazó la iteración anterior.
- Cambiar acento a azul Codex estricto (`#5c99d6`): descartado — se mantiene la
  familia indigo-violeta heredada de Qwen (D8), intercambiable solo en tokens.
- Reescribir bindings/contrato para exponer más datos: descartado — viola II y el
  spec 006 prohíbe cambios de backend (FR-012).

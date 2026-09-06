# Implementation Plan: Remodelado visual de la GUI Wails — apariencia CODEX GUI

**Branch**: `006-codex-gui-remodel` | **Date**: 2026-08-14 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/006-codex-gui-remodel/spec.md` +
directiva del usuario (2026-08-14): remodelar la GUI Wails para que **se vea como
la GUI desktop de OpenAI Codex** ("no se ve igual en lo absoluto actualmente").
Ejecutor del lenguaje visual definido en 005 US5/D8/D9, pero como **spec
independiente** con revisión del estado real (research D10) y workstreams propios.

## Summary

El feature 006 reemplaza el lenguaje visual "GitHub-style" actual (acento azul
`#2f81f7`/`#0969da`, tarjetas por mensaje, rail emoji 64px, pills flotantes,
spinners, overlays que oscurecen) por el lenguaje **Codex/Qwen** (dark-first,
chrome casi monocromo, transcript-canvas con burbuja de usuario + bloque asistente
full-width sin borde, bloques de código como objetos elevados con header+copy+
highlight, rail fino con SVG lineal + colapso 48↔56 + tooltip `Label Alt+n`,
composer "command shell" docked editable en streaming con Send⇄Stop, status line
inferior persistente, cero emoji en chrome). Es **solo frontend presentacional**
(Svelte + `app.css` + `lib/store.svelte.ts` para dedupe/agrupación de tool rows);
los bindings Go, servicios bound, contrato de eventos, rutas hash, atajos Alt+1..9
y el motor **no cambian** (guardarraíl UX-14, FR-012). Los tests vitest
pre-existentes pasan sin cambios y `e2e/onboarding.spec.ts` queda intacto (FR-013).

Enfoque técnico (detalle en [research.md](research.md) D10/D11):

- **Tokens**: reemplazo completo de `app.css` — un solo juego de semantic tokens
  dark-first (`html[data-theme="dark"|"light"]`), escala 4px, radius sm/md/lg/xl
  6/8/10/12 + pill, sombras solo en overlays, tipografía Inter + JetBrains Mono,
  acento único indigo `#8b94f8` (dark) / `#4f46e5` (light). Los componentes
  consumen `var(--token)`; hex literal prohibido fuera de `app.css` (G1).
- **Transcript**: burbuja de usuario a la derecha (`--bg-active`), bloque asistente
  full-width sin borde con header-per-message y cursor de escritura; sin fila
  "Streaming…" (la status line la absorbe).
- **CodeBlock.svelte**: header (language caps + botón Copy on hover/focus) +
  resaltado de sintaxis offline (highlight.js/shiki, tokens por tema) + overflow
  horizontal en el bloque; todo `<pre>` vía este componente (G8).
- **Rail**: SVG lineal stroke monocromo (sin emoji), marcador activo único (barra
  3px + tint `--accent-soft`), secciones caps, colapso 48↔56px con tooltip
  `Label Alt+n`; se consolida el rail duplicado (G6).
- **Composer**: command shell docked, textarea siempre activa, Enter=send /
  Shift+Enter=nueva línea / Enter=stop en stream, controles izquierda (chip de
  modelo) y derecha Send⇄Stop (misma ranura, accent fill / err-soft outline, min
  44px), hint row 11px fija.
- **Tool activity**: cards inline en el turno, colapsables, agrupadas (5+ → grupo
  compacto con contador), dedupe 1 row/call en `store.svelte.ts`; nunca un log que
  empuje al composer.
- **Status line** inferior 28px (proveedor · modelo · uso hoy · dot busy) que
  absorbe "Ready/Streaming" y elimina las pills flotantes (theme + avatar se
  integran en chrome).
- **Overlays**: Settings/Usage/Help como drawers/paneles con backdrop sutil que
  **no oscurecen** el chat, cierre con Esc/backdrop y foco restaurado al composer.
- **Estados**: empty + loading (skeleton) + error con CTA en las 7+ superficies de
  datos y chat streaming/error/empty/idle; motion ≤150ms + `prefers-reduced-motion`.

## Technical Context

**Language/Version**: Svelte 5 (runes) + TypeScript + Vite, frontend de la GUI
Wails (`cmd/wails/frontend/`); backend Go sin cambios.

**Primary Dependencies**: `streaming-markdown` + DOMPurify (render existente, se
mantiene); highlight.js o shiki **offline** para el CodeBlock (nueva dependencia
solo frontend). Sin framework UI nuevo (constitución V).

**Testing**: Vitest + jsdom (tests existentes sin cambios + nuevos por fase) y
Playwright (e2e `onboarding.spec.ts` intacto + smoke visual G10) — Test-Last, al
final de cada workstream (constitución I).

**Target Platform**: Windows 10/11 (WebView2) — misma que 005.

**Performance Goals**: render streaming memoizado del último bloque; sin re-parseo
O(n²) del markdown acumulado (SC-009 heredado); skeleton inmediato en cargas.

**Constraints**: cero diff de bindings/backend (UX-14); strings bloqueados por
tests intactos ("Welcome to LetsGO", "Save & start chatting", "API Key",
placeholder "Type your message...") salvo cambio deliberado documentado en el
mismo commit; tokens solo en `app.css`; cero emoji en chrome (G5).

**Scale/Scope**: 11 workstreams (R1..R11) sobre ~18 componentes Svelte +
`app.css` + `lib/store.svelte.ts`; ~1.500-3.000 LOC de cambios presentacionales.

## Constitution Check

*GATE: debe pasar antes de Phase 0 y se re-evalúa tras cada workstream.*

La constitución está **ratificada** (`.specify/memory/constitution.md` v1.0.0).
Verificación principio a principio:

- **I. Test-Last (NON-NEGOTIABLE)**: tests al final de cada workstream; los tests
  vitest pre-existentes y el e2e onboarding se mantienen como regresión. **Pasa**.
- **II. Contract-Only Core Access**: cero cambios en bindings Go, servicios bound
  y contrato de eventos; la UI consume el core solo vía el contrato existente
  (gate UX-14). **Pasa**.
- **III. One Source of Truth**: el estado de dominio no se duplica; el store solo
  añade presentación derivada (agrupación de tool rows, estado para status line).
  **Pasa**.
- **IV. Secure Rendering**: nunca `innerHTML`; se mantiene streaming-markdown +
  DOMPurify doble saneado; copy/highlight son DOM/CSS sin innerHTML (UX-07).
  **Pasa**.
- **V. Simplicity & YAGNI**: un solo set de tokens y primitivas; se consolida el
  rail duplicado, no se crean abstracciones. **Pasa**.

**Violaciones**: ninguna. No se requiere Complexity Tracking.

## Workstreams (R1..R11) — gates G1–G10 y UX-01..UX-14

Mismos gates G1–G10 y UX-01..UX-14 que 005 (gui-contract §8 y 006 contract §9),
con numeración de workstreams propia de 006. Cada workstream termina con sus tests
Vitest/Playwright (Test-Last) y cumple el gate de regresión UX-12..UX-14 acumulado.

| WS | Alcance | File(s) principal(es) | Gate(s) |
|----|---------|------------------------|---------|
| R1 | Tokens CSS completos en app.css (reemplazo total, dark-first, escala 4px, radius, sombras, tipografía) | `app.css` | G1, G2, G4 |
| R2 | Transcript: user bubble derecha + bloque asistente full-width sin borde + header-per-message + cursor stream | `message.svelte`, `chat.svelte`, `app.css` | G2, G4, G10 |
| R3 | CodeBlock component (header+lang+copy+highlight) acorde a `lib/markdown.ts` | `lib/markdown.ts`, `message.svelte`, nuevo `code-block.svelte` | G8, G2 |
| R4 | Rail → SVG lineal + colapso 48↔56 + tooltip `Label Alt+n`; consolidar rail duplicado | `App.svelte`, `rail.svelte` (consolidado), `app.css` | G5, G3, G4, G6 |
| R5 | Composer → command shell con Send⇄Stop y affordance streaming; hint row fija | `composer.svelte`, `app.css` | G2, G3, G10 |
| R6 | Tool activity → cards inline colapsables agrupadas; dedupe en store | `tool-activity.svelte`, `chat.svelte`, `lib/store.svelte.ts` | G7, UX-04, G9 |
| R7 | Status line inferior (proveedor·modelo·uso hoy·dot busy); eliminar pills flotantes de estado | `App.svelte`, nuevo `status-line.svelte`, `app.css` | G5, G2, G10 |
| R8 | Overlays → drawers sin oscurecer chat + Esc/backdrop + foco al composer | `settings-overlay.svelte`, `usage-overlay.svelte`, `help-overlay.svelte`, `App.svelte` | G3, UX-05 |
| R9 | Empty/loading(skeleton)/error en 7+ superficies; chat streaming/error/empty/idle | paneles + `chat.svelte`, `app.css`, primitivas | G7, UX-09, G9 |
| R10 | theme-switch/settings/usage/help + paleta consistentes con tokens (sin emoji, fila activa acento, grouped, footer hints) | `theme-switch.svelte`, `palette.svelte`, `settings-overlay.svelte`, `usage-overlay.svelte`, `help-overlay.svelte` | G5, G2, G3, G10 |
| R11 | Regresión: vitest pre-existentes pasan sin cambios; e2e onboarding intacto; demo funcional completa | todo el frontend | UX-12..UX-14, G1, G10 |

### R1 — Tokens CSS en app.css (reemplazo completo)

- [ ] R1.1 Reemplazar el bloque `:root`/`html[data-theme=...]` de `app.css` por el
      set de tokens de `contracts/gui-contract.md` §9 (bg, surfaces por luminancia,
      bordes, texto 3 niveles, acento + accent-on + accent-soft, semánticos,
      tipografía, spacing, radius, sombras, focus, motion, z-index) — dark-first.
- [ ] R1.2 Añadir `--font-ui` (Inter/system-ui) y `--font-mono` (JetBrains Mono);
      base 14px, mono 13px, caps 11px tracking 0.08em, tabular-nums.
- [ ] R1.3 Definir sombras `--shadow-sm/md/lg` (solo overlays/flotantes) y
      `--overlay` scrim sutil (< 0.25 para drawers; la página no lleva sombra).
- [ ] R1.4 Verificar G1: `rg "#[0-9a-fA-F]{3,8}" src/ --glob '*.svelte' --glob '*.ts'`
      → 0 hits fuera de `app.css`; G4: solo escala 4px; G2: pares AA en dark+light.
- [ ] R1.5 Vitest: test de token-set (las variables definidas existen y contrastan
      AA en ambos temas).

### R2 — Transcript: user bubble + assistant block sin borde

- [ ] R2.1 Reemplazar `.msg` (tarjeta) por `.msg--user` (burbuja derecha,
      `--bg-active`, radius-lg, max-width ~70%) y `.msg--assistant` (bloque
      full-width sin borde, sin fondo de panel).
- [ ] R2.2 Header-per-message en asistente: rol "Assistant" + timestamp + modelo
      (13px, `--text-tertiary`), separación por ritmo tipográfico (margen 24px).
- [ ] R2.3 Eliminar `.stream-indicator` con spinner y la fila "Streaming…/Ready"
      del transcript; el estado vive en la status line (R7); cursor de escritura
      `▍` + línea de progreso 2px bajo el bloque en curso.
- [ ] R2.4 Quitar `.error-banner` flotante del transcript en favor del banner con
      orientación + Retry (chat error state, R9).
- [ ] R2.5 Verificar G4 (chat ≤760px), G2, G10 con captura dark+light.

### R3 — CodeBlock component (header+lang+copy+highlight)

- [ ] R3.1 En `lib/markdown.ts`, post-procesar el HTML saneado: extraer `<pre><code
      class="language-*">` y emitir un marcador para el componente CodeBlock (sin
      tocar el saneado DOMPurify; sin innerHTML en Svelte).
- [ ] R3.2 Crear `code-block.svelte`: header (language caps + botón Copy on
      hover/focus con estado "Copied" temporal), body con scroll horizontal dentro
      del bloque (nunca scroll de página), radius-md, superficie elevada.
- [ ] R3.3 Resaltado de sintaxis offline (highlight.js/shiki, tokens por tema),
      sin CDN; fallback silencioso a `<code>` plano si el idioma no se detecta.
- [ ] R3.4 G8: todo `<pre>` vía CodeBlock; `rg "<pre>" src/ --glob '*.svelte'`
      → 0 en render directo; copiar con `navigator.clipboard` + `aria-label`.
- [ ] R3.5 Vitest: render de markdown con bloque de código → header + lang + copy
      presente; Playwright: captura del code block en ambos temas.

### R4 — Rail → SVG lineal + collapse 48↔56 + tooltip

- [ ] R4.1 Reemplazar emoji del rail (App.svelte `ICONS` y rail.svelte) por un set
      único de SVG stroke 16px (chat, sesiones, git, tasks, mcp, plugins, settings,
      usage, help) + account/theme en chrome (R7/R10).
- [ ] R4.2 Colapso 48↔56px: botón collapse al pie; colapsado = solo iconos con
      tooltip `Label Alt+n`; expandido = icono + label (13px) + hint `Alt+n`.
- [ ] R4.3 Marcador activo único: icono en `--accent` + barra 3px izquierda +
      tint `--accent-soft`; secciones WORK/TOOLS/SYSTEM en caps 11px.
- [ ] R4.4 Consolidar el rail duplicado (G6): eliminar `rail.svelte` standalone y
      dejar un único rail en `App.svelte` o un único componente.
- [ ] R4.5 G5 (cero emoji en chrome: `rg "💬|📁|🌿|✅|🔌|🧩|⚙️|📊|❓|☀|🌙|🎨|›"` → 0),
      G3 (focus-visible), G4 (targets ≥44px).

### R5 — Composer → command shell + Send⇄Stop + affordance streaming

- [ ] R5.1 Convertir `.composer` en shell docked: contenedor con radius-lg, borde
      hairline `--border-strong`, sombra sm; textarea sin borde interno.
- [ ] R5.2 Controles izquierda: chip de modelo (desplegable de modelos existente
      en config); mantener placeholder "Type your message..." (string bloqueado).
- [ ] R5.3 Send⇄Stop en la misma ranura derecha: Send = accent fill con
      `--accent-on`; Stop = err-soft outline; min 44px; deshabilitado solo si
      empty y no streaming (comportamiento actual intacto).
- [ ] R5.4 Hint row fija 11px bajo el campo: `Ctrl+K commands · / plan · @ files ·
      Alt+1-9 surfaces`; affordance "Streaming — Enter stops" visible solo en
      stream (string existente, se conserva).
- [ ] R5.5 G2/G3/G10: contraste del hint y del chip; captura del composer en dark+
      light y en stream (Stop).

### R6 — Tool activity → cards inline colapsables agrupadas

- [ ] R6.1 En `lib/store.svelte.ts`, mantener `upsertActivity` (1 row/call) y
      añadir derivación de **turno** (agrupar tool rows por turno de conversación)
      y contador de grupo (5+ → grupo compacto con contador).
- [ ] R6.2 Reescribir `tool-activity.svelte`: cards inline colapsables (header =
      icono SVG + nombre + estado live + chevron; body mono truncado 200ch +
      expand), success `--ok` / error `--err` + Retry; running edge accent 1px +
      pulse ≤150ms (off bajo reduced-motion).
- [ ] R6.3 Render inline en el turno (dentro del bloque del mensaje que las
      disparó), nunca después del transcript como log colgante; verificar que el
      composer nunca es empujado.
- [ ] R6.4 UX-04: 1 row por tool call ya agrupada; G7 (estado del grupo);
      G9 (pulse ≤150ms + reduced-motion).
- [ ] R6.5 Vitest: dedupe 1/call + agrupación por turno + colapso/expansión.

### R7 — Status line inferior (proveedor·modelo·uso hoy·dot busy)

- [ ] R7.1 Crear `status-line.svelte` (28px, `--bg-panel`, borde superior hairline,
      caps 11px): proveedor · modelo · uso hoy (tokens/coste de `useAccount`) ·
      dot busy (online/busy/offline con `aria-live`).
- [ ] R7.2 Mover el estado "Ready/Streaming" aquí (dot + texto corto) y **eliminar**
      las pills flotantes de estado del transcript; integrar theme + avatar en
      chrome (R10) — se retiran las pills flotantes top-right.
- [ ] R7.3 G5 (sin emoji, dot = CSS circle), G2 (AA del texto 11px), G10 (captura
      con status line en stream y en idle).
- [ ] R7.4 Vitest: la status line refleja provider/model/today y el dot cambia con
      el estado de cuenta; Playwright: smoke.

### R8 — Overlays → drawers sin oscurecer chat + foco

- [ ] R8.1 Convertir Settings/Usage/Help de overlay modal con scrim 0.45 a panel
      lateral/drawer con backdrop sutil (≤0.20) que deja el chat visible (UX-05).
- [ ] R8.2 Cierre con Esc y backdrop-click → restaurar foco al composer
      (`.composer textarea`); focus trap solo dentro del drawer mientras está
      abierto; `aria-modal="false"` con `aria-hidden` gestionado en el chat.
- [ ] R8.3 Reemplazar spinner de saving por skeleton/shimmer ≤150ms o estado
      inline; botón de cierre con SVG (no "✕" de texto puede quedarse si es
      accesible — preferir SVG stroke).
- [ ] R8.4 G3 (foco visible en ambos temas), UX-05 (Esc/backdrop + chat visible),
      G9 (transición drawer ≤150ms, slide en X).
- [ ] R8.5 Playwright: abrir Settings → el chat permanece visible y focuseable
      tras Esc; vitest: foco restaurado.

### R9 — Empty/loading(skeleton)/error en 7+ superficies + chat states

- [ ] R9.1 Crear primitivas reutilizables: `empty-state.svelte` (icono SVG + título
      + texto + CTA), `skeleton.svelte` (shimmer ≤150ms, off bajo reduced-motion),
      `error-banner.svelte` (icono + orientación + botón Retry).
- [ ] R9.2 Aplicar a 10 superficies del rail: chat, sessions, git, tasks, mcp,
      plugins, settings, usage, help, account — cada una con empty (CTA cuando
      aplica) + loading (skeleton) + error (orientación + retry).
- [ ] R9.3 Chat states: streaming (cursor ▍ + status line), error (banner con
      Retry que reproduce el último turno — UX-01), empty (CTA de bienvenida),
      idle (status line "Ready").
- [ ] R9.4 Sesiones: agrupar por fecha (Hoy/Ayer/Older) con caps 11px; "Refrescando…"
      preservando contenido; dot busy.
- [ ] R9.5 G7 (10 superficies), UX-09 (7/7 mínimas), G9; vitest por superficie.

### R10 — theme-switch/settings/usage/help + paleta consistentes

- [ ] R10.1 Theme-switch: integrar en el chrome (status line o footer del rail) con
      SVG sun/moon, sin pill flotante emoji; mostrar variante actual (UX-11).
- [ ] R10.2 Settings/usage/help: consumir primitivas (inputs, tabs, kpi, tabla de
      datos re-estilizada con tokens), sin hex, sin emoji.
- [ ] R10.3 Paleta Ctrl+K: 640px, radius-lg, `--shadow-lg`, **agrupada** (Ir a ·
      Acciones · Tema), iconos SVG, fila activa con acento (`--accent-soft` +
      ring), highlight substring con acento, footer hints, focus trap, Esc →
      foco al composer (comportamiento actual intacto).
- [ ] R10.4 G5 (cero emoji en palette/theme/settings/usage/help), G2, G3,
      G10 (captura palette dark+light).
- [ ] R10.5 Vitest: paleta grouped + fila activa + focus restore; theme-switch
      muestra variante actual.

### R11 — Regresión y demo funcional

- [ ] R11.1 `npm test -- --run`: los tests vitest pre-existentes
      (chat/composer/onboarding/errors/sessions/palette/tool-activity) pasan **sin
      cambios**; añadir solo los nuevos por fase (UX-12).
- [ ] R11.2 `npm run playwright`: `e2e/onboarding.spec.ts` **intacto** pasa; los
      nuevos specs de smoke visual G10 (chat+code+tool, palette, onboarding,
      empty, dark+light) pasan (UX-13).
- [ ] R11.3 `git diff` (fuera de `specs/`) toca solo `cmd/wails/frontend/**`;
      grep de call-sites a services/eventos sin cambios (UX-14).
- [ ] R11.4 Smoke manual `letsgo.exe gui` (J6): key → modelo → chat con tool
      inline → Ctrl+K → Esc overlays → sesión TUI retomada → tema alterno con AA.
- [ ] R11.5 Demo funcional completa: chat streaming, sesiones, herramientas,
      configuración, paleta, atajos Alt+1..9 y onboarding operativos.

## Gates del feature (mismos que 005, referenciados)

Los gates de diseño **G1–G10** y de UX **UX-01..UX-14** son los definidos en
`specs/005-bubbletea-tui-wails/contracts/gui-contract.md` §8, consolidados y
ampliados en `specs/006-codex-gui-remodel/contracts/gui-contract.md` §9 (token
purity G1, WCAG AA G2, foco G3, spacing G4, cero emoji G5, inventario G6, estados
G7, código G8, motion G9, smoke visual G10; UX-01..UX-11 de comportamiento y
UX-12..UX-14 de regresión). La verificación es idéntica a 005 (grep, Vitest,
Playwright contra el dev server, smoke `letsgo.exe gui`), de modo que 006 es
comparable 1:1 con la enmienda US5 de 005.

## Complexity Tracking

Sin violaciones de constitución (el remodelado es solo presentacional/frontend).
Complejidad justificada en D10/D11 (un solo set de tokens, un CodeBlock, un rail
consolidado, primitivas de estado; sin abstracciones nuevas). Se mantiene la tabla
vacía.

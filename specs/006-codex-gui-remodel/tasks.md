# Tasks: Remodelado visual de la GUI Wails — apariencia CODEX GUI

**Input**: Design documents from `/specs/006-codex-gui-remodel/`

**Prerequisites**: plan.md (required), spec.md (user stories), research.md, `contracts/gui-contract.md`, quickstart.md

**Tests**: SI — el SPEC lo requiere explícitamente (Vitest+jsdom con bindings mockeados, Playwright contra el dev server, gates G1–G10 y UX-01..UX-14). Según la constitución v1.0.0 (**Principio I, Test-Last, NON-NEGOTIABLE**), los tests se escriben y ejecutan **DESPUÉS** de la implementación, en batch al final de cada fase; los tests vitest pre-existentes y `e2e/onboarding.spec.ts` se mantienen intactos como regresión (FR-013, UX-12/UX-13).

**Organization**: Fases por workstream del plan.md (R1..R11), cada una implementación → tests → gate.

## Format: `[ID] [P?] [Fase] Descripción`

- **[P]**: Puede correr en paralelo (archivos distintos, sin dependencias)
- **[Fase]**: R1..R11 (plan.md). Las fases Setup/Polish no llevan label.
- Rutas exactas de fichero en cada descripción. Referencias a IDs de contrato §9, FR-xxx y SC-xxx donde aplican.

## Build / Test del repo (Windows)

- Frontend: en `cmd/wails/frontend/`: `npm run check` (svelte-check), `npm test -- --run` (vitest), `npx playwright test` (contra el dev server, channel msedge)
- Backend sin cambios (cero diff de bindings/backend — gate UX-14): `go build ./...` como humo
- GUI: `letsgo.exe gui` (J6)

---

## Phase 1: Setup (línea base verificada)

**Proposito**: confirmar la línea base verde (vitest + svelte-check + e2e + go build) ANTES de tocar el frontend, para que la regresión R11 sea atribuible solo al remodelado.

- [ ] T001 Verificar la línea base: en `cmd/wails/frontend/` — `npm run check` (0 errores), `npm test -- --run` (vitest verde), `npx playwright test` (e2e onboarding verde); `go build ./...` en la raíz verde (bindings intactos).
- [ ] T002 Añadir dependencia de resaltado de sintaxis offline al frontend en `cmd/wails/frontend/package.json`: `highlight.js` (runtime, sin CDN) para el CodeBlock (plan R3.3); `npm install` y `npm run build` de humo.

**Checkpoint**: línea base verde guardada y dependencia de highlight.js instalada.

---

## Phase 2: R1 — Tokens CSS en app.css (reemplazo completo)

**Proposito**: un solo set de semantic tokens dark-first en `app.css` (una sola fuente, contract §9.1); los componentes consumen `var(--token)` y cero hex fuera de `app.css` (G1).

- [X] T003 [R1] Reemplazar el bloque `:root`/`html[data-theme=...]` de `cmd/wails/frontend/src/app.css` por el set de tokens de `contracts/gui-contract.md` §9.1: `--font-ui` (Inter/system-ui), `--font-mono` (JetBrains Mono), escala 4px `--space-1..12`, `--radius-sm/md/lg/xl` (6/8/10/12 + pill), sombras solo overlays `--shadow-sm/md/lg`, foco/motion `--focus-ring`, `--dur-fast/normal`, `--ease`, capas `--z-*`; dark: `--bg #0e1116`, surfaces por luminancia `--bg-panel #14181e` → `--bg-elevated #1a1f26` → `--bg-input #1f252d` → `--bg-active #252c36`, `--border #2e3540`, `--border-strong #3d4654`, texto 3 niveles `#e8ebf0/#9aa3b0/#6f7885`, acento único `--accent #8b94f8` + `--accent-on` + `--accent-soft`, semánticos ok/warn/err (+ on, err-soft), `--overlay` y `--overlay-strong`; light: base `#f6f7f9`, `--accent #4f46e5`, texto `#1c2026/#5c6570/#8b939e` (G1, G2, G4).
- [X] T004 [R1] Renovar el resto de `app.css`: actualizar clases base (`button`, `.card`, `.pane`, `.caps`, `.tabular`, `.msg`, `.composer`, `.rail`, overlays) a los tokens §9.1 y la dirección por superficie §9.2 — chrome casi monocromo, superficie por luminancia, chat ≤760px, targets ≥44px, `:focus-visible` ring en todo interactivo, `prefers-reduced-motion` honrado (G2, G3, G4, G9).
- [X] T005 [R1] Test token-set (Vitest) en `cmd/wails/frontend/src/tests/tokens.test.ts`: las variables definidas existen en `app.css` y los pares texto/fondo y acento/fondo contrastan AA (texto ≥4.5:1, no-texto ≥3:1) en ambos temas (G2).

**Checkpoint**: tokens aplicados y verificados — G1 (rg hex = 0 fuera de app.css), G2 (AA), G4 (escala 4px).

---

## Phase 3: R2 — Transcript: user bubble + assistant block sin borde

**Proposito**: el transcript es un lienzo — usuario = burbuja derecha, asistente = bloque full-width sin borde, separación por ritmo tipográfico; el estado "Streaming…/Ready" sale del transcript (lo absorbe la status line R7) (FR-002).

- [X] T006 [R2] `cmd/wails/frontend/src/components/message.svelte`: reemplazar `<article class="msg">` (tarjeta con borde) por `.msg--user` (burbuja derecha, `--bg-active`, radius-lg, max-width ~70%, align-self flex-end) y `.msg--assistant` (bloque full-width sin borde, sin fondo de panel) con header-per-message discreto (mark + rol + timestamp + modelo, 13px `--text-tertiary`); cursor de escritura `▍` + línea de progreso 2px durante el stream (R2.1, R2.2, R2.3).
- [X] T007 [R2] `cmd/wails/frontend/src/components/chat.svelte`: eliminar `.stream-indicator` (spinner + "Streaming..."/"Ready") del transcript; mantener el error-banner (pasa a primitiva en R9); ajustar layout a chat-canvas (R2.3).
- [X] T008 [R2] Test (Vitest) en `cmd/wails/frontend/src/tests/chat.test.ts` + `message`-focused: un mensaje de usuario renderiza una única burbuja con clase `.msg--user`; el asistente renderiza `.msg--assistant` sin borde; sin `.stream-indicator` en el transcript; el streaming no rompe la acumulación (UX-10, I-8, G10).

**Checkpoint**: transcript-canvas funcionando — vitest verde, chat ≤760px (G4), G10 visual.

---

## Phase 4: R3 — CodeBlock component (header+lang+copy+highlight)

**Proposito**: todo `<pre>` es un objeto elevado con header (lenguaje caps + botón Copy on hover/focus + estado "Copied" ≤1.5s), resaltado de sintaxis offline y scroll horizontal dentro del bloque (FR-003, G8).

- [X] T009 [R3] `cmd/wails/frontend/src/lib/markdown.ts`: post-procesar el HTML saneado — detectar `<pre><code class="language-*">` y emitir un contenedor `.code-block` con header + botón copiar (sin tocar DOMPurify; sin innerHTML en Svelte) (R3.1).
- [X] T010 [R3] Crear `cmd/wails/frontend/src/components/code-block.svelte`: header (language caps + Copy on hover/focus, estado "Copied" 1.5s), body con scroll horizontal dentro del bloque (nunca scroll de página), radius-md, superficie `--bg-elevated` + `--border`; resaltado con highlight.js offline (R3.2, R3.3).
- [X] T011 [R3] `cmd/wails/frontend/src/components/message.svelte`: integrar el CodeBlock en el render del asistente; `rg "<pre>" src/ --glob '*.svelte'` → 0 en render directo; copiar con `navigator.clipboard` + `aria-label` (R3.4).
- [X] T012 [R3] Test (Vitest) en `cmd/wails/frontend/src/tests/code-block.test.ts`: render de markdown con bloque de código → header con lang + botón Copy presente; clic copia al portapapeles (mock `navigator.clipboard`) y muestra "Copied"; idioma no detectado → fallback `<code>` plano (G8, I-18).

**Checkpoint**: code blocks con header+copy+highlight — vitest verde, G8 (rg `<pre>` = 0 directo).

---

## Phase 5: R4 — Rail → SVG lineal + collapse 48↔56 + tooltip

**Proposito**: rail fino con iconografía SVG stroke monocroma (cero emoji), marcador activo único con acento, colapso 48↔56px con tooltip `Label Alt+n`, consolidado en un único componente (FR-004, G5, G6).

- [X] T013 [R4] Crear el set único de iconos SVG stroke 16px (chat, sessions, git, tasks, mcp, plugins, settings, usage, help, theme, account, collapse) en `cmd/wails/frontend/src/lib/icons.ts` (cero emoji, stroke currentColor) (R4.1, G5).
- [X] T014 [R4] `cmd/wails/frontend/src/components/rail.svelte`: reescribir como el ÚNICO rail — recibe la lista de destinos (9 tabs + secciones WORK/TOOLS/SYSTEM en caps 11px), marcador activo único (icono `--accent` + barra 3px izquierda + tint `--accent-soft`), colapso 48↔56px (botón collapse al pie), tooltip `Label Alt+n` en colapsado, hint `Alt+n` en expandido (R4.2, R4.3, R4.4).
- [X] T015 [R4] `cmd/wails/frontend/src/App.svelte`: eliminar el rail inline emoji (duplicado) y renderizar `<Rail>` único; actualizar `ICONS` → SVG de `lib/icons.ts` para la paleta; eliminar pills flotantes de tema/avatar (migran a chrome en R7/R10) (R4.4, R4.5).
- [X] T016 [R4] Test (Vitest) en `cmd/wails/frontend/src/tests/rail.test.ts`: el rail renderiza un único `<nav class="rail">` con 9 destinos en 3 secciones, marcador activo único, colapso 48↔56 con tooltip `Label Alt+`, y cero emoji en chrome (`rg` de iconos emoji → 0) (G5, G6, I-10, G3).

**Checkpoint**: rail consolidado en SVG — vitest verde, G5 (rg emoji = 0 en chrome), G6 (un solo rail).

---

## Phase 6: R5 — Composer → command shell + Send⇄Stop + affordance

**Proposito**: composer como "command shell" docked y anclado, editable durante el streaming, con chip de modelo, Send⇄Stop en la misma ranura, hint row 11px fija y affordance "Streaming — Enter stops" (FR-005, R5).

- [X] T017 [R5] `cmd/wails/frontend/src/components/composer.svelte`: convertir en shell — contenedor radius-lg, borde `--border-strong`, `--shadow-sm`; textarea sin borde interno siempre activa; chip de modelo a la izquierda (desplegable desde `useConfig`/config); Send (accent fill `--accent-on`) ⇄ Stop (err-soft outline) en la misma ranura derecha, min 44px, deshabilitado solo si empty y no streaming; placeholder "Type your message..." intacto (string bloqueado); hint row 11px `Ctrl+K commands · / plan · @ files · Alt+1-9 surfaces` + affordance "Streaming — Enter stops" en stream (R5.1-R5.4).
- [X] T018 [R5] Test (Vitest) en `cmd/wails/frontend/src/tests/composer.test.ts`: sigue alternando Send↔Stop y manteniendo el input activo durante el stream; la affordance "Streaming — Enter stops" aparece en stream; hint row presente (I-7, FR-008, G10).

**Checkpoint**: composer shell — vitest verde, G2/G3 (contraste hint/chip), G10 (captura dark+light y en stream).

---

## Phase 7: R6 — Tool activity → cards inline colapsables agrupadas

**Proposito**: la actividad de herramientas es inline en el turno, como cards colapsables agrupadas (5+ → grupo compacto con contador), 1 row por call (dedupe), nunca un log que empuje al composer (FR-006, UX-04).

- [X] T019 [R6] `cmd/wails/frontend/src/lib/store.svelte.ts`: mantener `upsertActivity` (1 row/call) y añadir derivación de turno (agrupar tool rows por turno de conversación) y contador de grupo (5+ → grupo compacto) (R6.1).
- [X] T020 [R6] `cmd/wails/frontend/src/components/tool-activity.svelte`: reescribir como cards inline colapsables (header = icono SVG + nombre + estado live + chevron; body mono truncado 200ch + expand), success `--ok` / error `--err` + Retry; running edge accent 1px + pulse ≤150ms (off bajo reduced-motion); agrupación 5+ → grupo compacto (R6.2, R6.4).
- [X] T021 [R6] `cmd/wails/frontend/src/components/chat.svelte`: renderizar `<ToolActivity />` inline en el turno correspondiente (dentro del bloque del mensaje que las disparó), nunca después del transcript como log colgante; el composer nunca es empujado (R6.3).
- [X] T022 [R6] Test (Vitest) en `cmd/wails/frontend/src/tests/tool-activity.test.ts`: dedupe 1 row/call sigue funcionando; agrupación por turno; colapso/expansión; 5+ tools → grupo compacto con contador (UX-04, I-9, G7, G9).

**Checkpoint**: tool cards inline agrupadas — vitest verde, UX-04, G9 (pulse ≤150ms + reduced-motion).

---

## Phase 8: R7 — Status line inferior (proveedor·modelo·uso hoy·dot busy)

**Proposito**: status line 28px persistente que absorbe "Ready/Streaming" y elimina las pills flotantes de estado (FR-007).

- [X] T023 [R7] Crear `cmd/wails/frontend/src/components/status-line.svelte` (28px, `--bg-panel`, borde superior hairline, caps 11px): proveedor · modelo · uso hoy (de `useAccount.today()`) · dot busy (CSS circle, `aria-live`); estado "Ready/Streaming" aquí (dot + texto corto) (R7.1, R7.2).
- [X] T024 [R7] `cmd/wails/frontend/src/App.svelte`: integrar `<StatusLine />` al pie del layout; retirar las pills flotantes top-right (ThemeSwitch pill + account avatar pill) — theme/account migran al chrome (rail footer R10 o status line) (R7.2).
- [X] T025 [R7] Test (Vitest) en `cmd/wails/frontend/src/tests/status-line.test.ts`: la status line refleja provider/model/today y el dot cambia con el estado de cuenta (online/busy/offline); sin emoji (G5, G2, R7.4).

**Checkpoint**: status line persistente — vitest verde, G5, G10 (captura con status line en stream e idle).

---

## Phase 9: R8 — Overlays → drawers que no oscurecen el chat + foco

**Proposito**: Settings/Usage/Help como drawers/paneles laterales con backdrop sutil (≤0.20) que deja el chat visible; cierre con Esc/backdrop y foco restaurado al composer (FR-009, UX-05).

- [X] T026 [R8] `cmd/wails/frontend/src/app.css`: `.overlay` con `--overlay` (scrim sutil ≤0.20) para drawers y `--overlay-strong` para la paleta; transición drawer ≤150ms slide en X (G9).
- [X] T027 [R8] `cmd/wails/frontend/src/components/settings-overlay.svelte`: convertir a drawer lateral (backdrop sutil, chat visible, `--shadow-lg`, radius-xl); Esc/backdrop cierran y restauran foco al composer; reemplazar spinner de saving por estado inline (skeleton/shimmer ≤150ms); botón de cierre con SVG stroke; labels/strings de test intactos (R8.1-R8.3).
- [X] T028 [R8] `cmd/wails/frontend/src/components/usage-overlay.svelte` y `help-overlay.svelte`: mismos patrones de drawer (backdrop sutil, cierre Esc/backdrop, foco al composer, botón SVG) (R8.1-R8.3).
- [X] T029 [R8] Test (Vitest) en `cmd/wails/frontend/src/tests/overlays.test.ts`: Esc cierra Settings/Usage/Help volviendo al chat (hash "") y restaurando foco al composer; backdrop-click cierra; el chat subyacente permanece visible (UX-05, G3, G9).

**Checkpoint**: drawers que no oscurecen — vitest verde, UX-05, G3 (foco), G9.

---

## Phase 10: R9 — Empty/loading(skeleton)/error en 10 superficies + chat states

**Proposito**: primitivas de estado reutilizables (empty-state, skeleton, error-banner) en las 10 superficies del rail; el chat define streaming/error/empty/idle; sesiones agrupadas por fecha (FR-010, G7, UX-09).

- [ ] T030 [R9] Crear primitivas `cmd/wails/frontend/src/components/empty-state.svelte` (icono SVG + título + texto + CTA), `skeleton.svelte` (shimmer ≤150ms, off bajo reduced-motion) y `error-banner.svelte` (icono + orientación + botón Retry que reproduce el último turno — UX-01) (R9.1).
- [ ] T031 [R9] Aplicar primitivas a las 10 superficies del rail (chat, sessions, git, tasks, mcp, plugins, settings, usage, help, account): cada una con empty (CTA cuando aplica) + loading (skeleton preservando contenido previo) + error (orientación + retry) (R9.2, R9.3).
- [ ] T032 [R9] `cmd/wails/frontend/src/components/sessions-panel.svelte`: agrupar sesiones por fecha (Hoy/Ayer/Older) con caps 11px; "Refrescando…" preservando contenido; dot busy; empty accionable con CTA "Nueva conversación" (R9.4).
- [ ] T033 [R9] Test (Vitest) en `cmd/wails/frontend/src/tests/empty-states.test.ts`: 7+ superficies muestran empty accionable y loading (skeleton) conservando previo; error-banner con Retry reproduce el último turno (UX-09, G7, UX-01).

**Checkpoint**: estados completos — vitest verde, G7 (10 superficies), UX-09, G9.

---

## Phase 11: R10 — theme-switch/settings/usage/help + paleta consistentes

**Proposito**: control de tema integrado en el chrome (sin pill emoji flotante), paneles y paleta con tokens, agrupada, fila activa con acento, footer hints, foco al composer al cerrar (FR-008, G5, G3).

- [ ] T034 [R10] `cmd/wails/frontend/src/components/theme-switch.svelte`: integrar en el chrome (footer del rail o status line) con SVG sun/moon (cero emoji), mostrar variante actual (UX-11); se retira la pill flotante (R10.1).
- [ ] T035 [R10] `cmd/wails/frontend/src/components/palette.svelte`: 640px, radius-lg, `--shadow-lg`, `--overlay-strong`; **agrupada** (Ir a · Acciones · Tema) con headers caps; iconos SVG; fila activa con acento (`--accent-soft` + ring); highlight substring con acento; footer hints; focus trap; Esc → foco al composer (comportamiento actual intacto) (R10.3).
- [ ] T036 [R10] `cmd/wails/frontend/src/components/settings-overlay.svelte` + `usage-overlay.svelte` + `help-overlay.svelte`: consumir primitivas (inputs, tabs, kpi, tabla de datos re-estilizada con tokens), sin hex, sin emoji (R10.2).
- [ ] T037 [R10] Test (Vitest) en `cmd/wails/frontend/src/tests/palette.test.ts`: la paleta sigue navegable con ↑/↓+Enter y Esc cierra; la fila activa usa `--accent-soft`; sin emoji en iconos; theme-switch muestra variante actual (G5, G3, R10.5).

**Checkpoint**: chrome y paleta consistentes — vitest verde, G5, G10 (captura palette dark+light).

---

## Phase 12: R11 — Regresión y demo funcional

**Proposito**: los tests vitest pre-existentes pasan sin cambios, `e2e/onboarding.spec.ts` intacto, cero diff de bindings/backend, demo funcional completa (FR-013, UX-12..UX-14).

- [ ] T038 [R11] `npm test -- --run` en `cmd/wails/frontend/`: los vitest pre-existentes (chat/composer/errors/onboarding/palette/sessions/tool-activity) pasan **sin cambios** + los nuevos de cada fase (R1..R10) (UX-12).
- [ ] T039 [R11] `npm run check` (svelte-check 0 errores) y `npm run build` (build de producción) en `cmd/wails/frontend/`.
- [ ] T040 [R11] `npx playwright test` en `cmd/wails/frontend/`: `e2e/onboarding.spec.ts` **intacto** pasa (strings "Welcome to LetsGO", "Save & start chatting", "API Key", placeholder intactos) (UX-13).
- [ ] T041 [R11] `git diff` (fuera de `specs/`) toca solo `cmd/wails/frontend/**`; `rg` de call-sites a services/eventos sin cambios (UX-14); `go build ./...` verde.
- [ ] T042 [R11] Smoke manual `letsgo.exe gui` (J6): key → modelo → chat con tool inline → Ctrl+K → Esc overlays → sesión TUI retomada → tema alterno con AA; checklist G10 (10 puntos) comparado contra captura Codex (SC-001..SC-008).

**Checkpoint**: feature 006 completo — vitest + svelte-check + build + e2e verdes, cero diff backend, demo J6 completa.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: sin dependencias — puede empezar de inmediato.
- **R1 (Phase 2)**: depende de Setup; es la base de todas las demás (tokens en app.css).
- **R2..R10 (Phases 3-11)**: dependen de R1 (los componentes consumen los tokens). R2/R3/R4/R5/R6 pueden trabajarse en paralelo tras R1 (ficheros distintos); R7 (status line) y R8 (overlays) dependen de R4 (rail/chrome) parcialmente; R9 y R10 dependen de R2-R8.
- **R11 (Phase 12)**: regresión final — depende de todas las fases.

### Within Each Phase

- Implementación → tests escritos y ejecutados al final (Test-Last, constitución I) → gate.
- Commit tras cada fase o grupo lógico; parar en cualquier checkpoint para validar.

### Parallel Opportunities

- [P] Tras R1: R2 (message/chat), R3 (code-block/markdown), R4 (rail/App/icons), R5 (composer), R6 (tool-activity/store) en paralelo por ficheros distintos.
- [P] Tras R2/R4: R7 (status-line/App) y R8 (overlays/app.css).
- [P] Tras R2-R8: R9 (estados) y R10 (palette/theme).
- [P] Tests de cada fase en paralelo sobre ficheros de test distintos.

---

## Implementation Strategy

### Entrega incremental

1. Setup (T001-T002) → línea base + highlight.js.
2. +R1 (T003-T005) → tokens Codex en app.css (base para todo).
3. +R2..R6 (T006-T022) → transcript-canvas, code blocks, rail SVG, composer shell, tool cards.
4. +R7..R8 (T023-T029) → status line + drawers que no oscurecen.
5. +R9..R10 (T030-T037) → estados completos + paleta/chrome consistentes.
6. R11 (T038-T042) → regresión y demo J6.

### Parallel Team Strategy

- Tras R1: A=R2+R3, B=R4+R7, C=R5+R6, D=R8, E=R9+R10.
- Nota de ficheros: `message.svelte`/`chat.svelte` solo R2/R3/R6; `App.svelte` solo R4/R7/R10; `app.css` R1/R8/R9; `store.svelte.ts` R6.

---

## Notas

- **Test-Last (constitución I)**: los tests se escriben y ejecutan al final de cada fase; no TDD-first.
- **Solo frontend presentacional**: cero cambios a bindings Go, servicios bound, contrato de eventos, rutas hash, atajos Alt+1..9 y motor (gate UX-14, FR-012).
- **Strings bloqueados por tests**: "Welcome to LetsGO", "Save & start chatting", "API Key", placeholder del composer "Type your message..." se conservan (UX-13, FR-013).
- **Token purity (G1)**: cero hex literal en `.svelte`/`.ts`; todo color vía `var(--token)` de `app.css`.
- **Cero emoji en chrome (G5)**: emoji solo permitido en el contenido markdown del usuario/asistente.
- **Render**: nunca `innerHTML`; DOMPurify + streaming-markdown sobre el acumulado (D3).
- Commit después de cada fase o grupo lógico; checkpoint al final de cada fase.

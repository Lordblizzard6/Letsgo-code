# Quickstart: Remodelado visual de la GUI Wails — apariencia CODEX GUI

**Branch**: `006-codex-gui-remodel` | **Fecha**: 2026-08-14 | **Spec**: [spec.md](spec.md)

Guía de validación end-to-end del remodelado visual del feature 006. Los detalles
de superficie y tokens están en [contracts/gui-contract.md](contracts/gui-contract.md)
§9 y el estado real en [research.md](research.md) D10. La validación funcional
base (J1..J5) se hereda de `specs/005-bubbletea-tui-wails/quickstart.md`; este
documento añade **J6**, que verifica los criterios de éxito SC-001..SC-008 del
spec 006.

## Prerrequisitos

- Implementación de los workstreams R1..R11 ([plan.md](plan.md)) con sus tests
  Vitest por fase (Test-Last) + `npm run playwright`.
- Mismo entorno que 005: Go 1.26, Wails v3, WebView2, key de proveedor configurada.
- Referencia visual: captura de la GUI desktop de OpenAI Codex (para la comparación
  del checklist G10 a 8ft).

## J6 — Codex remodel: demo + verificación (SC-001..SC-008)

**Objetivo**: verificar que la GUI **se ve como Codex** (SC-001) sin perder
funcionalidad (SC-007/SC-008), con los gates G1..G10 y UX-01..UX-14 del
[gui-contract §9](contracts/gui-contract.md) cumplidos.

```bash
cd cmd/wails/frontend
npm run check                                   # svelte-check: 0 errores, sin emoji, tokens en app.css
npm test -- --run                               # vitest existentes + nuevos (chat/composer/onboarding/errors + R1..R11)
npm run playwright                              # e2e onboarding intacto + smoke G10 (chat, palette, overlays, empty, dark+light)
rg "#[0-9a-fA-F]{3,8}" src/ --glob '*.svelte' --glob '*.ts'    # SC-002 (G1): 0 hits fuera de app.css
```

### Verificación por criterio de éxito

1. **SC-001 — Comparación visual Codex**: `letsgo.exe gui`, abrir una conversación
   con código y una herramienta; comparar contra la captura Codex y pasar los **10
   puntos** del checklist G10 (§9.4): composición general, cero emoji en chrome,
   transcript-canvas (burbuja usuario / bloque asistente sin borde), CodeBlock con
   header+copy+highlight, rail SVG con marcador único, composer shell con
   Send⇄Stop, status line inferior, paleta agrupada, drawers que no oscurecen,
   dark+light AA. Cobertura automatizada: smoke Playwright G10.
2. **SC-002 — Token purity (G1)**: `rg "#[0-9a-fA-F]{3,8}" src/ --glob '*.svelte'
   --glob '*.ts'` → 0; todo color vive en `app.css` (`html[data-theme=...]`).
3. **SC-003 — Keyboard 100% (G3)**: recorrer todo interactivo por teclado en dark
   y light; flujo Ctrl+K ⇄ paleta (↑/↓/Enter) ⇄ Esc ⇄ foco al composer, con focus
   visible (`:focus-visible` ring en ambos temas); Alt+1..9 sin cambios.
4. **SC-004 — Estados (G7/UX-04)**: las **10** superficies del rail definen empty +
   loading (skeleton) + error con CTA; el chat define streaming (cursor ▍ + status
   line) / error (banner + Retry que reproduce el último turno) / empty / idle;
   tool rows sin duplicados (1/call, agrupadas por turno).
5. **SC-005 — WCAG AA (G2)**: contrastes automáticos en ambos temas (body/meta
   ≥4.5:1, no-text ≥3:1, texto sobre acento ≥4.5:1, focus ring ≥3:1); smoke
   Playwright a 100/125/150% de escala.
6. **SC-006 — Código/iconos/spacing (G8/G5/G4)**: todo `<pre>` se renderiza vía
   CodeBlock (header+lang+copy+highlight); cero emoji en chrome; spacing solo en
   escala 4px y chat ≤760px.
7. **SC-007 — Regresión (UX-12..UX-14)**: `npm test -- --run` con los vitest
   pre-existentes pasando **sin cambios**; `e2e/onboarding.spec.ts` intacto;
   `git diff --stat` (fuera de `specs/`) solo toca `cmd/wails/frontend/**` (cero
   diff de bindings/backend).
8. **SC-008 — Demo funcional (J6 manual)**: jornada completa abajo.

### Jornada de demo (SC-008)

1. Abrir `letsgo.exe gui` con key configurada → onboarding ya completado, chat
   directo. Sin key → ProviderSetup split-screen con "Welcome to LetsGO",
   "API Key", "Save & start chatting" (strings bloqueados intactos).
2. Flujo key → modelo → chat en ≤4 acciones (configuración → guardar → enviar).
3. Conversar con una herramienta → **tool card inline** en el turno, colapsable,
   agrupada; el composer nunca se desplaza.
4. Durante el stream: el composer sigue **editable**, botón Send ⇄ Stop en la
   misma ranura, affordance "Streaming — Enter stops", status line con dot busy.
5. Ctrl+K → paleta agrupada navegable con ↑/↓ + Enter; Esc → cierra y devuelve el
   foco al composer.
6. Alt+1..9 recorren las superficies; Settings/Usage/Help abren **drawers** que no
   oscurecen el chat; Esc/backdrop las cierran.
7. Retomar una sesión creada en la TUI en 1 clic (historial compartido).
8. Alternar tema dark/light → ambos cumplen AA; bajo `prefers-reduced-motion` no
   hay shimmer/pulse/cursor blink.

**Esperado**: screenshot de la GUI vs. captura Codex sin fallos de token en el
checklist G10 (los gustos no son fallo de gate); suite vitest + Playwright verde;
`git diff` limitado a `cmd/wails/frontend/**`.

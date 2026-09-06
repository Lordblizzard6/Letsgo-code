# Implementation Plan: Doble interfaz — TUI Bubble Tea + GUI Wails

**Branch**: `005-bubbletea-tui-wails` | **Date**: 2026-08-13 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/005-bubbletea-tui-wails/spec.md` +
directiva del usuario (2026-08-13): **eliminar Fyne Y AHORA** y sustituirlo por
Wails, con el **mismo método de ejecución que Fyne hoy** (`letsgo.exe gui` vía
cobra -> `gui.Run()`), sin binario separado de Wails.

## Summary

El cliente LetsGo pasa a dos interfaces sobre un mismo core: la **TUI Bubble Tea**
(existente, sin cambios de tecnología) y una **GUI de escritorio Wails** que
**reemplaza por completo a la GUI Fyne** (`internal/gui` y `fyne.io/fyne` se
retiran de forma inmediata: primera tarea del feature, no gate final). El motor, la
base de datos sqlite y la configuración actuales se formalizan en un **contrato de
frontend** (comandos, eventos de stream, sesiones, configuración) con tests de
conformidad headless que ambas interfaces deben pasar. Enfoque técnico (detalle en
[research.md](research.md)):

- **Wails v3** (`v3.0.0-beta.7`, pin exacto, ya en go.mod) + Svelte-TS; streaming de
  chat vía evento `chat:delta` batcheado (~50 ms) + `streaming-markdown`/DOMPurify
  en el frontend.
- **Ejecución de la GUI por el mismo ejecutable**: el comando cobra `letsgo gui`
  (`cmd/gui.go`) lanza la app Wails (paquete reutilizable en `cmd/wails/gui`), igual
  que hoy lanza Fyne — sin binario Wails separado ni `wails3 dev` para arrancar.
- Retirada de Fyne **inmediata y total**: borrar `internal/gui`, `cmd/gui.go`
  (reescrito para Wails), imports `fyne.io/*`, assets fyne y `FyneApp.toml`;
  verificar `go mod why fyne.io/fyne/v2` y `fyne.io/systray` vacíos (Gate G-2/G-3).
- TUI sin cambios estructurales: tests de modelo puro (`Update` sintético) +
  teatest; GUI testeada con Vitest (mock de bindings) y Playwright contra el dev
  server de Vite en CI.

## Technical Context

**Language/Version**: Go 1.26.1 (módulo `github.com/user/go-claude-code`)

**Primary Dependencies**:
- TUI: charmbracelet bubbletea v1.3.10, bubbles v1.0.0, glamour v1.0.0, lipgloss (ya en go.mod; se mantienen).
- GUI: `github.com/wailsapp/wails/v3` (**pin `v3.0.0-beta.7`, ya en go.mod**), frontend Svelte + TypeScript + Vite; `@wailsio/runtime`.
- Core: engine/db/config actuales (sin nuevos frameworks).
- **Se eliminan AHORA**: `fyne.io/fyne/v2`, `fyne.io/systray` y transitivos (glfw/OpenGL) + `internal/gui` Fyne.

**Storage**: sqlite vía `internal/db` (esquema existente, sin tablas nuevas) + `internal/config`.

**Testing**: `go test` headless (engine/db/config + modelos TUI); teatest (charm) para programas Bubble Tea completos; Vitest + jsdom y Playwright (contra `wails dev`) para la GUI; suite Fyne headless se porta 1:1 al motor/servicios.

**Target Platform**: Windows 10/11 (WebView2 Evergreen/embedded, `-webview2 embed`); macOS (WKWebView) y Linux (webkit2gtk-4.1, `-tags webkit2_41`) soportados; CI en Linux + runner Windows para features Windows.

**Project Type**: aplicación de escritorio (GUI) + CLI interactivo (TUI), frontend web embebido en la GUI.

**Performance Goals**: primer token <3s; deltas continuos (batch ~50ms); scroll de ≥500 mensajes sin degradación; arranque TUI <1s y ventana GUI <5s (SC-003, SC-005, SC-006, SC-009).

**Constraints**: paridad funcional completa con la GUI Fyne actual antes de retirarla (FR-017/FR-018); ninguna interfaz duplica lógica de conversación (FR-001); mensajes en markdown crudo con saneado en el render (GUI: DOMPurify sobre acumulado); ejecución simultánea TUI+GUI fuera de v1.

**Scale/Scope**: 2 frontends nuevos/puertos: TUI (~sin cambio estructural) + GUI Wails (~10 paneles del rail, paleta Ctrl+K, atajos Alt+1..8, overlays, estados vacíos/de carga); ~2.000-5.000 LOC de frontend + capa de servicios bound; retirada completa de Fyne.

## Constitution Check

*GATE: debe pasar antes de Phase 0 y se re-evalúa tras Phase 1.*

La constitución está **ratificada** (`.specify/memory/constitution.md` v1.0.0,
2026-08-13). Verificación principio a principio:

- **I. Test-Last (NON-NEGOTIABLE)**: los tests se escriben y ejecutan DESPUÉS de la
  implementación, al final de cada historia/fase. **Pasa**: este plan pone la
  implementación primero (Wails → GUI) y los tests (Vitest/Playwright/Go) al final
  de cada US; los tests de conformidad C-001..C-006 ya existentes se mantienen como
  regresión (no se reordenan como "primero").
- **II. Contract-Only Core Access**: la GUI consume el core solo vía el contrato de
  frontend (C-001..C-006). **Pasa**: los services bound delegan al engine.
- **III. One Source of Truth**: motor/DB/config centralizados; la GUI no duplica
  lógica. **Pasa**: mismos paquetes, sin fork.
- **IV. Secure Rendering**: nunca `innerHTML`; DOMPurify + streaming-markdown sobre
  el acumulado; contraste AA en ambos temas. **Pasa**: scripting D3.
- **V. Simplicity & YAGNI**: una sola GUI; sin abstracciones nuevas. **Pasa**: se
  reutiliza el esqueleto existente de services/bindings.

**Violaciones**: ninguna. No se requiere Complexity Tracking.

*Re-evaluación tras Phase 1 (design): sin cambios — el diseño del contrato y la
retirada inmediata de Fyne refuerzan II (contract-only) y V (una sola GUI).
GATE: PASA.*

*Re-evaluación tras la enmienda US5 (2026-08-14): el restyle es **solo frontend
presentacional** (Svelte + app.css + lib/store.svelte.ts para dedupe de tool rows).
II se refuerza (los services bound no cambian; cero diff de bindings/backend, gate
UX-14), IV se ratifica (se mantiene streaming-markdown + DOMPurify doble saneado;
copy/highlight son DOM/CSS sin innerHTML; gate UX-07), V se cumple (un único set de
tokens y primitivas; se consolida el rail duplicado, no se crean abstracciones)
y I se mantiene (tests Vitest/Playwright al final de cada fase U5-A..D). GATE: PASA.*

## Project Structure

### Documentation (this feature)

```text
specs/005-bubbletea-tui-wails/
├── plan.md              # Este archivo (/speckit.plan)
├── research.md          # Phase 0: decisiones D1..D7 (Wails v3, Svelte-TS, streaming, testing, retirada Fyne)
├── data-model.md        # Phase 1: entidades/transiciones (sesiones, mensajes, config)
├── quickstart.md        # Phase 1: jornadas J1..J4 de validación
├── contracts/           # Phase 1:
│   ├── frontend-contract.md   # Superficie core compartida (comandos, eventos, sesiones/config, conformidad C-001..C-006)
│   └── gui-contract.md        # Inventario GUI Wails + checklist de paridad + Gate de retirada Fyne (G-1..G-3)
└── tasks.md             # Phase 2 (/speckit.tasks — no creado aquí)
```

### Source Code (repository root)

```text
cmd/
├── main.go              # entrada CLI
├── gui.go               # REWRITE: `letsgo gui` (cobra) lanza la GUI Wails (antes Fyne)
└── wails/               # NUEVO: GUI Wails
    ├── gui/             # paquete reutilizable `gui.Run()` llamado por cmd/gui.go (embed + application.New)
    ├── services/        # NUEVO: ChatService, SessionsService, GitService, TasksService,
    │                    #        MCPService, PluginsService, SettingsService, UsageService,
    │                    #        ThemeService, AccountService (bound, sin lógica de dominio)
    └── frontend/        # NUEVO: app Svelte-TS (Vite) → dist embebido
        ├── src/
        │   ├── components/   # rail, chat, composer, paneles, paleta, overlays, empty/loading
        │   ├── bindings/     # generado por wails3
        │   └── tests/        # vitest + playwright
internal/
├── engine/              # core conversacional (EXISTE — contrato sobre él; sin cambios de comportamiento)
├── db/                  # sqlite (EXISTE — sin tablas nuevas)
├── config/              # configuración compartida (EXISTE)
├── tui/                 # TUI Bubble Tea (EXISTE — consume el contrato; tests de modelo puro + teatest)
└── gui/                 # ELIMINAR de inmediato (Fyne out, primera tarea del feature)

tests/                   # N/A — los tests viven junto a cada paquete (convención Go del repo)
```

**Structure Decision**: la GUI Wails se ejecuta por el **mismo ejecutable** que la
TUI (`letsgo.exe gui`), reutilizando el comando cobra `cmd/gui.go` que hoy lanza
Fyne; `cmd/gui.go` importa el paquete `cmd/wails/gui` (`Run()`) en lugar de
`internal/gui`. `internal/gui` (Fyne) se elimina en la primera tarea del feature
con el pin de Wails ya presente (`v3.0.0-beta.7` en go.mod) y el esqueleto de
services/bindings ya generado. El contrato (`contracts/`) es la frontera entre
core e interfaces.

## US5 — Restyle & UX: diseño tipo Codex/Qwen (post-US4, enmienda 2026-08-14)

**Directiva del usuario (2026-08-14)**: "estilizar y mejorar la UI de Wails con
enfasis en similitud con Codex GUI y Qwen-GUI, misma filosofia de diseño y gates".
Investigacion a cargo de @UI Designer y @UX Researcher (hallazgos en
[research.md](research.md) D8/D9). Alcance del restyle: **solo frontend**
(Svelte + app.css); sin cambios en bindings, servicios bound, contrato de eventos,
rutas hash, atajos Alt+1..9 ni lógica del motor (guarda UX-14 en D9).

### Filosofía de diseño convergente (referencias)

1. **Chrome mínimo**: marco neutro, bordes hairline, superficies por luminancia (no
   por esquinas), sombras solo en overlays.
2. **Denso pero aireado**: UI 13–14px, paddings generosos 16–24px, jerarquía por
   tamaño/peso, no por color.
3. **Acento restringido**: un solo acento (activo/links/focus/CTA primario);
   semánticos sobrios.
4. **Dark-first**: dark es la navegación principal; light es traducción AA verificada.
5. **El chat es el canvas**: mensajes como superficies de lectura (bubble usuario /
   bloque asistente full-width, sin card por mensaje); bloques de código como
   objetos elevados con header + copiar + resaltado; actividad de herramientas
   **inline en el turno** (colapsable, agrupada), nunca un log colgante que empuje
   al composer.
6. **Keyboardo como carga estructural**: paleta, foco visible, Esc cierra overlays,
   restore de foco al composer.

### Workstreams (fases, cada una con su gate y tests Vitest/Playwright al final — Test-Last)

| Fase | Alcance (del backlog D9) | File(s) principal(es) |
|------|--------------------------|------------------------|
| U5-A (P0) | Composer always-on (I-1), Retry reproduce último turno (I-2), navegación ↑/↓+Enter en paleta (I-3), abrir sesión entra al chat (I-4), Esc cierra Settings/Usage/Help (I-5), tool dedupe 1 row/call (I-6), affordance "Streaming — Enter stops" (I-7) | `chat.svelte`, `composer.svelte`, `palette.svelte`, `App.svelte`, `sessions-panel.svelte`, `lib/store.svelte.ts` |
| U5-B (P1) | Densidad por turnos (I-8), tool activity inline/agrupada (I-9), iconos SVG + silicio en rail/chrome (I-10), copy Help alineado (I-11), sesiones serie (I-12), settings no-modal Esc-able (I-13) | `message.svelte`, `tool-activity.svelte`, `App.svelte`, `rail.svelte`, `help-overlay.svelte`, `sessions-panel.svelte`, `settings-overlay.svelte` |
| U5-C (P1) | Estados empty/loading con CTA en 7+ superficies (I-14), foco/a11y (I-15), header status strip (I-16) | paneles + `chat.svelte` + `app.css` |
| U5-D (P2) | Optimización render streaming (I-17), syntax highlight + copy (I-18), uso per-provider (I-19), deep-links flyout (I-20), onboarding autofocus/advanced (I-21), rename/delete sesiones (I-22), tips empty (I-23) | `lib/markdown.ts`, `message.svelte`, `usage-overlay.svelte`, `account-flyout.svelte`, `ProviderSetup.svelte`, `sessions-panel.svelte` |

**Tokens y sistema**: bloques de color/type/spacing/radius/shadow/estado definidos en
D8 → viviran ENTEROS en `app.css` (semantic tokens, `html[data-theme]`); los
componentes consumen `var(--token)`, nunca hex literal (gate G1/G2).

### Gates de diseño y UX (extienden los del gui-contract §8)

Gate de restyle (verificables en CI por grep/Playwright/vitest):

- **G1 Token purity** — `rg "#[0-9a-fA-F]{3,8}" src/ --glob "*.svelte" --glob "*.ts"` = 0 fuera de `app.css`.
- **G2 WCAG AA ambos temas** — contrastes automáticos (body/meta ≥4.5:1, no-text ≥3:1, texto sobre acento ≥4.5:1); smoke Playwright a 100/125/150%.
- **G3 Foco visible** — `:focus-visible` ring en todo interactivo, ambos temas; flujo completo por teclado.
- **G4 Disciplina de spacing** — lint px solo en escala 4px (4/8/12/16/20/24/32/40/48); chat ≤760px; targets ≥44px.
- **G5 Iconografía** — cero emoji en chrome, SVG stroke único set.
- **G6 Cobertura de inventario ≥95%** — toda superficie del gui-contract desde primitivas; sin rail duplicado.
- **G7 Estados completos** — 10 superficies con empty+loading(skeleton)+error; chat con streaming/error/empty/idle.
- **G8 Código** — todo `<pre>` vía componente code-block (header+lang+copy+highlight); sin `<pre>` desnudo.
- **G9 Movimiento** — transiciones ≤150ms (salvo progreso stream); `prefers-reduced-motion` honrado.
- **G10 Smoke visual** — screeenshots chat/palette/onboarding/empty/dark+light contra checklist.

Gates de UX (D9 §4): UX-01 (Retry reproduce último turno), UX-02 (composer siempre visible), UX-03 (resume 1 acción), UX-04 (1 row/tool + agrupado), UX-05 (Esc/backdrop cierran overlays + chat visible), UX-06 keyboard 100% + foco, UX-07 aria-live/busy + reduced-motion, UX-08 contraste token-pairs, UX-09 empty/loading 7/7, UX-10 densidad (sin card-anidada), UX-11 copy/estado veraces, UX-12..UX-14 regresión estabilizadores (test existentes pasan sin cambios; e2e onboarding intacto; cero diff de bindings/backend).

### Estructura documental (adiciones a este feature)

```text
specs/005-bubbletea-tui-wails/
├── plan.md            # + US5 (este sección, enmienda)
├── research.md        # + D8 (sistema de diseño token) y D9 (UX backlog/gates)
├── quickstart.md      # + J5 (validación visual/UX del restyle)
└── contracts/gui-contract.md  # + §8 Sistema de diseño y gates UX
```

## Complexity Tracking

Sin violaciones de constitución (la enmienda US5 es solo presentacional/frontend).
Complejidad justificada resumida en D8/D9 (un solo acento, una escala, un juego de
tokens; sin abstracciones nuevas). Se mantiene la tabla vacía.
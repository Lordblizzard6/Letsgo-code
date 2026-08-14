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

## Complexity Tracking

No aplica — Constitution Check sin violaciones (no se requiere tabla).
# Implementation Plan: GUI Overhaul — Codex-style App Shell

**Branch**: `003-gui-overhaul` | **Date**: 2026-08-06 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/003-gui-overhaul/spec.md`

## Summary

Rediseño estructural y visual del desktop GUI de LetsGO para alcanzar apariencia "tipo Codex app" pero original: una **barra vertical (rail)** fija a la izquierda con trabajo arriba (Conversación, Sesiones, Git, Tareas, MCP, Plugins) y control anclado abajo (Uso, Cuenta/Proveedores, Tema, Ajustes, Ayuda), colapsable ícono/ícono+etiqueta. Las seis vistas secundarias pasan de overlays a **paneles intercambiables** dentro del área principal, con vuelta al chat vía `Esc`/`Alt+Left`. Aplicación de un **tema oscuro escalonado "Deep Goblue"** (3 niveles de fondo, acento LetsGO único, tipografía dual sans/mono, foco visible) y preservación sin regresión de todos los flujos de 002 (aprobaciones inline, steers/queues, plans, fork, grants, picker, keymap `?`, paleta Ctrl+K). Implementación Fyne v2.8 con `terminalTheme` ampliado; la barra es un `container.NewBorder` alrededor del `viewStack` existente para no tocar la maquinaria de paneles ni los tests headless.

## Technical Context

**Language/Version**: Go 1.22+ (toolchain actual del repo). Verificado con `go build` CGO.

**Primary Dependencies**: Fyne v2.8.0 (`fyne.io/fyne/v2`), driver desktop para shortcuts (`driver/desktop`), CGO/GLFW toolchain en `D:\msys64\ucrt64\bin`. Fonts: Cascadia Mono (ya embebido, `embed`); se añade una face sans embebida (p. ej. Inter) para UI.

**Storage**: Configuración en `internal/config` (viper, archivo `~/.letsGo/config.yaml`) — nuevos campos `rail.collapsed`, `theme` (dark/light), anchor de ventana `~1280x760`. Persistencia vía `SaveConfig`/`LoadConfig` existentes.

**Testing**: `go test ./...` headless con `test.NewApp()` (patrón 002). Suite `internal/gui/gui_test.go`: 60+ tests montando vistas aisladas o Controller con `viewStack` manual. `internal/config/config_test.go`. NUNCA romper tests existentes; añadir nuevos para rail/theme/pane-nav.

**Target Platform**: Desktop Windows (build CGO msys64). App multiplaforma Fyne; validación principal en Windows.

**Project Type**: Desktop application (Fyne) — una sola app Go con GUI; estructura monorepo plano `internal/`.

**Performance Goals**: Abrir/cerrar panes <50ms; la barra NO scrollea (≤11 slots fijos). Interrupt Esc ≤500ms (regresión de 002). Aprobación p90 ≤2s. Sin impacto medible en render del transcript ventaneado.

**Constraints**: Conservar 100% de tests headless existentes verdes (no renombrar `viewStack`, mantener hijos tipados `*fyne.Container`). Keyboard-first (SC-009): el rail debe exponer foco nativo de `widget.Button`. Escavexex de ratón opcional. El colapso de rail se guarda por usuario. Acento LetsGO único (`#00ADD8`).

**Scale/Scope**: 1 pantalla shell + rail de ≤11 slots; 6 paneles re-encajados (settings/usage/git/mcp/plugins/tasks sobre índices fijos). Retoque visual de 5 superficies (approval, plan, chat bubbles, composer, palette). Sin features nuevas de agente.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

El `.specify/memory/constitution.md` es un template vacío (sin principios formales). Se aplican las normativas implícitas del repo (heredadas de 002): **test-first obligatorio**, **build CGO con `D:\msys64\ucrt64\bin` en PATH + `CGO_ENABLED=1`**, **specs bajo `specs/`**, **sin regresiones de la suite 002 (SC-001..010 preservados)**, **marcado de checklist al completar**.

- **Gate A (Test-first)**: Cada tarea: nuevo test → test falla → implementación → test verde. (PASS — patrón de 002.)
- **Gate B (Sin regresión)**: `go test ./...` verde antes/después, y 0 tests rotos del flujo 002. (PASS post-fase-1.)
- **Gate C (No regresión visual)**: dark-preserving sin romper; validación headless de tokens de tema. (PASS.)

## Project Structure

### Documentation (this feature)

```text
specs/003-gui-overhaul/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── gui-contract.md
├── checklists/
│   └── requirements.md
└── tasks.md             # (/speckit.tasks command output - NOT created here)
```

### Source Code (repository root)

```text
internal/
├── config/                    # viper config (statusline, auto_approve + NUEVOS rail/theme)
│   ├── config.go              # + Config.Rail{ Collapsed bool }, Config.ThemeVariant string
│   ├── config_test.go         # + tests rail.theme.
│   └── ...
├── gui/
│   ├── app.go                 # Controller: buildChat monta rail + root Border; showIndex con navHistory
│   ├── rail.go                # NUEVO: rail widget (rail Button items, grupos, icon-only/expanded)
│   ├── rail_test.go           # NUEVO: tests rail (colapso, slots, atajos, marca activa)
│   ├── theme.go               # ampliar: paleta DeepGoblue (3 niveles), font sans/mono split, railtema
│   ├── theme_test.go          # (existe) + checks de nuevos tokens
│   ├── keyboard.go            # + Alt+1..9 (rail), Alt+Left (back); cheatsheet `?` ampliado
│   ├── navigation.go          # NUEVO: showPane/paneByIndex/navHistory (Alt+Left/Esc-back)
│   ├── navigation_test.go     # NUEVO: tests de panorama/back
│   ├── settings.go            # + panel Appearance (theme, rail collapse), status checks (ya)
│   ├── ...
│   └── gui_test.go            # + tests de smoke del shell (existing tests intactos)
├── engine/                    # sin cambios funcionales (interfaces intactas)
└── ...
```

**Structure Decision**: Se extiende la estructura monoexistente. Todo el trabajo con el shell/rail/theme vive en `internal/gui/`, con la configuración y sus tests en `internal/config/`. El `viewStack` (índices 0=chat, 1=settings, 2=usage, 3=git, 4=mcp, 5=plugins, 6=tasks, + overlays) se conserva sin cambiar nombres para no romper `gui_test.go`; el rail se monta como borde izquierdo del stack a través de un nuevo contenedor raíz (`root = container.NewBorder(nil,nil,rail,nil,viewStack)`) que reemplaza `win.SetContent(c.viewStack)`.

## Complexity Tracking

> Constitution no define gates formales; no hay violaciones a justificar (repo plano, un solo módulo Go, sin capas extra). Se documenta en su lugar el riesgo máximo y su mitigación:

| Riesgo | Por qué existe | Mitigación |
|--------|----------------|------------|
| Refactor showIndex a navHistory accede `viewStack.Objects[i].(*fyne.Container)` | Los tests existentes asumen tipado `*fyne.Container` en el stack | Mantener el stack tal cual; el rail y el Border viven fuera; helper `contentAt(i)` |
| Colisión de atajos Alt vs Ctrl | los dominios existentes ya ocupan Ctrl+1..9 | rail usa `Alt+1..9`; `Alt+Left` back; cheat-sheet listado |
| Esc ambiguo (stream/paleta/vista) | el comportamiento cambia según widget enfocado | vuelta-vista solo push cuando índice activo != chat y ni palette ni picker abiertos |

## Phases

### Phase 0: Outline & Research

Investigación técnica consolidada (ver `research.md`):
- **Q1 — Estructura del rail**: `container.NewBorder(nil, nil, rail, nil, viewStack)`; botones `widget.Button` de importancia `LowImportance` con barra de acento izquierda (rect ocultable) — primera pasada sin BaseWidget custom (fase-2 polish). `rail` = `container.NewVBox` (grupo superior + separador + grupo inferior) con `MinSize` forzada. Colapsado = icono-only ~44–56px; expanded = icono+etiqueta ~200px; persistido en config.
- **Q2 — Vistas como panes**: las seis vistas ya son índices contiguos en `viewStack`; basta reutilizar showIndex + añadir navHistory (`[]int`) push/pop con `Alt+Left`/`Esc` (solo cuando índice activo y sin overlay). Overlays transitorios (palette, keymap, session picker, fork) permanecen como tal.
- **Q3 — Theme**: ampliar `terminalTheme.Color` con niveles de fondo (`ColorNameBackground`= abyss `#0A0E13`, `ColorNameButton/InputBackground/MenuBackground`= surface `#0F141B`, `ColorNameOverlayBackground`=raised `#161D26`, header=`#1E2733`, private `railBg`). Separador=`#242D38`, focus=`#00ADD8`. Sans font para UI + mono para Style{Monospace}. Importancia `High`/`Low` para unificar botones.
- **Q4 — Cuenta/Proveedores**: avatar (monograma GO sobre acento) + punto de estado en el pie del rail; abre panel in-pane (reuso settings provider section).
- **Q5 — Testing**: headless tests sobre `railView` (slots, colapso), `navHistory` (push/pop), `theme` (tokens nuevos), smoke Alt+1..9/Alt+Left, y regresión de la suite completa.
- **Q6 — Atajos**: `Alt+1..9` → showPane, `Alt+Left` → back; `?` cheat sheet crecer con categoría Rail; `Ctrl+K` intacto.

### Phase 1: Design & Contracts

- `data-model.md` — entidades RailSlot / PaneIndex / NavHistory / AppearancePrefs, con estados y reglas de validación (slot 1..11 fijos, atajo correlativo, rail nunca scrollea).
- `contracts/gui-contract.md` — contrato de la shell: funciones del Controller (`showPane`, `back`, `toggleRail`), eventos de cambio, e integración con app.go.
- `quickstart.md` — journeys (shell, rail-keyboard, checks de regresión), checklist SC-001..009 para verificación manual/testeable.
# Research: GUI Overhaul — Codex-style App Shell (Phase 0)

**Feature**: 003-gui-overhaul | **Date**: 2026-08-06

Base: investigación de diseño/UX/ingeniería sobre la GUI actual de LetsGO (Fyne v2.8.0, `internal/gui/`) y benchmark de apps de escritorio de agentes (OpenAI Codex GUI, ChatGPT desktop, Claude desktop 2026, Cursor/VS Code rail). Inputs de UI Designer, Desktop App Engineer y UX Researcher (verificado contra el código real y la API de Fyne v2.8 en module cache).

## 1. Problema y objetivo

- El usuario reporta que la GUI "no se ve nada appealing": estructura actual = menú superior + paneles overlay (viewStack de `container.NewStack`) sin barra lateral, tema plano monocromo `#0D0D0D`, tipografía mono en todo (incl. UI), sin jerarquía visual.
- Objetivo: shell "tipo Codex GUI" pero original — **barra vertical con trabajo arriba y cuentas/ajustes abajo**, paneles en-el-área, tema oscuro escalonado con acento LetsGO, keyboard-first preservado.

## 2. Decisiones de diseño (consolidadas)

### Q1 — Estructura del rail (barra vertical)

**Decision**: `container.NewBorder(nil, nil, rail, nil, viewStack)` — el rail es un `container.NewVBox` (grupo superior + separador hairline + grupo inferior) con `MinSize` forzada (~56px icon-only / ~200px expanded). Items = `widget.Button` `LowImportance` con barra de acento izquierda (`canvas.Rectangle` con `FillColor: AccentColor`, ocultable por slot activo).

**Rationale**: Border es el layout nativo "sidebar + content" de Fyne; no requiere layout custom. Botones nativos dan `fyne.Focusable` (Tab/Space/Enter), `desktop.Hoverable` y focus ring gratis — crítico para keyboard-first (SC-009). El `viewStack` existente (chat@0 … tasks@6 + overlays) queda intacto dentro del Border, así `showIndex` y los tests headless no cambian.

**Alternatives considered**: `widget.Toolbar` (horizontal, fondo sólido — pelea contra la verticalidad; descartado); `BaseWidget` custom con renderer propio (~150 líneas, fase-2 polish para hover/redondeo fino; primera pasada con botones nativos).

**Detalle colapso**: colapsado = solo íconos (44–56px); expanded = ícono+etiqueta (~200px); persistido en `config` (`rail.collapsed`). Tooltip siempre muestra "Etiqueta — Alt+n".

### Q2 — Vistas como paneles (no overlays)

**Decision**: Las seis vistas (settings/usage/git/mcp/plugins/tasks) ya viven en índices contiguos del viewStack; se convierten en **panes intercambiables** gobernados por el rail + `navHistory []int` en el Controller (`showPane(i)` hace push; `back()` pop con `Alt+Left` o `Esc` cuando el índice activo != chat y sin overlays abiertos).

**Rationale**: coherente con la filosofía no-modal de 002 (las aprobaciones inline nunca pelean con un modal); `Esc` → vuelve al chat con composer enfocado; paleta Ctrl+K salta a cualquier panel desde cualquier lugar.

**Alternatives considered**: mantener overlays y solo restilar (no cambia la sensación de estructura); modales de Fyne (tapan transcript, rompen flujo de teclado — descartado salvo confirmaciones destructivas: borrar sesión, revocar clave).

**Transitorios que PERMANECEN como overlays**: palette, keymap `?`, session picker, fork chooser.

### Q3 — Tema "Deep Goblue" (paleta escalonada + tipografía dual)

**Decision**: extender `terminalTheme` (theme.go) con:
- **3 niveles de fondo**: abyss `#0A0E13` (`ColorNameBackground`), surface `#0F141B` (`ColorNameButton`/`InputBackground`/`MenuBackground`), raised `#161D26` (`ColorNameOverlayBackground`), well `#1E2733` (`ColorNameHeaderBackground` + hover). Separador/borde `#242D38` (`ColorNameSeparator`/`Shadow`), input-border explícito `ColorNameInputBorder`.
- **Acento único**: Goblue `#00ADD8` ya existe (`AccentColor` + `accentThemeColor`); se reserva `ColorNamePrimary` `#4FC1FF` para acciones primarias; foco = acento. Semánticos: success `#2EA043`, warning `#D29922`, error `#F85149` (estados únicamente).
- **Tipografía dual**: `Theme.Font(style)` devuelve la face sans embebida (p. ej. Inter) para `Style{}` y Cascadia Mono (ya embebido) para `Style{Monospace:true}`. Mono se reserva a status line, código, payloads de herramientas, composer; sans para chrome/etiquetas/formularios.
- **Importancia unificada**: botones primarios `HighImportance`, secundarios `LowImportance`; `canvas.Rectangle` de acento para firmas (barra izquierda de approval, marca activa del rail) — Fyne NO tiene StyleBox; los buckets son colores de tema + importancia.

**Rationale**: luminosidad escalonada + hairlines = elevación en dark (sin sombras); acento único en momentos firmados = identidad LetsGO sin clonar Codex; foco visible = keyboard-first (SC-009).

**Caveats verificados en Fyne v2.8**: `fyne.IconThemeColorName` NO existe; íconos por `ThemeIconName` con `theme.NewColoredResource(icon, color)` para estado activo. `ThemeVariant` no debe cambiar por detección del SO (ambos variantes consistentes). Fyne no tiene StyleBox.

### Q4 — Cuenta/Proveedores en el pie del rail

**Decision**: avatar = monograma "GO" sobre círculo acento + punto de estado (verde/gris/rojo según proveedor activo), anclado en el borde inferior del rail; al pulsarlo abre el panel Cuenta (sección providers existente de settings, refactorizada a pane propio o reutilizada in-pane).

**Rationale**: convención "identidad en el pie" (VS Code, Slack, Codex) — estable y fácil de alcanzar; el estado de conexión del proveedor se ve de un vistazo.

### Q5 — Testing

**Decision**: headless `test.NewApp()`:
- `rail_test.go`: slots/colapso (MinSize, etiquetas visibles/ocultas), marca activa por índice, onTapped → callback showPane, tooltips.
- `navigation_test.go`: navHistory push/pop, `back()` con Esc/Alt+Left, guardas (no-pop cuando chat activo o overlay abierto).
- `theme_test.go`: tokens nuevos (tres niveles distintos, acento, separador, foco) en ambos variantes.
- `gui_test.go` (añadidos): smoke Alt+1..9 / Alt+Left con el shell montado; regresión: la suite completa de 002 debe seguir verde (los 60+ tests existentes no se tocan).
- Config: `config_test.go` para `rail.collapsed` y `theme` persistencia (round-trip viper).

**Patrón de verificación visual**: screenshots headless a resolución fija antes/después (opcional, `test.NewWindow` + `test.AssertRendersMarkdown` no aplica a canvas; se usan asserts de estructura/colores).

### Q6 — Atajos

**Decision**: `Alt+1..9` → slot del rail (`showPane`), `Alt+Left` → `back()`. `Ctrl+K` paleta, `Ctrl+1..9` sesiones, `Ctrl+,` settings, `Ctrl+U` usage, `Ctrl+L` focus, `Ctrl+Shift+A` auto-approve quedan INTACTOS. Cheat sheet `?` (keymap.go) crece con categoría "Rail" (Alt+n, Alt+Left) y "Vistas".

**Rationale**: Alt+ es un modificador libre (no usado hoy); evita colisión con Ctrl+1..9 (sesiones). Documentado en `keymapCatalog` + tooltips del rail.

## 3. Integración con el código actual

| Hoy | Objetivo | Fichero |
|------|----------|---------|
| `win.SetContent(c.viewStack)` | `c.root = container.NewBorder(nil,nil,rail,nil,viewStack)`; `SetContent(c.root)` | app.go |
| showIndex 0..6 | showPane(i) + navHistory push; back() | app.go + navigation.go (nuevo) |
| theme plano mono | paleta DeepGoblue + font split | theme.go |
| keyboard Ctrl-only | + Alt+1..9 / Alt+Left | keyboard.go |
| overlays settings/usage/git/mcp/plugins/tasks | panes del rail (mismos índices) | rail.go (nuevo) |
| settings genérico | + sección Appearance (theme claro/oscuro, colapso rail) | settings.go |
| keymapCatalog | + categoría Rail/Vistas | keymap.go |
| config | + `rail.collapsed`, `theme` | config.go + test |

## 4. Riesgos y mitigaciones

1. **Panic por type-assert en showIndex** (`obj.(*fyne.Container)`) si se cambia el tipo de hijos del stack → rail y Border FUERA del stack; helper `contentAt(i)`.
2. **Colisión de atajos** → Alt+ para rail; cheatsheet listado; regresión por smoke.
3. **Esc ambiguo** → regla "el widget enfocado decide"; back solo si pane activo y sin overlays.
4. **Tema: variante del SO** → ambos variantes consistentes; `ThemeVariant` establecido explícitamente por app (`Settings().SetThemeVariant`).
5. **Lista virtualizada (palette/sessions)** → restyle en template único + `Refresh()`; `HideSeparators` en v2.8.
6. **Foco con nav custom** → botones nativos (Tab/Alt navegan); foco inicial del pane por hook; vuelta al composer.

## 5. Fuera de alcance (confirmado)

Motor de conversación, API/providers, features de agente, análisis de uso. Solo presentación/estructura.

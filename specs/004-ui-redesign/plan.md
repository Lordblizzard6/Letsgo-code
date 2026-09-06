# Implementation Plan: UI/UX Refinement — Consistency, Legibility & Menu Architecture

**Branch**: `004-ui-redesign` | **Date**: 2026-08-07 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/004-ui-redesign/spec.md`

**Note**: Plan generado por `/speckit.plan`. Se construye sobre el shell heredado de `003-gui-overhaul`.

## Summary

Rediseño de presentación de la GUI de LetsGO (sin cambios funcionales del agente):

1. **Rail coherente** — tres zonas separadas por hairline (trabajo, herramientas con etiqueta, sistema anclado abajo), un destino único por slot, eliminación del slot `sessions` duplicado (mapeaba al pane 0 igual que `chat`), re-mapeo de atajos `Alt`, tooltips obligatorios y foco visible en modo colapsado.
2. **Profundidad correcta de superficies** — Settings como superficie no modal con secciones (Cuenta, Apariencia, Preferencias, Uso); Cuenta como flyout del avatar (estado, provider/model, resumen de uso, acciones); Uso como dashboard de tarjetas KPI.
3. **Legibilidad y elegancia** — medida de mensaje ~720px, marco único de tarjeta (border 1px, radius 6px, padding 16px), rampa tipográfica, focus ring 2px, Enviar/Detener en un mismo slot, microcopy en español.
4. **Accesibilidad y consistencia** — contraste AA (fix placeholder claro ~2.7:1), separadores >=3:1, iconos SVG semánticos con fallback, estados vacíos con CTA y refresco conservador.

Todo se apoya en la arquitectura de 003 (viewStack + railView + terminalTheme) sin tocar el modelo de datos ni el motor de streaming.

## Technical Context

**Language/Version**: Go 1.22+ (toolchain actual). Build CGO en Windows: `$env:PATH = "D:\msys64\ucrt64\bin;" + $env:PATH; $env:CGO_ENABLED = "1"`; `build.bat -n`.

**Primary Dependencies**: Fyne v2.8.0 (`fyne.io/fyne/v2`) con `driver/desktop`. Tema propio `terminalTheme` (ya ampliado en 003: tokens Deep Goblue dark y light). Iconos Fyne `theme.IconName*` con plan SVG semántico (fallback genérico + tooltip). Mono Cascadia ya embebido; la fuente sans usa la face del tema (se ajusta el tamaño de ramp, no se añaden fuentes nuevas).

**Storage**: `internal/config` (viper, `~/.letsGo/config.yaml`). US2 no exige nuevos campos persistidos: se reúsan `ThemeVariant` y `Rail.Collapsed`. El uso conserva el período actual (`startOfToday`). Sin migraciones de datos.

**Testing**: `go test ./...` headless con `test.NewApp()` (patrón 002/003). Suite `internal/gui/*_test.go` existente (gui_test, rail_test, navigation_test, theme_test, shortcuts_test, app_shell_test) mas tests nuevos: zonas del rail sin duplicados, tooltips y colapso con foco, settings por secciones, flyout de cuenta, dashboard de uso, CI de contraste de paleta, estados vacíos, y smoke de montaje completo. **Nunca romper tests existentes** (superficies preservadas en contracts/gui-contract.md de 002/003).

**Target Platform**: Desktop Windows (build CGO msys64). App multiplataforma Fyne; validación principal en Windows.

**Project Type**: Desktop application (Fyne), un solo módulo Go, estructura plana `internal/`.

**Performance Goals**: Abrir/cerrar panes <50ms. Rail fijo sin scroll (hasta 8 slots + pie). Transiciones <=250ms o sin animación con reduced-motion. Aprobación p90 <=2s (regresión preservada). Sin impacto medible en el render del transcript.

**Constraints**:

- `viewStack` con índices fijos 0=chat, 1=settings, 2=usage, 3=git, 4=mcp, 5=plugins, 6=tasks; overlays después. NO renombrar `viewStack` (los tests dependen).
- Rail y BorderRoot viven FUERA del `viewStack` (003). El bug de `slotForPane` con pane 0 duplicado se resuelve quitando el slot `sessions` del rail (US1), no moviendo el stack.
- Atajos invariables: `Ctrl+1..9` sesiones, `Ctrl+N`, `Ctrl+,`, `Ctrl+U`, `Ctrl+L`, `Ctrl+K`, `Shift+/`, paleta. `Alt+1..8` se re-mapean a la nueva composición (Chat, Git, Tareas, MCP, Plugins, Configuración, Uso, Ayuda). `Alt+Left`/`Esc` vuelven a chat (003) sin cambios.
- Keyboard-first (SC-003): foco nativo de `widget.Button`/`Entry`; ring accent visible siempre.
- Tokens de color: NO hardcodear `dark*` en widgets compartidos light/dark (ring del composer, avatar de perfil). Todo pasa por `terminalTheme`. Smoke test de paleta en CI.
- `lightPlaceholder` debe oscurecerse para AA (objetivo ~`#5C6B7A`). Los separadores se ajustan para >=3:1.
- Sincronizar keymap (`?`) y paleta (`Ctrl+K`) con la nueva composición de slots, nomenclatura y atajos.

**Scale/Scope**: 1 rail (8-9 slots + pie) · 1 Settings con pestañas · 1 flyout Cuenta · 1 dashboard Uso · medida/rampa tipográfica · suite de empty/loading · pack de iconos SVG (fallback) · CI de paleta. Sin funcionalidad nueva de agente.

## Constitution Check

*GATE: Debe pasar antes de Fase 0. Se re-checa tras Fase 1.*

El `.specify/memory/constitution.md` sigue siendo un template vacío (sin principios formales firmados). Como en 003, se aplican las normas implícitas heredadas de 002: **test-first** (test pasan después de implementar), **build CGO** con PATH msys64 + `CGO_ENABLED=1`, **docs bajo `specs/004-*`**, **cero regresiones** de 002/003, y **marcado de checklist/SC** en `quickstart.md` al completar.

- **Gate A (Test-first)**: cada tarea añade un test headless que falla, luego implementa, luego verde. PASS (patrón 002/003).
- **Gate B (Sin regresión)**: `go test ./...` verde antes y después; 0 tests rotos del flujo 002/003. PASS post-diseño (superficies 002/003 preservadas en contracts/ui-contract.md §1.1).
- **Gate C (Contraste AA automatizable)**: smoke de paleta en CI — si un token de texto en cualquier variante queda <4.5:1 o un no-texto <3:1, falla. PASS post-diseño (solo `lightPlaceholder` y boundary de `darkBorder` requieren ajuste).

**Re-chequeo post-Fase 1**: Confirmado, sin violaciones. research.md resuelve los unknowns del Technical Context; data-model.md + contracts/ui-contract.md no añaden datos persistidos ni rompen atajos/índices.

## Project Structure

### Documentation (this feature)

```text
specs/004-ui-redesign/
├── plan.md              # Este archivo (output /speckit.plan)
├── research.md          # Fase 0 (unknowns + decisiones)
├── data-model.md        # Fase 1 (entidades, sin datos nuevos)
├── quickstart.md        # Fase 1 (journeys + checklist SC-001..007)
├── contracts/
│   └── ui-contract.md   # Fase 1 (superficies nuevas y preservadas)
└── tasks.md             # Fase 2 (/speckit.tasks — no lo genera plan)
```

### Source Code (repository root)

```text
internal/
├── config/
│   ├── config.go        # reuso: ThemeVariant, Rail.Collapsed (003)
│   └── config_test.go
├── gui/
│   ├── app.go           # cableado: showSettings→overlay tabs, showAccount→flyout, focus-return
│   ├── rail.go          # refactor: zonas, sin slot sessions, mismo paneIndex unico, Alt remap
│   ├── rail_test.go     # + zonificación, sin duplicados, tooltips, focus colapsado
│   ├── tooltips.go      # NUEVO: helper tooltip(label + acceso)
│   ├── settings.go      # + selector de secciones (TabContainer) [Cuenta|Apariencia|Preferencias|Uso]
│   ├── settings_test.go # + secciones render, save/cancel, focus-return
│   ├── account.go       # NUEVO: flyout de la cuenta (estado, provider/model, uso, acciones)
│   ├── account_test.go  # NUEVO
│   ├── usage.go         # + dashboard KPI (tarjetas + tabla + barra) sin flat mono dump
│   ├── usage_test.go    # + KPI por encima de String()
│   ├── theme.go         # + lightPlaceholder fix, separador token para >=3:1
│   ├── theme_test.go    # + palette smoke (ambas variantes)
│   ├── message.go       # + columna max 720px
│   ├── composer.go      # + ring 2px, Enviar<->Detener mismo slot, placeholder ES
│   ├── approvalcard.go  # + marco común 1px/6px/16px, tooltip teclas
│   ├── icons.go         # NUEVO: iconos SVG semánticos con fallback
│   ├── empty.go         # NUEVO: EmptyState(msg,CTA) / LoadingState(keep+Refreshing)
│   ├── keyboard.go      # + Alt remap (sin sessions), clave de colisión
│   ├── keymap.go        # + categoría Rail actualizada
│   ├── palette.go       # + términos sincronizados (Chat, Configuración)
│   ├── navigation.go    # + showPane sin duplicar slot
│   ├── chatview.go      # + placeholders ES
│   ├── sessions.go      # + "Filtrar conversaciones…"
│   ├── mcpview.go       # + empty state MCP
│   ├── pluginsview.go   # + empty state plugins
│   ├── agenttasks.go    # + empty state tasks
│   ├── gitview.go       # + empty state git
│   ├── usageempty.go    # + empty state uso
│   ├── empty_test.go
│   └── shell_smoke_test.go # NUEVO: montar shell completo
├── engine/
└── tools/              # sin cambios de contrato (CostTracker reuso en usage)
```

**Structure Decision**: Se extiende la estructura plana existente. Todo el trabajo vive en `internal/gui/` (más ajustes de tema); `internal/config` solo reusa claves. Contrato documentado en `contracts/ui-contract.md`. Cada superficie (rail, settings, account, usage) es un widget headless testable cableado por el Controller (patrón 003). `viewStack` y sus índices NO cambian.

## Complexity Tracking

> No hay gates formales; no hay violaciones. Se listan riesgos y mitigaciones (igual 003).

| Riesgo | Por qué existe | Mitigación |
|--------|----------------|------------|
| Eliminar `sessions` del rail rompe tests que esperan 11 slots o Alt+2 | `scope rail_test.go` 003 hardcodea counts/mapeo | Actualizar rail_test para nueva composición; mapear Alt por `slot.shortcut`, no por índice literal |
| Settings con secciones rompe tests 002 de `showSettings` | Tests llaman `showSettings()` y asumen superficie in-pane | Mantener firma `showSettings()`; debajo muestra overlay con tabs; `Ctrl+,` invariante |
| Flyout de cuenta desde el avatar es un popover nuevo en Fyne | No se usa `widget.NewPopUp` hasta ahora; el anclaje cambia con el tema | `widget.NewPopUp(content, canvas)` con ancla al avatar; fallback in-pane si la ventana es pequeña |
| Usage KPI cambia el texto `usageView` que depende del dump | tests 003 pueden leer `.text.Text` | Mantener `usageSource`/`aggregateUsage`; solo cambia presentación; actualizar usage_test conservando invariantes de stats |
| Contraste: `lightPlaceholder` falla AA (2.76:1) | Identificado en research | Cambiar a ~`#5C6B7A`, add palette test |
| Alt remap colisiona con atajos existentes | `shortcuts_test.go` los registra hardcoded | Actualizar el catálogo de atajos y coordinarlo con el rail |
| `sessions` embebida en chat se conserva | US1 solo la quita del rail, no del lateral de la conversación | Revisar `sessionpicker.go` intacto; el contrato lo documenta |

## Phases

### Fase 0: Outline & Research

- **Q1 — Sustituir slot `sessions`**: `railSlots` pasa de 11 a 9-10: `chat` (pane 0) único dueño; eliminar `sessions`. Re-map Alt: Alt+1 Chat, Alt+2 Git, Alt+3 Tareas, Alt+4 MCP, Alt+5 Plugins, Alt+6 Configuración, Alt+7 Uso, Alt+8 Ayuda. Zona herramientas con etiqueta "HERRAMIENTAS"; sistema/pie: Tema, Configuración, Uso, Ayuda, avatar, colapso.
- **Q2 — Settings por secciones**: `container.NewAppTabs`/`TabContainer` con [Cuenta, Apariencia, Preferencias, Uso]. Revisar el patrón de superposición no modal (overlay de tamaño medio ~620x520 sobre el viewStack) que no desmonte la conversación: `widget.NewPopUp` con TabContainer, `Esc` cierra y enfoca el composer. Si el popup plantea problemas de layout en tests headless, un pane tabbed index 7 vía `showPane` (perdiendo compatibilidad visual FR-004) — se decide en `research.md`.
- **Q3 — Flyout Cuenta**: `widget.NewPopUp(content, canvas)` posicionado en el avatar; contenido: dot de estado, provider/model, resumen de uso del día, acciones ("API keys & providers…", "Uso…", "Cerrar sesión"). `Esc` cierra → `focusInput()`.
- **Q4 — Usage KPI**: conservar `usageSource` y `aggregateUsage`; nuevas superficies: `container.NewGridWithColumns(3)` con KPI cards, `widget.NewTable` per provider, barra de presupuesto (canvas sobre well). `usageHistogram.String` queda como debug/fallback.
- **Q5 — Empty states**: helper reutilizable `emptyState(msg, ctaText, onCTA)` y `loadingState(content, refreshing)`. Aplicado a sessions, mcp, plugins, tasks, git, usage. Los refrescos mantienen la vista previa + "Refrescando…" si ya hay datos.
- **Q6 — Iconos semánticos**: paquete `icons.go` con `iconFor(slotID) fyne.Resource`. SVG embebido (no base64). Fallback: `theme.Icon(slot.icon)` actual si el SVG no existe; tooltip obligatorio.
- **Q7 — Paleta CI**: `theme_test.go` itera pares definidos (texto y no-texto en light y dark) y computa contraste >=4.5/>=3. Fix: `lightPlaceholder` — `#5C6B7A`; `darkBorder` — si faltara para >=3:1.
- **Q8 — Tests de humo**: montar shell (rail + chat) → abrir Settings → flyout cuenta → usage → back → `focusInput()`. Suite completa 002/003.

### Fase 1: Design & Contracts

- `data-model.md` — entidades: RailSlot (zona, label ES, shortcut, destino único), SettingsTab, AccountFlyout, UsageKPI, EmptyState/LoadingState; reglas e invariantes (sin duplicados paneIndex, atajos Alt 1..8, ring accent, contraste).
- `contracts/ui-contract.md` — superficies nuevas (`showSettings` overlay tabbed, `showAccount` flyout, `showUsage` dashboard) + preservación de la superficie 002/003 (mismas firmas), invariantes de stack y focus-return.
- `quickstart.md` — 4 journeys de validación (rail coherente; settings/flyout; legibilidad; accesibilidad) y checklist SC-001..007.
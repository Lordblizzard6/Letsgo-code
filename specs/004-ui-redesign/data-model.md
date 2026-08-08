# Data Model: UI/UX Refinement

**Created**: 2026-08-07 | **Status**: Fase 1 — completo

## Resumen

Feature de presentación: el rail, settings, cuenta, uso, tipografía y estados se rediseñan **sin cambios de datos persistidos ni de telemetría**. Se reutilizan las estructuras existentes de `internal/config`, `internal/tools.CostTracker` y `viewStack`.

## Entidades

### RailSlot (rail.go)
- Campos: `id` (string estable), `labelES` (string), `icon` (glyph semántico con fallback `theme.Icon`), `shortcut` (string, `Alt+N`), `paneIndex` (int, -1 para acción), `zone` (work | tools | system).
- Reglas:
  - Cada índice de pane se asigna a UN solo slot (invariante "un destino por slot"). El slot `sessions` se ELIMINA del rail.
  - Orden de zonas: trabajo → herramientas (etiqueta "HERRAMIENTAS") → sistema (anclada abajo).
  - Los atajos `Alt+1..8` se leen del catálogo; no hay hardcode.

| id | labelES | shortcut | pane | zone |
|----|---------|----------|------|------|
| chat | Chat | Alt+1 | 0 | work |
| git | Git | Alt+2 | 3 | tools |
| tasks | Tareas | Alt+3 | 6 | tools |
| mcp | MCP | Alt+4 | 4 | tools |
| plugins | Plugins | Alt+5 | 5 | tools |
| settings | Configuración | Alt+6 | 1 | system |
| usage | Uso | Alt+7 | 2 | system |
| help | Ayuda | Alt+8 | -1 | system |
| theme | (icono Tema) | — | -1 | system |
| account | (avatar) | — | -1 (flyout) | system |

> Nota: la numeración final de `Alt+6`/`Alt+7` entre Configuración and Uso se fija aquí (6 Configuración, 7 Uso) como fuente de verdad; el catálogo de `railSlots` en código la reflejará igual, y `Ctrl+,` / `Ctrl+U` siguen como accesos secundarios invariantes.

## SettingsView (presentación)

- Secciones fijas: `cuenta`, `apariencia`, `preferencias`, `uso`.
- Reutiliza claves de config existentes (`ThemeVariant`, `Rail.Collapsed`, etc.); no añade claves nuevas.
- Estado del overlay: `visible bool`; guardar (Save) o descartar (Esc/Cancel) con `config.SaveConfig()`.
- Transición: mostrar → editar → save/cancel → `focusInput()`.

## AccountFlyout (presentación)

- Lectura: config actual (provider/model), `usageSource` (resumen de hoy), estado dot (derivado de `hasAnyKey`/engine).
- Acciones: `onKeys` (Settings · Cuenta), `onUsage` (Uso), `onSignOut`.
- Ciclo: `show()` (popup anclado al avatar) → interacción → `hide()` + `focusInput()`.

## UsageKPI (presentación)

- Fuente: `usageSource` / `aggregateUsage` / `startOfToday` (existentes).
- Superficie: 3 tarjetas KPI (coste hoy, requests, tokens) + tabla por provider + barra de presupuesto %.
- `usageHistogram.String()` sigue como fallback/debug; NO es superficie principal.

## EmptyState / LoadingState (presentación)

- `emptyState(msg, ctaText, onCTA)` y `loadingState(content, refreshing)`.
- Aplican a 6 superficies: sessions, mcp, plugins, tasks, git, usage.
- Regla: el refresh sobre datos previos conserva el contenido + "Refrescando…", el vacío solo cuando el resultado definitivo es vacío.

## Reglas de validación (desde spec)

- FR-003: tooltip obligatorio en colapsado con nombre y atajo.
- FR-004: overlay no modal; conversación visible detrás; focus-return.
- FR-002: no hay duplicados de destino en el rail.
- FR-013: contraste >=4.5 (texto) / >=3 (no-texto) en ambos themes (CI).

## Estado de persistencia

No hay migraciones, ni claves nuevas, ni cambios de esquema. Telemetría (`CostTracker`) sin cambios; solo presentación.
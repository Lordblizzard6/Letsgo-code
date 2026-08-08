# UI Contract: UI/UX Refinement (rail zones, settings tabs, account flyout, usage KPI)

**Feature**: 004-ui-redesign | **Date**: 2026-08-07

Este contrato complementa `specs/002-codex-style-ui/contracts/gui-contract.md` y `specs/003-gui-overhaul/contracts/gui-contract.md` **sin romperlos** (regla de compatibilidad 002/003). Define las superficies nuevas de este feature y las invariantes que deben mantener los tests.

## 1. Superficie del Controller (paquete `internal/gui`)

### 1.1 Firmas preservadas (002/003) — SIN cambio de firma ni comportamiento externo

- `showChat()`, `showSettings()`, `showUsage()`, `showGit()`, `showMCP()`, `showPlugins()`, `showTasks()`, `showAccount()`, `focusInput()`, `openPalette()`, `closePalette()`, `openKeymap()`, `closeKeymap()`, `showSessionPicker()`, `switchNth(n)`, `newSession()`, `back()`, `toggleRail()`, `toggleThemeVariant()`, `applyTheme()`, `showPane(i int)`.

### 1.2 Comportamientos redefinidos (este feature)

| Función | Comportamiento nuevo | Invariante |
|---------|----------------------|-----------|
| `showSettings()` | Abre un **overlay no modal ~620x520** con `widget.TabContainer` (4 secciones: Cuenta · Apariencia · Preferencias · Uso). `Esc`/Cancelar cierra y llama `focusInput()`. | El `viewStack` NO se desmonta; la conversación sigue visible tras el overlay. `Ctrl+,` mantiene su atajo. Fallback si el popup no rinde headless: pane tabbed índice 7 vía `showPane(7)` (documentado en research.md). |
| `showAccount()` | Abre el **flyout de cuenta** (`widget.NewPopUp` anclado al avatar): dot de estado, provider+model, resumen de uso de hoy, acciones "API keys & proveedores…", "Uso…", "Cerrar sesión". Ya NO llama a `showSettings()`. | `Esc` cierra el flyout y ejecuta `focusInput()`. |
| `showUsage()` | Muestra el **dashboard KPI** del panel Uso (mantiene índice 2). | `usageSource`/`aggregateUsage` intactos; presentación nueva, encima de datos sin cambios. |
| `showHelp()` | Abre keymap (`?`): categoría Rail actualizada (labels ES + nuevos atajos Alt). | Sin cambio de mecanismo (overlay keymap). |
| `showPane(i)` | Sin duplicidad: si dos slots apuntan al mismo índice (CASO `sessions`==`chat`==0) se resuelve eliminando el slot duplicado; `slotForPane` solo puede devolver UN slot por índice. | `back()`/`Esc`/`Alt+Left` inalterados. |

### 1.3 Rails zones (rail.go)

Orden nuevo (zona / id / label ES / shortcut / paneIndex):

- **Zona trabajo**: `chat` · "Chat" · `Alt+1` · pane 0
- **Zona herramientas** (etiqueta "HERRAMIENTAS"):
  - `git` · "Git" · `Alt+2` · pane 3
  - `tasks` · "Tareas" · `Alt+3` · pane 6
  - `mcp` · "MCP" · `Alt+4` · pane 4
  - `plugins` · "Plugins" · `Alt+5` · pane 5
- **Zona sistema** (anclada abajo):
  - `settings` · "Configuración" · `Alt+6` · pane 1
  - `usage` · "Uso" · `Alt+7` · pane 2
  - `help` · "Ayuda" · `Alt+8` · pane -1 (acción → keymap)
  - `theme` · (icono Tema) · sin shortcut · pane -1 (acción toggle)
  - `account` · (avatar, pie) · sin shortcut · pane -1 (acción → flyout)
  - colapso (toggle) en el pie, sin slot position 1.

**Eliminado del rail**: `sessions` (pane 0 duplicado). La funcionalidad de sesión/la lista de conversaciones permanece embebida en la vista Chat (sessionpicker intacto).

**Zonas**: separadas por `widget.Separator` (hairline). Modo colapsado: labels ocultos, `MinSize` =44px, tooltips obligatorios (`"Git (Alt+2)"`), focus ring visible.

## 2. Settings overlay (settings.go)

- `settingsView` se agrupa en `widget.TabContainer` (o `container.NewAppTabs`): `Cuenta`, `Apariencia`, `Preferencias`, `Uso`.
- **Apariencia** conserva: tema (Oscuro/Claro), rail colapsado, y un control de densidad opcional.
- **Cuenta/Uso**: enlaces hacia el flyout y el dashboard respetando config persistida anterior.
- Guarda/Cancela: `Esc` y botón Cancelar cierran sin aplicar cambios pendientes; `Save` persiste en bloque con `config.SaveConfig()`.
- **Foco**: en cierre (Save o Cancel) → `focusInput()`.

## 3. Account flyout (account.go)

```go
type accountFlyout struct {
    onKeys    func() // abre Settings → Cuenta
    onUsage   func() // abre Uso
    onSignOut func()
    show()    // widget.NewPopUp anclado al avatar
    hide()
}
```

Contenido: dot de estado (online/idle/error), provider+model, `$costeHoy · req hoy`, acciones. Sin datos guardados propios; lee de config/provider actual y de `usageSource`.

## 4. Usage KPI (usage.go)

- Mantiene `usageSource`, `startOfToday`, `aggregateUsage`.
- Superficie: `GridWithColumns(3)` tarjetas (coste de hoy / requests / tokens) + `widget.NewTable` per provider + barra de presupuesto porcentual (opcional si hay cuota).
- `usageHistogram.String()` SIEMPRE disponible como fallback/debug (uso en tests).
- No persiste estado; se refresca al entrar (igual 003).

## 5. Empty / Loading (empty.go)

```go
func emptyState(msg string, ctaText string, onCTA func()) fyne.CanvasObject
func loadingState(content fyne.CanvasObject, refreshing bool) fyne.CanvasObject
```

- Aplicado a: sessions list, mcp, plugins, tasks, git, usage.
- Refresco asíncrono: contenido previo + indicador "Refrescando…" (nunca vaciar el panel en blanco durante la recarga).
- Estados vacíos con mensaje + CTA (p.ej. "No hay conversaciones todavía." / "Nueva conversación").

## 6. Iconos semánticos (icons.go)

```go
func iconFor(id string) fyne.Resource // SVG embebido o fallback theme.Icon(name)
```

Mapeo: chat→burbuja, git→branch, tasks→checklist, mcp→servidor/cajas, plugins→puzzle, settings→gear, usage→gauge, theme→sun/moon (según variante), help→signo de pregunta. Fallback si no hay SVG: icono Fyne actual. Tooltip obligatorio en colapsado (label + shortcut).

## 7. Tema / contraste (theme.go)

- `lightPlaceholder` `#8A94A0` → ~`#5C6B7A` (cumple 4.5:1 con fondo claro).
- Separador `darkBorder` boundary lighten (→~`#2E3A48` si hace falta) para >=3:1 no-text.
- Ningún widget hardcodea colores en widgets compartidos light/dark (ring del composer y avatar deben tomar `theme.FocusColor()`, `theme.ForegroundColor()` y `ColorBackground` por surface).
- **Palette smoke test**: función en `theme_test.go` que itera los pares base (fondo/foreground secundario/accent/placeholder) en dark+light y falla si texto <4.5:1 o no-texto <3:1.

## 8. Teclado (keyboard.go / keymap.go)

| Atajo | Acción | Notas |
|-------|--------|-------|
| `Alt+1` | Chat | pane 0 |
| `Alt+2` | Git | pane 3 |
| `Alt+3` | Tareas | pane 6 |
| `Alt+4` | MCP | pane 4 |
| `Alt+5` | Plugins | pane 5 |
| `Alt+6` | Configuración | pane 1 (overlay) |
| `Alt+7` | Uso | pane 2 |
| `Alt+8` | Ayuda | action keymap |
| `Ctrl+,` / `Ctrl+U` | settings / usage | intactos |
| `Ctrl+K` paleta, `Ctrl+L` focus, `?` keymap, `Ctrl+1..9` sesiones | intactos | sin colisión |
| `Alt+Left` / `Esc` | back() | intactos (guarda overlay/foco) |

**Nota**: los atajos `Alt+1..8` se leen del catálogo de slots (NO hardcodear en test de reproducción); `shortcuts_test.go` se actualiza una sola vez.

## 9. Visual y estado (US3)

- Mensajes en prosa: columna max ~720px centrada; código ancho completo con wrap opcional.
- Tarjetas (usuario/asistente, aprobación, plan, tool panel): marco común border 1px, radius 6px, padding 16px.
- Composer: ring accent 2px en focus; `Enviar`↔`Detener` en el mismo slot durante streaming.
- Microcopy: chrome en español, términos técnicos en inglés; placeholders que describen acción.

## 10. Estado vacíos / carga (aplicados por superficie)

| Superficie | Vacío | CTA |
|-----------|-------|-----|
| Conversaciones (sidebar) | "No hay conversaciones todavía." | "Nueva conversación" |
| MCP | "No hay servidores MCP configurados." | "Añadir servidor" |
| Plugins | "No hay plugins instalados." | "Instalar…" |
| Tareas | "Todavía no hay tareas." | "Nueva tarea" |
| Git | "No hay repositorio." | "Abrir carpeta…" |
| Uso | "Sin actividad registrada hoy." | — |

## 11. Tests obligatorios (contrato verificado)

- `TestRailZonesAndSingleDestination`: 3 zonas, sin panes duplicados, atajos Alt correctos.
- `TestRailTooltipCollapsed`: tooltip presente en colapsado con nombre+acceso; focus visible.
- `TestSettingsTabsRender`: el overlay de settings muestra 4 secciones; save/cancel cierran y focusInput.
- `TestAccountFlyout`: clic avatar abre popup; Esc cierra y enfoca composer; onUsage/onKeys enrutan.
- `TestUsageKPISurfaces`: dashboard construido sin depender de `usageHistogram.String` (existe texto KPI).
- `TestPaletteContrast`: pares (texto>=4.5, no-texto>=3) en ambos themes — FAIL si token rompe (palette smoke test).
- `TestEmptyStatesAllSurfaces`: 6 superficies muestran empty/CTA; refresco mantiene vista previa.
- `TestMessageMeasure`: columna prosa <=720px (headless, measurable).
- `TestComposerSwapSendStop`: el mismo slot cambia en streaming.
- **Regresión**: NINGÚN test de 002/003 se rompe (superficie 002/003 preservadas). Actualizar solo tests que agreguen asserts explícitos (p.ej. `rail_test` a la nueva lista u offset del catálogo).

## 12. Semantics de cierre

- `Esc`: cierra overlay settings / flyout / keymap / palette / picker con prioridad de foco; si no overlay, `back()` para pane != chat; en composer, cancela stream. Sin cambio (003).
- Al cerrar cualquier superficie no modal: `focusInput()` (composer).
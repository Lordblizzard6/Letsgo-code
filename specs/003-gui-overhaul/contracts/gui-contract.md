# GUI Contract: Codex-style App Shell (rail + panes + theme)

**Feature**: 003-gui-overhaul | **Date**: 2026-08-06

Este contrato documenta las superficies externas de la shell: qué funciones expone el Controller, qué eventos consume, qué atajos quedan registrados y qué invariantes deben mantener los tests. Complementa el contrato 002 (`specs/002-codex-style-ui/contracts/gui-contract.md`) sin romperlo: las funciones de 002 (`showSettings`, `showUsage`, `showGit`, `showMCP`, `showPlugins`, `showTasks`, `showChat`, `focusInput`, `openPalette`, `openKeymap`, `switchNth`, `newSession`) permanecen con la MISMA firma y comportamiento, pero ahora se renderizan como panes del rail.

## 1. Controller surface (paquete `internal/gui`)

### 1.1 Nuevas funciones

| Función | Firma | Contrato |
|---------|-------|----------|
| `showPane(i int)` | `func (c *Controller) showPane(i int)` | Muestra el pane `i` del viewStack y hace push de `i` anterior a `navHistory` (si `i != Current` y `i` es pane, 0..6). Marca activo el slot del rail correspondiente. No-op si viewStack nil. |
| `back()` | `func (c *Controller) back()` | Vuelve al pane anterior: pop `navHistory` si `Current != 0` y no hay overlay (palette/picker/keymap) abierto; si la pila está vacía o `Current == 0`, no-op. Al volver, foco al composer del chat si el destino es 0. |
| `toggleRail()` | `func (c *Controller) toggleRail()` | Alterna expandido↔colapsado del rail y persiste `config.AppConfig.Rail.Collapsed` vía `SaveConfig`. Refresca el rail. |
| `showAccount()` | `func (c *Controller) showAccount()` | Abre el pane de Cuenta/Proveedores (panel in-pane con sección providers existente). |

### 1.2 Funciones 002 preservadas (sin cambio de firma)

`showChat()`, `showSettings()`, `showUsage()`, `showGit()`, `showMCP()`, `showPlugins()`, `showTasks()`, `focusInput()`, `openPalette()`, `closePalette()`, `openKeymap()`, `closeKeymap()`, `showSessionPicker()`, `switchNth(n)`, `newSession()`.

- `showSettings`/`showUsage`/… DELEGAN en `showPane(i)` (mismo índice que hoy: 1/2/3/4/5/6) para mantener atajos Ctrl+,, Ctrl+U y tests 002 verdes.
- `showChat()` set Current=0 y restablece foco al composer.

### 1.3 Invariantes del stack

- `viewStack.Objects` MUST contener `*fyne.Container` en TODOS sus índices (tests 002 y showIndex dependen de ello). El rail y el Border raíz viven FUERA del stack.
- Índices fijos: 0=chat, 1=settings, 2=usage, 3=git, 4=mcp, 5=plugins, 6=tasks; después (dinámico) palette, keymap, picker.
- `showPane` MUST NO tocar índices ≥7 (overlays) — `showIndex` los oculta todos igual que hoy, pero `back()` respeta la guarda de overlay abierto.
- La marca activa del rail (`rail.activeID`) MUST coincidir con `Current` pane tras cualquier `showPane`.

## 2. Rail widget (`rail.go`)

```go
type railView struct {
    onSelect func(paneIndex int)     // 0..6; -1 para acciones (theme/help/account por separado)
    onToggle func()                  // colapso
    activeID string                  // slot activo
    collapsed bool                   // icon-only vs expanded
    content() fyne.CanvasObject
}
```

**Reglas**:
- `content()` MUST NO depender de `viewStack` (el rail es autónomo; el Controller cablea `onSelect`).
- Slots en orden: **work**: chat, sessions, git, tasks, mcp, plugins · **control**: usage, account, theme, settings, help.
- `Shortcut` de cada slot: `Alt+1..9` en orden de lista (work primero); account=pie (sin atajo o `Alt+0` — verificación: si `Alt+0` se usa, listar en cheat-sheet).
- En modo collapsed los labels se ocultan; los tooltips se conservan; `MinSize` ≥44px.
- Estado activo: barra de acento izquierda visible SOLO en el slot cuyo `activeID == Current`.
- En el pie (group control), account se posiciona al fondo (order fijo: usage, theme, settings, help, account? — decisión: account último, pegado al borde, FR-005).

## 3. Teclado (FR-023 ampliado)

| Atajo | Acción | Notas |
|-------|--------|-------|
| `Alt+1..9` | `showPane(i)` slot correspondiente | NUEVO — rail |
| `Alt+Left` | `back()` | NUEVO — pane back |
| `Ctrl+K` | `openPalette()` | intacto |
| `Ctrl+1..9` | `switchNth` (sesiones) | intacto — sin colisión |
| `Ctrl+,` / `Ctrl+U` | settings / usage | intactos (delegan showPane) |
| `Ctrl+L` | focus composer | intacto |
| `Ctrl+Shift+A` | toggle auto-approve | intacto |
| `?` | keymap | cheat-sheet crece: categoría "Rail" |

**Guarda Esc**: el composer (`Esc` = cancel stream), palette/picker/keymap (`Esc` = cerrar) tienen prioridad por foco; `Esc` = back() SOLO cuando el foco está en el rail o en un pane y `Current != 0` y sin overlays.

## 4. Tema (theme.go)

Nuevos tokens públicos (constantes exportadas en `internal/gui`):

```go
const (
    ThemeColorNameRail    fyne.ThemeColorName = "letsgo-rail"
    ThemeColorNameSurface fyne.ThemeColorName = "letsgo-surface"
    ThemeColorNameRaised  fyne.ThemeColorName = "letsgo-raised"
)
```

Mapeos requeridos en `terminalTheme.Color(name, variant)`:

| ColorName | Dark | Light | Rol |
|-----------|------|-------|-----|
| `Background` | `#0A0E13` | `#F5F7FA` | abyss / base |
| `Button, InputBackground, MenuBackground` | `#0F141B` | `#FFFFFF` | surface |
| `OverlayBackground, ThemeColorNameRaised` | `#161D26` | `#FFFFFF` | raised |
| `HeaderBackground, Hover, ThemeColorNameRail` | `#1E2733` | `#E8ECF1` | well / rail bg |
| `Separator, Shadow, InputBorder` | `#242D38` | `#C9D2DC` | hairline |
| `accentThemeColor` | `#00ADD8` | `#0087A9` | acento LetsGO |
| `Primary` | `#4FC1FF` | `#0068A8` | acciones primarias |
| `Focus, Pressed` | `#00ADD8` | `#0087A9` | foco |
| `Success / Warning / Error` | `#2EA043 / #D29922 / #F85149` | `#1F7A37 / #9A6700 / #D1242F` | semánticos |

**Contrato de tipografía**: `Font(TextStyle{})` → face sans (UI); `Font(TextStyle{Monospace:true})` → Cascadia Mono. Labels que hoy son mono y representan código/telemetría siguen mono; chrome pasa a sans.

**Invariante visual**: al menos 3 niveles de fondo distinguibles entre sí en ambos variantes (assert en tests).

## 5. Config (config.go)

```go
type RailConfig struct {
    Collapsed bool `mapstructure:"collapsed"` // default false
}
// Config:
//   Rail RailConfig `mapstructure:"rail"`
//   ThemeVariant string `mapstructure:"theme"` // "dark" | "light", default "dark"
```

- `SaveConfig`/`LoadConfig` persisten ambos (round-trip test).
- `ThemeVariant` se aplica vía `app.Settings().SetThemeVariant(...)` en `start()`; no depende del SO (FR-007).

## 6. Eventos/estado consumidos (sin cambios)

`streamPump` (StreamStart/Delta/Done, ToolRequested/Executing, Idle, ErrorEvent), `guiPermissionDriver` (SessionGranted/onShow/onDismiss), grants, plan mode: **intactos**. La shell solo reubica/reestila. La franja de estado sigue recibiendo los mismos updates (`statusbar.go` sin cambios de contrato).

## 7. Tests obligatorios (contrato verificado)

- `TestRailSlotsAndCollapse`: lista de slots correcta (11), colapso oculta labels mantiene MinSize ≥44, marca activa se mueve.
- `TestRailSelectDispatches`: onSelect con índice correcto por slot (chat=0 … tasks=6).
- `TestPaneBackReturnsToChat`: showPane(3) → back() → Current=0 y foco composer; showPane(2) → showPane(5) → back() → 2 (pila).
- `TestBackGuardsOverlay`: con palette abierta, back() no cambia Current.
- `TestAltShortcutsSmoke`: Alt+1..9 / Alt+Left despachan a showPane/back (ShortcutHandler).
- `TestThemeTokensDistinct`: 3 niveles de fondo ≠ entre sí, en dark y light; acento ≠ primary.
- `TestRailPreferencePersists`: toggleRail → SaveConfig → LoadConfig → collapsed == true.
- **Regresión**: toda la suite existente de `internal/gui` y `internal/config` verde sin editar (excepto tests que explícitamente añadan asserts).

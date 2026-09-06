# Tasks: GUI Overhaul — Codex-style App Shell (003)

**Input**: spec.md + plan.md + research.md + data-model.md + contracts/gui-contract.md + quickstart.md de `/specs/003-gui-overhaul/`

**Prerequisites**: 002 completo (status line, keymap `?`, panes 0..6, tests headless verdes).

**Tests**: Obligatorios (test-first) — headless `test.NewApp()` en `internal/gui/*_test.go`; round-trip viper en `internal/config/config_test.go`. NO romper tests existentes de 002 (viewStack índices 0..6, `*fyne.Container` en todos los hijos).

**Build**: `$env:PATH = "D:\msys64\ucrt64\bin;" + $env:PATH; $env:CGO_ENABLED = "1"` antes de `go build`/`go test`/`go vet`.

---

## Fase 1: Config + Tema (Shared)

- [X] T001 Añadir `RailConfig{Collapsed bool}` + `ThemeVariant string` a `internal/config/config.go`, defaults (`rail.collapsed=false`, `theme="dark"`), persistencia en `SaveConfig`/`LoadConfig`. Tests round-trip en `config_test.go`.
- [X] T002 Ampliar `internal/gui/theme.go` a "Deep Goblue": 3 niveles de fondo (abyss `#0A0E13` / surface `#0F141B` / raised `#161D26`, well `#1E2733`), separador `#242D38`, foco = acento, semánticos (success `#2EA043`, warning `#D29922`, error `#F85149`) en dark y light; consts `ThemeColorNameRail`/`Surface`/`Raised`; font split (sans embebido para `Style{}`, Cascadia Mono para `{Monospace:true}`). Tests: 3 niveles distintos en ambos variantes, acento≠primary, font split.

## Fase 2: Shell (Rail + Navegación)

- [X] T003 `internal/gui/rail.go`: `railView` con slots work (chat/sessions/git/tasks/mcp/plugins) + control (usage/account/theme/settings/help), separador hairline, colapso icon-only (≥44px) vs expanded (~200px), marca activa (barra acento izq.), tooltips con atajo, `onSelect(paneIndex)`. Tests `rail_test.go`: slots orden, colapso oculta labels y mantiene MinSize, onSelect por slot.
- [X] T004 `internal/gui/navigation.go`: `showPane(i)` (push navHistory, set Current, marca rail), `back()` (Esc/Alt+Left; guarda: Current!=0 y sin overlay), `showChat` → Current=0 + foco composer. Tests `navigation_test.go`: push/pop, vuelta a chat, guardas overlay, pila vacía → chat.
- [X] T005 Integrar shell en `internal/gui/app.go`: `root = container.NewBorder(nil,nil,rail,nil,viewStack)`; `win.SetContent(c.root)`; `showSettings/Usage/Git/MCP/Plugins/Tasks` delegan en `showPane(i)` (índices 1..6 iguales); toggleRail en rail + persistencia. Smoke tests: rail montado, panes via rail, Alt+1..9/Alt+Left dispatch (ShortcutHandler), regresión 002 intacta.

## Fase 3: Cuenta + Settings + Keymap

- [X] T006 Pie del rail: avatar "GO" (círculo acento + punto de estado online/idle/error según `hasAnyKey`/engine) anclado al borde inferior → `showAccount()` abre pane Cuenta (sección providers). Sección Appearance en settings: tema claro/oscuro (aplica `SetThemeVariant` + guarda), colapso rail. Tests: `TestRailAvatarAndStatus` (foot render, tap→account, dot color), `TestToggleThemeVariantPersists` (flip + round-trip), settings evalúa `ThemeVariant`/`Rail.Collapsed` del config.
- [X] T007 `keymap.go`: añadir categoría "Rail" (`Alt+1..9` → panes/usage/settings/help, `Alt+Left` → back). Test: `TestKeymapCatalogHasRailCategory` (catálogo contiene las entradas Rail y el filtro las resuelve).

## Fase 4: Polish + Validación

- [X] T008 Restyle 002 en el nuevo shell: approval card con barra acento izquierda + importancia botones, composer con anillo de foco (InputBorder+Focus), mensajes usuario/asistente diferenciados, statusbar mantiene piezas (solo re-bosquejo). Sin cambio de contrato. Tests: `TestMessageRoleBubbles` (barras de rol + grant button), `TestComposerFocusRing` (anillo accent↔border).
- [X] T009 Validación final: `go vet ./...`, `go test ./...` (CGO), `build.bat -n` → todo verde; checklist quickstart SC-001..009 actualizado; tasks.md marcado [X].

## Notes

- Test-first: nuevo test → falla → implementar → verde.
- No renombrar `viewStack` ni cambiar el tipo de sus hijos; rail y Border viven FUERA del stack.
- Esc guarda: overlay/palette/picker/keymap tienen prioridad; back() solo con pane activo != chat.
- Colisión de atajos: rail usa `Alt+` (Ctrl+1..9 = sesiones intacto).
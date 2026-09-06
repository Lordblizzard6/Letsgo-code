# Data Model: GUI Overhaul — Codex-style App Shell

**Feature**: 003-gui-overhaul | **Date**: 2026-08-06 | **Spec**: [spec.md](spec.md)

La shell introduce tres entidades nuevas (RailSlot, NavHistory, AppearancePrefs) y reutiliza la vista activa del Controller (ya indexada). No hay entidades de dominio nuevas: la conversación, sesiones, tools, grants y config del engine se mantienen intactas (ver data-model.md de 002).

## 1. RailSlot (Slot de la barra)

Representa un elemento de navegación de la barra vertical. Es estático y determinista (mismo orden siempre).

| Campo | Tipo | Descripción | Reglas |
|-------|------|-------------|--------|
| `ID` | string | Identificador estable (ej. `"chat"`, `"sessions"`, `"git"`, `"tasks"`, `"mcp"`, `"plugins"`, `"usage"`, `"account"`, `"theme"`, `"settings"`, `"help"`) | Único; ≤11 slots (SC-008) |
| `Label` | string | Etiqueta visible en modo expandido (ej. "Conversación", "Ajustes") | No vacío |
| `PaneIndex` | int | Índice del pane en el viewStack a mostrar, o -1 si es acción (theme toggle, help) | Chat=0, Settings=1, Usage=2, Git=3, MCP=4, Plugins=5, Tasks=6; cuenta/ayuda/tema externos |
| `Shortcut` | string | Atajo de teclado (ej. `Alt+1`) mostrado en tooltip y cheat-sheet | Correlativo 1..9; sin colisión con Ctrl+* |
| `Icon` | fyne.ThemeIconName | Ícono plano del slot | Necesario (mín 44px en colapso) |
| `Group` | string | `"work"` (arriba) o `"control"` (abajo, anclado) | Separador hairline entre grupos (FR-001) |

**Reglas de validación**:
- `index` MUST exist: uno por slot; sin duplicados.
- Rail NO scrollea: lista completa visible sin scroll (SC-008). Si el contenido excede, se colapsa la etiqueta antes que el slot (icon-only).
- `Shortcut` MUST ser `Alt+n` (n=1..9) para slots; nunca `Ctrl+*` (reservado).
- El pie (`Group=control`) MUST contener al menos: Uso, Cuenta/Proveedores, Tema, Ajustes, Ayuda (FR-001/FR-005). Cuenta anclada en el borde inferior.
- `PaneIndex` -1 (action slots: Tema=alterna dark/light, Ayuda=abre keymap) no debe ocultar el viewStack.

## 2. PaneIndex (Vista activa)

Representa el pane visible en el área principal.

| Campo | Tipo | Descripción | Reglas |
|-------|------|-------------|--------|
| `Current` | int | Índice del pane visible (0..6; 0=chat) | Una sola activa |
| `InViewStack` | bool | Si el pane es interior del stack (true para panes; false para overlays transitorios) | define el comportamiento de Esc |
| `History` | []int | Pila de navegación (sin el actual) | Top = destino de `back()` |

**Reglas**:
- `Chat=0` es la raíz; nunca se saca de la pila.
- Cambiar de pane (rail, paleta, `Alt+n`, menú) hace push del índice anterior (salvo sea el mismo) y set Current.
- `back()` (Alt+Left/Esc) pop del History si `Current != 0` y no hay overlay (palette/picker/keymap) abierto; si History vacío → vuelve a chat (0).
- `Current >= 0 && Current <= 6` siempre (bounds).

## 3. AppearancePrefs (Preferencias de apariencia)

Persistidas vía config existente (viper). 

| Campo | Tipo | Default | Descripción |
|-------|------|---------|-------------|
| `rail.collapsed` | bool | false (expandido) | Estado colapsado (icon-only) de la barra |
| `theme.variant` | string | `"dark"` | `"dark"` / `"light"` explícito para la app |
| `statusline.show` | bool | true | (ya en 002: superficie) |

**Reglas de validación**:
- Se guarda con el flujo `SaveConfig`/`LoadConfig` existente (round-trip testado).
- `rail.collapsed` es GLOBAL (no por vista) — FR-002.
- `theme.variant` fuerza que `Color()` reciba el variante consistente; no depende de la detección del SO (FR-007).
- El colapso del rail está permitido en cualquier pane; la marca activa se mantiene.

## 4. Account/Provider (Cuenta/proveedor)

Derivación visual del pie del rail; no nueva persistencia (las claves ya viven en config).

| Campo | Tipo | Descripción |
|-------|------|-------------|
| `ProviderName` | string | Proveedor activo (p. ej. groq, openai) — leído de config `provider` |
| `Status` | enum{online, idle, error, none} | Estado de conexión (derivado de clave configurada / última respuesta del engine) |
| `Avatar` | render | Monograma "GO" con acento; punto de estado (verde/gris/rojo) sobre él |

**Reglas**: anclado al borde inferior (FR-005); al pulsar abre el pane Cuenta (providers). Tapping punto → panel en-el-área, nunca modal.

## 5. Estado transitorio (no persistido)

| Objeto | Estado relevante | Contrato con |
|--------|------------------|--------------|
| `navHistory []int` | pila push/pop | navigation.go |
| `rail.activeID` | slot activo (barra de acento visible) | rail.go |
| `overlayOpen` (global) | nivel de overlay (palette/picker/keymap) | app.go — guarda de back() |

## 6. Mapa de dependencias

```text
config.AppConfig
  ├─ rail.collapsed ───────────► rail.go (expandir/colapsar)
  ├─ theme.variant ────────────► theme.go (Color(variant))
  ├─ statusline.* ─────────────► statusbar.go (sin cambio)
  └─ provider/api keys ────────► settings.go (panel Cuenta)

Controller
  ├─ viewStack (0..6 + overlays) ──► showPane(i)/showIndex(i)
  ├─ navHistory []int ────────────► back() (Alt+Left/Esc)
  ├─ rail *railView ───────────────► onSelect(pane) → showPane
  └─ keyboard Alt+1..9/Alt+Left ───► attachShortcuts

keymapCatalog ──► keymapOverlay (categoría Rail/Vistas)
terminalTheme ──► theme tokens (3 niveles, acento, font split)
```

## 7. No-cambios (preservados)

- `viewStack` (campo + números de índice 0..6) — no renombrar; tests headless lo usan.
- SQLite `db` (sesiones/grants/mensajes) — intacto.
- Engine commands/events (SendMessage/Steer/Queue/… ) — intactos; solo se cambia presentación.
- La franja de estado y sus piezas configurables — intactas (solo re-bosquejo pixels).

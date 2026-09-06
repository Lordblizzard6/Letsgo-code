# Contrato de la GUI Wails

**Branch**: `005-bubbletea-tui-wails` | **Fecha**: 2026-08-11 | **Spec**: [spec.md](../spec.md)

**Propósito**: inventario de la superficie GUI que la interfaz Wails debe
implementar para alcanzar **paridad funcional completa** con la GUI Fyne actual
(FR-017) y permitir la **retirada de Fyne** (FR-018). La GUI Fyne existente
(`internal/gui`) es la referencia de comportamiento; este contrato es el checklist
de paridad.

## 1. Estructura de navegación (rail)

La GUI Wails replica el rail actual de 10 destinos en 3 zonas, con estado
colapsado/expandido y marca activa:

| Slot | Etiqueta | Pane | Atajo | Zona |
|------|----------|------|-------|------|
| chat | Chat | 0 (chat principal) | Alt+1 | work |
| git | Git | Git | Alt+2 | tools |
| tasks | Tareas | Tareas (subagentes) | Alt+3 | tools |
| mcp | MCP | Servidores MCP | Alt+4 | tools |
| plugins | Plugins | Plugins | Alt+5 | tools |
| settings | Configuración | Configuración | Alt+6 | system |
| usage | Uso | KPI uso | Alt+7 | system |
| help | Ayuda | Ayuda | Alt+8 | system |
| theme | Tema | Cambio dark/light | — | system |
| account | Cuenta | Flyout/panel de cuenta | — | system |

**Paridad**:
- Zona "HERRAMIENTAS" con separadores; colapso del rail conserva tooltips *label
  (atalo)* en los iconos.
- Marca/accent del slot activo única (no hay dos marcas a la vez).
- El avatar de cuenta muestra estado online/ocupado (dot) y abre el flyout con
  datos de cuenta + acciones (claves, uso, salir).

## 2. Chat (pane principal)

- Composer con botón **Enviar ↔ Detener** según estado de stream (mismo patrón
  que la TUI), placeholder "Escribe a LetsGO…", input siempre activo durante el
  stream (FR-008: se puede escribir mientras se cancela/genera).
- Streaming incremental con indicador de estado y cancelación en cualquier momento.
- Render de markdown incremental + saneado (double sanitize acumulado), código con
  wrap opcional, columna máxima ~720 px, tarjetas con borde 1px/radius 6px/padding
  16px (tokens visuales heredados de 004).
- Actividad de herramientas (tool:start/end) mostrada dentro del chat sin bloquear.

## 3. Historial de sesiones (panel lateral)

- Crear / listar / retomar sesiones; historial completo compartido con la TUI.
- Estado vacío: "No hay conversaciones todavía." + CTA "Nueva conversación".
- Carga asíncrona con "Refrescando…" conservando el contenido previo.

## 4. Paneles del rail (paridad Fyne)

| Pane | Contenido mínimo |
|------|------------------|
| Git | branches, status, commit/diff/review/tags/hooks/undo/redo, PR + push |
| Tareas | lista de sub-agentes con estado; vacío: "Todavía no hay tareas." |
| MCP | lista de servidores + detalle (estado, tools, start/stop); vacío con CTA "Añadir servidor" |
| Plugins | lista + instalar/cargar/habilitar/deshabilitar; vacío con CTA "Instalar…" |
| Configuración | tabs Cuenta / Apariencia / Preferencias / Uso (overlay) |
| Uso | KPI hoy (coste, requests, tokens) + tabla por proveedor + presupuesto; vacío: "Sin actividad registrada hoy." |
| Ayuda | guía de atajos y comandos |
| Tema | cambio dark/light aplicado al motor y persistido |

## 5. Accesos rápidos y estados

- **Paleta Ctrl+K** (comando + salto a superficies), atajos Alt+1..8 y Esc para
  cerrar overlays/flyout, todo 100% operable por teclado con focus visible en ambos
  temas (SC-003).
- Estados vacíos/de carga en las 6+ superficies (contrato §5/§10 del 004) y
  contraste AA en ambos temas (palette smoke CI).

## 6. Primer arranque (onboarding)

- Sin API key configurada → flujo de configuración de proveedor/key como pantalla
  inicial; solo tras guardar una key válida se habilita el chat (FR-011).

## 7. Retirada de Fyne (inmediata)

La retirada de `internal/gui` y de `fyne.io/fyne` es **inmediata** (primera tarea
del feature, directiva del usuario 2026-08-13: "eliminar Fyne ya mismo, mismo
método de ejecución `letsgo.exe gui`"). Gates de verificación de la eliminación:

- G-1: Los flujos portados tienen test equivalente (engine/servicios + Vitest/
  Playwright), escrito AL FINAL de cada historia (constitución I, Test-Last).
- G-2: El proyecto compila y las suites (engine, TUI, GUI, config) pasan **sin**
  importar fyne (`go mod why fyne.io/fyne/v2` vacío, `rg fyne` limpio fuera de `specs/`).
- G-3: `go.mod` limpio: `fyne.io/systray` y demás transitivos retirados; assets
  `fyne bundle`/`FyneApp.toml` eliminados; iconos en `build/` de Wails.

## 8. Sistema de diseño y gates UX (enmienda US5, 2026-08-14)

### 8.1 Dirección de diseño por superficie

- **Rail (48px)**: iconos SVG stroke (sin emoji), secciones WORK/TOOLS/SYSTEM en
  caps 11px, marcador activo único (accent icon + barra 3px + tint `--accent-soft`);
  collapse 48↔56px con tooltip *Label Alt+n*. Footer del rail: collapse + chip de
  cuenta (absorbe avatar y theme flotantes).
- **Status line inferior (28px)**: provider · modelo · branch · uso hoy · dot
  busy/sync — canal persistente de bajo ruido (patrón Codex); el estado "Ready/
  Streaming" deja de vivir en el transcript.
- **Chat**: columna ≤760px centered; **user = bubble derecha** (`--bg-active`),
  **asistente = bloque full-width sin borde** con header-per-message (mark +
  "Assistant" + timestamp + meta modelo/durar). Stream = cursor `▍` + línea 2px de
  progreso; sin fila "Streaming…".
- **Code blocks = objetos elevados**: `CodeBlock.svelte` con header (language caps +
  Copy on hover/focus), resaltado syntax (highlight.js/shiki offline, tokens por
  tema), overflow horizontal en el bloque (nunca scroll de página). Todo `<pre>`
  vía este componente (G8).
- **Composer = command shell docked**: textarea siempre activa, Enter=send /
  Shift+Enter=nueva línea / Enter=stop en stream; controles izquierda (model chip ·
  plan toggle · @file · auto-approve) y derecha Send⇄Stop (accent fill / err-soft
  outline, min 44px); hint row 11px `Ctrl+K commands · / plan · @ files ·
  Alt+1-8 surfaces`.
- **Tool activity inline en el turno**: cards colapsables (header = icono + nombre +
  estado live + chevron; body mono truncado 200ch + expand). 5+ tools → grupo
  compacto con contador (compactInline Qwen). Running = edge accent 1px + pulse
  (≤150ms, off bajo reduced-motion); success `--ok`, error `--err` + Retry. Nunca
  un log colgante que empuje al composer.
- **Sesiones**: agrupadas por fecha (Hoy/Ayer/Older) con caps 11px; rows hover
  `--bg-hover`, activo `--accent-soft`; dot busy; empty `"No hay conversaciones
  todavía."` + CTA "Nueva conversación"; "Refrescando…" preservando contenido.
- **Paleta (Ctrl+K)**: 640×~50vh, radius-lg, `--shadow-lg`, grouped (Ir a · Acciones
  · Tema), navegación ↑/↓+Enter, highlight substring accent, footer hints, focus
  trap, Esc cierra y re-foca el composer.
- **Overlays**: Settings/Usage/Help **no-oscurecen el chat** (drawer/panel con
  backdrop sutil) y **Esc/backdrop las cierran** con foco al composer; sombra
  `--shadow-lg`, radius 12px.
- **Onboarding (ProviderSetup)**: split-screen brand (45%: mark + "Welcome to
  LetsGO" + value prop) / form (55%); **gate FR-011 sin cambios**; strings de test
  intactos ("Welcome to LetsGO", "Save & start chatting", "API Key").
- **Estados**: empty/loading(skeleton)/error con CTA en las 7+ superficies de
  datos (G7); skeletons en paneles pesados; errores con icono + orientación + Retry.

### 8.2 Gates de restyle (verificables en CI: grep, Vitest, Playwright)

| Gate | Criterio |
|------|----------|
| G1 | Token purity: `rg "#[0-9a-fA-F]{3,8}" src/ --glob '*.svelte' --glob '*.ts'` → 0 hits fuera de `app.css`. |
| G2 | WCAG AA ambos temas: body/meta ≥4.5:1, no-text ≥3:1, texto sobre acento ≥4.5:1; smoke Playwright 100/125/150%. |
| G3 | `:focus-visible` ring (2px/offset 2) en todo interactivo, ambos temas; flujo 100% por teclado. |
| G4 | Spacing solo escala 4px (4/8/12/16/20/24/32/40/48); chat ≤760px; targets ≥44px. |
| G5 | Cero emoji en chrome (rail, botones, headers, status); SVG stroke único set. |
| G6 | Inventario ≥95% desde primitivas; sin rail duplicado (`rail.svelte` consolidado). |
| G7 | 10 superficies con empty+loading(skeleton)+error; chat streaming/error/empty/idle. |
| G8 | Todo `<pre>` vía `CodeBlock` (header+lang+copy+highlight); sin `<pre>` desnudo. |
| G9 | Transiciones ≤150ms (salvo progreso stream); `prefers-reduced-motion` honrado. |
| G10 | Smoke visual Playwright (chat+code+tool, palette, onboarding, empty, dark+light) contra checklist. |

### 8.3 Gates de UX (medibles, de D9)

UX-01 Retry reproduce el último turno de usuario (FR-013) · UX-02 composer siempre
visible/enabled (FR-008) · UX-03 resume de sesión en 1 acción (navega a chat) ·
UX-04 1 row por tool call + agrupado por turno · UX-05 Esc/backdrop cierran
overlays y el chat subyace · UX-06 cada interactivo operable/focuseable, sin traps ·
UX-07 `aria-live`/`aria-busy` en stream + reduced-motion · UX-08 contraste de los
token-pairs en ambos temas · UX-09 empty+loading en 7/7 superficies · UX-10
densidad: sin card-anidada en el transcript · UX-11 Help == palette hints == keymap
real; theme muestra variante actual · UX-12..UX-14 estabilizadores: tests vitest
existentes sin cambios, e2e `onboarding.spec.ts` intacto, **cero diff en
bindings/backend** (grep de call-sites a services/eventos).

Ejecución de la GUI: `letsgo.exe gui` (cobra → `cmd/wails/gui.Run()`), mismo método
que lanzaba Fyne; sin binario Wails separado.
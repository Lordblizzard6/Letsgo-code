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

Ejecución de la GUI: `letsgo.exe gui` (cobra → `cmd/wails/gui.Run()`), mismo método
que lanzaba Fyne; sin binario Wails separado.
# Research: UI/UX Refinement

**Created**: 2026-08-07 | **Status**: Completo — Fase 0 (unknowns resueltos)

## Metodologia

Auditoria del codigo en `internal/gui/` y de `specs/003-gui-overhaul/`, mas sintesis de dos subagentes (UX Researcher y UI Designer) y benchmark de GUIs de agentes IA de escritorio (Codex, Claude Desktop, Cursor). Este archivo resuelve los unknowns del Technical Context del plan.

## Decisiones (formato Decision / Rationale / Alternatives)

### D1 — Estructura del rail
- **Decision**: tres zonas separadas: trabajo (Chat), herramientas con etiqueta "HERRAMIENTAS" (Git, Tareas, MCP, Plugins) y sistema anclado abajo (Uso, Configuracion, Ayuda, Tema, avatar, colapso). Se elimina el slot `sessions` (duplicaba pane 0). Atajos Alt: 1 Chat, 2 Git, 3 Tareas, 4 MCP, 5 Plugins, 6 Configuracion, 7 Uso, 8 Ayuda.
- **Rationale:** el patron de zonas por frecuencia de uso sigue los estandares de VS Code, Slack y Codex; corrige la marca activa (hoy `sessions` compite con `chat` por el pane 0); la lista de conversaciones sigue embebida en Chat y no se pierde. Los indices del viewStack no cambian.
- **Alternatives:** slot `sessions` con pane propio (fuera de alcance) o un rail plano sin zonas (no resuelve el desorden).

### D2 — Configuracion en un overlay con secciones
- **Decision:** `showSettings()` conserva su firma (02) y abre un overlay no modal de ~620x520 con `widget.TabContainer` de 4 secciones (Cuenta, Apariencia, Preferencias, Uso). `Esc`/Cancelar cierra y devuelve el focus al composer.
- **Rationale:** la conversacion queda montada detras (FR-004) y el patron de overlays (palette, keymap, picker) ya existe y es testable headless.
- **Alternatives:** modal bloqueante (viola FR-004); pane tabbed indice 7 (fallback si el popup falla en headless).

### D3 — Flyout de cuenta
- **Decision:** el avatar abre `widget.NewPopUp(content, canvas)` anclado al propio avatar con: dot de estado (online/idle/error), proveedor y modelo activos, resumen de uso del dia y acciones ("API keys & proveedores...", "Uso...", "Cerrar sesion"). `Esc` cierra y vuelve el foco al composer.
- **Rationale:** la cuenta es identidad, no navegacion; separar de Settings (hoy `showAccount` llama a `showSettings`) y no perder contexto.
- **Alternatives:** un panel in-pane para cuenta (mezcla cuenta con settings) o un modal bloqueante.

### D4 — Uso como dashboard KPI
- **Decision:** reutilizar `usageSource`, `aggregateUsage` y `startOfToday`. Presentar `GridWithColumns(3)` con tarjetas KPI (coste de hoy, requests, tokens), tabla por proveedor (`widget.NewTable`) y barra de presupuesto porcentual si existe plan.
- **Rationale:** responde "cuanto gaste hoy" de un vistazo (SC-004) sin cambios de modelo de datos. `usageHistogram.String` pasa a fallback/debug.
- **Alternatives:** el dump monoespacial actual (fuente de "antiguo/disperso") o un enlace a una vista en profundidad.

### D5 — Estados vacios y carga
- **Decision:** helpers `emptyState(msg, ctaText, onCTA)` y `loadingState(content, refreshing)` en `empty.go`, aplicados a sessions, mcp, plugins, tasks, git y usage. El refresco asincrono conserva el contenido previo y muestra "Refrescando...".
- **Rationale:** una lista vacia sin CTA parece un bug; la CTA mejora onboarding. Mantener datos previos evita parpadeos.
- **Alternatives:** sin estado (menos onboarding); limpiar el panel en cada refresh (posicion perdida).

### D6 — Iconos semanticos
- **Decision:** `icons.go` con `iconFor(slotID) fyne.Resource` que devuelve SVGs embebidos (burbuja de chat, branch, checklist, caja/sevidor, puzzle, gear, sol/moon, gauge) para el rail y superficies; fallback a `theme.Icon(slot.icon)` y obligatorio: tooltip en colapsado.
- **Rationale:** los iconos Fyne actuales confunden (Chat = MailCompose, Git = ViewRefresh); un glifo que recuerda al destino mejora el indice visual.
- **Alternatives:** mantener el tema actual (no corrige el problema); fuente de iconos (peso y mantenimiento).

### D7 — Contraste y paleta CI
- **Decision:** token `lightPlaceholder` pasa de `#8A94A0` (2.76:1) a ~`#5C6B7A` (>=4.5:1); se ajusta el boundary del separador (darkBorder) para >=3:1. `theme_test.go` itera todos los pares base (dark+light) y falla si un par de texto queda <4.5:1 o un no-texto <3:1.
- **Rationale:** convierte la garantia de accesibilidad en CI automatizado (SC-005) y corrige el unico fallo conocido.
- **Alternatives:** oscurecer toda la paleta (rompe jerarquia); ignorar AA.
- **rat**: cambio dos tokens + smoke test.

### D8 — Tipografia y medida
- **Decision:** los tipos de la rampa 12/14/18/20/24 (sans) y mono 12-13 solo para codigo/telemetria; la prosa de los mensajes max ~720px centrada; el codigo ancho completo con wrap opcional; espaciado sobre base de 4px.
- **Rationale:** la columna corta se lee mejor (45-80 cpl) y restringir mono a contenido tecnico hace el resto legible.
- **Alternatives:** medida fluida; doble fuente (ya hay sans+mono del tema).

### D9 — Composer y foco
- **Decision:** ring de foco accent de 2px; el boton principal alterna Enviar y Detener en el mismo slot durante streaming; placeholder "Escribe a LetsGO... (Enter enviar · Tab encolar · Esc detener)". Estados hover/focus/active distintos.
- **Rationale:** prominencia del input y feed de teclado (US-003); el swap de boton evita botones redundant.
- **Alternatives:** dos botones separados; ring 1px como hoy.

## Auditoria (hallazgos clave)

| Hallazgo | Evidencia | Decision |
|----------|-----------|----------|
| `chat` y `sessions` ambos pane 0 | rail.go | D1: quitar `sessions` |
| Grupo de control mixto, 4 propósitos | rail.go | D1: zonas |
| `showAccount` llama a `showSettings` | app.go | D3: flyout |
| Settings es un form lineal (~20 controles) | settings.go | D2: overlay tabs |
| Uso es dump monoespacial | usage.go | D4: dashboard KPI |
| Iconos sin tooltips | rail 44px | D6: SVG+tooltip |
| Microcopy inconsistente | sessions.go, chatview.go | D8: placeholders ES |
| `lightPlaceholder` falla AA 2.76:1 | theme.go | D7: fix token |
| Sin estados vacios/carga | sessions.go, mcpview.go | D5: helpers |
| Ring composer 1px | composer.go | D9: 2px |

## Verificacion de contraste (WCAG AA)

| Par | Ratio | Veredicto |
|-----|-------|-----------|
| oscuro: foreground/abyss | 16.5 | PASS |
| oscuro: secondary/surface | 8.4 | PASS |
| oscuro: accent/background | 7.3 | PASS |
| oscuro: placeholder/input | 6.1 | PASS |
| claro: foreground/bg | 13.1 | PASS |
| claro: secondary/bg | 7.4 | PASS |
| claro: primary/bg | 5.3 | PASS |
| claro: placeholder/bg | 2.76 | FAIL -> fix `#5C6B7A` |
| separator darkBorder/well | ~0.9 | FAIL no-texto >=3 -> lighten |

## Reconocimientos del entorno

- Abrir/cerrar panes <50ms; overlays transicion <=250ms o 0 con reduced-motion.
- Aprobacion p90 <=2s (no regresion 003).
- Sin nuevos campos persistentes; solo presentacion.
- Los tests de la suite headless de 002/003 se mantienen verdes como puerta de entrada.
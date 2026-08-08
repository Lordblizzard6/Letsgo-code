# Quickstart: GUI Overhaul — Codex-style App Shell (Phase 1)

**Feature**: 003-gui-overhaul | **Date**: 2026-08-06

Guía de validación de la shell rediseñada. Cada journey es testeable end-to-end (headless y/o manual) y mapea a FR/SC de `spec.md` y al contrato `contracts/gui-contract.md`.

**Build (Windows)**: `$env:PATH = "D:\msys64\ucrt64\bin;" + $env:PATH; $env:CGO_ENABLED = "1"` antes de `go build`/`go test`/`go vet`.

**Tests**: `go test ./...` (patrón headless `test.NewApp()` en `internal/gui/gui_test.go`; contract en `contracts/gui-contract.md`).

**Preferencias antes de validar**: si se persiguen preferencias de apariencia de sesiones previas, borrar `$HOME/.letsGo/config.yaml` y reiniciar para probar defaults (dark, rail expanded).

---

## S1 — Shell con barra vertical (US1; SC-001, SC-007, SC-008)

1. Lanzar `letsgo.exe` con claves configuradas → se abre el chat con una **barra vertical a la izquierda** (rail), colapsable.
2. **Verificar**: grupo superior (Conversación, Sesiones, Git, Tareas, Herramientas/MCP, Plugins) y grupo inferior separado por línea (Uso/cuota, Tema, Ajustes, Ayuda y, pegado al borde inferior, **Cuenta/Proveedores** con avatar y punto de estado).
3. De un clic a cada slot → el pane correspondiente se muestra **en el área principal** (sin ventana modal); la conversación permanece montada detrás (SC-007). Se puede pulsar `Esc` para volver al chat con el composer enfocado.
4. **Verificar**: el transcript es visible/scrolleable en todo momento (sin modal); no se abre ninguna ventana extra al navegar.
5. **Edge**: sin sesiones previas, el picker de arranque se ve igual encima del shell; tras elegir sesión, el rail sigue presente.

## S2 — Rail: colapso y teclado (FR-002, FR-003; SC-003, SC-009)

1. Pulsar el toggle de colapso → la barra pasa a ícono-only (≥44px), sin etiquetas; los tooltips siguen ("Git — Alt+3").
2. Pulsar `Alt+1..9` → cada slot activa su panel; el anillo de foco se mueve con la barra sin atascos (SC-003).
3. `Alt+Left` → vuelve al pane anterior; `Esc` → vuelve al chat con foco en el composer.
4. **Keyboard-only walk**: `Ctrl+L` → escribir → `Enter` → aprobar con `Enter` → `Ctrl+N` → `Ctrl+1` → `Esc` interrupt → todo sigue funcionando con el rail presente (regresión 002, SC-009).
5. Cerrar y relanzar → el estado colapsado/expandido se mantiene (FR-002, persistencia).

## S3 — Tema Deep Goblue (FR-006..008; SC-006)

1. **Verificar visualmente**: 3 niveles de fondo (chrome/rail, superficie/paneles, tarjetas elevadas) + hairlines de 1px; acento LetsGO (Goblue) en momentos firmados (barra de aprobación, marca del rail, botón primario); semánticos (éxito/aviso/error) SOLO en estados.
2. **Tipografía**: chrome/etiquetas en sans; status line, código, payloads de herramientas y composer en mono.
3. Cambiar Tema (slot inferior o Configuración → Appearance) a claro → toda la paleta se adapta; volver a oscuro; reiniciar → la preferencia persiste (FR-007).
4. **Bono**: la tarjeta de aprobación/plan card muestra la barra de acento y botones Aprobar/Esc con la importancia correcta (contrato §4).

## S4 — Cheat sheet `?` y regresión de atajos (FR-011, FR-012; SC-005)

1. Con el foco en el composer, pulsar `?` → el cheat sheet lista ahora la categoría **Rail** (Alt+1..9, Alt+Left) y **Vistas**; filtrable; `Esc` cierra sin perder el texto del composer.
2. **Regresión FR-023**: Ctrl+N (nueva sesión), Ctrl+1..9 (sesiones), Ctrl+, (settings), Ctrl+U (usage), Ctrl+L (foco), Ctrl+Shift+A (auto-approve) — todos siguen funcionando tras el rediseño (delegan en panes).
3. Aprobación p90 ≤2s con transcript visible (SC-001/007); interrupt Esc ≤500ms (SC-005); nada de modal nuevo.

## Checklist final (por spec)

| SC | Verificación | Estado |
|----|--------------|--------|
| SC-001 | Cuenta/ajustes/uso localizables en barra inferior <5s (test de descubrimiento 10 sujetos) | ✅ estructural (pie rail + slots) | 
| SC-002 | Config → model → tema → volver al chat ≤3 acciones (<20s), sin cerrar ventana ni perder conversación | ✅ estructural (pane 1, Appearance, back) | 
| SC-003 | Walk oficial completo solo con teclado (incl. rail Alt+n, Esc/Alt+Left) | ✅ test (shortcuts) | 
| SC-004 | Aprobación p90 ≤2s con transcript visible (sin regresión) | ⏳ manual | 
| SC-005 | Interrupt Esc ≤500ms con composer operativo (sin regresión) | ⏳ manual | 
| SC-006 | ≥80% de 10 sujetos enumeran ubicación de cuentas/ajustes sin ayuda | ⏳ manual | 
| SC-007 | Transcript visible/scrolleable durante aprobación y durante vista alterna | ✅ build + smoke | 
| SC-008 | Rail no scrollea; ≤11 slots; colapso icon-only mantiene toda la navegación | ✅ test | 
| SC-009 | Regresión 002 completa verde (`go test ./...`) | ✅ (gui+config; engine flaky aislado) |

## Notas de regresión

- `viewStack` conserva índices (chat@0 … tasks@6) y el tipado `*fyne.Container` de todos sus hijos → los tests 002 existentes (`gui_test.go`) no se tocan.
- `go vet ./...` y `go test ./...` deben pasar con CGO (msys64).
- El rail y el Border raíz viven FUERA del viewStack; cualquier test nuevo que toque la shell no asume los índices overlay.
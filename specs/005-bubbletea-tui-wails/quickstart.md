# Quickstart: Doble interfaz — TUI Bubble Tea + GUI Wails

**Branch**: `005-bubbletea-tui-wails` | **Fecha**: 2026-08-13 | **Spec**: [spec.md](spec.md)

Guía de validación end-to-end. Los detalles de superficie están en
[contracts/frontend-contract.md](contracts/frontend-contract.md) y
[contracts/gui-contract.md](contracts/gui-contract.md); el modelo de datos en
[data-model.md](data-model.md). No duplica implementación.

## Prerrequisitos

- Go 1.26, entorno Windows (este repo se desarrolla/valida en Windows; Linux/macOS
  soportados por Wails con webkit2gtk/Command Line Tools).
- Wails CLI v3 (`go install github.com/wailsapp/wails/v3/cmd/wails3@latest`) solo
  para regenerar bindings/generar instalador; la app se arranca con `letsgo.exe gui`.
- WebView2 Runtime en Windows (preinstalado en Windows 11; `wails3 doctor` lo
  verifica).
- API key de proveedor configurada (`letsgo settings` o fichero de config) — salvo
  en J4, que arranca sin key a propósito.

## J1 — Core compartido: una fuente de verdad (P1)

**Objetivo**: verificar que TUI y GUI consumen el mismo motor, DB y configuración
(contrato C-001…C-006).

```bash
go test ./internal/engine/... ./internal/db/... ./internal/config/...   # conformidad del contrato (headless)
go test ./internal/tui/...                                              # modelos TUI (Update sintético + teatest)
```

**Esperado**: suite verde sin pantalla; el contrato falla si una interfaz deja de
respetar la superficie (comandos, eventos, sesiones, config).

## J2 — TUI Bubble Tea completa (P2)

**Objetivo**: ciclo de chat completo en terminal 80x24 (FR-003…FR-007).

```bash
letsgo chat
```

1. Escribir un mensaje y Enter → streaming con markdown renderizado (glamour).
2. Durante el stream: `Ctrl+C` → la respuesta se cancela, prompt de vuelta en <2s,
   sin mensaje parcial persistido (SC-008).
3. `Ctrl+S` → selector de modelo; confirmar → persiste y el motor se refresca sin
   reiniciar la TUI.
4. `/help`, `/clear`, `/cost`, `/tokens`, `/compact`, `/quit` responden; las
   sugerencias aparecen mientras se escribe `/`.
5. Salir y volver a entrar → retomar una sesión anterior (historial compartido).

**Esperado**: cada paso completable por teclado, estados legibles en 80x24 y en
redimensionado durante un stream activo (edge cases).

## J3 — GUI Wails y eliminación de Fyne (P3)

**Objetivo**: la GUI reemplaza a Fyne con paridad funcional (FR-008…FR-017) y el
proyecto queda sin Fyne (FR-018, retirada inmediata + Gate G-2/G-3).

```bash
letsgo gui                       # lanza la GUI Wails — MISMO método que Fyne hoy
# desarrollo frontend (opcional): wails3 dev  (hot-reload del frontend en :9245)
```

1. Enviar un mensaje → streaming con indicador + botón Detener; cancelar en
   cualquier momento; conversación persistida.
2. Crear una sesión en la GUI, retomarla desde `letsgo chat` (y viceversa) con el
   historial completo (SC-002).
3. Ejecutar una herramienta → actividad y resultado (éxito/error) dentro del chat
   sin bloqueo.
4. Recorrer los 10 destinos del rail (chat, git, tasks, mcp, plugins, settings,
   usage, help, theme, account) con Alt+1..8 y la paleta Ctrl+K, en ambos temas.
5. Redimensionar la ventana → layout operativo (chat, sesiones, composer).
6. Cerrar la ventana con un stream activo → sin procesos colgados; la próxima
   apertura recupera la sesión intacta.

**Retirada de Fyne (inmediata)**:

```bash
rm -r internal/gui                # GUI Fyne eliminada (código y tests)
go mod tidy                       # retira fyne.io/* y transitivos (glfw/OpenGL)
rg -n "fyne|Fyne" --glob '!specs/**' .  # sin referencias (código, docs, assets)
go test ./... && go build ./...   # suite headless y build verdes sin Fyne
go mod why fyne.io/fyne/v2        # debe devolver "no required" (Gate G-2)
go mod why fyne.io/systray        # idem (Gate G-3)
```

**Esperado**: `letsgo gui` abre la ventana Wails; `go mod why fyne.io/fyne/v2`
vacío; `rg fyne` solo devuelve coincidencias en `specs/` y el checklist de
[gui-contract.md](contracts/gui-contract.md) cubierto 1:1 por tests.

## J4 — Primer arranque y configuración (P3)

**Objetivo**: onboarding y configuración compartida (FR-010, FR-011).

```bash
# borrar la config de keys a propósito
Remove-Item $env:APPDATA\letsgo\config.* -ErrorAction SilentlyContinue
letsgo gui
```

1. Abrir la GUI sin key → flujo de configuración de proveedor/key como pantalla
   inicial; chatear solo tras guardar una key válida.
2. Cambiar modelo en Configuración → guardar → el motor lo aplica sin reiniciar.
3. Verificar en la TUI que el modelo/config persistida es la misma (SC-007: flujo
   key→modelo→chat en ≤4 acciones).
4. Introducir una key inválida o sin conexión → error claro con orientación y
   historial intacto (SC-011).

**Esperado**: los 4 pasos completables; la TUI refleja la misma configuración.

## Escenarios de regresión (Gate B del feature)

- Suite completa: `go test ./internal/... ./cmd/...` verde antes y después de la
  retirada de Fyne.
- Contraste AA de la paleta en ambos temas (smoke CI heredado del 004).
- Transcripts ≥500 mensajes con scroll sin degradación (SC-009).

## J5 — Restyle & UX tipo Codex/Qwen (US5)

**Objetivo**: validar que el restyle (enmienda 2026-08-14) alcanza la filosofía de
diseño de Codex/Qwen sin regresión funcional; dirigido por los gates de
[gui-contract.md](contracts/gui-contract.md) §8 (G1–G10, UX-01..UX-14).

**Prerrequisitos**: implementación de las fases U5-A..D (plan.md US5) + tests
Vitest de cada fase (Test-Last) + `npm run playwright`.

```bash
cd cmd/wails/frontend
npm run check          # svelte-check: 0 errores, sin emoji, tokens en app.css
npm test -- --run      # vitest existentes + nuevos (chat/composer/onboarding/errors + U5)
npm run playwright     # e2e onboarding intacto + nuevos (paleta, overlays, composer)
rg "#[0-9a-fA-F]{3,8}" src/ --glob '*.svelte' --glob '*.ts'   # gate G1: solo app.css
```

**Esperado (gates)**:

1. **G1/G4/G5 (estática)**: cero hex en componentes (solo tokens), spacing en
   escala 4px, cero emoji en chrome, chat ≤760px.
2. **G2/G8 (visual)**: Playwright captura dark+light de (a) chat con bubble +
   bloque asistente + code-block con Copy, (b) paleta, (c) onboarding split,
   (d) empty/loading, (e) overlays no-oscurecen el chat → contraste AA y foco
   visibles en ambos temas; todo `<pre>` es `CodeBlock`.
3. **G3/G6/G7 (integridad)**: flujo 100% por teclado (Ctrl+K ⇄ paleta ↑/↓/Enter ⇄
   Esc, Alt+1..9, Enter send/stop, Tab en settings), sin rail duplicado, y las 10
   superficies con empty+loading+error.
4. **UX-01..UX-11 (comportamiento)**: Retry re-envía el último turno; composer
   nunca se pierde y sigue editable en stream (Send⇄Stop); abrir una sesión entra
   al chat; 1 row por tool ya agrupada; Esc/backdrop cierran Settings/Usage/Help;
   foco restaurado; `aria-busy` en stream; empty+loading en 7/7 superficies.
5. **UX-12..UX-14 (regresión)**: los vitest pre-existentes pasan sin cambios; el
   e2e `onboarding.spec.ts` no se toca; `git diff --stat` (fuera de `specs/`) solo
   toca `cmd/wails/frontend/**` (cero diff de bindings/backend).

**Esperado final**: smoke `letsgo.exe gui` con el dist nuevo; screenshot de
referencia vs. checklist G10 sin fallos de token (los gustos no son fallo de gate).

**Jornada de demo**: abrir la GUI con key → flujo key→modelo→chat (≤4 acciones,
SC-007) → conversar con una herramienta (tool card inline) → Ctrl+K navegar entre
superficies → Esc cierra overlays → sesión de la TUI retomada en 1 clic → guardar
modelo en Configuración verifica en TUI.

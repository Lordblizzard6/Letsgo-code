# Investigación: Doble interfaz — TUI Bubble Tea + GUI Wails

**Branch**: `005-bubbletea-tui-wails` | **Fecha**: 2026-08-11 | **Spec**: [spec.md](spec.md)

## Resumen

El cliente LetsGo pasa a ofrecer dos interfaces sobre un mismo core: la TUI actual
(Bubble Tea, ya en el repo) y una GUI de escritorio **Wails** que **reemplaza por
completo a Fyne** (decisión del usuario: Fyne se elimina, `internal/gui` y
`fyne.io/fyne` se retiran). Este documento resuelve cada [NEEDS CLARIFICATION] o
decisión técnica del plan con su elección, racional y alternativas.

---

## D1: Versión de Wails

**Decision**: Wails **v3** (`v3.0.0-beta.7`, pin exacto — ya en go.mod), API
desktop estable.

**Rationale**: v2 está estable (2.13.0) pero v3 es el rewrite en marcha (beta con
API desktop estable, runtime ESM moderno `@wailsio/runtime`, services sin ctx
oculto → testing trivial). El bump beta.6→beta.7 ya está consolidado en go.mod y el
skeleton (`cmd/wails/main.go`, bindings) se generó con esa versión. Proyecto en
2026 = v3.

**Alternatives considered**:
- v2.13.0 (estable, no-beta): `townhouse` — fallback si el equipo no acepta beta.
- RFC/plugin alternativos (fyne v2.8, walk, webview directo): descartados (Fyne
  se elimina por decisión del usuario; walk/webview directo = reinventar Wails).

## D2: Scaffold, frontend y método de ejecución de la GUI

**Decision**: `wails3 init` con plantilla **Svelte + TypeScript** (build estático a
`frontend/dist`). **La app Wails se ejecuta por el MISMO ejecutable de siempre**:
el comando cobra `letsgo gui` (`cmd/gui.go`) importa un paquete reutilizable
(`cmd/wails/gui`, que contiene el `//go:embed` de `frontend/dist` +
`application.New` + `Run()`), igual que hoy lanza Fyne. `wails3 dev` queda solo
para desarrollo frontend con hot-reload; la distribución es el mismo `letsgo.exe`.

**Rationale**: el usuario exige "que su método de ejecución sea el mismo (mediante
el exe gui)". Cobra ya resuelve `letsgo gui`; exponer `Run()` de Wails como paquete
evita un segundo binario y mantiene un único entry (`cmd/gui.go`). Wails v3 permite
iniciar la app desde cualquier función (no exige `package main`).

**Alternatives considered**: binario Wails separado (`wails3 build -nsis` como
entry propio) — descartado por directiva del usuario (una sola GUI, mismo exe);
React-TS (peso similar, más ecosistema); vanilla-TS y Vue (equivalente).

## D3: Arquitectura de streaming (TUI y GUI)

**Decision**: Dos adaptadores sobre **un único core de chat** (`internal/engine`):

- **TUI**: mantiene su loop Bubble Tea actual (bubbles viewport + glamour para
  markdown) consumiendo el mismo bus de eventos del core.
- **GUI**: servicios Wails bound (`ChatService` etc.); el stream se emite desde Go
  con **eventos batcheados** `chat:delta` (~50 ms de acumulación) vía Event bus de
  Wails; el frontend renderiza con **streaming-markdown + DOMPurify** (nunca
  reparsear el acumulado ni `innerHTML`); los mensajes se guardan en markdown
  crudo (una sola fuente), glamour en TUI, streaming-markdown en GUI.

**Rationale**: `internal/engine` ya existe y ya sirve a TUI + GUI Fyne; el patrón
"one engine, multiple frontends" está probado (Charm soft-serve). El rendering
incremental + sanitización sobre el texto acumulado evita tanto el lag de
re-render de markdown como el riesgo de prompt-injection del contenido del LLM.

**Alternatives considered**: render markdown en Go con glamour y enviar HTML
(requiere sanitizador igualmente y pierde render incremental); eventos por token
(IPC JSON flood — se descarta, se batean).

## D4: Requisito de plataforma (WebView2) y distribución

**Decision**: Estrategia **Evergreen** con `-webview2 embed` (bootstrapper
embebido, ~150 KB) para Windows; Windows 11 lo trae preinstalado. El binario de
distribución es `letsgo.exe` (con el frontend embebido y el bootstrapper), con
instalador opcional vía `wails3 build -nsis` apuntando al entry de la app.

**Rationale**: Un cliente de chat para desarrolladores puede asumir WebView2 o la
instalación automática vía bootstrapper. Fixed Version (~250 MB) es excesivo salvo
despliegue air-gapped.

**Alternatives considered**: Fixed Version (descartada por tamaño), `Error`
(descartada por UX).

## D5: Testing

**Decision**: Tres capas, todas en CI:

1. **Core/motor**: tests Go headless (la lógica de la suite Fyne `test` se porta a
   tests de servicio/engine — misma cobertura, ~1:1).
2. **TUI**: tests de modelo puro (drive `Update`/`Msg` sintético, sin terminal) +
   un puñado de tests con `teatest`/programa completo headless.
3. **GUI**: **Vitest + jsdom** para componentes (mock de `bindings/`) y
   **Playwright** contra el servidor web de `wails dev` (puerto Vite) en CI.

**Rationale**: Los métodos bound de Wails v3 son métodos Go normales sin ctx oculto
→ unit test directo. El dev server web permite E2E headless sin ventana nativa.

**Alternatives considered**: runner nativo headless de Wails (no existe de
primera mano); E2E solo manual (insuficiente para Gate B: "0 regresiones").

## D6: Retirada de Fyne (INMEDIATA)

**Decision**: **Retirada inmediata y total** (no gate final): borrar `internal/gui`
(código y tests Fyne), reescribir `cmd/gui.go` para lanzar el paquete Wails
(`cmd/wails/gui.Run()`), eliminar todos los imports de `fyne.io/*`, `go mod tidy`,
retirar también `fyne.io/systray`, borrar assets generados por `fyne bundle` y
`FyneApp.toml`, migrar iconos a `build/` de Wails y portar la suite headless Fyne
(`fyne test`) a tests Go de engine/servicio + Vitest/Playwright.

**Rationale**: el usuario pidió "remover toda referencia a Fyne YA" y pasar
directamente a implementar Wails con el pin `v3.0.0-beta.7` ya en go.mod y el
esqueleto de services/bindings ya presente. `go mod tidy` + `go mod why` verifican
que nada arrastra Fyne (glfw/OpenGL desaparecen del build; Wails solo necesita C en
Linux por webkit2gtk). Eliminar primero evita que el frontend Fyne y el Wails
coexistan y dupliquen mantenimiento.

**Alternatives considered**: mantener Fyne como fallback (descartado por el
usuario: una sola GUI); migración incremental pane a pane (descartada: el esqueleto
Wails ya está y el usuario pide eliminación directa).

## D7: Sincronización TUI ↔ GUI

**Decision**: Misma sqlite + misma configuración (`internal/db`, `internal/config`)
ya compartidas; ejecución simultánea de ambas interfaces **fuera del alcance v1**
(cada lanzamiento abre su own lock/uso; compatible con la arquitectura actual).

**Rationale**: Es lo que ya hace el proyecto (TUI y GUI Fyne comparten DB/config
hoy). No añadir sincronización cross-proceso en procesos separados para v1.

**Alternatives considered**: proceso único con reuso de core en memoria
(descartado para v1: mayor riesgo, sin requisito de usuario).

## Retroalimentación al SPEC

- La GUI Wails cubre el **rail completo** de 10 destinos (chat, git, tasks, mcp,
  plugins, settings, usage, help, theme, account) + paleta Ctrl+K + atajos
  Alt+1..8 + estados vacíos/de carga → reflejado en FR-017/FR-018, SC-010/SC-012.
- Wails v3 en beta → riesgo asumido con pin exacto (D1).
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

---

# Enmienda US5 — Restyle & UX (2026-08-14)

Investigación de @UI Designer y @UX Researcher sobre las GUI de **OpenAI Codex
desktop** y **Qwen Code (Desktop/Web Shell)** para estilizar la GUI Wails de
LetsGO con la misma filosofía de diseño y gates. Alcance: **solo frontend**
(Svelte + `app.css` + `lib/store.svelte.ts`); sin cambios en bindings/servicios/
eventos/motor.

## D8: Sistema de diseño (tokens, tipografía, layout y componentes)

**Decision**: Adoptar la filosofía convergente Codex/Qwen — **dark-first, chrome
mínimo, acento único restringido (indigo-violeta), el chat como canvas (bubble
usuario / bloque asistente sin card por mensaje), bloques de código como objetos
elevados con header+copy+highlight, tool/plan actividad inline en el turno**. Todos
los valores viven en `app.css` como semantic tokens (`var(--bg-surface)`,
`var(--text-secondary)`, `html[data-theme="dark"|"light"]`); hex literal prohibido
en componentes.

**Coordenadas de datos (propuestas; verificación final como gate G2 en CI)**:

- **Dark (base `#0e1116`)** · surfaces `#14181e`→`#222833` · borde hairline `#2e3540`
  (input `#3d4654`) · texto `#e8ebf0`/`#9aa3b0`/`#6f7885` (15.8/7.4/3.8:1) ·
  acento indigo `#8b94f8` (6.9:1) con `--accent-on #101218` · semánticos
  ok `#4ade80`, warn `#fbbf24`, err `#f87171` (siempre con icono+texto, no color solo).
- **Light (base `#f6f7f9`)** · surface `#ffffff`/`#eef0f3` · borde `#dfe3e8`/`#c4cbd4`
  · texto `#1c2026`/`#5c6570`/`#8b939e` (16.4/5.9/3.2:1) · acento `#4f46e5` (6.3:1)
  · semánticos ok `#15803d`, warn `#b45309`, err `#dc2626`.
- **Sombras**: `--shadow-sm/md/lg` (overlays llevan sombra dura; la página surface no)
  y `--overlay` scrim.
- **Tipografía**: stack UI Inter/system-ui (peso 400 base, 500/600 para énfasis) +
  mono JetBrains Mono (sustituye a Cascadia Mono); escala 11/13/14/15/16/18/20/24
  (base chat 14px/22); prosa line-height 1.6; `tabular-nums` en números;
  etiquetas caps 11px tracking 0.08em.
- **Spacing/radius/breakpoints**: escala 4px (4/8/12/16/20/24/32/40/48); radius
  sm/md/lg/xl 6/8/10/12 + pill; chat col ≤760px; rail 48px (icon-only) colapsable a
  56px con tooltip `Label Alt+n`; ≥1600px modo dos paneles (chat + sesiones).
- **Componentes**: rail con SVG stroke icons (sin emoji) + marcador activo único
  (accent icon + barra 3px + `--accent-soft`); status line inferior 28px (proveedor·
  modelo·branch·uso·dot busy) — elimina las pills flotantes; composer como command
  shell docked (model chip · plan toggle · @file · Send⇄Stop · hint row `Ctrl+K
  commands · / plan · @ files · Alt+1-8 surfaces`); message bubble (usuario) /
  bloque full-width (asistente) con header-per-message (mark+Assistant+timestamp+
  estado/modelo) y cursor stream; CodeBlock.svelte (header+language+Copy+highlight,
  cursor `▍`); tool cards inline colapsables + group-compaction (Qwen
  `compactInline`); sesiones agrupadas por fecha + status dot; paleta 640px
  grouped + highlight + footer hints; onboarding split-screen (brand 45% + form 55%);
  skeletons (sin spinners salvo <150ms).

**Rationale**: los dos referentes convergieron (chrome quieto, densidad aireada,
acento sobrio, chat como canvas, keyboard-first) y ambos exponen su tema como
**tokens semánticos** (`codex-theme-v1`; Qwen key set) → el sistema de tokens CSS
es el mecanismo correcto y deja swap futuro del acento sin tocar componentes.
Un solo acento indigo-violeta hereda la familia Qwen sin perder el neutralismo de
Codex. Preserva las coords AA ya presentes en app.css (mismo rango de contraste).

**Alternatives considered**: acento estrictamente azul Codex (`#5c99d6`) — descartado
por herencia Qwen del producto (se puede swap solo en tokens); doble acento
(brand + status) — descartado (restricción de un solo acento); card-por-mensaje
(mantener actual) — descartado por P8 (densidad tranquila exige quitar el
card-every-message); emoji como iconografía — descartado por G5 (inconsistente,
"toy-grade", rompe densidad).

## D9: Backlog de UX priorizado y gates (backlog: prioridades)

**Decision**: Backlog de 23 mejoras priorizadas **P0/P1/P2** del @UX Researcher
(files y notas en el reporte; resumen abajo). Principios benchmark: P1 chat-first,
P2 composer siempre-en y editable durante stream (Send↔Stop un slot), P3 tool
inline agrupada por turno, P4 disclosure progresiva, P5 errores calm/actionable
(Retry reproduce el turno), P6 continuidad de sesión (resume 1 acción), P7
keyboard-first, P8 densidad tranquila, P9 onboarding de un solo objetivo.

- **P0 (U5-A)**: I-1 composer pinned; I-2 Retry re-envía el último mensaje de
  usuario; I-3 paleta navegable ↑/↓+Enter+focus restore; I-4 abrir sesión navega a
  chat (`window.location.hash=""`); I-5 Esc cierra Settings/Usage/Help + refocus
  composer (backdrop-click también); I-6 dedupe tool rows (1 row/call en
  `store.svelte.ts`, update por `callId`); I-7 affordance "Streaming — Enter stops".
- **P1 (U5-B)**: I-8 densidad por turnos (sin cards por mensaje; código = el card,
  con Copy); I-9 tool activity inline/agrupada colapsable; I-10 iconos SVG + mover
  theme/account a chrome (rail/status line) y consolidar `rail.svelte` duplicado;
  I-11 Help alignado con TABS; I-12 sesiones empty CTA + "Refrescando…" + fecha/
  recuento; I-13 settings no-modal drawer Esc-able (mismos calls `SaveConfig`).
- **P1 (U5-C)**: I-14 empty+loading con CTA en 7+ superficies; I-15 foco/a11y
  (aria-live/busy, focus trap paleta, reduced-motion); I-16 header status strip
  (provider·model·hoy — datos de useAccount/useConfig).
- **P2 (U5-D)**: I-17 render streaming memoizado; I-18 syntax highlight + toggle
  raw (Alt+M); I-19 tabla per-provider + wording budget; I-20 deep-links flyout;
  I-21 onboarding autofocus + model "Advanced"; I-22 rename/delete sesión; I-23
  tips en empty chat.

**Rationale**: los 7 P0 son violaciones directas de los principios benchmark
(F-008 composer siempre-en, FR-013 retry actionable, SC-003 keyboard 100%,
continuidad de sesión, tool no duplicado) y todos son solo-frontend. La
estabilidad de tests (UX-12..UX-14: vitest existentes sin cambios, e2e onboarding
intacto, cero diff de bindings/backend) permite iterar el restyle en fases con
gate de regresión en cada una.

**Alternatives considered**: regrabar el contrato de eventos o services para
"mejorar" UX — descartado (viola II, fuera de alcance); adoptar un framework de
UI (Tailwind/shadcn) — descartado (V: un solo set de tokens CSS ya estandariza,
sin nueva toolchain); rediseñar la navegación (sidebar grande al estilo Codex)
— descartado para v1 (el rail ya cumple el gui-contract y P4 referee disclosure;
solo se re-viste).
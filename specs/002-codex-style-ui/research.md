# Research: Codex-Style UI (Phase 0)

**Feature**: 002-codex-style-ui | **Date**: 2026-08-06

Base: investigación previa del UI/UX de Codex (repo openai/codex, TUI + GUI de escritorio) + análisis del estado actual de LetsGO (Fyne v2.8).

## 1. Qué hace Codex (referencia de diseño)

Hallazgos clave de la investigación (TUI y GUI desktop):

- **Transcript como interfaz única**: el chat es el centro; herramientas, errores y aprobaciones viven dentro del transcript/composer, no en ventanas separadas.
- **Composer como sala de control**: Enter = steer (inyecta instrucción al turno en curso), Tab = queue (encola como siguiente turno), `@`/`!`/`/` = file picker / shell passthrough / comandos. El composer **nunca se deshabilita**.
- **Approvals inline sobre el composer**: el overlay de aprobación cubre solo la zona del input; el transcript permanece visible y scrolleable. Enter aprobar, Esc rechazar. Cola visible ("N pendientes").
- **Status indicator**: spinner + tiempo transcurrido + paso actual + `esc to interrupt`; se oculta durante streaming de texto y reaparece entre ráfagas de actividad.
- **Plan mode**: modo visible; el agente propone pasos; el usuario aprueba/rechaza/edita antes de escribir archivos.
- **Session picker al inicio**: Nueva / Reanudar última / Elegir sesión, filtrable por cwd y fecha; forks como copias independientes.
- **Grants por sesión**: "Permitir siempre en esta sesión" para categorías de herramientas; revocables; nunca globales.
- **Status line configurable**: modelo, rama, modo, uso de contexto, rate limit, versión.
- **Paleta de comandos** (Ctrl+K / `/`): acciones de la app, filtrable, solo teclado.
- **Cheat sheet `?`**: atajos categorizados, filtrable.
- **Identidad**: el TUI usa ANSI magenta (sin hex custom, enforced por clippy); la GUI de escritorio es la referencia de panes/jerarquía/densidad. **Decisión Q1=C**: LetsGO no replica el magenta; adopta identidad propia con acento distintivo, tema oscuro.

## 2. Estado actual de LetsGO (Fyne) — brecha

| Capacidad | Hoy | Objetivo |
|---|---|---|
| Aprobaciones | `dialog.NewCustomConfirm` modal (tapa transcript) | Tarjetas inline sobre composer + cola visible |
| Composer | Se deshabilita durante turno activo | Siempre editable: Enter steer, Tab queue |
| Estado de trabajo | Barra simple (statusbar.go) | Franja Working con spinner/tiempo/paso/esc-to-interrupt |
| Modo Plan | No existe como estado visible | Modo Plan + tarjeta de plan |
| Inicio | Sesión directa / lista en menú | Picker Nueva/Reanudar/Elegir/Fork |
| Grants | Auto-approve global por categoría (config) | + grants por sesión (DB), revocables |
| Paleta / cheat sheet | No existen | Ctrl+K paleta + overlay `?` |
| Status line | No configurable | Campos configurables (viper) |

## 3. Decisiones técnicas para Fyne v2.8

### 3.1 Superficies y layout (guía de diseño propia, no réplica literal)

- **Layout**: `Border` de 3 franjas: transcript (centro), franja de estado (top, colapsable/auto), composer + zona de tarjetas de aprobación (bottom). Sin paneles laterales fijos.
- **Acento distintivo LetsGO**: color de acento único definido en `theme.go` (constante `AccentColor`) usado en: estado "Working", foco del composer, borde de tarjetas de aprobación, selección de paleta. El resto de la paleta se deriva del tema oscuro existente para preservar contraste/legibilidad.
- **Densidad**: padding/espaciado reducidos respecto a widgets por defecto vía `widget.NewRichText`/estilos propios (patrón ya usado en chatview.go/markdown.go).

### 3.2 Aprobaciones inline (FR-001..004, 016, 017, 021, 022)

- **Modelo de cola en GUI**: `approvalQueue []ApprovalDecision` en una superficie dedicada (`approvalcard.go`) sobre el composer; el transcript nunca se tapa (no usa `dialog`).
- **Operación solo teclado**: foco en la tarjeta activa; Enter → `ApproveTool`, Esc → `RejectTool` con motivo estructurado por defecto ("user rejected with Esc"). Dif f resaltado: reutilizar renderer markdown existente.
- **Cierre de ventana con pendientes** (FR-022): el hook `Window.OnClosed` (o `fyne.CurrentApp().Driver().Canvas` teardown) emite `RejectTool` por cada pendiente antes de `Stop`.
- **Grants (FR-016/017)**: el botón "Permitir siempre en esta sesión" en la tarjeta → comando `GrantSession{Category, SessionID}` → `internal/db` tabla `session_grants`. El PermissionDriver consulta primero la negación persistente (siempre gana), luego grant de sesión, luego auto-approve global, luego pregunta.
- **Semántica del motor**: los comandos nuevos (`Steer`, `Queue`, `GrantSession`, `ForkSession`) se añaden al contrato **interno** (ver 3.3). No cambia el contrato público de eventos existente.

### 3.3 Engine: contrato interno ampliado

Comandos nuevos (en `commands.go`, interfaz interna del loop):

- `Steer{Text, SessionID}` — inyecta instrucción al turno en curso (FR-009).
- `Queue{Text, SessionID}` — encola texto como siguiente turno sin ejecutar (FR-010).
- `SetPlanMode{Mode}` — Plan/Execute/Auto; en Plan, la UI presenta la tarjeta y el engine **no ejecuta herramientas de escritura** hasta aprobación (FR-012).
- `ApprovePlan{PlanID}`, `RejectPlan{PlanID}`, `EditPlan{PlanID, Instructions}` — ciclo de la tarjeta de plan (FR-013).
- `GrantSession{Category, SessionID, Allow}` — persistir/revocar grant (FR-016/017).
- `ForkSession{SourceID, MessageID}` — clonar sesión (FR-015).

Eventos nuevos (sin romper el público): `PlanProposed`, `PlanDecided`, `ModeChanged`, `GrantsChanged`.

Cómo encaja con el loop existente: `SendMessage`/`SendTask` ya persisten el mensaje y abren stream; `Steer` sigue el mismo camino de persistencia (mensaje con flag `is_steer`) y reanuda el stream del turno en curso. `Queue` persiste el mensaje con flag `is_queued` y lo ejecuta al llegar `Idle`.

### 3.4 Franja de estado (FR-005, 006, 020)

- Derivada de eventos existentes: `StreamStart`/`StreamDelta`/`StreamDone`, `ToolRequested`/`ToolExecuting`/`ToolResult`, `AgentTaskUpdate`, `Idle`, `ErrorEvent`.
- Ocultar en streaming de texto (`StreamDelta` activo) y reaparecer entre ráfagas de herramientas (`ToolExecuting`).
- Desactualización (FR-020): timer de 15 min sin actualización de contexto/rate → marca "desactualizado" (estilo atenuado + icono).
- Esc → `Cancel` existente (ya detiene stream); verificar ≤500ms en tests (SC-005).

### 3.5 Composer: steer/queue (FR-008, 009, 010)

- `Entry` nunca `Disable()` durante turno activo.
- En estado activo: Enter con texto → `Steer`; Tab con texto → `Queue` (con indicación visual "encolado →"); en reposo: Enter → `SendMessage`/`SendTask` (comportamiento actual).
- Indicador de modo en el composer (steer/queue/plan) para evitar ambigüedad con Tab (que también navega foco): Tab solo actúa como queue si hay texto y turno activo; si no, comportamiento normal de foco.

### 3.6 Paleta Ctrl+K / `/` (FR-011, 026)

- Superficie `palette.go`: overlay `widget.NewCard`/Panel con `Entry` de filtro + lista filtrable; solo teclado (↑↓ navega, Enter ejecuta, Esc cierra).
- Acciones internas: registradas desde un catálogo (nueva sesión, cambiar sesión, compactar, copiar, fork, grants, settings, atajos, etc.), categorizadas como Codex.
- `@`: file picker con búsqueda difusa sobre el proyecto (walk del cwd respetando `.gitignore`; coincidencia por subcadena/prefix + ranking simple; sin dependencias externas). Inserta `@ruta` en el composer.
- `!`: shell passthrough — mapeado al comando existente de ejecución con permisos; **nunca** exime del PermissionDriver (assumption de seguridad).

### 3.7 Picker de inicio y fork (FR-014, 015)

- `sessionpicker.go`: al arrancar (si hay sesiones previas) overlay con Nueva / Reanudar última / Elegir (lista filtrable por cwd y fecha — `internal/db.ListSessions` ya devuelve estos campos).
- Fork: `ForkSession` → nueva fila en `sessions` (copia de metadata) + copia de mensajes hasta `MessageID`; la original intacta.

### 3.8 Config (FR-018) y cheat sheet (FR-019)

- Status line fields: claves viper `statusline.show`, `statusline.fields[]` (modelo/rama/modo/contexto/rate/versión) persistidas con el mecanismo existente (`internal/config.SaveConfig`, chmod 0600).
- `keymap.go`: overlay `?` con atajos categorizados (navegación, edición, sesiones, herramientas, paleta) filtrable; Esc cierra.

## 4. Riesgos y mitigaciones

| Riesgo | Mitigación |
|---|---|
| Enter/Tab en composer chocan con comportamiento de foco de Fyne | Detección de turno activo + texto no vacío; tests headless de teclado |
| Grants por sesión podrían sugerir debilitamiento de seguridad | Regla dura: negación persistente > grant de sesión > auto-approve global (test de contrato) |
| Paleta `@` con proyecto grande (walk lento) | Walk en goroutine con cancelación; resultado incremental; límite de resultados |
| Rediseño rompe atajos existentes | Test de humo sobre FR-023 (todos los atajos registrados en keyboard.go) |
| Overlay de aprobación sobre composer sin modal complica el foco | Una sola tarjeta activa con foco; las demás en cola (numeradas) |

## 5. Fuentes

- Repo openai/codex (TUI/GUI): transcript-first, composer steer/queue, approvals inline, plan mode, session grants, statusline.
- Especificación 002 (FR-001..026, SC-001..010, entidades).
- Código existente: `internal/gui/*`, `internal/engine/{commands,events,permission_driver}.go`, `internal/db/database.go`, `internal/config/config.go`.

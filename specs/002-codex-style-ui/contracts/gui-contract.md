# Contract: GUI Surface v2 — Codex-Style (`internal/gui`)

**Feature**: 002-codex-style-ui | **Date**: 2026-08-06 | **Supersedes**: contracts/gui-contract.md de 001 (rediseño)

La superficie Fyne rediseñada: layout, acciones, mapa de teclado y garantías de comportamiento para tests de aceptación. Detalles de widgets internos van a `tasks.md`.

## Layout (una sola ventana, sin sidebar)

```text
┌──────────────────────────────────────────────────────────────┐
│ Work status strip (auto): Working ● spinner · paso actual    │
│ · tiempo · esc to interrupt        [statusline configurable] │
├──────────────────────────────────────────────────────────────┤
│ Transcript (chat view + inline tool panels + diffs)          │
│  — siempre visible y scrolleable (FR-001, SC-007)            │
│  — marcas sutiles: steer / en cola / plan instructions       │
├──────────────────────────────────────────────────────────────┤
│ Approval queue zone (0..n cards, inline sobre el composer)   │
│  Card: tool name · summary · diff · [Enter] [Esc] [Session]  │
├──────────────────────────────────────────────────────────────┤
│ Composer (SIEMPRE editable)                                   │
│  [@ file picker] [! shell] [/ comandos] [texto multiline]    │
│  Enter=steer (activo) · Tab=queue (activo) · Ctrl+Enter=newline
└──────────────────────────────────────────────────────────────┘
```

- Sin ventanas modales que tapen el transcript; el approval queue vive en la franja inferior (FR-001, SC-007).
- Menús existentes se mantienen como ruta secundaria (assumption).

## Views

| View | Trigger | Contents |
|------|---------|----------|
| Transcript | default | Conversation thread markdown, tool panels inline, marcas steer/queue/plan |
| Approval card (inline) | `engine.ToolRequested` | Tool name, acción, diff resaltado, cola "N pendientes", botón grant de sesión (FR-002/016/021) |
| Plan card (inline) | `engine.PlanProposed` | Pasos/archivos/criterios, Aprobar / Rechazar / Editar (FR-013) |
| Command palette | `Ctrl+K` o `/` en composer | Acciones de app + `@` file picker + `!` shell (FR-011/026) |
| Session picker | arranque con sesiones previas | Nueva / Reanudar última / Elegir (filtro cwd+fecha) / Fork (FR-014/015) |
| Keymap overlay | `?` | Cheat sheet de atajos categorizado, filtrable (FR-019) |
| Status line config | menu/preferencias | Campos on/off: modelo, rama, modo, contexto, rate, versión (FR-018) |
| Settings | `Ctrl+,` | Provider/model/API key; auto-approve; grants por sesión listados/revocables (FR-017) |
| Usage / Git / Plugins / MCP / Agent tasks | menús (ruta secundaria) | Sin cambios funcionales respecto a 001 |

## Actions (mouse and/or keyboard)

| Action | Keyboard | Notas |
|--------|----------|-------|
| Send / steer message | `Enter` (composer) | Reposo → `SendMessage`/`SendTask`; turno activo → `Steer` (FR-009) |
| Queue message | `Tab` con texto en turno activo | Encela como siguiente turno; indicador "en cola" (FR-010) |
| Newline | `Ctrl+Enter` | Igual que 001 |
| Interrupt stream | `Esc` (turno activo) | ≤500 ms; composer operativo (FR-005/007, SC-005) |
| Approve tool card | `Enter` (card enfocada) | FR-003 |
| Reject tool card | `Esc` (card enfocada) | FR-003; motivo estructurado por defecto |
| Grant "permitir en sesión" | `Ctrl+Enter` (card) o botón | FR-016 |
| Command palette | `Ctrl+K` | FR-011; también `/` en composer |
| Keymap overlay | `?` | FR-019; Esc cierra |
| New session | `Ctrl+N` | FR-023 (preservado) |
| Switch session | `Ctrl+1..9` | FR-023 (preservado) |
| Focus composer | `Ctrl+L` | FR-023 (preservado) |
| Settings | `Ctrl+,` | FR-023 (preservado) |
| Usage | `Ctrl+U` | FR-023 (preservado) |
| Auto-approve toggle | `Ctrl+Shift+A` | FR-023 (preservado) |
| Fork session | desde picker/paleta | FR-015 |
| Esc en paleta/picker/plan/overlay | `Esc` | Cierra sin acción |

## Appearance (FR-024 — identidad propia, no réplica ANSI)

- Tema oscuro preservado; **acento distintivo LetsGO** (constante `AccentColor` en theme.go) para: estado Working, foco del composer, borde de tarjetas activas, selección de paleta.
- Monospaced en transcript/composer; densidad reducida en tarjetas y franja de estado.
- Marca "desactualizado" en contexto/rate tras 15 min sin actualización (FR-020).
- Sin light theme en v1.

## Behavior guarantees (mapped a FRs)

- Composer NUNCA deshabilitado durante turno activo (FR-008); Enter/Tab behavior contextual.
- Franja de trabajo visible durante actividad de herramientas; oculta durante streaming de texto (FR-006).
- Cierre de ventana con aprobaciones pendientes → rechazo estructurado de cada una + `Stop` (FR-022).
- En modo Plan, el engine no ejecuta escrituras hasta aprobación de la tarjeta (FR-012).
- Todos los atajos 001 (Ctrl+N, Ctrl+1..9, Ctrl+,, Ctrl+U, Ctrl+L, Ctrl+Shift+A) siguen funcionando (FR-023, test de humo).
- Flujo completo solo con teclado: Ctrl+L → enviar → aprobar Enter → Ctrl+N → Ctrl+1 → Esc (SC-003/009).
- `!` shell passthrough: ejecuta únicamente lo permitido por el PermissionDriver (security assumption).

# Quickstart: UI/UX Refinement

**Created**: 2026-08-07 | **Status**: Validación del feature

Referencia: [spec.md](spec.md) · [plan.md](plan.md) · [research.md](research.md) · [contracts/ui-contract.md](contracts/ui-contract.md)

## Prerequisitos

- Windows, toolchain CGO: antes de cualquier comando Go ejecutar en PowerShell (7+):
  ```powershell
  $env:PATH = "D:\msys64\ucrt64\bin;" + $env:PATH
  $env:CGO_ENABLED = "1"
  ```
- Compilar el binario: `build.bat -n` → `letsgo.exe`
- Ejecutar la app: `go run . gui`

## Validación rápida (smoke)

```powershell
go test ./internal/gui/... ./internal/config/...  # suite headless (002/003 + nuevos)
go build ./...                                      # todo compila
```

Los contratos detallados (firmas, tabs, invariantes, tests obligatorios) están en [contracts/ui-contract.md](contracts/ui-contract.md); los journeys de abajo son la comprobación manual/funcional.

## Journey 1 — Rail coherente (US1)

**Paso a paso**
1. Abrir la app; identificar en la barra izquierda tres zonas separadas por hairline.
2. Recorrer la zona superior (Chat), la de herramientas (Git · Tareas · MCP · Plugins, etiqueta "HERRAMIENTAS") y la inferior (Uso · Configuración · Ayuda · Tema · avatar). Verificar que ningún slot duplica otro y solo uno muestra la accent bar activa.
3. Pulsar `Alt+1`..`Alt+8` y comprobar que cada pane muestra el panel esperado (Chat, Git, Tareas, MCP, Plugins, Configuración, Uso, Ayuda).
4. Colapsar el rail (toggle del pie): los iconos muestran tooltip con nombre + atajo ("Git (Alt+2)") al pasar el cursor; el foco por teclado sigue visible sobre el slot.
5. `Alt+Left`/`Esc` vuelve al Chat desde cualquier pane, con el composer enfocado.

**Comprueba**: SC-001, SC-002, SC-003 (parcial), FR-001 a FR-003.

## Journey 2 — Settings por secciones y flyout de cuenta (US2)

1. `Ctrl+,` — abre el overlay de configuración no modal con 4 secciones (Cuenta · Apariencia · Preferencias · Uso). La conversación sigue visible detrás.
2. Cambia el tema en Apariencia y colapsa el rail; `Save` persiste. `Esc`/Cancelar descarta y devuelve el foco al composer.
3. Haz clic en el avatar (abajo-izquierda) — se abre el flyout con estado, provider+model, resumen de uso de hoy y las acciones "API keys & proveedores…", "Uso…", "Cerrar sesión".
4. Desde `Alt+7` abre Uso: tarjetas KPI, tabla por proveedor y barra de presupuesto (si hay cuota) — nada de volcado monoespacial.
5. Cierra todo con `Esc` y comprueba que el foco está en el composer.

**Comprueba**: SC-002, SC-004 (panel Uso), las FR-004 a FR-007.

## Journey 3 — Legibilidad y elegancia (US3)

1. Conversación con mensajes largos: la prosa no supera ~720px, centrada; los bloques de mensaje llevan ritmo (24px).
2. Código en bloques: ancho completo con wrap opcional.
3. Composer: enfoca → ring accent 2px; envía un mensaje largo y observa que el botón principal cambia a "Detener" mientras transmite, en el mismo lugar.
4. Inspección visual: tarjetas (usuario/asistente, aprobación, plan, tool panel) con el mismo marco border 1px radius 6px padding 16px.
5. Microcopy: chrome en español; términos técnicos en inglés; placeholder del composer "Escribe a LetsGo… (Enter enviar · Tab encolar · Esc detener)" y filtro de conversaciones "Filtrar conversaciones…".

**Comprueba**: SC-003 (parcial: foco en conversación), SC-007, FR de legibilidad.

## Journey 4 — Accesibilidad y estados (US4)

1. Pon el tema en claro: todo el texto secundario >=4.5:1, separadores >=3:1 (verificar visualmente o con el smoke test de paleta).
2. Recorre todos los focusable con Tab: siempre se ve un ring accent visible.
3. Vacía una lista (conversaciones, MCP, plugins…) y observa el estado "No hay … todavía." + CTA; recarga alguna lista asíncrona y verifica que no se vacía en blanco (indica "Refrescando…").
4. Iconografías: cada icono del rail corresponde al destino (burbuja = Chat, branch = Git…); en colapsado siempre hay tooltip sobre todos.

**Comprueba**: SC-005, SC-006, FR-011 a FR-013.

## Checklist SC-001..007

- [ ] SC-001: los 5 destinos top (Chat, Git, Configuración, Uso, Cuenta) se localizan ≤2 clics la primera vez; sin slot "Sesiones" que confunda.
- [ ] SC-002: ajusta modelo → cambia tema → vuelve, en ≤3 acciones (rail/Ctrl+, → section → volver).
- [ ] SC-003: recorrido completo (rail, panes, paleta Ctrl+K, keymap) 100% por teclado, focus visible en ambos temas.
- [ ] SC-004: el panel Uso responde "¿cuánto gasté hoy?" ≤2s (KPI visible a primer golpe de vista).
- [ ] SC-005: pares de texto ≥4.5:1 y no-texto ≥3:1 en ambos temas (palette smoke CI).
- [ ] SC-006: ≥6 superficies con estado vacío + CTA; ninguna se vacía en blanco en refresh.
- [ ] SC-007: decisión de aprobación ≤2s (p90), transcript siempre visible (sin regresión con 003).

## Out of scope (futuro)

- Gestión global de sesiones como pane independiente (extensión futura documentada).
- Migración del tema claro completo en Apariencia (ya existe; solo se corrige el contraste de `lightPlaceholder`).
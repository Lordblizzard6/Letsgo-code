# Implementation Plan: Modernización de la GUI — Layout de3 columnas

**Feature Branch**: `007-gui-modernization`

**Created**: 2026-08-27

**Reference**: `specs/007-gui-modernization/spec.md`

## Technical Context

- **Framework**: Svelte5 (runes mode) — se mantiene, no se migra a React
- **Build**: Vite + TypeScript
- **Backend**: Go (Wails v2) — sin cambios a bindings
- **Tests**: Vitest + Playwright
- **Reference Design**: `ejemplo de gui/gui-shell/` (React19 + Vite7)

## Architecture

### Layout de3 columnas

```
┌─────────────────────────────────────────────────────────┐
│ #app-header (50px, flex, bg-panel, border-bottom)       │
│ [menu] [logo brand] [project] [account▾] [model▾]     │
│ [usage-chip] [explorer] [settings] [palette Ctrl+K]    │
├──────────┬──────────────────────────┬───────────────────┤
│ #sidebar │ #chat-pane               │ #right-pane       │
│ (260px)  │ (flex:1)                 │ (320px)           │
│          │                          │                   │
│ Reciente │ #tabs-bar                │ .explorer-panel   │
│ ──────── │ ─────────────────        │ .activity-panel   │
│ Sesiones │ #chat-viewport           │   .tool-card      │
│          │   .msg.user/.assistant   │   .tool-card      │
│          │   .tool-line             │                   │
│          │ .disclaimer              │                   │
│          │ #composer                │                   │
│          │   #cmd-palette           │                   │
│          │   #attachment-list       │                   │
│          │   #composer-box          │                   │
│          │   #composer-hint         │                   │
└──────────┴──────────────────────────┴───────────────────┘
```

### Design Tokens (CSS Custom Properties)

Referencia: `ejemplo de gui/gui-shell/src/App.css:1-80`

```css
:root {
  /* Dark theme (default) */
  --bg: #0a0e15;
  --bg-panel: #212631;
  --bg-elevated: #373f4e;
  --border: #4e576a;
  --text: #e6eaf0;
  --accent: #7c5cff;
  --accent-soft: rgba(124, 92, 255, 0.16);
  --danger: #e5484d;
  --success: #34c98a;
  --code-bg: #07090e;
  --mono: "JetBrains Mono", "Consolas", monospace;
}

[data-theme="light"] {
  --bg: #f4f6f9;
  --bg-panel: #e6e9ee;
  --accent: #6a4bef;
  /* ... */
}
```

### Componentes principales

| Componente | Archivo | Descripción |
|-----------|---------|-------------|
| App | `App.svelte` | Layout de3 columnas + header |
| Sidebar | `sidebar.svelte` | Proyectos recientes + sesiones |
| TabsBar | `tabs-bar.svelte` | Barra de pestañas de chat |
| ChatViewport | `chat-viewport.svelte` | Viewport de mensajes |
| MessageRow | `message-row.svelte` | Fila de mensaje (user/assistant/tool) |
| Composer | `composer.svelte` | Áreá de entrada de texto |
| CmdPalette | `cmd-palette.svelte` | Paleta de comandos / |
| GlobalPalette | `global-palette.svelte` | Paleta Ctrl+K |
| ActivityPane | `activity-pane.svelte` | Panel de actividad de herramientas |
| ExplorerPanel | `explorer-panel.svelte` | Explorador de archivos |
| ToolCard | `tool-card.svelte` | Tarjeta de herramienta |
| SettingsModal | `settings-modal.svelte` | Modal de configuración |
| SpendModal | `spend-modal.svelte` | Modal de gasto |
| Header | `header.svelte` | Header de50px |

## Phases

### Phase 1: Design Tokens + Layout Base (R1)

**Propósito**: Establecer los design tokens del gui-shell y el layout de3 columnas.

- [ ] R1.1 Reescribir `app.css` con los design tokens del gui-shell (dark/light)
- [ ] R1.2 Crear `header.svelte` (50px, flex, controles)
- [ ] R1.3 Crear `sidebar.svelte` (260px, colapsable, proyectos + sesiones)
- [ ] R1.4 Crear `activity-pane.svelte` (320px, explorador + actividad)
- [ ] R1.5 Reescribir `App.svelte` con layout de3 columnas
- [ ] R1.6 Test: vitest verde, svelte-check 0 errores

**Gate**: Layout de3 columnas visible, header funcional, sidebar colapsable.

### Phase 2: Header + Selectors (R2)

**Propósito**: Implementar los selectors de cuenta, modelo y proyecto en el header.

- [ ] R2.1 Crear `account-selector.svelte` (dropdown de cuentas)
- [ ] R2.2 Crear `model-selector.svelte` (dropdown de modelos)
- [ ] R2.3 Crear `project-selector.svelte` (dropdown de proyectos)
- [ ] R2.4 Crear `usage-chip.svelte` (tokens + costo)
- [ ] R2.5 Integrar selectors en `header.svelte`
- [ ] R2.6 Test: selectors abren/cierran, selección funciona

**Gate**: Header con todos los selectors operativos.

### Phase 3: Tabs + Chat Viewport (R3)

**Propósito**: Implementar barra de tabs y viewport de mensajes.

- [ ] R3.1 Crear `tabs-bar.svelte` (múltiples pestañas, add/close)
- [ ] R3.2 Crear `chat-viewport.svelte` (scroll, stick-to-bottom)
- [ ] R3.3 Reescribir `message-row.svelte` (user burbuja, assistant full-width, tool línea)
- [ ] R3.4 Integrar markdown rendering (highlight.js)
- [ ] R3.5 Test: tabs funcionan, mensajes se renderizan

**Gate**: Chat con múltiples pestañas y mensajes correctos.

### Phase 4: Composer + Mode Selector + Cmd Palette (R4)

**Propósito**: Modernizar el composer con attach, mode selector y command palette.

- [ ] R4.1 Reescribir `composer.svelte` (attach + textarea + send/stop)
- [ ] R4.2 Crear `mode-selector.svelte` (dropdown code/arch/planning/research)
- [ ] R4.3 Crear `cmd-palette.svelte` (paleta / con filtrado)
- [ ] R4.4 Crear `attachment-chip.svelte` (chips de archivos adjuntos)
- [ ] R4.5 Integrar composer en chat pane
- [ ] R4.6 Test: composer funciona, cmd palette navegable

**Gate**: Composer con attach, send/stop, mode selector y cmd palette.

### Phase 5: Global Palette Ctrl+K (R5)

**Propósito**: Implementar la paleta de comandos global.

- [ ] R5.1 Crear `global-palette.svelte` (overlay + input + items agrupados)
- [ ] R5.2 Integrar Ctrl+K handler en App.svelte
- [ ] R5.3 Implementar focus trap y keyboard navigation
- [ ] R5.4 Test: Ctrl+K abre, ↑/↓ navega, Enter ejecuta, Esc cierra

**Gate**: Paleta global funcional con keyboard navigation.

### Phase 6: Activity Pane + Tool Cards (R6)

**Propósito**: Implementar el panel de actividad de herramientas.

- [ ] R6.1 Crear `tool-card.svelte` (running/done/error states)
- [ ] R6.2 Crear `explorer-panel.svelte` (file browser)
- [ ] R6.3 Integrar activity pane en right-pane
- [ ] R6.4 Test: tool cards muestran estados correctamente

**Gate**: Activity pane con tool cards y explorer funcional.

### Phase 7: Modales Settings + Spend (R7)

**Propósito**: Implementar los modales de configuración y gasto.

- [ ] R7.1 Crear `settings-modal.svelte` (overlay + tabs + body)
- [ ] R7.2 Crear `spend-modal.svelte` (overlay + gráfico + tabla)
- [ ] R7.3 Implementar focus trap y Esc close
- [ ] R7.4 Test: modales abren/cierran, focus trap funciona

**Gate**: Modales settings y spend funcionales.

### Phase 8: Estados Vacíos + Loading + Error (R8)

**Propósito**: Implementar estados de UI en todas las superficies.

- [ ] R8.1 Crear `empty-state.svelte` (icono + título + CTA)
- [ ] R8.2 Crear `skeleton.svelte` (shimmer ≤150ms)
- [ ] R8.3 Crear `error-banner.svelte` (orientación + retry)
- [ ] R8.4 Crear `thinking-dots.svelte` (3 puntos parpadeantes)
- [ ] R8.5 Aplicar estados a las10 superficies del rail
- [ ] R8.6 Test: empty/loading/error en7+ superficies

**Gate**: Estados de UI consistentes en todas las superficies.

### Phase 9: Sidebar Proyectos + Sesiones (R9)

**Propósito**: Implementar la lista de proyectos recientes y sesiones en el sidebar.

- [ ] R9.1 Implementar `recent-projects.svelte` (folder icon + nombre + delete)
- [ ] R9.2 Implementar `sessions-list.svelte` (título + badge + fecha + delete)
- [ ] R9.3 Implementar empty states accionables
- [ ] R9.4 Implementar colapsado del sidebar con animación220ms
- [ ] R9.5 Test: sidebar muestra proyectos y sesiones

**Gate**: Sidebar con proyectos y sesiones funcionales.

### Phase 10: Regresión y Polish (R10)

**Propósito**: Verificar que todo funciona y aplicar polish final.

- [ ] R10.1 Ejecutar vitest completo (sin cambios en tests pre-existentes)
- [ ] R10.2 Ejecutar svelte-check (0 errores)
- [ ] R10.3 Ejecutar npm run build (build exitoso)
- [ ] R10.4 Ejecutar playwright e2e (onboarding intacto)
- [ ] R10.5 Verificar gates G1-G10
- [ ] R10.6 Aplicar polish visual (sombras, bordes, transiciones)

**Gate**: Todos los tests pasan, gates verificados, polish aplicado.

## Checkpoints

- **Checkpoint1** (R1-R2): Layout de3 columnas + header con selectors
- **Checkpoint2** (R3-R4): Chat con tabs + composer modernizado
- **Checkpoint3** (R5-R6): Global palette + activity pane
- **Checkpoint4** (R7-R8): Modales + estados de UI
- **Checkpoint5** (R9-R10): Sidebar + regresión completa

## Risks

1. **Riesgo**: El layout de3 columnas puede romper el e2e existente
   - **Mitigación**: Mantener las rutas hash y atajos intactos; testear después de cada fase

2. **Riesgo**: Los design tokens del gui-shell pueden no ser compatibles con los existentes
   - **Mitigación**: Migrar tokens gradualmente; mantener alias legacy durante la transición

3. **Riesgo**: El composer con mode selector puede Confundir a usuarios existentes
   - **Mitigación**: El mode selector es opcional; el composer funciona sin él

4. **Riesgo**: Los modales pueden no ser accesibles
   - **Mitigación**: Implementar focus trap, aria-modal, y Esc close desde el inicio

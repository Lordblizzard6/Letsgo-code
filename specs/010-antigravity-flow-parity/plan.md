# Implementation Plan: Paridad de Flujo Antigravity y Opencode

**Branch**: `010-antigravity-flow-parity` | **Date**: 2026-09-04 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/010-antigravity-flow-parity/spec.md`

---

## Summary

Alinear la experiencia de usuario tanto en la aplicación gráfica (Wails v3 GUI) como en la terminal (Bubbletea TUI) con el flujo de trabajo de Antigravity y Opencode. Las transformaciones clave incluyen:
1. **Compositor integrado**: Mover los selectores de modelos y modo de trabajo directamente a la barra inferior del contenedor del compositor de chat.
2. **Capacidad de Rollback**: Permitir rebobinar la conversación a cualquier turno anterior descartando en cascada los mensajes posteriores en SQLite y recargando el prompt en el editor.
3. **Panel Lateral de Git Diff**: Proporcionar un panel lateral derecho desplegable para auditar cambios y archivos modificados con resaltado de sintaxis.
4. **Settings en el pie de la barra lateral**: Reubicar el acceso a Configuración al footer de la barra lateral de proyectos.
5. **Paridad en Bubbletea TUI**: Soporte de `/rollback`, `/diff` y visualización compacta del modelo y modo en la línea de entrada.

---

## Technical Context

**Language/Version**: Go 1.26, TypeScript 5.x, React 18 / Svelte / HTML5  
**Primary Dependencies**:
- Desktop GUI: Wails v3 (`github.com/wailsapp/wails/v3 v3.0.0-beta.7`), `@wailsio/runtime`
- Terminal TUI: `charmbracelet/bubbletea`, `charmbracelet/lipgloss`, `charmbracelet/glamour`, `charmbracelet/bubbles`
- Frontend: Vite, CSS Modules / CSS Variables  
**Storage**: SQLite 3 vía `internal/db/database.go`  
**Testing**: `go test ./...`, Vitest (`npm test`) en `cmd/wails/frontend`  
**Target Platform**: Windows (WebView2), Linux, macOS  
**Project Type**: Dual interface (Desktop GUI + Terminal CLI/TUI)  
**Performance Goals**:
- Apertura del panel de Git Diff < 300ms
- Rollback de mensajes en SQLite < 100ms
- Cambio de modelo y renderizado reactivo instantáneo (< 50ms)  
**Constraints**:
- Sin librerías GPL3
- Respetar estricta separación de capas (Principio II: Contract-Only Core Access)
- Paletas de contraste WCAG AA en temas `goulm`, `dark`, `light`

---

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Principio I (Test-Last Verification)**: Las pruebas se ejecutan y validan al final de la construcción de cada historia de usuario. Cumple.
- **Principio II (Contract-Only Core Access)**: Las operaciones de rollback y git diff se exponen a través de `cmd/wails/services/sessions_service.go` y `cmd/wails/services/git_service.go`. Cumple.
- **Principio III (One Source of Truth)**: La verdad del historial persiste en SQLite (`internal/db/database.go`) y la memoria de ejecución en `internal/engine`. Cumple.
- **Principio IV (Secure Rendering)**: La visualización de diffs y markdown continúa utilizando saneamiento sin `innerHTML` arbitrario. Cumple.
- **Principio V (Simplicity & YAGNI)**: Se reutilizan las capacidades existentes de `GitService`, adaptando la presentación y los contratos mínimos necesarios. Cumple.

---

## Project Structure

### Documentation (this feature)

```text
specs/010-antigravity-flow-parity/
├── plan.md              # Este archivo
├── research.md          # Investigación de patrones y alternativas
├── data-model.md        # Modelos y extensiones SQLite
├── quickstart.md        # Guía de validación y escenarios ejecutables
└── contracts/
    └── antigravity_flow_contract.md  # Definición de RPCs y eventos
```

### Source Code

```text
internal/
├── db/
│   └── database.go           # RollbackSession: eliminación atómica >= messageID
├── engine/
│   └── engine.go             # Sincronización de memoria tras rollback
└── tui/
    └── ui.go                 # Comandos /rollback, /diff, statusbar en prompt

cmd/wails/
├── services/
│   ├── sessions_service.go   # Rollback(sessionID, messageID)
│   └── git_service.go        # DiffSummary() estructurado
└── frontend/
    ├── bindings/             # Bindings generados / sincronizados
    └── src/
        ├── App.tsx           # Integración compositor, git panel, rollback, footer settings
        ├── app.css           # Estilos para panel git, tarjetas de herramientas, composer toolbar
        └── SettingsView.tsx  # Vista de ajustes
```

---

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
| :--- | :--- | :--- |
| Ninguna | Arquitectura estándar | No se introducen dependencias externas ni patrones superfluos |

---

## Implementation Roadmap

### Phase 1: Backend Foundations (Rollback & Git Services)
1. Implementar `RollbackSession(sessionID string, targetMessageID int64)` en `internal/db/database.go`.
2. Exponer `SessionsService.Rollback(sessionID string, messageID int64)` en `cmd/wails/services/sessions_service.go`.
3. Exponer `GitService.DiffSummary()` o adaptar `Diff()` para entregar datos listos para el panel en `cmd/wails/services/git_service.go`.
4. Exportar métodos en `cmd/wails/frontend/bindings/` y `bindings.ts`.

### Phase 2: Frontend GUI - Compositor & Settings Footer
1. Refactorizar el contenedor de entrada `#chat-input-pane` en `App.tsx`:
   - Mover el selector de modelo (`#header-model-select`) a la barra de herramientas del compositor (`.composer-toolbar`).
   - Mover el selector de modo de trabajo (`.mode-switch`) a la barra de herramientas del compositor.
2. Mover el disparador de Ajustes (`Settings`) al pie de la barra lateral (`#sidebar-footer`).
3. Añadir botón de acción "Rebobinar" en los mensajes del chat que invoca `SessionsService.Rollback(...)` y carga el texto original en el `input`.

### Phase 3: Frontend GUI - Panel Lateral de Git Diff
1. Crear el componente / subvista de panel lateral de Git Diff en `App.tsx` y `app.css`.
2. Añadir botón de alternancia en la barra superior o atajo `Ctrl+D` para abrir/cerrar el panel lateral.
3. Conectar la llamada a `GitService.Diff()` para renderizar la lista de archivos y el visor de diff unificado coloreado.

### Phase 4: Bubbletea TUI Parity
1. Añadir comando `/rollback` en `internal/tui/ui.go` para descartar el último turno y cargar el prompt en el `textarea`.
2. Añadir comando `/diff` en `internal/tui/ui.go` para visualizar el diff de git en el viewport.
3. Reflejar modelo y modo activo de forma limpia en el prompt de entrada.

### Phase 5: Verification & Build
1. Compilar y probar frontend: `npm run build` y `npm test`.
2. Ejecutar pruebas unitarias de Go: `go test ./...`.
3. Compilar ejecutable `letsgo.exe` y realizar verificación de los escenarios de `quickstart.md`.

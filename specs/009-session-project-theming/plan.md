# Implementation Plan: Revisión de Proyectos, Sesiones y Temas

**Branch**: `009-session-project-theming` | **Date**: 2026-09-03 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/009-session-project-theming/spec.md`

## Summary

Implementar la visibilidad permanente de proyectos en la barra lateral aunque no contengan sesiones, dotar a cada ítem de proyecto de un botón contextual para iniciar una nueva conversación directamente en esa carpeta, asegurar el registro inmediato de nuevas sesiones en SQLite sin estados huérfanos, limpiar controles redundantes de nueva conversación en el espacio de chat, y consolidar la persistencia y reactividad total de temas (`goulm`, `dark`, `light`) en la ventana de configuración y entre reinicios de la aplicación.

## Technical Context

**Language/Version**: Go 1.26 (motor y backend RPC), TypeScript 5.2 + React 18 (frontend Wails)

**Primary Dependencies**: Wails v3 (`github.com/wailsapp/wails/v3`), `@wailsio/runtime`, Vite 8, React, Lucide/SVG Icons

**Storage**: SQLite (`internal/db`) para sesiones y proyectos; Viper + AES-256 (`internal/config`) para configuración global; `localStorage` para hidratación síncrona en frontend

**Testing**: `go test ./...` para paquetes Go (`internal/db`, `cmd/wails/services`); `npm run build` para Vite/TypeScript

**Target Platform**: Windows desktop (WebView2), macOS, Linux

**Project Type**: Desktop GUI Application (Wails v3)

**Performance Goals**: Creación de sesión < 100ms, alternancia de temas instantánea (< 16ms / 60fps), apertura de proyecto < 200ms

**Constraints**: Sin librerías GPL3, acceso a núcleo exclusivamente mediante contratos, política Test-Last de la constitución

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Principio I (Test-Last Verification)**: Las pruebas se ejecutan al finalizar la implementación de los componentes. Aprobado.
- **Principio II (Contract-Only Core Access)**: Todos los accesos a SQLite y configuración se realizan a través de `SessionsService`, `ThemeService` y `SettingsService`. Aprobado.
- **Principio III (One Source of Truth)**: La persistencia de proyectos y sesiones radica en SQLite y la configuración en Viper. Aprobado.
- **Principio IV (Secure Rendering)**: Sanitización activa de contenidos; contraste WCAG AA en temas `goulm`, `dark` y `light`. Aprobado.
- **Principio V (Simplicity & YAGNI)**: Cambios mínimos quirúrgicos sin capas de abstracción innecesarias. Aprobado.

## Project Structure

### Documentation (this feature)

```text
specs/009-session-project-theming/
├── plan.md              # Este archivo (plan de implementación)
├── research.md          # Investigación y decisiones arquitectónicas
├── data-model.md        # Definición de entidades (Project, Session, Theme)
├── quickstart.md        # Guía paso a paso de verificación
├── contracts/           # Contratos de servicio RPC y eventos Wails
│   └── session_contract.md
└── checklists/
    └── requirements.md  # Validación de requisitos de la especificación
```

### Source Code Layout

```text
cmd/wails/
├── frontend/src/
│   ├── App.tsx          # Árbol de proyectos, gestión de pestañas, zoom, eventos
│   ├── SettingsView.tsx # Modal de configuración adaptado a variables de tema
│   ├── bindings.ts      # Enlaces generados de RPC
│   └── app.css          # Paleta de variables CSS para temas (goulm, dark, light)
├── services/
│   ├── sessions_service.go # RPCs de proyectos y sesiones (Create, List, SelectFolder)
│   ├── theme_service.go    # RPC de persistencia de tema (Set, Get)
│   └── settings_service.go # RPC de configuración
internal/
├── config/config.go     # Almacenamiento Viper de tema y preferencias
└── db/database.go       # Persistencia SQLite de sesiones y mensajes
```

## Implementation Phases

### Phase 1: Árbol de Proyectos y Creación Contextual de Sesiones
1. Modificar `App.tsx` para combinar la lista de `recentProjects` con las sesiones cargadas desde SQLite en `groupedSessions`.
2. Mostrar cada carpeta de proyecto con su cabecera, nombre y contador (incluso si el contador es `0`).
3. Añadir botón `+` ("Nueva conversación") en la cabecera de cada proyecto que ejecute la creación y apertura de una conversación vinculada a la ruta del proyecto.
4. Remover botones redundantes de nueva conversación en el panel de chat y compositor.

### Phase 2: Registro Inmediato y Garantizado de Sesiones
1. En `App.tsx`, asegurar que al pulsar el botón `+` o abrir una pestaña se invoque sincrónicamente `SessionsService.Create(...)`.
2. Guardar la sesión directamente en SQLite antes de aceptar mensajes para garantizar que no existan sesiones sin registrar.
3. Asegurar que el evento `session:list` actualice inmediatamente la barra lateral.

### Phase 3: Persistencia y Adaptación Dinámica de Temas
1. Verificar que `SettingsView.tsx` use exclusivamente variables CSS semánticas (`var(--bg-panel)`, `var(--bg-elevated)`, `var(--border)`, `var(--text)`, etc.) en todas sus secciones.
2. Garantizar que el cambio de tema guarde inmediatamente en `localStorage` y llame a `ThemeService.Set`.
3. Validar que al iniciar la aplicación, `App.tsx` lea `localStorage.getItem("app_theme")` y configure `document.documentElement.dataset.theme` antes de renderizar la UI.

## Complexity Tracking

*No se detectan violaciones a los principios constitucionales.*

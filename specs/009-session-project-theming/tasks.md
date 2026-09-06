# Tasks: Revisión de Proyectos, Sesiones y Temas

**Input**: Documentos de diseño en `/specs/009-session-project-theming/`
**Prerequisites**: [plan.md](plan.md), [spec.md](spec.md), [research.md](research.md), [data-model.md](data-model.md), [contracts/session_contract.md](contracts/session_contract.md), [quickstart.md](quickstart.md)

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Ejecutable en paralelo (archivos independientes)
- **[Story]**: Historia de usuario correspondiente (`[US1]`, `[US2]`, `[US3]`, `[US4]`)
- Rutas exactas en cada tarea

---

## Phase 1: Setup (Alineación de Infraestructura)

**Purpose**: Verificación de contratos y esquemas base para proyectos y sesiones

- [x] T001 Auditar esquema y métodos de proyectos en `internal/db/database.go` para admitir proyectos sin sesiones previas
- [x] T002 [P] Verificar enlaces RPC generados en `cmd/wails/services/sessions_service.go` y `cmd/wails/frontend/src/bindings.ts`

---

## Phase 2: Foundational (Prerrequisitos Bloqueantes)

**Purpose**: Infraestructura compartida para la visualización unificada de proyectos y creación de sesiones

- [x] T003 Exponer o estructurar la lista unificada de proyectos en `cmd/wails/services/sessions_service.go`
- [x] T004 Estructurar el estado unificado de proyectos y sesiones en `cmd/wails/frontend/src/App.tsx` para preservar carpetas vacías

**Checkpoint**: Base lista para la integración visual y funcional de las historias de usuario

---

## Phase 3: User Story 1 - Proyectos Visibles sin Sesiones y Creación Contextual (Priority: P1) 🎯 MVP

**Goal**: Mostrar todos los proyectos registrados en la barra lateral aunque tengan 0 conversaciones y permitir crear nuevas sesiones con un botón `+` en cada ítem de proyecto.

**Independent Test**: Añadir un nuevo proyecto desde el explorador; verificar que aparece en el árbol lateral con 0 conversaciones y hacer clic en su botón `+` para abrir una sesión vinculada a esa carpeta.

- [x] T005 [US1] Modificar `groupedSessions` en `cmd/wails/frontend/src/App.tsx` para incluir todos los proyectos de `recentProjects` aunque tengan 0 sesiones
- [x] T006 [US1] Añadir botón interactivo `+` ("Nueva conversación") en la cabecera de cada carpeta en `cmd/wails/frontend/src/App.tsx`
- [x] T007 [US1] Implementar `createSessionForProject(projectPath)` en `cmd/wails/frontend/src/App.tsx` para crear y abrir la sesión en la carpeta destino
- [x] T008 [P] [US1] Añadir estilos para el botón de nueva conversación en cabeceras de proyecto en `cmd/wails/frontend/src/app.css`

**Checkpoint**: Los proyectos vacíos son visibles y operativos; cada proyecto permite crear sesiones directamente.

---

## Phase 4: User Story 2 - Registro Confiable e Inmediato de Nuevas Sesiones (Priority: P1)

**Goal**: Asegurar que toda nueva conversación quede registrada de inmediato en SQLite sin pestañas huérfanas ni pérdida de sesiones al reiniciar.

**Independent Test**: Abrir una nueva sesión, enviar un mensaje, reiniciar la aplicación y comprobar que la conversación figura en la base de datos y en la barra lateral.

- [x] T009 [US2] Actualizar `addTab()` en `cmd/wails/frontend/src/App.tsx` para llamar de inmediato a `SessionsService.Create(...)` garantizando `sessionId` válido
- [x] T010 [US2] Asegurar inserción síncrona con metadatos completos en `internal/db/database.go` al invocar `CreateSession`
- [x] T011 [US2] Emitir evento `session:list` al crear cualquier sesión en `cmd/wails/services/sessions_service.go`
- [x] T012 [US2] Sincronizar el manejador de `Events.On("session:list")` en `cmd/wails/frontend/src/App.tsx` para reflejar sesiones recién creadas al instante

**Checkpoint**: Ninguna sesión creada se pierde o queda sin registrar.

---

## Phase 5: User Story 3 - Eliminación de Controles Redundantes en el Espacio de Chat (Priority: P2)

**Goal**: Limpiar el espacio de conversación eliminando botones redundantes de "Nueva conversación" que dispersen el flujo de trabajo.

**Independent Test**: Inspeccionar visualmente la cabecera del chat, cuerpo de conversación y compositor para verificar que no existen botones redundantes de nueva conversación.

- [x] T013 [US3] Eliminar botones redundantes de "nueva conversación" en el espacio de chat en `cmd/wails/frontend/src/App.tsx`
- [x] T014 [P] [US3] Ajustar espaciado y estilos en `cmd/wails/frontend/src/app.css` tras retirar los controles redundantes

**Checkpoint**: El espacio de chat queda despejado y enfocado en la interacción del hilo activo.

---

## Phase 6: User Story 4 - Persistencia Absoluta de Temas y Adaptación de la Ventana de Configuración (Priority: P2)

**Goal**: Garantizar persistencia infalible del tema seleccionado y adaptación cromática total del modal de Ajustes.

**Independent Test**: Cambiar entre temas en Ajustes, verificar adaptación instantánea de colores del modal, reiniciar la aplicación y verificar que se mantiene el tema seleccionado.

- [x] T015 [US4] Auditar y asegurar uso estricto de variables CSS de diseño en `cmd/wails/frontend/src/SettingsView.tsx` para todas las secciones
- [x] T016 [US4] Consolidar la hidratación síncrona y sincronización de tema en `cmd/wails/frontend/src/App.tsx` desde `localStorage` y `GetConfig()`
- [x] T017 [US4] Verificar la persistencia en disco de `ThemeVariant` en `internal/config/config.go` y `cmd/wails/services/theme_service.go`
- [x] T018 [P] [US4] Revisar contraste y estilos de tokens para los tres temas (`goulm`, `dark`, `light`) en `cmd/wails/frontend/src/app.css`

**Checkpoint**: Los tres temas persisten tras el reinicio y el modal de Ajustes adapta su apariencia sin fallos.

---

## Phase 7: Polish & Validación Cruzada

**Purpose**: Verificación final, compilación cruzada y pruebas de extremo a extremo

- [x] T019 [P] Compilar frontend con `npm run build` en `cmd/wails/frontend`
- [x] T020 [P] Ejecutar suite de pruebas de Go con `go test ./...`
- [x] T021 Ejecutar escenarios de validación descritos en `specs/009-session-project-theming/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies
- **Setup (Phase 1)**: Sin dependencias.
- **Foundational (Phase 2)**: Depende de Phase 1; bloquea el desarrollo de las historias de usuario.
- **User Story 1 & User Story 2 (Phases 3 & 4)**: Prioridad P1; pueden implementarse en secuencia o en paralelo sobre la base completada.
- **User Story 3 & User Story 4 (Phases 5 & 6)**: Prioridad P2; dependen de la base completada.
- **Polish (Phase 7)**: Depende de todas las historias completadas.

### Parallel Opportunities
- T001 y T002 en Setup pueden ejecutarse en paralelo.
- T008 y T014 en estilos CSS pueden ajustarse en paralelo con la lógica de componentes.
- T019 y T020 en Polish pueden ejecutarse simultáneamente.

---

## Implementation Strategy

### MVP First (User Story 1 & 2)
1. Completar Setup y Foundational (T001 - T004).
2. Implementar User Story 1 (T005 - T008): Proyectos visibles sin sesiones y botón `+` contextual.
3. Implementar User Story 2 (T009 - T012): Registro inmediato de sesiones en SQLite.
4. Validar que los proyectos no se ocultan y las sesiones no se pierden.

### Incremental Delivery (User Story 3 & 4)
5. Aplicar limpieza del espacio de chat (T013 - T014).
6. Consolidar persistencia y estilos del modal de configuración (T015 - T018).
7. Ejecutar compilación y validación final (T019 - T021).

# Tasks: UI/UX Refinement — Consistency, Legibility & Menu Architecture

**Input**: Design documents from `/specs/004-ui-redesign/`

**Prerequisites**: plan.md (required), spec.md (user stories), research.md, data-model.md, `contracts/ui-contract.md`

**Tests**: SI — el `contracts/ui-contract.md` seccion 11 exige tests headless por superficie y una suite de contraste de paleta; se incluyen como tareas de test antes de implementar cada story (TDD).

**Org**: Tareas agrupadas por user story para implementacion y validacion independiente.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Paralela (archivos distintos, sin dependencias pendientes)
- **[Story]**: US1..US4 (spec.md). Fases Setup/Foundational/Polish sin label.
- Incluir rutas exactas. Todo bajo `internal/gui/` salvo indicacion.

## Build / Test del repo (Windows)

- CGO obligatorio: `$env:PATH = "D:\msys64\ucrt64\bin;" + $env:PATH; $env:CGO_ENABLED = "1"`
- Compilar: `build.bat -n` → `letsgo.exe` · Test headless: `go test ./internal/gui/... ./internal/config/...`

---

## Phase 1: Setup (Infraestructura compartida)

**Proposito**: verificar linea base y preparar los helpers visuales compartidos que usan todos los stories.

- [x] T001 Verificar la linea base: ejecutar `go vet ./...` y `go build ./...`; guardar el resultado de la suite headless actual (`go test ./internal/gui/... ./internal/config/...`) como referencia verde para las puertas de regresion (Gate B).
- [x] T002 Implementar `tooltip(label, shortcut) fyne.CanvasObject` en `internal/gui/tooltips.go` con el formato "Label (Alt+N)" y su test en `internal/gui/tooltips_test.go`.
- [x] T003 [P] Implementar `iconFor(id string) fyne.Resource` en `internal/gui/icons.go` con el mapeo de los 10 destinos (chat, git, tasks, mcp, plugins, settings, usage, help, theme, account) y fallback a `theme.IconName` correspondiente cuando el SVG no existe; test en `icons_test.go`.
- [x] T004 [P] Implementar `emptyState(msg, ctaText, onCTA)` y `loadingState(content, refreshing)` en `internal/gui/empty.go` conforme a `ui-contract.md` seccion 10, con test base de cada helper en `empty_test.go`.

**Checkpoint**: helpers (tooltip, iconos, empty/loading) disponibles y testeados; todos los stories pueden apoyarse en ellos.

---

## Phase 2 — Foundational (Pre-requisitos bloqueantes)

**Proposito**: la re-estructuracion del rail (zonas, un destino por slot, remap Alt) es requisito de los demas (Settings/Cuenta/Uso y la navegacion coherente).

- [x] T005 Refactorizar el catalogo `railSlots` en `internal/gui/rail.go`: anadir campo `zone` (`work|tools|system`) y **eliminar** el slot `sessions` (duplicaba pane 0); orden definitivo segun `data-model.md`: chat(0), git(3), tasks(6), mcp(4), plugins(5), settings(1), usage(2), help(-1), theme(-1), account(-1).
- [x] T006 Re-mapear los atajos `Alt+1..8` en `internal/gui/keyboard.go` y actualizar el catalogo de `internal/gui/keymap.go` + labels de `internal/gui/palette.go`; siempre leyendo el catareal (nunca hardcode).
- [x] T007 [P] Actualizar `internal/gui/rail_test.go` a la lista nueva (zonas, sin duplicados, `slotForPane(0)` unico) y `internal/gui/shortcuts_test.go` a la nueva tabla Alt.

**Checkpoint**: Foundation del rail completa; los stories pueden empezar sobre una navegacion estable.

---

## Phase 3 — US Story 1 — Rail coherente: zonas y un destino por slot (P1) 🎯 MVP

**Goal**: el usuario ve una barra en 3 zonas, sin duplicados, con tooltips y foco en modo colapsado. Es el esqueleto del shell actualizado y, por si solo, un MVP valido.

**Independent Test**: abrir la app, recorrer con raton y `Alt+1..8`, ver marca activa unica, tooltip ("Git (Alt+2)") en colapsado, foco visible con teclado y vuelta al Chat con `Alt+Left`/`Esc`. (Journey 1 de quickstart.)

### Tests (contrato §11)

- [x] T008 [P] [US1] Anadir `TestRailZonesAndSingleDestination` en `internal/gui/rail_test.go`: 3 zonas, un destino por pane, `slotForRail(0)` unico, atajos Alt correctos.
- [x] T009 [P] [US1] Anadir `TestRailTooltipCollapsed` en `internal/gui/rail_test.go`: tooltip con nombre + shortcut en colapsado y focus ring visible.

### Implementacion

- [x] T010 [P] [US1] Renderizar las 3 zonas separadas (`widget.separator`) con etiqueta "HERRAMIENTAS" en `internal/gui/rail.go` (rebuild con 3 VBox y hairline).
- [x] T011 [US1] Aplicar los tooltips de T002 a cada slot en estado colapsado de `rail.go`, incluido el toggle y el avatar (con icon del T003).
- [x] T012 [US1] Accionar `slotForPane`/`slotByID` en `rail.go` para la lista nueva sin hardcode de indices; la marca activa (accent bar) solo en el slot cuyo pane == panel actual.
- [x] T013 [US1] Ajustar `internal/gui/navigation.go` (`showPane`, `back`) a la nueva composicion sin romper `navHistory`; anadir smoke `internal/gui/shell_smoke_test.go` que monta Controller, navega por Alt y vuelve con Esc.

**Checkpoint**: US1 completo y testeable aislado (MVP). Los demas stories se pueden sumar sin tocar la navegacion base.

---

## Phase 4 — US Story 2 (Profundidad correcta: Settings, Cuenta y Uso) (Prioridad P2)

**Goal**: Configuracion como overlay con secciones, cuenta en flyout del avatar y Uso como dashboard KPI.

**Independent Test**: abrir `Ctrl+,` (overlay no modal, 4 secciones, chat detras, foco al composer al cerrar); avatar -> flyout (estado, provider/model, uso, acciones); `Alt+7` Uso con tarjetas KPI/tabla/barra. (Journey 2.)

### Tests (contrato §11)

- [x] T014 [P] [US2] Anadir `TestSettingsOverlayRender` en `internal/gui/settings_test.go`: overlay abre 4 secciones; Esc/Cancel cierran con focus; Save persiste.
- [x] T015 [P] [US2] Anadir `TestAccountFlyout` en `internal/gui/account_test.go`: click avatar abre popup; Esc cierra y enfoca; acciones rutadas.
- [x] T016 [P] [US2] Anadir `TestUsageKPI` en `internal/gui/usage_test.go`: surfaces KPI construidas sin depender de `usageHistogram.String`.

### Implementacion

- [x] T017 [US2] Reagrupar `internal/gui/settings.go` en `widget.TabContainer` [Cuenta|Apariencia|Preferencias|Uso] conservando campos y persistencia; `Save` ejecuta `config.SaveConfig()`.
- [x] T018 [US2] Implementar el contenedor overlay no modal en `internal/gui/app.go`: `showSettings()` abre un overlay ~660x520 con TabContainer; fallback a pane 1 si el entorno de tests no monta overlay (research D2).
- [x] T019 [P] [US2] Implementar `accountFlyout` en `internal/gui/account.go` con `widget.NewPopUp` anclado al avatar (datos: provider/model, resumen `usageSource`) y acciones `onKeys/onUsage/onSignOut` cableadas desde `app.go`.
- [x] T020 [US] Cambiar `showAccount()` en `internal/gui/app.go` para abrir el flyout (no `showSettings`); conectar el click del avatar en `rail.go`.
- [x] T021 [US] Rediseñar `internal/gui/usage.go` a dashboard KPI: `GridWithColumns(3)` (coste hoy, requests, tokens) + `widget.NewTable` per provider + barra de presupuesto %; mantener `refresh` y `startOfToday`; actualizar `usage_test`.

**Checkpoint**: US2 independiente (overlay settings, flyout cuenta, uso KPI).

---

## Phase 5 — US Story 3 (Legibilidad y elegancia) (P2)

**Goal**: media de mensaje ~720px, marco común de tarjeta, rampa tipografica, composer con ring 2px y Enviar↔Detener, microcopy ES.

**Independent Test**: columna prosa <=720px; tarjetas con marco comun; composer ring al focus; swap Enviar/Detener en streaming; placeholders ES. (Journey 3.)

### Tests (contrato §11)

- [x] T022 [P] [US3] Anadir `TestMessageMeasure` en `internal/gui/message_test.go`: la prosa no supera ~720px en headless.
- [x] T023 [P] [US3] Anadir `TestComposerSwapSendStop` en `internal/gui/composer_test.go`: el mismo boton alterna label/icon en streaming.

### Implementacion

- [x] T024 [US3] Limitar la columna de pronto (max ~720px centrado, codigo ancho completo con wrap opcional) en `internal/gui/message.go`, `markdown.go` y `chatview.go`.
- [x] T025 [P] [US3] Aplicar el mar/mar comunicato de tarjetas (border 1px, radius 6px, padding 16px) en `message.go`, `approvalcard.go` y `toolpanel.go` via tokens del tema.
- [x] T026 [US3] Anadir la rampa tipografica (12/14/16/20/24 sans + mono 12-13) como constantes en `internal/gui/theme.go` y aplicarla en cabeceras/titulos; quitar tamanos literales.
- [x] T027 [US3] Composer: ring focus de **2px** accent (`composer.go` via `theme.FocusColor()`), boton principal alterna Enviar <-> Detener en el mismo lugar (depende del estado `busy`), placeholder ES "Escribe a LetsGO...".

**Checkpoint**: US3 completo; la conversion se lee comoda y de forma coherente.

---

## Phase 6 — US Story 4 (Accesibilidad y estados) (P3)

**Goal:** [PRACTICE] contraste AA en CI, focus visible, iconos SVG semanticos y estados vacios/loading en 6 superficies.

**Independent Test**: tema claro cumple >=4.5/>=3 (palette test); Tab recorre focus visible; listas vacias muestran CTA; iconos semanticos con tooltip en colapsado. (Journey 4.)

### Tests (contrato §11)

- [x] T028 [P] [US4] Anadir `TestPaletteContrast` en `internal/gui/theme_test.go` — itera pares base (dark+light) y falla si texto <4.5:1 o no-texto <3:1 (palette smoke).
- [x] T029 [P] [US4] Anadir `TestEmptyStatesAllSurfaces` en `internal/gui/empty_test.go`: 6 superficies con empty+CTA; refresh conserva el contenido previo.

### Implementacion

- [x] T030 [US4] Fix contraste: `lightPlaceholder` -> ~`#5C6B7A` y ajustar `darkBorder` (-> ~`#2E3A48` si hace falta) en `internal/gui/theme.go` hasta pasar `TestPaletteContrast`.
- [x] T031 [US4] Composer ring y avatar: reemplazar constantes oscuras por tokens del tema (`theme.FocusColor()`, `theme.ForegroundColor()`, surface) en `composer.go` y `rail.go` para tema claro.
- [x] T032 [P] [US4] Aplicar los helpers de T004 a las 6 superficies (`sessions.go`, `mcpview.go`, `pluginsview.go`, `agenttasks.go`, `gitview.go`, `usage.go`) con textos/CTAs de `ui-contract` seccion 10; el `loadingState` envuelve los refrescos async (conservar contenido previo + "Refrescando...").
- [x] T033 [P] [US4] Sustituir iconos genericos por los SVG de `icons.go` en superficies (rail, approval, chips) sin romper el fallback + tooltips.

**Checkpoint**: todos los stories implementados y verificables.

---

## Phase 7 — Polish & Cross-Cutting

**Propor**: regresion completa, documentacion y checklist.

- [x] T034 [P] Actualizar `quickstart.md` (journeys) y `contracts/ui-contract.md` si algo cambia de firma tras implementacion.
- [x] T035 [P] Limpiar: recorrer `internal/gui/*.go` para borrar constantes muertas y mantener `usageHistogram.String` solo como fallback/debug.
- [x] T036 [P] Smoke de regresion: ampliar `shell_smoke_test.go` y ejecutar toda la suite (gui y config) + `go build ./...`: 0 regresiones (Gate B).
- [x] T037 Ejecutar las validaciones de `quickstart.md` (J1-J4) y marcar los checkboxes SC-001..007.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (P1)**: sin deps; T001-T004 paralelos.
- **Foundational (P2)**: depende de Setup; T005-T007 bloquean todos los stories.
- **User Stories (P3-P6)**: dependen de Foundation; pueden ir en paralelo o secuencial por prioridad (P1 → P2).
- **Ray: Polish (P7)**: depende de los stories deseados completos.

### User Story Dependencies

- **US1 (P1)**: tras Foundation; sin dep de otros (re-hace la navegacion).
- **US2 (P2)**: tras Foundation; usa tooltips/iconos/empty del Setup y el rail de US1, pero sus ficheros (settings/account/usage) son propios y puede ir en paralelo.
- **US3 (P2)**: tras Foundation; usa tokens de tipografia y marco común; no depende de las superficies US2.
- **US4 (P3)**: tras Foundation; usa `emptyState`/`icons`/`tooltips` del Setup y el rail de US1 (foco), puede ir en paralelo.

### Parallel Opportunities

- [P] Setup: T002, T003, T004.
- [P] dentro de cada story: los tests primero y en paralelo (T008/T009, T014/T015/T016, T022/T023, T028/T029).
- Los stories en paralelo por equipos despues de Foundation (US1 rail, US2 settings/account/usage, US3 message, US4 contrast/states).

---

## Ejemplo paralelo: US1

```bash
# Tests first (TDD) antes de implementar:
# T008 rail_test: TestRailZonesAndSingleDestination
# T009 rail_test: TestRailTooltipCollapsed
# Despues, en paralelo:
# T010 rail.go: render zonas + hairline
# T011 rail.go: tooltips collapsed
# T012 rail.go: slotForPane unico
# T013 navigation.go + shell_smoke_test.go
```

## Implementation Strategy

### MVP first (solo US1)

1. P1 Setup (T001-T004)
2. P2 Foundation (T005-T007)
3. P3 US1 (T008-T013) y **STOP VALIDATE**: rail en 3 zonas, sin duplicados, Alt+1..8, tooltips/colapso, foco — MVP navegable.
4. Demostrar/desplegar el rail coherente.

### Entrega incremental

1. Setup + Foundation → base lista.
2. +US1 → MVP (barra en 3 zonas + teclado).
3. +US2 (settings overlay, fly account, dashboard KPI) → prueba independiente.
4. +US3 (legibilidad) → +US4 (accesibilidad/estados) → polish P7.
5. Cada story aporta valor sin romper los previos; Gates B y C verifican.

### Parallel team

- A: US1 rail · B: US2 (después de Foundation) · C: US3 (mensajes/composer, no tocar rail) · D: US4 (empty/contraste) una vez presentes los helpers de Setup.
- Nota: `rail.go` lo edita solo A durante US1; B/C/D respetan rail.go.

---

## Notas

- **Story label**: cada tarea de phases lleva [US#] exacto al spec.
- **TDD**: escribir los tests, verlos fallar, implementar, verlos pasar.
- **Regresion**: contratos 002/003 se mantienen verdes sin tocar comportamientos preservados.
- **Colores**: ningun widget hardcodea `dark*` en clases compartidas (Gate C).
- **No tocar** los indices `viewStack` 0..6 ni `navHistory`, salvo el catalogo de rail (T005).
- Commit despues de cada tarea o grupo logico; checkpoint al final de cada story.
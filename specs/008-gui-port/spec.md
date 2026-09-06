# Spec 008: GUI Port — Copy gui-shell example 1:1

## Overview

Port the entire GUI from `ejemplo de gui/gui-shell/` (React 19 + Vite 7) into the LetsGO frontend (`cmd/wails/frontend/`), replacing the current Svelte-based frontend. The goal is a **faithful 1:1 visual copy** of the example's design, layout, themes, and component structure — adapted to work with the Wails backend instead of mock data.

## User Story

**As a** LetsGO developer,
**I want** the GUI to match the `ejemplo de gui/gui-shell/` design exactly,
**So that** the application has a polished, professional appearance identical to the reference.

## Scope

### In Scope
1. **CSS port**: Copy `App.css` (3441 lines) and `style.css` (44 lines) verbatim — themes, layout, tokens, all component styles
2. **React components**: Port `App.tsx`, `SettingsView.tsx`, `SpendView.tsx`, `Markdown.tsx`, `Combobox.tsx` — adapting imports and removing mock data
3. **Types**: Port `types.ts` with any additions needed for Wails backend integration
4. **i18n**: Port `i18n.ts` (ES + EN dictionaries)
5. **Assets**: Port fonts (Nunito woff2), images (goulm.png), and highlight.js CSS
6. **Build config**: Update `vite.config.ts` for React plugin, update `package.json` with React deps, update `tsconfig.json`/`tsconfig.node.json`, update `index.html`
7. **Entry point**: Replace `main.ts` (Svelte) with `main.tsx` (React)
8. **Backend integration**: Replace mock data with Wails bindings (`@wailsio/runtime`) — events, commands, store
9. **Remove Svelte**: Delete all `.svelte` files, Svelte-specific tests, Svelte config

### Out of Scope
- New features not in the example (keep the example's feature set)
- New themes beyond dark/light
- Mobile responsive design (desktop only, matching example)
- New accessibility features beyond what the example provides

## Technical Context

### Current State
- Frontend: Svelte 5 + TypeScript + Vite under `cmd/wails/frontend/`
- Backend: Go with Wails v3, events/commands contract in `specs/005-bubbletea-tui-wails/contracts/`
- Build: `npm run build` + `go build -o lets-go.exe .`
- Embed: `//go:embed all:wails/frontend/dist` in `cmd/gui.go`

### Target State
- Frontend: React 19 + TypeScript + Vite 7 (matching example exactly)
- Backend: Same Go/Wails — frontend consumes events/commands via `@wailsio/runtime`
- Build: Same workflow — `npm run build` + `go build -o lets-go.exe .`
- Embed: Same path — `cmd/wails/frontend/dist`

### Key Decisions
1. **React over Svelte**: The example is React; porting it faithfully requires React, not a Svelte re-implementation
2. **Monolithic App.tsx**: Keep the example's single-file architecture (2042 lines) — split later if needed
3. **Mock data removal**: Strip all MOCK_* constants and `demoRespond()` — replace with Wails event listeners
4. **Backend binding**: Use `@wailsio/runtime` for events (`chat:delta`, `session:*`, `tool:*`, etc.) and commands

## Functional Requirements

| ID | Requirement |
|----|-------------|
| FR-01 | Three-column layout: sidebar (260px) + chat-pane (flex) + right-pane (320px) |
| FR-02 | Dark theme (default) and light theme with CSS custom properties |
| FR-03 | Header with brand, project selector, account/model selectors, usage chip, settings button, palette button |
| FR-04 | Sidebar with recent projects and saved sessions |
| FR-05 | Tabbed chat with web/test/skills special panes |
| FR-06 | Composer with slash commands, attachments, mode selector |
| FR-07 | Activity panel with tool cards (running/done/error states) |
| FR-08 | Settings modal with 6 tabs (Accounts, General, Budget, Safety, Search, Paths) |
| FR-09 | Spend modal with heatmap, summary, and entry list |
| FR-10 | Command palette (Ctrl+K) for global actions |
| FR-11 | Markdown rendering with syntax highlighting (react-markdown + rehype-highlight) |
| FR-12 | Bilingual i18n (ES/EN) with ~150 translation keys |
| FR-13 | File explorer panel |
| FR-14 | Nunito font family (regular/600/700) |
| FR-15 | Streaming cursor animation for assistant messages |
| FR-16 | Copy/retry actions on messages |
| FR-17 | Zoom slider (70-150%) |

## Non-Functional Requirements

| ID | Requirement |
|----|-------------|
| NFR-01 | Build must succeed (`npm run build` + `go build`) |
| NFR-02 | Frontend must embed in Go binary via `//go:embed` |
| NFR-03 | Must consume Wails events, not mock data |
| NFR-04 | Dark/light themes must maintain WCAG contrast AA |
| NFR-05 | Transitions ≤150ms, prefers-reduced-motion respected |
| NFR-06 | No emoji in chrome UI elements |
| NFR-07 | C-zero changes to Go bindings/contract |

## Constitution Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Test-Last | ✅ | Tests written after implementation |
| II. Contract-Only Core Access | ✅ | Frontend consumes Wails events/commands only |
| III. One Source of Truth | ✅ | Backend state from engine, frontend is presentation |
| IV. Secure Rendering | ⚠️ | Must verify DOMPurify integration with react-markdown |
| V. Simplicity & YAGNI | ✅ | Copying existing example, not adding new features |

## Complexity Tracking

| Complexity | Justification |
|-----------|---------------|
| React migration from Svelte | Example is React; faithful port requires same framework |
| Mock data removal | Example uses hardcoded data; real app needs Wails events |
| Backend binding | Must replace `demoRespond()` with actual Wails event listeners |

## Deliverables

1. `specs/008-gui-port/spec.md` — This file
2. `specs/008-gui-port/plan.md` — Implementation plan
3. `specs/008-gui-port/tasks.md` — Task breakdown
4. `specs/008-gui-port/research.md` — Technical research
5. `specs/008-gui-port/data-model.md` — Data model
6. `specs/008-gui-port/contracts/gui-contract.md` — GUI contract
7. `specs/008-gui-port/quickstart.md` — Validation guide

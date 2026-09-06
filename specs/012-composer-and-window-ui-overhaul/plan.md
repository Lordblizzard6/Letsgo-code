# Implementation Plan: Composer Layout and Window Sizing Antigravity Parity

**Branch**: `012-composer-and-window-ui-overhaul` | **Date**: 2026-09-04 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/012-composer-and-window-ui-overhaul/spec.md`

## Summary

Modernize the GUI window and composer interface to achieve Antigravity parity:
1. Update initial window size to 1280x850 with minimum size 960x600 in `cmd/wails/gui/gui.go`.
2. Move "Nueva conversación" action to a prominent button at the top of the sidebar navbar.
3. Position work modes (`Plan`, `Código`, `Arch`, `Research`) in the top corner of the composer box, cyclable via `Tab` (and `Shift+Tab` backwards) when autocomplete is idle.
4. Redesign composer bottom bar with `[📎 Attach]` and `[🤖 Model Selector ▾]` on the left, and `[↑ Enviar]` on the right.
5. Enhance `/` command viewer popup with icons, descriptions, and keyboard shortcuts matching Antigravity style.

## Technical Context

**Language/Version**: Go 1.26, TypeScript / React 18
**Primary Dependencies**: Wails v3, Lucide React
**Storage**: SQLite (`internal/db`)
**Testing**: `npm test` (Vitest), `go test ./...`
**Target Platform**: Windows (desktop), cross-platform compatible
**Project Type**: Desktop GUI presentation
**Constraints**: No GPL3 libraries, test-last principle, surgical diffs

## Constitution Check

- **I. Test-Last Verification**: Vitest component/logic tests and Go service tests run after implementation.
- **II. Contract-Only Core Access**: No core internal bypass; only presentation state in `App.tsx` and window options in `gui.go`.
- **III. One Source of Truth**: Backend services (`SettingsService`, `ChatService`) remain authoritative for model selection and modes.
- **IV. Secure Rendering**: All command descriptions and model names rendered with safe React DOM bindings.
- **V. Simplicity & YAGNI**: Clean CSS and React state management without introducing external heavyweight UI libraries.

## Project Structure

### Documentation (this feature)

```text
specs/012-composer-and-window-ui-overhaul/
├── spec.md
├── checklists/
│   └── requirements.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
└── contracts/
    └── ui-composer-contract.md
```

### Source Code Changes

```text
cmd/wails/
├── gui/
│   └── gui.go                  # Window size 1280x850, MinWidth 960, MinHeight 600
└── frontend/src/
    ├── App.tsx                 # Composer layout, Tab mode cycling, sidebar new chat button, command viewer
    ├── app.css                 # Styling for composer header, mode tabs, bottom toolbar, command viewer
    └── i18n.ts                 # Localized strings for modes, buttons, and hints
```

## Planned Phases

- **Phase 0**: Window Dimensions & Sidebar Navigation (`gui.go`, `App.tsx`)
- **Phase 1**: Composer Header & Work Mode Tab Cycling (`App.tsx`, `app.css`)
- **Phase 2**: Bottom Controls Layout Parity (`[Attach] [Model] ... [Send]`)
- **Phase 3**: Antigravity-style Command Viewer Popup
- **Phase 4**: Verification & Test Gate (`npm test`, `npm run build`, `go build`)

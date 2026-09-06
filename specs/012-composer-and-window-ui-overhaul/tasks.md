# Tasks: Composer Layout and Window Sizing Antigravity Parity

**Feature Branch**: `012-composer-and-window-ui-overhaul` | **Date**: 2026-09-04 | **Spec**: [spec.md](spec.md) | **Plan**: [plan.md](plan.md)

## Phase 1: Setup & Foundational
- [x] T001 Configure window dimensions (1280x850, min 960x600) in `cmd/wails/gui/gui.go`
- [x] T002 Implement real AI provider SVG logo icons (Anthropic, OpenAI, Gemini, Groq, Ollama, DeepSeek, OpenRouter) in `cmd/wails/frontend/src/ProviderIcons.tsx`

## Phase 2: User Story 1 - Sidebar New Chat Button & Typography Scale (P1)
- [x] T003 [US1] Add prominent `+ Nueva conversación` action button in sidebar above conversations in `cmd/wails/frontend/src/App.tsx` and style in `cmd/wails/frontend/src/app.css`

## Phase 3: User Story 2 - Work Modes in Top Corner of Composer with Tab Cycling (P1)
- [x] T004 [US2] Implement top-corner segmented modes (`Plan`, `Código`, `Arch`, `Research`) inside composer box in `cmd/wails/frontend/src/App.tsx` and style in `cmd/wails/frontend/src/app.css`
- [x] T005 [US2] Implement `Tab` and `Shift+Tab` keyboard cycling between modes in `onKeyDown`, preserving suggestion completion when autocomplete is open in `cmd/wails/frontend/src/App.tsx`

## Phase 4: User Story 3 - Composer Bottom Bar Layout Parity (P1)
- [x] T006 [US3] Redesign bottom toolbar: `+` icon for attach, Model Selector with real AI logo on left, and icon-only send button on right in `cmd/wails/frontend/src/App.tsx` and `cmd/wails/frontend/src/app.css`

## Phase 5: User Story 4 - Antigravity Style Command Viewer Popup (P2)
- [x] T007 [US4] Elevate `/` command popup with icons, descriptions, and keyboard cues in `cmd/wails/frontend/src/App.tsx` and `cmd/wails/frontend/src/app.css`

## Phase 6: Polish & Verification (Test-Last per Constitution Principle I)
- [x] T008 Run frontend tests and bundle build (`npm test` and `npm run build` in `cmd/wails/frontend`)
- [x] T009 Run Go test suite (`go test ./cmd/...`)
- [x] T010 Build binary and verify end-to-end (`go build -o letsgo.exe .`)

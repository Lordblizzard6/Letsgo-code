# Research: TUI Modernization & Feature Architecture

## 1. Dynamic Model Discovery in Terminal

### Decision
Integrate `internal/api/discovery.go` and `internal/config/config.go` into the TUI's model picker instead of the static hardcoded arrays in `internal/tui/ui.go`.

### Rationale
`internal/api/discovery.go` already implements live model querying for Gemini, OpenRouter, Groq, Ollama, OpenAI, and Anthropic. Reusing this gives the terminal user identical dynamic model listing as the GUI, including locally downloaded Ollama models and all OpenRouter endpoints, without maintaining separate hardcoded lists.

### Alternatives Considered
- *Static configuration list*: Easy to implement but perpetually out-of-date and unable to see local Ollama models.
- *External CLI call to open settings*: Disrupts the terminal flow.

---

## 2. Modal & Overlay Architecture in Bubble Tea

### Decision
Refactor `model.currentMode` (which only had `chatMode` and `settingsMode`) into an explicit overlay state enum:
```go
type tuiOverlay int

const (
    overlayNone tuiOverlay = iota
    overlayModelPicker
    overlayGitReview
    overlayActivityInspector
    overlaySessionPicker
)
```

### Rationale
A single clean state variable makes keyboard routing straightforward. When `overlay != overlayNone`, the overlay intercepts keystrokes (`Esc` closes, arrows navigate, `Enter` confirms). When `overlay == overlayNone`, keystrokes route to normal chat and the textarea.

### Alternatives Considered
- *Multiple boolean flags (`inModelMenu`, `inGitMenu`, `inActivityMenu`)*: Prone to state conflicts and desynchronization.
- *Multiple bubble tea sub-programs*: Destroys terminal alternate screen buffers and complicates state sharing.

---

## 3. Interactive Autocomplete Popup

### Decision
When the cursor in `textarea` is at word starting with `/` or `@`, compute matches against `AvailableSlashCommands` or project files/mentions. Display a floating box positioned above the composer with up to 6 matches, showing icon, name, and description.

### Rationale
Matches the Antigravity and modern CLI UX (like Claude Code and OpenCode). Allows fast exploration of commands without consulting documentation.

---

## 4. Git Review & Staging in Terminal

### Decision
Implement terminal-native staged and unstaged file lists with +/- indicators, line stats (`+Additions / -Deletions`), interactive unified diff preview with line numbers, and quick commit action.

### Rationale
Terminal users need to quickly inspect diffs after code editing tools run, stage appropriate files, and commit without having to exit or open another terminal window.

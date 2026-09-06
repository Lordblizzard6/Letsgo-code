# Data Model: TUI Modernization & Feature Extensions

## TUI State Enums & Types

### 1. `tuiOverlay`
Defines which full-view or popover modal is currently active over the chat canvas.
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

### 2. `AutocompleteState`
Tracks real-time popup suggestion state.
```go
type AutocompleteItem struct {
    Type        string // "cmd" | "file" | "skill" | "tool"
    Icon        string
    Title       string
    Description string
    Value       string
}

type AutocompleteState struct {
    Active        bool
    Trigger       string // "/" or "@"
    Query         string
    SelectedIndex int
    Items         []AutocompleteItem
}
```

### 3. `TUIGitFile` & `TUIGitSummary`
Structured representation of Git changes for terminal review.
```go
type TUIGitFile struct {
    Path      string
    Name      string
    Dir       string
    Status    string // "M", "A", "D", "??"
    Staged    bool
    Additions int
    Deletions int
}

type TUIGitSummary struct {
    Branch           string
    Clean            bool
    UncommittedCount int
    CommittedCount   int
    Files            []TUIGitFile
    SelectedFile     string
    DiffContent      string
}
```

### 4. `TUIActivityItem`
Tracks commands and tool calls for the Activity Inspector.
```go
type TUIActivityItem struct {
    ID        string
    Type      string // "tool" | "command" | "task"
    Name      string
    Status    string // "running" | "done" | "failed"
    Input     string
    Output    string
    StartTime time.Time
    Duration  time.Duration
    Diff      string
}
```

### 5. `TUIModelItem`
Dynamic model representation in the model switcher.
```go
type TUIModelItem struct {
    Provider string // "gemini" | "anthropic" | "openai" | "groq" | "openrouter" | "ollama" | "deepseek"
    ID       string
    Name     string
    Desc     string
    IsActive bool
}
```

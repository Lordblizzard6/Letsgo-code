# Phase 1 Data Model: Opencode Parity & Antigravity Activity Panel

## Entities & Types

### 1. `GitFileDiff` (Go & TypeScript)
Represents a file modified in the git working tree.

```go
type GitFileDiff struct {
    Path      string `json:"path"`
    Name      string `json:"name"`
    Dir       string `json:"dir"`
    Status    string `json:"status"`    // "M", "A", "D", "U", "??"
    Additions int    `json:"additions"` // Lines added
    Deletions int    `json:"deletions"` // Lines deleted
    IsStaged  bool   `json:"is_staged"` // True if staged in git index
}
```

```typescript
export interface GitFileDiff {
    path: string;
    name: string;
    dir: string;
    status: string;
    additions: number;
    deletions: number;
    is_staged: boolean;
}
```

### 2. `GitDiffSummary` (Go & TypeScript)
Summary of git state for the activity inspector.

```go
type GitDiffSummary struct {
    IsRepo           bool          `json:"is_repo"`
    Clean            bool          `json:"clean"`
    Branch           string        `json:"branch"`
    UncommittedCount int           `json:"uncommitted_count"`
    CommittedCount   int           `json:"committed_count"`
    Files            []GitFileDiff `json:"files"`
    RawDiff          string        `json:"raw_diff"`
}
```

```typescript
export interface GitDiffSummary {
    is_repo: boolean;
    clean: boolean;
    branch: string;
    uncommitted_count: number;
    committed_count: number;
    files: GitFileDiff[];
    raw_diff: string;
}
```

### 3. `ActivityTab` (TypeScript)
Active tab within the Activity Inspector.

```typescript
export type ActivityTab = "overview" | "git" | "commands";
```

### 4. `CommandLogItem` (TypeScript)
Represents a recorded command, tool call, or MCP call.

```typescript
export interface CommandLogItem {
    id: string;
    name: string;
    kind: "tool" | "bash" | "mcp" | "skill";
    args?: string;
    status: "running" | "done" | "error";
    durationMs?: number;
    timestamp: number;
    output?: string;
    diff?: string; // Embedded git diff if files were modified
}
```

### 5. `BackgroundTask` (TypeScript)
Represents background task in the Overview panel.

```typescript
export interface BackgroundTask {
    id: string;
    title: string;
    status: "running" | "done" | "error";
    startTime: number;
}
```

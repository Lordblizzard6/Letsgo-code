# Data Model: Composer Layout and Window Sizing

## Entities & Enums

### WorkMode
```typescript
type WorkModeId = "code" | "architecture" | "planning" | "research";

interface WorkMode {
  id: WorkModeId;
  icon: React.ComponentType;
  i18n: string;
  hint: string;
}
```

Order of Tab cycling:
1. `code` (Código)
2. `architecture` (Arch)
3. `planning` (Plan)
4. `research` (Research)

### WindowConfig (Go)
```go
type WindowConfig struct {
    Width     int
    Height    int
    MinWidth  int
    MinHeight int
}
```
Default: `Width: 1280`, `Height: 850`, `MinWidth: 960`, `MinHeight: 600`.

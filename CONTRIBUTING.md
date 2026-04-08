# Contributing to Let's Go Code

Thank you for your interest in contributing to **Let's Go Code**! This document provides guidelines and instructions for contributing to the project.

## Development Setup

### Prerequisites
- Go 1.21 or later
- Git
- Make (optional)

### Getting Started

1. **Fork and clone**
   ```bash
   git clone https://github.com/your-username/go-claude-code.git
   cd go-claude-code
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Build the project**
   ```bash
   go build -o claudego
   ```

4. **Run tests**
   ```bash
   go test ./...
   ```

## Project Structure

```
go-claude-code/
├── cmd/              # CLI command implementations
│   ├── add.go
│   ├── chat.go
│   └── ...
├── internal/
│   ├── api/          # API client for AI providers
│   ├── config/       # Configuration management
│   ├── db/           # SQLite database
│   ├── tui/          # Terminal UI (Bubble Tea)
│   └── tools/        # Tool implementations
├── main.go           # Entry point
└── go.mod            # Go module definition
```

## Adding a New Command

1. Create a new file in `cmd/` directory:
   ```go
   // cmd/mycommand.go
   package cmd

   import "github.com/spf13/cobra"

   func init() {
       rootCmd.AddCommand(myCommand)
   }

   var myCommand = &cobra.Command{
       Use:   "mycommand [args]",
       Short: "Brief description",
       Long:  `Longer description`,
       Run: func(cmd *cobra.Command, args []string) {
           // Implementation
       },
   }
   ```

2. Register flags in `init()`:
   ```go
   func init() {
       rootCmd.AddCommand(myCommand)
       myCommand.Flags().StringP("flag", "f", "", "Flag description")
   }
   ```

3. Build and test:
   ```bash
   go build -o claudego
   ./claudego mycommand --help
   ```

## Adding a New Tool

Tools are implemented in `internal/tools/`:

```go
// internal/tools/my_tool.go
package tools

type MyTool struct{}

func (t MyTool) Name() string {
    return "my_tool"
}

func (t MyTool) Description() string {
    return "What this tool does"
}

func (t MyTool) Schema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "param": map[string]interface{}{
                "type":        "string",
                "description": "Parameter description",
            },
        },
        "required": []string{"param"},
    }
}

func (t MyTool) Execute(input map[string]interface{}) (string, error) {
    // Implementation
    return "result", nil
}
```

Register in `internal/tools/registry.go`:
```go
func init() {
    RegisterTool(MyTool{})
}
```

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Run `go vet` for static analysis
- Add comments for exported functions
- Use meaningful variable names

### Naming Conventions
- **Files:** `snake_case.go`
- **Functions:** `PascalCase` for exported, `camelCase` for private
- **Variables:** `camelCase`
- **Constants:** `UPPER_SNAKE_CASE` or `PascalCase`
- **Types:** `PascalCase`

## Testing

### Unit Tests
```bash
go test ./internal/...
go test ./cmd/...
```

### Integration Tests
```bash
go test -tags=integration ./...
```

### Coverage
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Commit Messages

Use conventional commits format:
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `refactor:` Code refactoring
- `test:` Test changes
- `chore:` Build/dependency changes

Examples:
```
feat: add branch command with subcommands
fix: resolve issue with session resume
docs: update README with new examples
```

## Pull Request Process

1. **Create a branch**
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make changes**
   - Write code
   - Add tests
   - Update documentation

3. **Run checks**
   ```bash
   go build ./...
   go test ./...
   go vet ./...
   ```

4. **Commit and push**
   ```bash
   git add .
   git commit -m "feat: add new feature"
   git push origin feature/my-feature
   ```

5. **Create PR**
   - Fill out PR template
   - Link related issues
   - Request review

## Documentation

- Update README.md if adding user-facing features
- Add godoc comments for exported functions
- Update PARITY_REPORT.md if implementing TypeScript features
- Update PROJECT_STATUS.md with current state

## Questions?

- Open an issue for bugs
- Start a discussion for features
- Check existing documentation

## Code of Conduct

- Be respectful and constructive
- Welcome newcomers
- Focus on what is best for the community
- Show empathy towards others

Thank you for contributing! 🎉

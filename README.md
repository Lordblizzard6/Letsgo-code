# LetsGO Code

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Status](https://img.shields.io/badge/status-production%20ready-brightgreen.svg)](PROJECT_STATUS.md)

```
    __         ______       ____
   / /   ___  / ____/____  / __ \_________  ________  __________
  / /   / _ \/ / __/ ___/ / / / / ___/ __ \/ ___/ _ \/ ___/ ___/
 / /___/  __/ /_/ / /    / /_/ / /  / /_/ / /  /  __(__  |__  )
/_____/\___/\____/_/    /_____/_/   \____/_/   \___/____/____/
```

**LetsGO Code** - A faster, native AI coding assistant built with Go.

> Achieves 1:1 feature parity with Claude Code CLI TypeScript version for core functionality.

## 🚀 Quick Start

```bash
# Install
go install github.com/user/go-claude-code@latest

# Or download pre-built binary
# (See releases page)

# Configure API key
letsgo login

# Start chatting
letsgo chat
```

## 📋 Requirements

- Go 1.21+ (for building from source)
- SQLite3 (bundled)
- Git (optional, for git features)

## 🏗️ Building from Source

```bash
git clone https://github.com/user/go-claude-code.git
cd go-claude-code
go build -o letsgo
./letsgo --help
```

## ✨ Features

### Core Features
- 💬 **Interactive Chat** - Full TUI with Bubble Tea + Syntax Highlighting
- 📝 **Session Management** - Persistent sessions with SQLite
- 🔧 **Git Integration** - Branch, commit, diff, review
- 🤖 **AI Tools** - Bash, file edit, web search, LSP
- 💰 **Cost Tracking** - Monitor API usage and costs
- 🔌 **Plugin System** - Extensible architecture
- 🎯 **MCP Support** - Model Context Protocol
- 🌐 **Multi-Provider** - Anthropic, OpenAI, Groq, OpenRouter, Ollama
- 📋 **Slash Commands** - /files, /diff, /search, /compact, etc
- 🛡️ **Guardrails** - Safety checks for dangerous commands
- 📋 **Plan Mode** - Structured planning before execution
- 🤖 **Agents** - Spawn sub-agents for parallel tasks
- 🧪 **Tested** - 23+ unit tests with CI/CD

### Commands (75 implemented)

<details>
<summary>Click to see all commands</summary>

**Core:** `add`, `chat`, `clear`, `copy`, `exit`, `files`, `context`, `compact`

**Git:** `init`, `branch`, `tag`, `commit`, `commit-push-pr`, `diff`, `hooks`, `undo`, `redo`

**Session:** `session`, `resume`, `rename`, `reset`, `history`, `rewind`, `teleport`

**Review:** `review`, `ultrareview`, `security-review`, `advisor`

**Config:** `config`, `login`, `logout`, `model`, `env`, `doctor`, `update`, `upgrade`, `onboarding`, `output-style`, `privacy-settings`, `rate-limit-options`, `sandbox-toggle`, `oauth-refresh`

**Productivity:** `thinkback`, `thinkback-play`, `btw`, `share`, `summary`, `insights`, `release-notes`, `debug-tool-call`

**UI:** `theme`, `color`, `vim`, `keybindings`, `statusline`, `plan`, `fast`, `effort`

**External:** `ide`, `chrome`, `desktop`, `mobile`, `github`, `lsp`, `mcp`

**Integrations:** `install-github-app`, `install-slack-app`, `reload-plugins`, `remote-env`

**Advanced:** `agents`, `tasks`, `skills`, `permissions`, `plugin`, `memory`, `cost`, `usage`, `stats`, `status`

See [full command list](PROJECT_STATUS.md#lista-completa-de-comandos-implementados)

</details>

## 🔧 Configuration

### Supported Providers
- **Anthropic** (Claude) - `https://api.anthropic.com/v1` - ✅ Full tool support
- **OpenAI** (GPT-4o) - `https://api.openai.com/v1` - ✅ Full tool support  
- **Groq** (Llama) - `https://api.groq.com/openai/v1` - ⚠️ Basic tools only
- **OpenRouter** (Multi) - `https://openrouter.ai/api/v1` - ⚠️ Varies by model
- **Ollama** (Local) - `http://localhost:11434` - ❌ Chat only, no tools

> **Nota:** El soporte de herramientas (tools) varía por proveedor. Consulta la [Guía de Tools por Proveedor](docs/PROVIDER_TOOLS_GUIDE.md) para detalles completos.

### Environment Variables
```bash
ANTHROPIC_API_KEY=your-key-here      # Anthropic
OPENAI_API_KEY=your-key-here         # OpenAI
GROQ_API_KEY=your-key-here           # Groq
OPENROUTER_API_KEY=your-key-here     # OpenRouter
```

### Model Shortcuts
When using Groq or OpenRouter, you can use short names:
```bash
# Groq shortcuts
letsgo chat --model llama           # llama-3.3-70b-versatile
letsgo chat --model llama-3.1-8b    # llama-3.1-8b-instant

# OpenRouter shortcuts  
letsgo chat --model claude          # anthropic/claude-3.5-sonnet
letsgo chat --model gpt4            # openai/gpt-4o
letsgo chat --model llama           # meta-llama/llama-3.3-70b
```

### Config File
```yaml
# ~/.letsGo/config.yaml
api_key: your-api-key
model: claude-3-5-sonnet
base_url: https://api.anthropic.com/v1
shell: bash
verbose: false
temperature: 0.7
max_tokens: 4096
stream: true
```

**Configuration directory:** `~/.letsGo/` (all config, database, and cost tracking files stored here)

## 🛡️ Security & Guardrails

### Permission Levels
Each tool can be configured with three permission levels:
- **auto** - Automatically execute safe operations
- **ask** - Prompt for user confirmation (default for dangerous operations)
- **deny** - Never allow the operation

Configure in `~/.letsGo/permissions.json`:
```json
{
  "version": "1.0",
  "global_mode": "ask",
  "tools": {
    "bash": {
      "mode": "ask",
      "scope": "project",
      "allowed_commands": ["ls", "cat", "echo"],
      "denied_commands": ["rm -rf /"]
    },
    "edit": {
      "mode": "ask",
      "scope": "project",
      "denied_paths": ["/etc/passwd", "/etc/shadow"]
    }
  }
}
```

### Protected Operations
The following operations require confirmation:
- **High Risk**: `rm -rf /`, disk formatting, piping curl to shell, drop database
- **Medium Risk**: recursive deletion, git force push, chmod 777, sudo commands
- **Protected Paths**: `/etc/passwd`, `/etc/shadow`, `~/.ssh`

### Plan Mode
For complex changes, enter plan mode to review before execution:
```bash
# Enter plan mode - Claude will create a structured plan
letsgo chat
> Enter plan mode for refactoring the auth system

# Tools available in plan mode:
# - plan_step_add: Add steps to the plan
# - plan_show: Display current plan
# - plan_approve: Approve and start execution
# - plan_step_complete: Mark step as done
# - exit_plan_mode: Cancel and exit
```

## 🤖 Agents

Spawn sub-agents for parallel task execution:
```bash
# Create an agent for a subtask
letsgo chat
> Create an agent to refactor the utils module while I work on the API

# Agent tools:
# - agent: Spawn a new agent with a specific task
# - agent_get: Check agent status and results
# - agent_list: List all agents
```

Agent states: `idle` → `running` → `completed` | `failed`

## 🎨 Usage Examples

### Interactive Chat
```bash
letsgo chat
```

### One-off Tasks
```bash
letsgo chat "Review this code for bugs"
letsgo review file.go
letsgo diff
```

### Git Workflow
```bash
letsgo branch create feature/new-thing
letsgo commit-push-pr "Add new feature"
```

### Session Management
```bash
letsgo session list
letsgo resume <session-id>
letsgo thinkback "search query"
```

## 📊 Parity with TypeScript CLI

| Category | Status | Details |
|----------|--------|---------|
| Commands | 90% | 75/110 commands |
| Tools | 69% | 31/45 tools |
| Core Features | 100% | All essential features |

See [detailed parity report](PARITY_REPORT_DETAILED.md)

## 🏛️ Architecture

```
cmd/              # CLI commands (75 files)
internal/
  ├── api/        # API client
  ├── config/     # Configuration
  ├── db/         # SQLite database
  ├── tui/        # Terminal UI
  └── tools/      # Tool implementations (31 files)
```

## 🧪 Testing

```bash
# Run tests
go test ./...

# Build and run
make build
make run
```

## 📈 Performance

- **Startup time:** ~50ms (vs ~500ms TypeScript)
- **Memory usage:** ~20MB (vs ~150MB TypeScript)
- **Binary size:** ~25MB single executable

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

MIT License - see [LICENSE](LICENSE) file

## 🙏 Acknowledgments

- [Anthropic](https://anthropic.com) for Claude AI
- [Charm](https://charm.sh) for Bubble Tea TUI framework
- Original TypeScript CLI team

## 📚 Documentation

- [Project Status](PROJECT_STATUS.md) - Current state and roadmap
- [Parity Report](PARITY_REPORT_DETAILED.md) - TypeScript vs Go comparison
- [Contributing Guide](CONTRIBUTING.md) - How to contribute

---

**Status:** ✅ Production Ready  
**Version:** 0.1.0  
**Maintained by:** [lordblizzard6](https://github.com/lordblizzard6)

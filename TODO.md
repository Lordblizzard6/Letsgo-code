# TODO COMPLETO - Let's Go Code

> **Fecha:** Abril 2026  
> **Estado:** 75 comandos ✅, 31 tools ✅, Infraestructura ❌

---

## 🔴 CRÍTICO - Infraestructura de Producción

### 1. Sistema de Tests (0% → 70% objetivo)

#### Tests Unitarios
```
cmd/
  config_test.go          - Test de carga/guardado de config
  chat_test.go            - Test de inicialización
  doctor_test.go          - Test de chequeos de salud
  ... (un test por comando)

internal/
  api/
    client_test.go        - Test de API calls con mocks
    types_test.go         - Test de serialización
  
  tools/
    bash_test.go          - Test de ejecución de comandos
    file_ops_test.go      - Test de read/write/edit
    grep_test.go          - Test de búsqueda
    registry_test.go      - Test de registro de tools
  
  db/
    database_test.go      - Test de operaciones SQL
  
  tui/
    ui_test.go            - Test de renderizado
    input_test.go         - Test de manejo de input
  
  config/
    config_test.go        - Test de configuración
```

#### Tests de Integración
```
tests/integration/
  chat_session_test.go    - Flujo completo de chat
  git_workflow_test.go    - Flujo de git (add/commit/push)
  tool_execution_test.go  - Ejecución de tools
  config_persistence_test.go - Persistencia de config
```

#### Tests E2E
```
tests/e2e/
  cli_test.go             - Test del CLI completo
  install_test.go         - Test de instalación
```

#### Framework de Test
- Usar `testify` para assertions
- Usar `gomock` o `mockery` para mocks
- Coverage mínimo: 70% de funciones públicas

---

### 2. CI/CD Pipeline (GitHub Actions)

#### Archivo: `.github/workflows/ci.yml`
```yaml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - run: go mod download
      - run: go test -v -race -coverprofile=coverage.out ./...
      - run: go tool cover -html=coverage.out -o coverage.html
      
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          
  build:
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
        arch: [amd64, arm64]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - run: go build -o letsgo-${{ matrix.os }}-${{ matrix.arch }}
```

#### Archivo: `.github/workflows/release.yml`
```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - name: Build all platforms
        run: |
          GOOS=linux GOARCH=amd64 go build -o dist/letsgo-linux-amd64
          GOOS=linux GOARCH=arm64 go build -o dist/letsgo-linux-arm64
          GOOS=darwin GOARCH=amd64 go build -o dist/letsgo-darwin-amd64
          GOOS=darwin GOARCH=arm64 go build -o dist/letsgo-darwin-arm64
          GOOS=windows GOARCH=amd64 go build -o dist/letsgo-windows-amd64.exe
      - uses: softprops/action-gh-release@v1
        with:
          files: dist/*
```

---

### 3. Documentación godoc

Comentar todas las funciones y tipos públicos:
```go
// ExecuteTool ejecuta una herramienta por su nombre con los parámetros dados.
// Retorna el resultado como string o un error si la ejecución falla.
// Ejemplo:
//   result, err := ExecuteTool("BashTool", map[string]interface{}{"command": "ls"})
func ExecuteTool(name string, input interface{}) (string, error)
```

Archivos a documentar:
- [ ] `internal/api/*.go` - 15 funciones
- [ ] `internal/tools/*.go` - 35 tools
- [ ] `internal/db/*.go` - 20 funciones
- [ ] `internal/tui/*.go` - 10 funciones
- [ ] `internal/config/*.go` - 8 funciones
- [ ] `cmd/*.go` - 75 comandos (solo los públicos)

---

## 🟡 MEDIA - Mejoras de UX/UI

### 4. Mejoras TUI (Terminal User Interface)

#### a) Header persistente con información
```
┌─────────────────────────────────────────────────────────────┐
│ Let's Go Code v0.1.0    Session: abc123    Model: Claude-3.5 │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ User: Hola                                                  │
│ Let'sGo(Claude-3.5): ¡Hola! ¿En qué puedo ayudarte?        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
> █
```

#### b) Split panels (sidebar + chat)
- Sidebar izquierda: historial de sesiones, archivos recientes
- Panel derecha: chat principal

#### c) Syntax highlighting
- Resaltar código en las respuestas del modelo
- Soporte para Go, TypeScript, Python, etc.

#### d) Vim mode completo
- Modos normal, insert, visual
- Comandos vim básicos (hjkl, dd, yy, p)
- Atajos de navegación

#### e) Mouse support
- Click para posicionar cursor
- Scroll con rueda del mouse
- Selección de texto

#### f) Tema personalizable
- Archivo de tema: `~/.letsgo/theme.yml`
- Colores configurables para cada elemento
- Temas predefinidos: default, dark, light, high-contrast

### 5. Mensajes de Bienvenida Mejorados

Al iniciar el chat:
```
🚀 Let's Go Code v0.1.0
Session: abc123 | Model: Claude-3.5 Sonnet

💡 Tips:
   • Type /help for commands
   • Use Ctrl+S for settings
   • Press Ctrl+C to exit
   • Try /skills to see available skills

> █
```

### 6. Comandos Slash Mejorados

| Comando | Descripción |
|---------|-------------|
| `/help` | Mostrar ayuda interactiva |
| `/model` | Cambiar modelo con picker |
| `/settings` | Abrir menú de configuración |
| `/clear` | Limpiar pantalla |
| `/save` | Guardar sesión actual |
| `/export` | Exportar conversación |
| `/cost` | Mostrar costo acumulado |
| `/tokens` | Mostrar uso de tokens |
| `/files` | Listar archivos en contexto |
| `/compact` | Compactar contexto |
| `/undo` | Deshacer último mensaje |
| `/redo` | Rehacer mensaje |

---

## 🟢 BAJA - Funcionalidades Adicionales

### 7. Sistema de Plugins Completo

#### Cargar plugins dinámicamente:
```go
type Plugin interface {
    Name() string
    Version() string
    Init() error
    GetCommands() []Command
    GetTools() []Tool
}
```

- Directorio de plugins: `~/.letsgo/plugins/`
- Formatos soportados: `.so` (Linux), `.dylib` (macOS), `.dll` (Windows)
- Hot-reload de plugins

### 8. Workflow System Avanzado

```yaml
# .github/workflows/letsgo-ci.yml
name: Code Review
on:
  workflow_call:
    
steps:
  - run: |
      letsgo chat << 'EOF'
      Please review the code in this PR and provide feedback on:
      1. Code quality
      2. Potential bugs
      3. Performance issues
      4. Security concerns
      EOF
```

- Parámetros en workflows
- Validación de workflows
- Workflow marketplace

### 9. Integraciones Enterprise

#### GitHub App Real
- OAuth flow completo
- Webhooks para PRs
- Checks API integration
- Review comments automation

#### Slack App Real
- OAuth con workspaces
- Slash commands (/letsgo)
- Notificaciones de sesiones
- Thread replies

### 10. Feature Flags Experimentales

Evaluación caso a caso:
- `FORK_SUBAGENT` - Fork de subagentes
- `TORCH` - Búsqueda semántica avanzada
- `ULTRAPLAN` - Planificación avanzada

---

## 📊 Métricas de Implementación

| Área | Actual | Objetivo | Esfuerzo Est. |
|------|--------|----------|---------------|
| Tests Unitarios | 0% | 70% | 2-3 semanas |
| CI/CD | 0% | 100% | 3-5 días |
| godoc | 20% | 100% | 1 semana |
| TUI Mejoras | 60% | 90% | 1-2 semanas |
| Plugins | 30% | 100% | 2 semanas |
| Integraciones | Mock | Real | 2-3 semanas |

---

## 🎯 Roadmap Priorizado

### Sprint 1 (1 semana)
1. ✅ Mostrar ASCII art en chat
2. 🔄 Mejorar header de TUI
3. 🔄 Comandos slash básicos
4. 🔄 Setup CI/CD básico

### Sprint 2 (1 semana)
1. Tests unitarios core (api, tools)
2. godoc básico
3. Colores y tema en TUI
4. Mensajes de bienvenida

### Sprint 3 (1 semana)
1. Tests de integración
2. Vim mode básico
3. Settings interactivos
4. Mejoras de UX

### Sprint 4 (1-2 semanas)
1. Plugin system dinámico
2. Workflow avanzado
3. Integraciones enterprise
4. Benchmarks

---

## ✅ Checklist de Completitud

### Infraestructura
- [ ] Tests unitarios > 70%
- [ ] CI/CD GitHub Actions
- [ ] godoc completo
- [ ] Release automation

### UI/UX
- [ ] ASCII art en chat
- [ ] Header persistente con info
- [ ] Colores consistentes
- [ ] Mensaje de bienvenida
- [ ] Comandos slash
- [ ] Settings interactivo

### Avanzado
- [ ] Plugin system completo
- [ ] Workflow system
- [ ] Integraciones reales
- [ ] Feature flags evaluados

---

*Última actualización: Abril 2026*

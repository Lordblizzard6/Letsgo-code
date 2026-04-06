# MyCLI - CLI Asistente de Código Multiproveedor

Un CLI tipo Claude Code que soporta múltiples proveedores de LLM (OpenAI, Anthropic, Gemini) con capacidades de agente autónomo.

## 🚀 Instalación

```bash
go mod download
go build -o mycli
```

## ⚙️ Configuración

### Opción 1: Variables de entorno

```bash
export ANTHROPIC_API_KEY=sk-ant-...
# o
export OPENAI_API_KEY=sk-...
# o
export GEMINI_API_KEY=...
```

### Opción 2: Archivo de configuración

Copia `config.yaml.example` a `config.yaml` o `~/.mycli/config.yaml`:

```yaml
provider:
  name: anthropic
  api_key: ""  # O usar variable de entorno
  model: claude-sonnet-4-20250514
  temperature: 0.7
  max_tokens: 4096

agent:
  max_iterations: 10

tools:
  require_confirmation: true
```

## 📖 Uso

### Modo interactivo

```bash
./mycli -i
```

### Modo comando directo

```bash
./mycli "crea un archivo hello.go con un Hola Mundo"
```

### Especificar proveedor

```bash
./mycli -p openai -m gpt-4o "tu prompt"
```

## 🛠️ Herramientas Disponibles

El agente puede usar las siguientes herramientas automáticamente:

- **read_file**: Leer archivos
- **write_file**: Escribir/crear archivos
- **list_dir**: Listar directorios
- **run_command**: Ejecutar comandos de shell (con sandbox de seguridad)
- **git_status**: Estado del repositorio git
- **git_diff**: Cambios no commitados
- **git_log**: Historial de commits

## 🔒 Seguridad

- Comandos peligrosos bloqueados (`rm -rf`, `sudo`, etc.)
- Path traversal prevenido en operaciones de archivos
- Timeout de 60s en comandos de shell
- Confirmación requerida para ejecución de herramientas (configurable)

## 🏗️ Arquitectura

```
mycli/
├── cmd/              # Punto de entrada CLI
├── internal/
│   ├── llm/          # Proveedores LLM (OpenAI, Anthropic, Gemini)
│   ├── agent/        # Bucle del agente y gestión de contexto
│   ├── tools/        # Herramientas ejecutables
│   ├── ui/           # Interfaz TUI (Bubbletea)
│   └── config/       # Configuración
└── main.go
```

## 📝 Licencia

MIT

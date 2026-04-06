package llm

import (
	"context"
	"encoding/json"
)

// ToolCall representa una llamada a herramienta desde el LLM
type ToolCall struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Args  json.RawMessage `json:"args"`
}

// ContentBlock representa un bloque de contenido (texto, imagen, archivo, etc.)
type ContentBlock struct {
	Type     string          `json:"type"` // "text", "image", "file", "tool_use", "tool_result"
	Text     string          `json:"text,omitempty"`
	ImageURL string          `json:"image_url,omitempty"`
	Data     []byte          `json:"data,omitempty"`
	MIMEType string          `json:"mime_type,omitempty"`
	ToolCall *ToolCall       `json:"tool_call,omitempty"`
	ToolResult *ToolResult   `json:"tool_result,omitempty"`
}

// ToolResult representa el resultado de una herramienta ejecutada
type ToolResult struct {
	ToolID    string `json:"tool_id"`
	ToolName  string `json:"tool_name"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error,omitempty"`
}

// Message representa un mensaje en el historial de chat
type Message struct {
	Role        string         `json:"role"`           // "system", "user", "assistant", "tool"
	Content     string         `json:"content"`        // Contenido texto plano (legacy)
	Blocks      []ContentBlock `json:"blocks,omitempty"` // Contenido estructurado multimodal
	ToolCalls   []ToolCall     `json:"tool_calls,omitempty"`
	ToolID      string         `json:"tool_id,omitempty"`
	ToolName    string         `json:"tool_name,omitempty"`
	Attachments []Attachment   `json:"attachments,omitempty"`
}

// Attachment representa un archivo adjunto
type Attachment struct {
	Filename string `json:"filename"`
	Path     string `json:"path"`
	MIMEType string `json:"mime_type"`
	Size     int64  `json:"size"`
}

// Chunk representa un fragmento de respuesta streaming
type Chunk struct {
	Content     string
	Blocks      []ContentBlock
	ToolCalls   []ToolCall
	Done        bool
	Err         error
	Usage       UsageInfo
	Thought     string // Para modelos que exponen chain-of-thought
}

// Tool define una herramienta disponible para el agente
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Schema      json.RawMessage `json:"schema"` // JSON Schema de los parámetros
	Handler     ToolHandler     `json:"-"`
	Category    string          `json:"category"` // "fs", "shell", "git", "mcp", "custom"
	RequiresAuth bool           `json:"requires_auth,omitempty"`
	Timeout     int             `json:"timeout,omitempty"` // segundos, 0 = default
	Hidden      bool            `json:"hidden,omitempty"`  // No mostrar al usuario
}

// ToolHandler es la función que ejecuta la herramienta
type ToolHandler func(ctx context.Context, args json.RawMessage) (string, error)

// ToolHandlerWithProgress es para herramientas que reportan progreso
type ToolHandlerWithProgress func(ctx context.Context, args json.RawMessage, progress chan<- string) (string, error)

// Response representa una respuesta completa no-streaming
type Response struct {
	Content     string
	Blocks      []ContentBlock
	ToolCalls   []ToolCall
	Usage       UsageInfo
	Model       string
	FinishReason string
}

// UsageInfo información de tokens usados
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
	CacheReadTokens  int `json:"cache_read_tokens,omitempty"`  // Para cache de prompts
	CacheWriteTokens int `json:"cache_write_tokens,omitempty"` // Para cache de prompts
}

// Provider es la interface que todos los proveedores deben implementar
type Provider interface {
	Name() string
	Model() string
	SupportsStreaming() bool
	SupportsTools() bool
	SupportsMultimodal() bool
	SupportsPromptCaching() bool
	
	// Chat básico
	Stream(ctx context.Context, messages []Message, tools []Tool) (<-chan Chunk, error)
	Chat(ctx context.Context, messages []Message, tools []Tool) (*Response, error)
	
	// Con sistema de archivos como contexto
	StreamWithFiles(ctx context.Context, messages []Message, tools []Tool, files []FileContext) (<-chan Chunk, error)
	
	// Contadores y límites
	GetTokenCount(messages []Message) int
	GetMaxTokens() int
}

// FileContext representa un archivo como contexto para el LLM
type FileContext struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Lines   LineRange `json:"lines,omitempty"`
}

// LineRange especifica un rango de líneas
type LineRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// MCPConfig configuración para Model Context Protocol
type MCPConfig struct {
	ServerName string `json:"server_name"`
	Command    string `json:"command"`
	Args       []string `json:"args,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
	Timeout    int    `json:"timeout,omitempty"`
}

// SecurityPolicy políticas de seguridad para ejecución de herramientas
type SecurityPolicy struct {
	AllowedPaths      []string `json:"allowed_paths"`       // Paths permitidos
	DeniedPaths       []string `json:"denied_paths"`        // Paths denegados
	AllowedCommands   []string `json:"allowed_commands"`    // Comandos permitidos (whitelist)
	DeniedCommands    []string `json:"denied_commands"`     // Comandos denegados (blacklist)
	MaxFileSize       int64    `json:"max_file_size"`       // Tamaño máximo de archivo
	RequireConfirmFor []string `json:"require_confirm_for"` // Herramientas que requieren confirmación
	EnableAuditLog    bool     `json:"enable_audit_log"`
}

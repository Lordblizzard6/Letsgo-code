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

// Message representa un mensaje en el historial de chat
type Message struct {
	Role      string     `json:"role"`      // "system", "user", "assistant", "tool"
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	ToolID    string     `json:"tool_id,omitempty"` // ID de la herramienta para respuestas
	ToolName  string     `json:"tool_name,omitempty"` // Nombre para respuestas tool
}

// Chunk representa un fragmento de respuesta streaming
type Chunk struct {
	Content   string
	ToolCalls []ToolCall
	Done      bool
	Err       error
}

// Tool define una herramienta disponible para el agente
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Schema      json.RawMessage `json:"schema"` // JSON Schema de los parámetros
	Handler     ToolHandler     `json:"-"`
}

// ToolHandler es la función que ejecuta la herramienta
type ToolHandler func(args json.RawMessage) (string, error)

// Response representa una respuesta completa no-streaming
type Response struct {
	Content   string
	ToolCalls []ToolCall
	Usage     UsageInfo
}

// UsageInfo información de tokens usados
type UsageInfo struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// Provider es la interface que todos los proveedores deben implementar
type Provider interface {
	Name() string
	SupportsStreaming() bool
	SupportsTools() bool
	Stream(ctx context.Context, messages []Message, tools []Tool) (<-chan Chunk, error)
	Chat(ctx context.Context, messages []Message, tools []Tool) (*Response, error)
}

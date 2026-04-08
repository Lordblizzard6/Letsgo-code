package api

type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // Can be string or array of blocks
}

type ContentBlock struct {
	Type       string       `json:"type"`
	Text       string       `json:"text,omitempty"`
	Source     *ImageSource `json:"source,omitempty"`
	ToolUse    *ToolUse     `json:"tool_use,omitempty"`
	ToolResult *ToolResult  `json:"tool_result,omitempty"`
}

type ImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type ToolUse struct {
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	Input interface{} `json:"input"`
}

type ToolResult struct {
	ToolUseID string `json:"tool_use_id"`
	ToolName  string `json:"tool_name,omitempty"` // Added to support Gemini/OpenAI requirements
	Content   string `json:"content"`
	IsError   bool   `json:"is_error,omitempty"`
}

type Tool struct {
	Name        string      `json:"name"`
	Type        string      `json:"type,omitempty"` // Added for OpenAI/Groq parity
	Description string      `json:"description"`
	InputSchema interface{} `json:"input_schema,omitempty"`
	Function    *struct {
		Name        string      `json:"name"`
		Description string      `json:"description"`
		Parameters  interface{} `json:"parameters"`
	} `json:"function,omitempty"` // For OpenAI format
}

type Request struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
	Stream    bool      `json:"stream"`
	System    string    `json:"system,omitempty"`
	Tools     []Tool    `json:"tools,omitempty"`
}

type Response struct {
	ID      string         `json:"id"`
	Type    string         `json:"type"`
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
	Usage   Usage          `json:"usage"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// OpenAI / Groq compatibility structures
type OpenAIStreamResponse struct {
	Choices []struct {
		Delta struct {
			Content   string       `json:"content,omitempty"`
			ToolCalls []OpenAITool `json:"tool_calls,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"choices"`
	Error *OpenAIError `json:"error,omitempty"`
}

type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

type OpenAITool struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// Ollama response structures
type OllamaStreamResponse struct {
	Model     string        `json:"model"`
	CreatedAt string        `json:"created_at"`
	Message   OllamaMessage `json:"message"`
	Done      bool          `json:"done"`
}

type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

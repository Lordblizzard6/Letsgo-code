package agent

import (
	"github.com/tuusuario/mycli/internal/llm"
)

type ContextManager struct {
	messages      []llm.Message
	systemPrompt  string
	maxTokens     int
	currentTokens int
}

func NewContextManager(systemPrompt string, maxTokens int) *ContextManager {
	return &ContextManager{
		messages:      make([]llm.Message, 0),
		systemPrompt:  systemPrompt,
		maxTokens:     maxTokens,
		currentTokens: 0,
	}
}

func (c *ContextManager) AddMessage(msg llm.Message) {
	c.messages = append(c.messages, msg)
	c.currentTokens += estimateTokens(msg.Content)

	// Truncar si excede límite
	if c.currentTokens > c.maxTokens {
		c.truncate()
	}
}

func (c *ContextManager) GetMessages() []llm.Message {
	var result []llm.Message
	if c.systemPrompt != "" {
		result = append(result, llm.Message{
			Role:    "system",
			Content: c.systemPrompt,
		})
	}
	result = append(result, c.messages...)
	return result
}

func (c *ContextManager) Clear() {
	c.messages = make([]llm.Message, 0)
	c.currentTokens = 0
}

func (c *ContextManager) truncate() {
	// Estrategia: mantener system + últimos N mensajes
	// Eliminar mensajes del medio o resumirlos

	// Mantener al menos 5 mensajes recientes
	if len(c.messages) > 5 {
		removed := c.messages[:len(c.messages)-5]
		c.messages = c.messages[len(c.messages)-5:]

		for _, m := range removed {
			c.currentTokens -= estimateTokens(m.Content)
		}
	}
}

func estimateTokens(text string) int {
	// Aproximación: ~4 caracteres por token en promedio
	// Para producción, usar tiktoken-go
	return len(text) / 4
}

func (c *ContextManager) GetTokenCount() int {
	return c.currentTokens
}

func (c *ContextManager) GetMessageCount() int {
	return len(c.messages)
}

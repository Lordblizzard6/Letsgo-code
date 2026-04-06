package agent

import (
"github.com/tuusuario/mycli/internal/llm"
)

// ContextManager gestiona el contexto de conversación
type ContextManager struct {
messages     []llm.Message
systemPrompt string
maxTokens    int
currentTokens int
}

// NewContextManager crea un nuevo manejador de contexto
func NewContextManager(systemPrompt string, maxTokens int) *ContextManager {
cm := &ContextManager{
messages:     make([]llm.Message, 0),
systemPrompt: systemPrompt,
maxTokens:    maxTokens,
}

// Añadir system prompt si existe
if systemPrompt != "" {
cm.messages = append(cm.messages, llm.Message{
Role:    "system",
Content: systemPrompt,
})
}

return cm
}

// AddMessage añade un mensaje al contexto
func (cm *ContextManager) AddMessage(msg llm.Message) {
cm.messages = append(cm.messages, msg)
cm.currentTokens += estimateTokens(msg.Content)

// Si excedemos max tokens, truncar mensajes antiguos (excepto system)
for cm.currentTokens > cm.maxTokens && len(cm.messages) > 1 {
// No eliminar el primer mensaje si es system
if cm.messages[0].Role == "system" && len(cm.messages) == 2 {
break
}

removed := cm.messages[1]
cm.messages = append(cm.messages[:1], cm.messages[2:]...)
cm.currentTokens -= estimateTokens(removed.Content)
}
}

// GetMessages devuelve todos los mensajes
func (cm *ContextManager) GetMessages() []llm.Message {
return cm.messages
}

// GetMessageCount devuelve el número de mensajes
func (cm *ContextManager) GetMessageCount() int {
return len(cm.messages)
}

// GetTokenCount devuelve el conteo aproximado de tokens
func (cm *ContextManager) GetTokenCount() int {
return cm.currentTokens
}

// Clear limpia el contexto (manteniendo system prompt)
func (cm *ContextManager) Clear() {
if cm.systemPrompt != "" {
cm.messages = []llm.Message{{
Role:    "system",
Content: cm.systemPrompt,
}}
} else {
cm.messages = make([]llm.Message, 0)
}
cm.currentTokens = 0
}

// SetSystemPrompt actualiza el system prompt
func (cm *ContextManager) SetSystemPrompt(prompt string) {
cm.systemPrompt = prompt

// Reemplazar o añadir system prompt
if len(cm.messages) > 0 && cm.messages[0].Role == "system" {
cm.messages[0].Content = prompt
} else {
cm.messages = append([]llm.Message{{
Role:    "system",
Content: prompt,
}}, cm.messages...)
}
}

// GetLastMessages devuelve los últimos N mensajes
func (cm *ContextManager) GetLastMessages(n int) []llm.Message {
if n >= len(cm.messages) {
return cm.messages
}
return cm.messages[len(cm.messages)-n:]
}

// RemoveToolResults elimina resultados de herramientas para ahorrar tokens
func (cm *ContextManager) RemoveToolResults() {
var filtered []llm.Message
for _, msg := range cm.messages {
if msg.Role != "tool" || len(msg.Content) < 1000 {
filtered = append(filtered, msg)
} else {
// Mantener solo resumen
msg.Content = "[Tool result truncated]"
filtered = append(filtered, msg)
}
}
cm.messages = filtered
}

// estimateTokens estima tokens de forma simple (~4 chars por token)
func estimateTokens(text string) int {
return len(text) / 4
}

package tools

import (
"context"
"encoding/json"
"fmt"

"github.com/tuusuario/mycli/internal/llm"
)

// RegisterThinkTool registra la herramienta think para razonamiento interno
func (r *Registry) RegisterThinkTool() {
r.Register(llm.Tool{
Name:        "think",
Description: "Herramienta interna para razonamiento y planificación. No muestra resultados al usuario directamente.",
Schema:      json.RawMessage(`{"type":"object","properties":{"thought":{"type":"string","description":"Pensamiento o plan interno"},"action":{"type":"string","description":"Acción a tomar después de pensar: 'plan', 'reflect', 'decide'"}},"required":["thought"]}`),
Category:    "internal",
Hidden:      true,
Handler: func(ctx context.Context, args json.RawMessage) (string, error) {
// Esta herramienta es un no-op, solo sirve para que el modelo pueda "pensar"
var params struct {
Thought string `json:"thought"`
Action  string `json:"action,omitempty"`
}
if err := json.Unmarshal(args, &params); err != nil {
return "", err
}

// En el futuro podríamos loggear los pensamientos para debugging
_ = params.Thought
_ = params.Action

return "Thinking recorded", nil
},
})

r.Register(llm.Tool{
Name:        "ask_user",
Description: "Pregunta al usuario para obtener información adicional o confirmación.",
Schema:      json.RawMessage(`{"type":"object","properties":{"question":{"type":"string","description":"Pregunta para el usuario"},"options":{"type":"array","items":{"type":"string"},"description":"Opciones múltiples (opcional)"}},"required":["question"]}`),
Category:    "interaction",
Handler: func(ctx context.Context, args json.RawMessage) (string, error) {
// Esta herramienta será manejada especialmente por el agent loop
var params struct {
Question string   `json:"question"`
Options  []string `json:"options,omitempty"`
}
if err := json.Unmarshal(args, &params); err != nil {
return "", err
}

if len(params.Options) > 0 {
return fmt.Sprintf("Waiting for user response to: %s\nOptions: %v", params.Question, params.Options), nil
}
return fmt.Sprintf("Waiting for user response to: %s", params.Question), nil
},
})
}

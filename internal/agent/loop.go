package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tuusuario/mycli/internal/llm"
	"github.com/tuusuario/mycli/internal/tools"
)

type Agent struct {
	provider      llm.Provider
	ctxManager    *ContextManager
	tools         *tools.Registry
	maxIterations int
	onStream      func(string)
	onToolCall    func(llm.ToolCall) bool // retorna false si cancela
	onToolResult  func(string)
}

func NewAgent(
	provider llm.Provider,
	systemPrompt string,
	tools *tools.Registry,
	maxIterations int,
) *Agent {
	return &Agent{
		provider:      provider,
		ctxManager:    NewContextManager(systemPrompt, 100000),
		tools:         tools,
		maxIterations: maxIterations,
	}
}

func (a *Agent) SetStreamCallback(fn func(string)) {
	a.onStream = fn
}

func (a *Agent) SetToolCallCallback(fn func(llm.ToolCall) bool) {
	a.onToolCall = fn
}

func (a *Agent) SetToolResultCallback(fn func(string)) {
	a.onToolResult = fn
}

func (a *Agent) Run(ctx context.Context, input string) error {
	// Añadir mensaje del usuario
	a.ctxManager.AddMessage(llm.Message{
		Role:    "user",
		Content: input,
	})

	for iteration := 0; iteration < a.maxIterations; iteration++ {
		messages := a.ctxManager.GetMessages()
		availableTools := a.tools.GetDefinitions()

		// Streaming
		stream, err := a.provider.Stream(ctx, messages, availableTools)
		if err != nil {
			return fmt.Errorf("stream error: %w", err)
		}

		var fullResponse strings.Builder
		var pendingCalls []llm.ToolCall

		for chunk := range stream {
			if chunk.Err != nil {
				return fmt.Errorf("chunk error: %w", chunk.Err)
			}

			if chunk.Content != "" {
				fullResponse.WriteString(chunk.Content)
				if a.onStream != nil {
					a.onStream(chunk.Content)
				}
			}

			if len(chunk.ToolCalls) > 0 {
				pendingCalls = append(pendingCalls, chunk.ToolCalls...)
			}

			if chunk.Done {
				break
			}
		}

		// Guardar respuesta del asistente
		a.ctxManager.AddMessage(llm.Message{
			Role:      "assistant",
			Content:   fullResponse.String(),
			ToolCalls: pendingCalls,
		})

		// Si no hay tool calls, terminar
		if len(pendingCalls) == 0 {
			return nil
		}

		// Ejecutar herramientas
		for _, call := range pendingCalls {
			// Confirmación
			if a.onToolCall != nil {
				if !a.onToolCall(call) {
					// Usuario canceló
					a.ctxManager.AddMessage(llm.Message{
						Role:     "tool",
						ToolID:   call.ID,
						ToolName: call.Name,
						Content:  "Tool execution cancelled by user",
					})
					continue
				}
			}

			// Ejecutar
			result, err := a.tools.Execute(call.Name, call.Args)
			if err != nil {
				result = fmt.Sprintf("Error: %v", err)
			}

			if a.onToolResult != nil {
				a.onToolResult(result)
			}

			// Guardar resultado
			a.ctxManager.AddMessage(llm.Message{
				Role:     "tool",
				ToolID:   call.ID,
				ToolName: call.Name,
				Content:  result,
			})
		}
	}

	return fmt.Errorf("max iterations (%d) reached", a.maxIterations)
}

func (a *Agent) GetContext() *ContextManager {
	return a.ctxManager
}

package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type AnthropicProvider struct {
	client      *anthropic.Client
	model       string
	maxTokens   int
	temperature float32
}

func NewAnthropic(apiKey, model string) (*AnthropicProvider, error) {
	return &AnthropicProvider{
		client:      anthropic.NewClient(option.WithAPIKey(apiKey)),
		model:       model,
		maxTokens:   4096,
		temperature: 0.7,
	}, nil
}

func (a *AnthropicProvider) Name() string            { return "anthropic" }
func (a *AnthropicProvider) SupportsStreaming() bool { return true }
func (a *AnthropicProvider) SupportsTools() bool     { return true }

func (a *AnthropicProvider) Stream(ctx context.Context, messages []Message, tools []Tool) (<-chan Chunk, error) {
	// Separar system prompt del historial
	var systemPrompt string
	var chatMsgs []anthropic.MessageParam

	for _, m := range messages {
		if m.Role == "system" {
			systemPrompt = m.Content
			continue
		}
		chatMsgs = append(chatMsgs, toAnthropicParam(m))
	}

	req := anthropic.MessageNewParams{
		Model:     anthropic.F(a.model),
		MaxTokens: anthropic.Int(int64(a.maxTokens)),
		Messages:  anthropic.F(chatMsgs),
	}

	if systemPrompt != "" {
		req.System = anthropic.F([]anthropic.TextBlockParam{
			{Type: "text", Text: anthropic.F(systemPrompt)},
		})
	}

	if len(tools) > 0 {
		req.Tools = anthropic.F(toAnthropicTools(tools))
	}

	stream := a.client.Messages.NewStreaming(ctx, req)
	ch := make(chan Chunk, 64)

	go func() {
		defer close(ch)

		var toolCallBuffer map[int]*ToolCall = make(map[int]*ToolCall)
		var toolCallIndex int

		for stream.Next() {
			event := stream.Current()

			switch e := event.(type) {
			case anthropic.ContentBlockDeltaEvent:
				if e.Delta.Text != "" {
					ch <- Chunk{Content: e.Delta.Text}
				}
				if e.Delta.InputJSON != "" {
					// Acumular JSON de tool call
					if tc, exists := toolCallBuffer[toolCallIndex]; exists {
						existing := string(tc.Args)
						tc.Args = json.RawMessage(existing + e.Delta.InputJSON)
					}
				}
			case anthropic.ContentBlockStartEvent:
				if e.ContentBlock.Type == "tool_use" {
					toolCallBuffer[toolCallIndex] = &ToolCall{
						ID:   e.ContentBlock.ID,
						Name: e.ContentBlock.Name,
						Args: json.RawMessage(""),
					}
				}
				toolCallIndex++
			}
		}

		// Enviar tool calls acumulados
		var allToolCalls []ToolCall
		for _, tc := range toolCallBuffer {
			if tc.Name != "" {
				allToolCalls = append(allToolCalls, *tc)
			}
		}
		if len(allToolCalls) > 0 {
			ch <- Chunk{ToolCalls: allToolCalls}
		}

		ch <- Chunk{Done: true}

		if err := stream.Err(); err != nil {
			ch <- Chunk{Err: err}
		}
	}()

	return ch, nil
}

func (a *AnthropicProvider) Chat(ctx context.Context, messages []Message, tools []Tool) (*Response, error) {
	// Implementación sincrónica similar
	return nil, fmt.Errorf("no implementado")
}

func toAnthropicParam(m Message) anthropic.MessageParam {
	role := m.Role
	if role == "system" {
		role = "user" // Anthropic maneja system aparte
	}

	if m.ToolName != "" || m.Role == "tool" {
		// Respuesta de herramienta
		return anthropic.NewToolResultMessage(m.ToolID, m.Content)
	}

	if len(m.ToolCalls) > 0 {
		var blocks []anthropic.ContentBlockParamUnion
		if m.Content != "" {
			blocks = append(blocks, anthropic.NewTextBlock(m.Content))
		}
		for _, tc := range m.ToolCalls {
			blocks = append(blocks, anthropic.NewToolUseBlockParam(tc.ID, tc.Name, tc.Args))
		}
		return anthropic.NewAssistantMessage(blocks...)
	}

	return anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content))
}

func toAnthropicTools(tools []Tool) []anthropic.ToolParam {
	var out []anthropic.ToolParam
	for _, t := range tools {
		out = append(out, anthropic.ToolParam{
			Name:        anthropic.F(t.Name),
			Description: anthropic.F(t.Description),
			InputSchema: anthropic.F(t.Schema),
		})
	}
	return out
}

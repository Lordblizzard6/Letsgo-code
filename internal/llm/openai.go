package llm

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type OpenAIProvider struct {
	client       *openai.Client
	model        string
	maxTokens    int
	temperature  float32
	isOpenRouter bool
}

func NewOpenAI(apiKey, model, baseURL string) (*OpenAIProvider, error) {
	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}

	isOpenRouter := false
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
		isOpenRouter = strings.Contains(baseURL, "openrouter")
	}

	if isOpenRouter {
		opts = append(opts, option.WithMiddleware(func(req openai.Request) (openai.Response, error) {
			req.Header.Set("HTTP-Referer", "https://mycli.dev")
			req.Header.Set("X-Title", "MyCLI")
			return req.Next(req)
		}))
	}

	client := openai.NewClient(opts...)

	return &OpenAIProvider{
		client:       client,
		model:        model,
		maxTokens:    4096,
		temperature:  0.7,
		isOpenRouter: isOpenRouter,
	}, nil
}

func (o *OpenAIProvider) Name() string {
	if o.isOpenRouter {
		return "openrouter"
	}
	return "openai"
}

func (o *OpenAIProvider) SupportsStreaming() bool { return true }
func (o *OpenAIProvider) SupportsTools() bool     { return true }

func (o *OpenAIProvider) Stream(ctx context.Context, messages []Message, tools []Tool) (<-chan Chunk, error) {
	params := openai.ChatCompletionNewParams{
		Model:       openai.F(o.model),
		MaxTokens:   openai.Int(o.maxTokens),
		Temperature: openai.Float(float64(o.temperature)),
		Stream:      openai.F(true),
		Messages:    toOpenAIMessages(messages),
	}

	if len(tools) > 0 {
		params.Tools = toOpenAITools(tools)
	}

	stream := o.client.Chat.Completions.NewStreaming(ctx, params)
	ch := make(chan Chunk, 64)

	go func() {
		defer close(ch)

		var toolCallBuffer map[int]*ToolCall = make(map[int]*ToolCall)

		for stream.Next() {
			event := stream.Current()

			for _, choice := range event.Choices {
				delta := choice.Delta

				// Contenido de texto
				if delta.Content != nil {
					ch <- Chunk{Content: *delta.Content}
				}

				// Tool calls (pueden venir fragmentados)
				for _, tc := range delta.ToolCalls {
					idx := int(tc.Index)
					if _, exists := toolCallBuffer[idx]; !exists {
						toolCallBuffer[idx] = &ToolCall{}
					}

					if tc.ID != nil {
						toolCallBuffer[idx].ID = *tc.ID
					}
					if tc.Function.Name != nil {
						toolCallBuffer[idx].Name = *tc.Function.Name
					}
					if tc.Function.Arguments != nil {
						existing := string(toolCallBuffer[idx].Args)
						toolCallBuffer[idx].Args = json.RawMessage(existing + *tc.Function.Arguments)
					}
				}
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

func (o *OpenAIProvider) Chat(ctx context.Context, messages []Message, tools []Tool) (*Response, error) {
	params := openai.ChatCompletionNewParams{
		Model:       openai.F(o.model),
		MaxTokens:   openai.Int(o.maxTokens),
		Temperature: openai.Float(float64(o.temperature)),
		Messages:    toOpenAIMessages(messages),
	}

	if len(tools) > 0 {
		params.Tools = toOpenAITools(tools)
	}

	resp, err := o.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, err
	}

	var toolCalls []ToolCall
	for _, tc := range resp.Choices[0].Message.ToolCalls {
		toolCalls = append(toolCalls, ToolCall{
			ID:   tc.ID,
			Name: tc.Function.Name,
			Args: json.RawMessage(tc.Function.Arguments),
		})
	}

	return &Response{
		Content:   resp.Choices[0].Message.Content,
		ToolCalls: toolCalls,
		Usage: UsageInfo{
			PromptTokens:     int(resp.Usage.PromptTokens),
			CompletionTokens: int(resp.Usage.CompletionTokens),
			TotalTokens:      int(resp.Usage.TotalTokens),
		},
	}, nil
}

// Helpers de conversión
func toOpenAIMessages(messages []Message) []openai.ChatCompletionMessageParamUnion {
	var out []openai.ChatCompletionMessageParamUnion

	for _, m := range messages {
		switch m.Role {
		case "system":
			out = append(out, openai.SystemMessage(m.Content))
		case "user":
			if m.ToolName != "" {
				// Respuesta de herramienta
				out = append(out, openai.ToolMessage(m.Content, m.ToolID))
			} else {
				out = append(out, openai.UserMessage(m.Content))
			}
		case "assistant":
			if len(m.ToolCalls) > 0 {
				toolCalls := make([]openai.ChatCompletionMessageToolCallParam, len(m.ToolCalls))
				for i, tc := range m.ToolCalls {
					toolCalls[i] = openai.ChatCompletionMessageToolCallParam{
						ID:   tc.ID,
						Type: "function",
						Function: openai.ChatCompletionMessageToolCallFunctionParam{
							Name:      tc.Name,
							Arguments: string(tc.Args),
						},
					}
				}
				out = append(out, openai.AssistantMessage(openai.F(toolCalls)))
			} else {
				out = append(out, openai.AssistantMessage(m.Content))
			}
		case "tool":
			out = append(out, openai.ToolMessage(m.Content, m.ToolID))
		}
	}

	return out
}

func toOpenAITools(tools []Tool) []openai.ChatCompletionToolParam {
	var out []openai.ChatCompletionToolParam
	for _, t := range tools {
		out = append(out, openai.ChatCompletionToolParam{
			Type: "function",
			Function: openai.FunctionDefinitionParam{
				Name:        openai.F(t.Name),
				Description: openai.F(t.Description),
				Parameters:  json.RawMessage(t.Schema),
			},
		})
	}
	return out
}

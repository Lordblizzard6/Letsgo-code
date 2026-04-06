package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiProvider struct {
	client      *genai.Client
	model       string
	maxTokens   int
	temperature float32
}

func NewGemini(apiKey, model string) (*GeminiProvider, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("gemini client: %w", err)
	}

	return &GeminiProvider{
		client:      client,
		model:       model,
		maxTokens:   4096,
		temperature: 0.7,
	}, nil
}

func (g *GeminiProvider) Name() string            { return "gemini" }
func (g *GeminiProvider) SupportsStreaming() bool { return true }
func (g *GeminiProvider) SupportsTools() bool     { return true }

func (g *GeminiProvider) Stream(ctx context.Context, messages []Message, tools []Tool) (<-chan Chunk, error) {
	model := g.client.GenerativeModel(g.model)
	model.SetMaxOutputTokens(int32(g.maxTokens))
	model.SetTemperature(g.temperature)

	// Separar system prompt
	var systemPrompt string
	var chatMsgs []*genai.Content

	for _, m := range messages {
		if m.Role == "system" {
			systemPrompt = m.Content
			continue
		}
		content := toGeminiContent(m)
		chatMsgs = append(chatMsgs, content)
	}

	if systemPrompt != "" {
		model.SystemInstruction = genai.NewText(systemPrompt)
	}

	if len(tools) > 0 {
		model.Tools = []*genai.Tool{{FunctionDeclarations: toGeminiFuncs(tools)}}
	}

	// Desactivar safety settings para CLI de código
	model.SafetySettings = []*genai.SafetySetting{
		{Category: genai.HarmCategoryHarassment, Threshold: genai.HarmBlockOnlyHigh},
		{Category: genai.HarmCategoryHateSpeech, Threshold: genai.HarmBlockOnlyHigh},
		{Category: genai.HarmCategorySexuallyExplicit, Threshold: genai.HarmBlockOnlyHigh},
		{Category: genai.HarmCategoryDangerousContent, Threshold: genai.HarmBlockOnlyHigh},
	}

	iter := model.GenerateContentStream(ctx, chatMsgs...)
	ch := make(chan Chunk, 64)

	go func() {
		defer close(ch)

		var toolCallBuffer []ToolCall

		for {
			resp, err := iter.Next()
			if err != nil {
				if err.Error() != "EOF" {
					ch <- Chunk{Err: err}
				}
				break
			}

			for _, cand := range resp.Candidates {
				if cand.Content == nil {
					continue
				}
				for _, part := range cand.Content.Parts {
					switch p := part.(type) {
					case genai.Text:
						ch <- Chunk{Content: string(p)}
					case genai.FunctionCall:
						argsJSON, _ := json.Marshal(p.Args)
						toolCallBuffer = append(toolCallBuffer, ToolCall{
							ID:   p.ID,
							Name: p.Name,
							Args: argsJSON,
						})
					}
				}
			}
		}

		if len(toolCallBuffer) > 0 {
			ch <- Chunk{ToolCalls: toolCallBuffer}
		}
		ch <- Chunk{Done: true}
	}()

	return ch, nil
}

func (g *GeminiProvider) Chat(ctx context.Context, messages []Message, tools []Tool) (*Response, error) {
	return nil, fmt.Errorf("no implementado")
}

func toGeminiContent(m Message) *genai.Content {
	role := m.Role
	if role == "assistant" {
		role = "model"
	}
	if role == "system" {
		role = "user"
	}

	content := &genai.Content{Role: role}

	if m.ToolName != "" || m.Role == "tool" {
		content.Parts = []genai.Part{
			genai.FunctionResponse{
				Name:     m.ToolName,
				Response: map[string]interface{}{"result": m.Content},
			},
		}
	} else if len(m.ToolCalls) > 0 {
		var parts []genai.Part
		if m.Content != "" {
			parts = append(parts, genai.Text(m.Content))
		}
		for _, tc := range m.ToolCalls {
			var args map[string]interface{}
			json.Unmarshal(tc.Args, &args)
			parts = append(parts, genai.FunctionCall{
				ID:   tc.ID,
				Name: tc.Name,
				Args: args,
			})
		}
		content.Parts = parts
	} else {
		content.Parts = []genai.Part{genai.Text(m.Content)}
	}

	return content
}

func toGeminiFuncs(tools []Tool) []*genai.FunctionDeclaration {
	var out []*genai.FunctionDeclaration
	for _, t := range tools {
		var schema genai.Schema
		json.Unmarshal(t.Schema, &schema)
		out = append(out, &genai.FunctionDeclaration{
			Name:        t.Name,
			Description: t.Description,
			Parameters:  &schema,
		})
	}
	return out
}

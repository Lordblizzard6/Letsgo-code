package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Provider constants
const (
	ProviderAnthropic  = "anthropic"
	ProviderOpenAI     = "openai"
	ProviderGroq       = "groq"
	ProviderOllama     = "ollama"
	ProviderOpenRouter = "openrouter"
)

type Client struct {
	APIKey   string
	BaseURL  string
	Provider string
}

const defaultAnthropicVersion = "2023-06-01"

func NewClient(apiKey string, baseURL string) *Client {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1/messages"
	}

	// Detect provider from baseURL
	provider := detectProvider(baseURL)

	// Adjust URL for Ollama if needed
	if provider == ProviderOllama && !strings.Contains(baseURL, "/api/chat") {
		// Ollama chat endpoint
		baseURL = strings.TrimSuffix(baseURL, "/") + "/api/chat"
	}

	return &Client{
		APIKey:   apiKey,
		BaseURL:  baseURL,
		Provider: provider,
	}
}

func detectProvider(baseURL string) string {
	lowerURL := strings.ToLower(baseURL)

	switch {
	case strings.Contains(lowerURL, "anthropic.com"):
		return ProviderAnthropic
	case strings.Contains(lowerURL, "api.openai.com"):
		return ProviderOpenAI
	case strings.Contains(lowerURL, "api.groq.com"):
		return ProviderGroq
	case strings.Contains(lowerURL, "ollama") || strings.Contains(lowerURL, "localhost:11434"):
		return ProviderOllama
	case strings.Contains(lowerURL, "openrouter.ai"):
		return ProviderOpenRouter
	default:
		// Default to OpenAI format for unknown providers
		return ProviderOpenAI
	}
}

// StreamRequestWithContext performs streaming request with context support and retry logic
func (c *Client) StreamRequestWithContext(ctx context.Context, req Request, onDelta func(string), onToolUse func(ToolUse), onToolInput func(string, string), onUsage func(int, int)) error {
	const maxRetries = 3
	const baseDelay = 1 * time.Second

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			delay := baseDelay * time.Duration(1<<attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		err := c.streamRequestInternal(ctx, req, onDelta, onToolUse, onToolInput, onUsage)
		if err == nil {
			return nil
		}

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Don't retry on 4xx errors (client errors)
		if strings.Contains(err.Error(), "API error (4") {
			return err
		}

		lastErr = err
	}

	return fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// streamRequestInternal is the actual request implementation
func (c *Client) streamRequestInternal(ctx context.Context, req Request, onDelta func(string), onToolUse func(ToolUse), onToolInput func(string, string), onUsage func(int, int)) error {
	// Resolve model names for specific providers
	model := req.Model
	if c.Provider == ProviderOpenRouter {
		model = ResolveOpenRouterModel(req.Model)
	} else if c.Provider == ProviderGroq {
		model = ResolveGroqModel(req.Model)
	}

	// NORMALIZATION FOR PROVIDERS
	var finalReq interface{}

	if !strings.Contains(c.BaseURL, "anthropic.com") {
		type OaiMessage struct {
			Role    string      `json:"role"`
			Content interface{} `json:"content"`
		}
		type OaiRequest struct {
			Model     string       `json:"model"`
			Messages  []OaiMessage `json:"messages"`
			MaxTokens int          `json:"max_tokens"`
			Stream    bool         `json:"stream"`
			Tools     interface{}  `json:"tools,omitempty"`
		}

		oaiMessages := []OaiMessage{}
		if req.System != "" {
			oaiMessages = append(oaiMessages, OaiMessage{Role: "system", Content: req.System})
		}
		for _, m := range req.Messages {
			role := m.Role
			var content interface{}

			switch c := m.Content.(type) {
			case string:
				content = c
			case []ContentBlock:
				// OpenAI uses a different structure for tool calls/results
				// For simple text, we can still use a string or array of parts
				text := ""
				var toolCalls []interface{}
				for _, block := range c {
					if block.Type == "text" {
						text += block.Text
					} else if block.Type == "tool_use" && block.ToolUse != nil {
						// In OAI, tool calls are a separate field in the assistant message
						// but here we are simplifying to match the OAI API expectations
						// for history when it's already a ToolCall object.
						argsMap := make(map[string]interface{})
						if argsStr, ok := block.ToolUse.Input.(string); ok {
							json.Unmarshal([]byte(argsStr), &argsMap)
						}
						toolCalls = append(toolCalls, map[string]interface{}{
							"id":   block.ToolUse.ID,
							"type": "function",
							"function": map[string]interface{}{
								"name":      block.ToolUse.Name,
								"arguments": block.ToolUse.Input,
							},
						})
					} else if block.Type == "tool_result" && block.ToolResult != nil {
						// Tool results in OAI are separate messages with role "tool"
						oaiMessages = append(oaiMessages, OaiMessage{
							Role:    "tool",
							Content: block.ToolResult.Content,
						})
						// Since OAI requires a tool_call_id, we'd need to handle that.
						// This is a simplification.
						continue
					}
				}
				if text != "" {
					content = text
				}
			default:
				content = fmt.Sprintf("%v", m.Content)
			}

			if content != nil {
				oaiMessages = append(oaiMessages, OaiMessage{Role: role, Content: content})
			}
		}

		oaiTools := []interface{}{}
		for _, t := range req.Tools {
			oaiTools = append(oaiTools, map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"name":        t.Name,
					"description": t.Description,
					"parameters":  t.InputSchema,
				},
			})
		}

		finalReq = OaiRequest{
			Model:     model,
			Messages:  oaiMessages,
			MaxTokens: req.MaxTokens,
			Stream:    req.Stream,
			Tools:     oaiTools,
		}
	} else {
		finalReq = req
	}

	jsonData, err := json.Marshal(finalReq)
	if err != nil {
		return err
	}

	// Build request URL
	requestURL := c.BaseURL

	httpReq, err := http.NewRequestWithContext(ctx, "POST", requestURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Ollama doesn't use Authorization header, it's local
	if c.Provider == ProviderOllama {
		// Ollama no auth needed for local instances
	} else if c.Provider == ProviderAnthropic {
		httpReq.Header.Set("x-api-key", c.APIKey)
		httpReq.Header.Set("anthropic-version", defaultAnthropicVersion)
	} else {
		// OpenAI, Groq, OpenRouter all use Bearer auth
		httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
		// OpenRouter specific headers
		if c.Provider == ProviderOpenRouter {
			httpReq.Header.Set("HTTP-Referer", "https://claude-code-go.dev")
			httpReq.Header.Set("X-Title", "Claude Code Go")
		}
	}

	client := &http.Client{
		Timeout: 60 * time.Second, // Reducido de 120s a 60s
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var currentToolID string
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for context cancellation in the streaming loop
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Ollama usa NDJSON (líneas JSON completas, no SSE)
		if c.Provider == ProviderOllama {
			var ollamaResp OllamaStreamResponse
			if err := json.Unmarshal([]byte(line), &ollamaResp); err == nil {
				if ollamaResp.Message.Content != "" {
					onDelta(ollamaResp.Message.Content)
				}
				if ollamaResp.Done {
					// Estimate token usage for Ollama (no native token count)
					onUsage(0, 0)
					return nil
				}
			}
			continue
		}

		// SSE format (OpenAI, Anthropic, etc.)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			return nil
		}

		// OpenAI / Groq / OpenRouter Usage handling
		if !strings.Contains(c.BaseURL, "anthropic.com") {
			var oai OpenAIStreamResponse
			if err := json.Unmarshal([]byte(data), &oai); err == nil && len(oai.Choices) > 0 {
				// Check for API errors in stream
				if oai.Error != nil {
					return fmt.Errorf("API error: %s (type: %s, code: %s)", oai.Error.Message, oai.Error.Type, oai.Error.Code)
				}

				choice := oai.Choices[0]

				// Check finish_reason - stream ended
				if choice.FinishReason != "" && choice.FinishReason != "null" {
					// Stream finished normally
					return nil
				}

				if content := choice.Delta.Content; content != "" {
					onDelta(content)
					continue
				}
				if len(choice.Delta.ToolCalls) > 0 {
					tc := choice.Delta.ToolCalls[0]
					// Handle case where tool call ID might be null in some chunks
					if tc.ID != "" {
						currentToolID = tc.ID
					}
					if tc.Function.Name != "" {
						onToolUse(ToolUse{ID: currentToolID, Name: tc.Function.Name})
					}
					if tc.Function.Arguments != "" {
						onToolInput(currentToolID, tc.Function.Arguments)
					}
					continue
				}
				// Empty delta but no finish_reason yet, continue reading
				continue
			} else if err != nil {
				// Log parsing errors for debugging
				// fmt.Printf("Debug: JSON parse error: %v\n", err)
				continue
			}
		}

		var event map[string]interface{}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}

		eventType, _ := event["type"].(string)
		switch eventType {
		case "message_start":
			if msg, ok := event["message"].(map[string]interface{}); ok {
				if usage, ok := msg["usage"].(map[string]interface{}); ok {
					inputTokens := int(usage["input_tokens"].(float64))
					onUsage(inputTokens, 0)
				}
			}
		case "content_block_start":
			block := event["content_block"].(map[string]interface{})
			if block["type"] == "tool_use" {
				var tu ToolUse
				tu.ID = block["id"].(string)
				tu.Name = block["name"].(string)
				tu.Input = make(map[string]interface{})
				currentToolID = tu.ID
				onToolUse(tu)
			}
		case "content_block_delta":
			delta := event["delta"].(map[string]interface{})
			if delta["type"] == "text_delta" {
				if text, ok := delta["text"].(string); ok {
					onDelta(text)
				}
			} else if delta["type"] == "input_json_delta" {
				if partial, ok := delta["partial_json"].(string); ok {
					onToolInput(currentToolID, partial)
				}
			}
		case "message_delta":
			if usage, ok := event["usage"].(map[string]interface{}); ok {
				outputTokens := int(usage["output_tokens"].(float64))
				onUsage(0, outputTokens)
			}
		case "message_stop":
			return nil
		}
	}
	return nil
}

// StreamRequest performs streaming request (backward compatibility - uses background context)
func (c *Client) StreamRequest(req Request, onDelta func(string), onToolUse func(ToolUse), onToolInput func(string, string), onUsage func(int, int)) error {
	return c.StreamRequestWithContext(context.Background(), req, onDelta, onToolUse, onToolInput, onUsage)
}

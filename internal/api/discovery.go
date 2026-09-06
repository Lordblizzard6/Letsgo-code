package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// DiscoveredModel represents an AI model discovered dynamically or from a catalog.
type DiscoveredModel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Provider    string `json:"provider"`
	Description string `json:"description,omitempty"`
	ContextSize int    `json:"context_size,omitempty"`
}

// DiscoverModels queries provider endpoints dynamically to discover models.
// If the network call fails or credentials are not yet set, it falls back to a curated catalog.
func DiscoverModels(provider, apiKey, baseURL string) ([]DiscoveredModel, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	client := &http.Client{Timeout: 5 * time.Second}

	switch provider {
	case "ollama":
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		endpoint := strings.TrimRight(baseURL, "/") + "/api/tags"
		req, err := http.NewRequestWithContext(context.Background(), "GET", endpoint, nil)
		if err == nil {
			resp, err := client.Do(req)
			if err == nil && resp.StatusCode == http.StatusOK {
				defer resp.Body.Close()
				var payload struct {
					Models []struct {
						Name string `json:"name"`
					} `json:"models"`
				}
				if json.NewDecoder(resp.Body).Decode(&payload) == nil && len(payload.Models) > 0 {
					var out []DiscoveredModel
					for _, m := range payload.Models {
						out = append(out, DiscoveredModel{
							ID:       m.Name,
							Name:     m.Name,
							Provider: "ollama",
						})
					}
					return out, nil
				}
			}
		}
		// Fallback for Ollama
		return []DiscoveredModel{
			{ID: "llama3.2", Name: "Llama 3.2", Provider: "ollama"},
			{ID: "qwen2.5-coder", Name: "Qwen 2.5 Coder", Provider: "ollama"},
			{ID: "deepseek-r1", Name: "DeepSeek R1", Provider: "ollama"},
			{ID: "mistral", Name: "Mistral", Provider: "ollama"},
		}, nil

	case "groq":
		if apiKey != "" {
			req, err := http.NewRequestWithContext(context.Background(), "GET", "https://api.groq.com/openai/v1/models", nil)
			if err == nil {
				req.Header.Set("Authorization", "Bearer "+apiKey)
				resp, err := client.Do(req)
				if err == nil && resp.StatusCode == http.StatusOK {
					defer resp.Body.Close()
					var payload struct {
						Data []struct {
							ID string `json:"id"`
						} `json:"data"`
					}
					if json.NewDecoder(resp.Body).Decode(&payload) == nil && len(payload.Data) > 0 {
						var out []DiscoveredModel
						for _, m := range payload.Data {
							out = append(out, DiscoveredModel{
								ID:       m.ID,
								Name:     m.ID,
								Provider: "groq",
							})
						}
						return out, nil
					}
				}
			}
		}
		// Fallback for Groq
		var out []DiscoveredModel
		for _, m := range GroqModelCatalog {
			out = append(out, DiscoveredModel{
				ID:          m.ID,
				Name:        m.Name,
				Provider:    "groq",
				Description: m.Description,
				ContextSize: m.Context,
			})
		}
		return out, nil

	case "openrouter":
		if apiKey != "" {
			req, err := http.NewRequestWithContext(context.Background(), "GET", "https://openrouter.ai/api/v1/models", nil)
			if err == nil {
				req.Header.Set("Authorization", "Bearer "+apiKey)
				resp, err := client.Do(req)
				if err == nil && resp.StatusCode == http.StatusOK {
					defer resp.Body.Close()
					var payload struct {
						Data []struct {
							ID          string `json:"id"`
							Name        string `json:"name"`
							ContextLen  int    `json:"context_length"`
							Description string `json:"description"`
						} `json:"data"`
					}
					if json.NewDecoder(resp.Body).Decode(&payload) == nil && len(payload.Data) > 0 {
						var out []DiscoveredModel
						for _, m := range payload.Data {
							out = append(out, DiscoveredModel{
								ID:          m.ID,
								Name:        m.Name,
								Provider:    "openrouter",
								Description: m.Description,
								ContextSize: m.ContextLen,
							})
						}
						return out, nil
					}
				}
			}
		}
		// Fallback for OpenRouter
		var out []DiscoveredModel
		for _, m := range OpenRouterModelCatalog {
			out = append(out, DiscoveredModel{
				ID:          m.ID,
				Name:        m.Name,
				Provider:    "openrouter",
				Description: m.Description,
				ContextSize: m.Context,
			})
		}
		return out, nil

	case "openai":
		if apiKey != "" {
			req, err := http.NewRequestWithContext(context.Background(), "GET", "https://api.openai.com/v1/models", nil)
			if err == nil {
				req.Header.Set("Authorization", "Bearer "+apiKey)
				resp, err := client.Do(req)
				if err == nil && resp.StatusCode == http.StatusOK {
					defer resp.Body.Close()
					var payload struct {
						Data []struct {
							ID string `json:"id"`
						} `json:"data"`
					}
					if json.NewDecoder(resp.Body).Decode(&payload) == nil && len(payload.Data) > 0 {
						var out []DiscoveredModel
						for _, m := range payload.Data {
							// Filter out audio, whisper, tts, embeddings, moderation to show chat/reasoning models
							if strings.HasPrefix(m.ID, "gpt-") || strings.HasPrefix(m.ID, "o1") || strings.HasPrefix(m.ID, "o3") || strings.HasPrefix(m.ID, "chatgpt-") {
								out = append(out, DiscoveredModel{
									ID:       m.ID,
									Name:     m.ID,
									Provider: "openai",
								})
							}
						}
						if len(out) > 0 {
							return out, nil
						}
					}
				}
			}
		}
		return []DiscoveredModel{
			{ID: "gpt-4o", Name: "GPT-4o", Provider: "openai"},
			{ID: "gpt-4o-mini", Name: "GPT-4o Mini", Provider: "openai"},
			{ID: "o1", Name: "o1", Provider: "openai"},
			{ID: "o3-mini", Name: "o3-mini", Provider: "openai"},
		}, nil

	case "gemini", "google":
		if apiKey != "" {
			endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", apiKey)
			req, err := http.NewRequestWithContext(context.Background(), "GET", endpoint, nil)
			if err == nil {
				resp, err := client.Do(req)
				if err == nil && resp.StatusCode == http.StatusOK {
					defer resp.Body.Close()
					var payload struct {
						Models []struct {
							Name        string `json:"name"`
							DisplayName string `json:"displayName"`
							Description string `json:"description"`
						} `json:"models"`
					}
					if json.NewDecoder(resp.Body).Decode(&payload) == nil && len(payload.Models) > 0 {
						var out []DiscoveredModel
						for _, m := range payload.Models {
							modelID := strings.TrimPrefix(m.Name, "models/")
							if strings.Contains(modelID, "gemini") {
								out = append(out, DiscoveredModel{
									ID:          modelID,
									Name:        m.DisplayName,
									Provider:    "gemini",
									Description: m.Description,
								})
							}
						}
						if len(out) > 0 {
							return out, nil
						}
					}
				}
			}
		}
		return []DiscoveredModel{
			{ID: "gemini-2.0-flash", Name: "Gemini 2.0 Flash", Provider: "gemini"},
			{ID: "gemini-1.5-pro", Name: "Gemini 1.5 Pro", Provider: "gemini"},
			{ID: "gemini-1.5-flash", Name: "Gemini 1.5 Flash", Provider: "gemini"},
		}, nil

	case "anthropic":
		return []DiscoveredModel{
			{ID: "claude-3-7-sonnet-20250219", Name: "Claude 3.7 Sonnet (Hybrid Reasoning)", Provider: "anthropic", ContextSize: 200000},
			{ID: "claude-3-5-sonnet-20241022", Name: "Claude 3.5 Sonnet v2", Provider: "anthropic", ContextSize: 200000},
			{ID: "claude-3-5-haiku-20241022", Name: "Claude 3.5 Haiku", Provider: "anthropic", ContextSize: 200000},
			{ID: "claude-3-opus-20240229", Name: "Claude 3 Opus", Provider: "anthropic", ContextSize: 200000},
		}, nil

	case "deepseek":
		return []DiscoveredModel{
			{ID: "deepseek-chat", Name: "DeepSeek V3", Provider: "deepseek", ContextSize: 64000},
			{ID: "deepseek-reasoner", Name: "DeepSeek R1", Provider: "deepseek", ContextSize: 64000},
		}, nil

	default:
		return nil, fmt.Errorf("unknown provider %q", provider)
	}
}

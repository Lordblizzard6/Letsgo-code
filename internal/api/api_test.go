package api

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-api-key", "https://api.anthropic.com/v1/messages")

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.APIKey != "test-api-key" {
		t.Errorf("Expected APIKey 'test-api-key', got '%s'", client.APIKey)
	}

	if client.BaseURL != "https://api.anthropic.com/v1/messages" {
		t.Errorf("Expected BaseURL 'https://api.anthropic.com/v1/messages', got '%s'", client.BaseURL)
	}

	if client.Provider != ProviderAnthropic {
		t.Errorf("Expected Provider '%s', got '%s'", ProviderAnthropic, client.Provider)
	}
}

func TestNewClient_DefaultURL(t *testing.T) {
	client := NewClient("test-key", "")

	if client.BaseURL != "https://api.anthropic.com/v1/messages" {
		t.Errorf("Expected default URL, got '%s'", client.BaseURL)
	}
}

func TestNewClient_Groq(t *testing.T) {
	client := NewClient("test-key", "https://api.groq.com/openai/v1")

	if client.Provider != ProviderGroq {
		t.Errorf("Expected Provider '%s' for Groq URL, got '%s'", ProviderGroq, client.Provider)
	}
}

func TestNewClient_OpenAI(t *testing.T) {
	client := NewClient("test-key", "https://api.openai.com/v1")

	if client.Provider != ProviderOpenAI {
		t.Errorf("Expected Provider '%s' for OpenAI URL, got '%s'", ProviderOpenAI, client.Provider)
	}
}

func TestNewClient_OpenRouter(t *testing.T) {
	client := NewClient("test-key", "https://openrouter.ai/api/v1")

	if client.Provider != ProviderOpenRouter {
		t.Errorf("Expected Provider '%s' for OpenRouter URL, got '%s'", ProviderOpenRouter, client.Provider)
	}
}

func TestResolveOpenRouterModel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"claude", "anthropic/claude-3.5-sonnet"},
		{"gpt4", "openai/gpt-4o"},
		{"llama", "meta-llama/llama-3.3-70b-instruct"},
		{"anthropic/claude-3-opus", "anthropic/claude-3-opus"}, // Already full ID
	}

	for _, test := range tests {
		result := ResolveOpenRouterModel(test.input)
		if result != test.expected {
			t.Errorf("ResolveOpenRouterModel(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestResolveGroqModel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"llama", "llama-3.3-70b-versatile"},
		{"llama-3.3-70b", "llama-3.3-70b-versatile"},
		{"llama-3.1-8b", "llama-3.1-8b-instant"},
		{"custom-model", "custom-model"}, // Passthrough
	}

	for _, test := range tests {
		result := ResolveGroqModel(test.input)
		if result != test.expected {
			t.Errorf("ResolveGroqModel(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestDiscoverModels(t *testing.T) {
	providers := []string{"ollama", "groq", "openrouter", "openai", "gemini", "anthropic", "deepseek"}
	for _, p := range providers {
		models, err := DiscoverModels(p, "", "")
		if err != nil {
			t.Fatalf("DiscoverModels(%q) failed: %v", p, err)
		}
		if len(models) == 0 {
			t.Errorf("DiscoverModels(%q) returned empty list", p)
		}
	}
}

func TestResolveOpenRouterModel_Gemini(t *testing.T) {
	m := ResolveOpenRouterModel("gemini")
	if m != "google/gemini-2.0-flash-001" {
		t.Errorf("Expected google/gemini-2.0-flash-001, got %s", m)
	}
}

func TestSystemPrompt_ProjectRules(t *testing.T) {
	prompt := GetBuildSystemPrompt("claude-3-5-sonnet", ".")
	if len(prompt) == 0 {
		t.Fatal("System prompt should not be empty")
	}
}


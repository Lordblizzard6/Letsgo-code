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

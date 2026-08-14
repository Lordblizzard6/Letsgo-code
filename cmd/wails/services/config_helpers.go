package services

import (
	"github.com/user/go-claude-code/internal/config"
)

// pascalKey maps the PascalCase field names produced by the Wails bindings
// (Config.APIKey, Config.AnthropicAPIKey, ...) to the snake_case keys of the
// core contract (frontend-contract §3 SaveConfig).
var pascalKey = map[string]string{
	"APIKey":           "api_key",
	"AnthropicAPIKey":  "anthropic_api_key",
	"OpenAIAPIKey":     "openai_api_key",
	"GroqAPIKey":       "groq_api_key",
	"OpenRouterAPIKey": "openrouter_api_key",
	"Model":            "model",
	"BaseURL":          "base_url",
	"Shell":            "shell",
	"Verbose":          "verbose",
	"Temperature":      "temperature",
	"MaxTokens":        "max_tokens",
	"Stream":           "stream",
	"AutoApprove":      "auto_approve",
	"ThemeVariant":     "theme",
	"Statusline":       "statusline",
	"Rail":             "rail",
}

// applyConfig merges a JSON payload (as produced by the Wails bindings) over
// config.AppConfig. It maps only the well-known keys of the core contract so
// arbitrary payloads cannot corrupt the config file. Both the snake_case keys
// of the core contract and the PascalCase field names of the generated
// bindings are accepted.
func applyConfig(partial map[string]any) error {
	c := &config.AppConfig
	for key, val := range partial {
		if snake, ok := pascalKey[key]; ok {
			key = snake
		}
		switch key {
		case "api_key":
			c.APIKey = asString(val)
		case "anthropic_api_key":
			c.AnthropicAPIKey = asString(val)
		case "openai_api_key":
			c.OpenAIAPIKey = asString(val)
		case "groq_api_key":
			c.GroqAPIKey = asString(val)
		case "openrouter_api_key":
			c.OpenRouterAPIKey = asString(val)
		case "model":
			c.Model = asString(val)
		case "base_url":
			c.BaseURL = asString(val)
		case "shell":
			c.Shell = asString(val)
		case "verbose":
			c.Verbose = asBool(val)
		case "temperature":
			c.Temperature = asFloat(val)
		case "max_tokens":
			c.MaxTokens = asInt(val)
		case "stream":
			c.Stream = asBool(val)
		case "auto_approve":
			c.AutoApprove = asBoolMap(val)
		case "theme":
			c.ThemeVariant = asString(val)
		case "statusline":
			c.Statusline = asStatusline(val)
		case "rail":
			c.Rail = asRail(val)
		}
	}
	return nil
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func asBool(v any) bool {
	b, _ := v.(bool)
	return b
}

func asFloat(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	}
	return 0
}

func asInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	}
	return 0
}

func asBoolMap(v any) map[string]bool {
	out := make(map[string]bool)
	if m, ok := v.(map[string]any); ok {
		for k, item := range m {
			out[k] = asBool(item)
		}
	}
	return out
}

func asStatusline(v any) config.StatuslineConfig {
	out := config.AppConfig.Statusline
	if m, ok := v.(map[string]any); ok {
		if show, ok := m["show"]; ok {
			out.Show = asBool(show)
		}
		if fields, ok := m["fields"].([]any); ok {
			out.Fields = nil
			for _, f := range fields {
				out.Fields = append(out.Fields, asString(f))
			}
		}
	}
	return out
}

func asRail(v any) config.RailConfig {
	out := config.AppConfig.Rail
	if m, ok := v.(map[string]any); ok {
		if collapsed, ok := m["collapsed"]; ok {
			out.Collapsed = asBool(collapsed)
		}
	}
	return out
}
package api

// OpenRouterModels contiene el mapeo de nombres amigables a IDs completos
type OpenRouterModel struct {
	ID          string
	Name        string
	Provider    string
	Description string
	Pricing     string
	Context     int
}

// OpenRouterModelMap mapea nombres cortos a IDs completos
var OpenRouterModelMap = map[string]string{
	// Anthropic
	"claude":        "anthropic/claude-3.5-sonnet",
	"claude-sonnet": "anthropic/claude-3.5-sonnet",
	"claude-opus":   "anthropic/claude-3-opus",
	"claude-haiku":  "anthropic/claude-3-haiku",
	"claude-3.5":    "anthropic/claude-3.5-sonnet",

	// OpenAI
	"gpt4":        "openai/gpt-4o",
	"gpt-4":       "openai/gpt-4o",
	"gpt4o":       "openai/gpt-4o",
	"gpt-4o":      "openai/gpt-4o",
	"gpt4o-mini":  "openai/gpt-4o-mini",
	"gpt-4o-mini": "openai/gpt-4o-mini",

	// Meta
	"llama":     "meta-llama/llama-3.3-70b-instruct",
	"llama-3.3": "meta-llama/llama-3.3-70b-instruct",
	"llama-3.1": "meta-llama/llama-3.1-70b-instruct",
	"llama-70b": "meta-llama/llama-3.3-70b-instruct",
	"llama-8b":  "meta-llama/llama-3.1-8b-instruct",

	// Mistral
	"mistral":        "mistralai/mistral-large",
	"mistral-large":  "mistralai/mistral-large",
	"mistral-medium": "mistralai/mistral-medium",

	// DeepSeek - Actualizado
	"deepseek":       "deepseek/deepseek-chat-v3",
	"deepseek-chat":  "deepseek/deepseek-chat-v3",
	"deepseek-coder": "deepseek/deepseek-coder-v2",
	"deepseek-v3":    "deepseek/deepseek-chat-v3",
	"deepseek-r1":    "deepseek/deepseek-r1",

	// Qwen - Actualizado
	"qwen":         "qwen/qwen-2.5-72b-instruct",
	"qwen-2.5":     "qwen/qwen-2.5-72b-instruct",
	"qwen-2.5-72b": "qwen/qwen-2.5-72b-instruct",
	"qwen-3":       "qwen/qwen3-235b-a22b",
	"qwen-3-235b":  "qwen/qwen3-235b-a22b",

	// Nous
	"hermes":   "nousresearch/hermes-3-llama-3.1-70b",
	"hermes-3": "nousresearch/hermes-3-llama-3.1-70b",

	// Microsoft
	"wizard":   "microsoft/wizardlm-2-8x22b",
	"wizardlm": "microsoft/wizardlm-2-8x22b",
}

// ResolveOpenRouterModel resuelve un nombre de modelo a su ID completo
func ResolveOpenRouterModel(model string) string {
	if fullID, ok := OpenRouterModelMap[model]; ok {
		return fullID
	}
	// Si ya contiene '/', asumimos que es un ID completo
	if len(model) > 0 && model[0] != '/' && len(model) > 3 {
		for i := 0; i < len(model)-1; i++ {
			if model[i] == '/' {
				return model
			}
		}
	}
	// Default fallback
	return "anthropic/claude-3.5-sonnet"
}

// OpenRouterModelCatalog lista todos los modelos disponibles con metadata
// Actualizado: Abril 2026
var OpenRouterModelCatalog = []OpenRouterModel{
	{ID: "anthropic/claude-3.5-sonnet", Name: "Claude 3.5 Sonnet", Provider: "Anthropic", Description: "Balance de velocidad y calidad", Context: 200000},
	{ID: "anthropic/claude-3-opus", Name: "Claude 3 Opus", Provider: "Anthropic", Description: "Máxima calidad", Context: 200000},
	{ID: "anthropic/claude-3.7-sonnet", Name: "Claude 3.7 Sonnet", Provider: "Anthropic", Description: "Latest with extended thinking", Context: 200000},
	{ID: "openai/gpt-4o", Name: "GPT-4o", Provider: "OpenAI", Description: "Versátil y potente", Context: 128000},
	{ID: "openai/gpt-4o-mini", Name: "GPT-4o Mini", Provider: "OpenAI", Description: "Económico", Context: 128000},
	{ID: "openai/o3-mini", Name: "o3 Mini", Provider: "OpenAI", Description: "Reasoning model", Context: 200000},
	{ID: "meta-llama/llama-3.3-70b-instruct", Name: "Llama 3.3 70B", Provider: "Meta", Description: "Open source líder", Context: 128000},
	{ID: "meta-llama/llama-4-maverick", Name: "Llama 4 Maverick", Provider: "Meta", Description: "Latest multimodal", Context: 256000},
	{ID: "meta-llama/llama-4-scout", Name: "Llama 4 Scout", Provider: "Meta", Description: "Efficient multimodal", Context: 128000},
	{ID: "mistralai/mistral-large", Name: "Mistral Large", Provider: "Mistral", Description: "Potente", Context: 128000},
	{ID: "mistralai/mistral-small", Name: "Mistral Small", Provider: "Mistral", Description: "Rápido", Context: 32000},
	{ID: "deepseek/deepseek-chat-v3", Name: "DeepSeek V3", Provider: "DeepSeek", Description: "Chat avanzado", Context: 64000},
	{ID: "deepseek/deepseek-r1", Name: "DeepSeek R1", Provider: "DeepSeek", Description: "Razonamiento", Context: 64000},
	{ID: "qwen/qwen3-235b-a22b", Name: "Qwen 3 235B", Provider: "Alibaba", Description: "MoE de alto rendimiento", Context: 128000},
	{ID: "qwen/qwen-2.5-72b-instruct", Name: "Qwen 2.5 72B", Provider: "Alibaba", Description: "Fuerte rendimiento", Context: 128000},
}

// GetOpenRouterModelByProvider retorna modelos filtrados por proveedor
func GetOpenRouterModelByProvider(provider string) []OpenRouterModel {
	var result []OpenRouterModel
	for _, m := range OpenRouterModelCatalog {
		if m.Provider == provider {
			result = append(result, m)
		}
	}
	return result
}

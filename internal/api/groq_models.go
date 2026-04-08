package api

// GroqModels contiene el mapeo de nombres amigables a IDs completos
type GroqModel struct {
	ID          string
	Name        string
	Description string
	Speed       string // "ultra", "fast", "medium"
	Context     int
}

// GroqModelMap mapea nombres cortos a IDs completos
var GroqModelMap = map[string]string{
	// Llama 3.x series
	"llama":         "llama-3.3-70b-versatile",
	"llama-3.3":     "llama-3.3-70b-versatile",
	"llama-3.1":     "llama-3.1-8b-instant",
	"llama-3.1-8b":  "llama-3.1-8b-instant",
	"llama-3.3-70b": "llama-3.3-70b-versatile",
	"llama-70b":     "llama-3.3-70b-versatile",
	"llama-8b":      "llama-3.1-8b-instant",

	// Llama 4 (Preview)
	"llama-4-scout": "meta-llama/llama-4-scout-17b-16e-instruct",
	"llama-4":       "meta-llama/llama-4-scout-17b-16e-instruct",

	// GPT-OSS (OpenAI open weight models on Groq)
	"gpt-oss":      "openai/gpt-oss-120b",
	"gpt-oss-120b": "openai/gpt-oss-120b",
	"gpt-oss-20b":  "openai/gpt-oss-20b",

	// Gemma
	"gemma":      "gemma2-9b-it",
	"gemma-2":    "gemma2-9b-it",
	"gemma-2-9b": "gemma2-9b-it",

	// Qwen
	"qwen":       "qwen/qwen3-32b",
	"qwen-3":     "qwen/qwen3-32b",
	"qwen-3-32b": "qwen/qwen3-32b",
}

// ResolveGroqModel resuelve un nombre de modelo a su ID completo
func ResolveGroqModel(model string) string {
	if fullID, ok := GroqModelMap[model]; ok {
		return fullID
	}
	// Si contiene el formato esperado, asumimos que es válido
	return model
}

// GroqModelCatalog lista todos los modelos disponibles con metadata
// Actualizado: Abril 2026 - Eliminado Mixtral 8x7B (descontinuado)
var GroqModelCatalog = []GroqModel{
	{ID: "llama-3.3-70b-versatile", Name: "Llama 3.3 70B", Description: "Versátil y potente", Speed: "fast", Context: 128000},
	{ID: "llama-3.1-8b-instant", Name: "Llama 3.1 8B", Description: "Ultra rápido", Speed: "ultra", Context: 128000},
	{ID: "openai/gpt-oss-120b", Name: "GPT-OSS 120B", Description: "OpenAI open weight", Speed: "fast", Context: 128000},
	{ID: "openai/gpt-oss-20b", Name: "GPT-OSS 20B", Description: "Rápido y eficiente", Speed: "ultra", Context: 128000},
	{ID: "meta-llama/llama-4-scout-17b-16e-instruct", Name: "Llama 4 Scout", Description: "Preview - Multimodal", Speed: "ultra", Context: 128000},
	{ID: "gemma2-9b-it", Name: "Gemma 2 9B", Description: "Ligero y eficiente", Speed: "ultra", Context: 8192},
	{ID: "qwen/qwen3-32b", Name: "Qwen 3 32B", Description: "Fuerte en razonamiento", Speed: "fast", Context: 128000},
}

// GetGroqFastModels retorna modelos rápidos para tareas simples
func GetGroqFastModels() []GroqModel {
	var result []GroqModel
	for _, m := range GroqModelCatalog {
		if m.Speed == "ultra" || m.Speed == "fast" {
			result = append(result, m)
		}
	}
	return result
}

// GetGroqCodingModels retorna modelos recomendados para código
func GetGroqCodingModels() []GroqModel {
	codingModels := []string{
		"openai/gpt-oss-120b",
		"qwen/qwen3-32b",
		"llama-3.3-70b-versatile",
	}
	var result []GroqModel
	for _, id := range codingModels {
		for _, m := range GroqModelCatalog {
			if m.ID == id {
				result = append(result, m)
				break
			}
		}
	}
	return result
}

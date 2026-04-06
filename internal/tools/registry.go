package tools

import (
	"encoding/json"
	"fmt"

	"github.com/tuusuario/mycli/internal/llm"
)

type Registry struct {
	tools map[string]llm.Tool
}

func NewRegistry() *Registry {
	r := &Registry{
		tools: make(map[string]llm.Tool),
	}

	// Registrar herramientas por defecto
	r.RegisterFileSystemTools()
	r.RegisterShellTools()
	r.RegisterGitTools()

	return r
}

func (r *Registry) Register(tool llm.Tool) {
	r.tools[tool.Name] = tool
}

func (r *Registry) GetDefinitions() []llm.Tool {
	var out []llm.Tool
	for _, t := range r.tools {
		out = append(out, t)
	}
	return out
}

func (r *Registry) Execute(name string, args json.RawMessage) (string, error) {
	tool, exists := r.tools[name]
	if !exists {
		return "", fmt.Errorf("tool not found: %s", name)
	}
	return tool.Handler(args)
}

func (r *Registry) HasTool(name string) bool {
	_, exists := r.tools[name]
	return exists
}

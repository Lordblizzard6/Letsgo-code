package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tuusuario/mycli/internal/llm"
)

func (r *Registry) RegisterFileSystemTools() {
	r.Register(llm.Tool{
		Name:        "read_file",
		Description: "Lee el contenido de un archivo. Usa para entender código existente.",
		Schema:      json.RawMessage(`{"type":"object","properties":{"path":{"type":"string","description":"Ruta del archivo a leer"}},"required":["path"]}`),
		Handler: func(args json.RawMessage) (string, error) {
			var params struct {
				Path string `json:"path"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return "", err
			}

			// Seguridad: prevenir path traversal
			cleanPath := filepath.Clean(params.Path)
			if strings.HasPrefix(cleanPath, "..") {
				return "", fmt.Errorf("path traversal no permitido")
			}

			content, err := os.ReadFile(cleanPath)
			if err != nil {
				return "", err
			}

			return fmt.Sprintf("```\n%s\n```", string(content)), nil
		},
	})

	r.Register(llm.Tool{
		Name:        "write_file",
		Description: "Escribe contenido en un archivo. Crea el archivo si no existe.",
		Schema:      json.RawMessage(`{"type":"object","properties":{"path":{"type":"string","description":"Ruta del archivo"},"content":{"type":"string","description":"Contenido a escribir"}},"required":["path","content"]}`),
		Handler: func(args json.RawMessage) (string, error) {
			var params struct {
				Path    string `json:"path"`
				Content string `json:"content"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return "", err
			}

			cleanPath := filepath.Clean(params.Path)
			if strings.HasPrefix(cleanPath, "..") {
				return "", fmt.Errorf("path traversal no permitido")
			}

			// Crear directorios si no existen
			dir := filepath.Dir(cleanPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return "", err
			}

			if err := os.WriteFile(cleanPath, []byte(params.Content), 0644); err != nil {
				return "", err
			}

			return fmt.Sprintf("Archivo escrito: %s (%d bytes)", cleanPath, len(params.Content)), nil
		},
	})

	r.Register(llm.Tool{
		Name:        "list_dir",
		Description: "Lista el contenido de un directorio.",
		Schema:      json.RawMessage(`{"type":"object","properties":{"path":{"type":"string","description":"Ruta del directorio"}},"required":["path"]}`),
		Handler: func(args json.RawMessage) (string, error) {
			var params struct {
				Path string `json:"path"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return "", err
			}

			cleanPath := filepath.Clean(params.Path)
			entries, err := os.ReadDir(cleanPath)
			if err != nil {
				return "", err
			}

			var result strings.Builder
			for _, e := range entries {
				result.WriteString(fmt.Sprintf("%s %s\n",
					map[bool]string{true: "📁", false: "📄"}[e.IsDir()],
					e.Name()))
			}

			return result.String(), nil
		},
	})
}

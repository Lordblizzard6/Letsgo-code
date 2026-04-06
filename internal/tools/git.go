package tools

import (
	"encoding/json"
	"os/exec"
	"strconv"

	"github.com/tuusuario/mycli/internal/llm"
)

func (r *Registry) RegisterGitTools() {
	r.Register(llm.Tool{
		Name:        "git_status",
		Description: "Obtiene el estado del repositorio git.",
		Schema:      json.RawMessage(`{"type":"object","properties":{}}`),
		Handler: func(args json.RawMessage) (string, error) {
			cmd := exec.Command("git", "status")
			output, err := cmd.CombinedOutput()
			if err != nil {
				return "", err
			}
			return string(output), nil
		},
	})

	r.Register(llm.Tool{
		Name:        "git_diff",
		Description: "Obtiene los cambios no commitados.",
		Schema:      json.RawMessage(`{"type":"object","properties":{}}`),
		Handler: func(args json.RawMessage) (string, error) {
			cmd := exec.Command("git", "diff")
			output, err := cmd.CombinedOutput()
			if err != nil {
				return "", err
			}
			return string(output), nil
		},
	})

	r.Register(llm.Tool{
		Name:        "git_log",
		Description: "Obtiene el historial de commits.",
		Schema:      json.RawMessage(`{"type":"object","properties":{"limit":{"type":"integer","description":"Número de commits"}},"required":[]}`),
		Handler: func(args json.RawMessage) (string, error) {
			var params struct {
				Limit int `json:"limit,omitempty"`
			}
			json.Unmarshal(args, &params)
			if params.Limit == 0 {
				params.Limit = 10
			}

			cmd := exec.Command("git", "log", "--oneline", "-n", strconv.Itoa(params.Limit))
			output, err := cmd.CombinedOutput()
			if err != nil {
				return "", err
			}
			return string(output), nil
		},
	})
}

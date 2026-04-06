package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/tuusuario/mycli/internal/llm"
)

func (r *Registry) RegisterShellTools() {
	r.Register(llm.Tool{
		Name:        "run_command",
		Description: "Ejecuta un comando de shell. Usa para compilar, testear, git, etc.",
		Schema:      json.RawMessage(`{"type":"object","properties":{"command":{"type":"string","description":"Comando a ejecutar"},"cwd":{"type":"string","description":"Directorio de trabajo"}},"required":["command"]}`),
		Handler: func(args json.RawMessage) (string, error) {
			var params struct {
				Command string `json:"command"`
				Cwd     string `json:"cwd,omitempty"`
			}
			if err := json.Unmarshal(args, &params); err != nil {
				return "", err
			}

			// Lista blanca de comandos peligrosos
			dangerous := []string{"rm -rf", "sudo", "dd", "mkfs", ":(){:|:&};:"}
			for _, d := range dangerous {
				if strings.Contains(params.Command, d) {
					return "", fmt.Errorf("comando peligroso bloqueado: %s", d)
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, "sh", "-c", params.Command)
			if params.Cwd != "" {
				cmd.Dir = params.Cwd
			}

			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Sprintf("Exit code: %d\nOutput:\n%s", cmd.ProcessState.ExitCode(), string(output)), nil
			}

			return string(output), nil
		},
	})
}

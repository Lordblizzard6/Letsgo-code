package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/tuusuario/mycli/internal/agent"
	"github.com/tuusuario/mycli/internal/config"
	"github.com/tuusuario/mycli/internal/llm"
	"github.com/tuusuario/mycli/internal/tools"
)

var (
	cfgFile     string
	provider    string
	model       string
	interactive bool
)

var rootCmd = &cobra.Command{
	Use:   "mycli",
	Short: "CLI asistente de código multiproveedor",
	Long:  `Un CLI tipo Claude Code que soporta OpenAI, Anthropic y Gemini.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Cargar configuración
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("config: %w", err)
		}

		// Overrides de CLI
		if provider != "" {
			cfg.Provider.Name = provider
		}
		if model != "" {
			cfg.Provider.Model = model
		}

		// Crear proveedor
		llmProvider, err := llm.NewProvider(llm.ProviderConfig{
			Name:      cfg.Provider.Name,
			APIKey:    cfg.Provider.APIKey,
			Model:     cfg.Provider.Model,
			BaseURL:   cfg.Provider.BaseURL,
			Temperature: cfg.Provider.Temperature,
			MaxTokens: cfg.Provider.MaxTokens,
		})
		if err != nil {
			return fmt.Errorf("provider: %w", err)
		}

		// Crear registry de herramientas
		toolRegistry := tools.NewRegistry()

		// Crear agente
		ag := agent.NewAgent(
			llmProvider,
			config.GetSystemPrompt(),
			toolRegistry,
			cfg.Agent.MaxIterations,
		)

		if interactive {
			// Modo interactivo simple (sin TUI por ahora)
			fmt.Println("🤖 MyCLI - Modo interactivo")
			fmt.Println("Escribe tu pregunta o 'quit' para salir")
			fmt.Println()

			for {
				fmt.Print("❯ ")
				var input string
				fmt.Scanln(&input)

				if input == "quit" || input == "exit" {
					break
				}

				if input == "" {
					continue
				}

				ctx, cancel := context.WithCancel(context.Background())

				// Handle Ctrl+C
				sigChan := make(chan os.Signal, 1)
				signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
				go func() {
					<-sigChan
					cancel()
				}()

				ag.SetStreamCallback(func(content string) {
					fmt.Print(content)
				})

				ag.SetToolCallCallback(func(call llm.ToolCall) bool {
					fmt.Printf("\n🔧 Ejecutar herramienta: %s\n", call.Name)
					return true
				})

				ag.SetToolResultCallback(func(result string) {
					fmt.Printf("\n✅ Resultado: %s\n", result)
				})

				if err := ag.Run(ctx, input); err != nil {
					fmt.Printf("\nError: %v\n", err)
				}
				fmt.Println()

				cancel()
			}
		} else {
			// Modo comando directo
			if len(args) == 0 {
				return fmt.Errorf("proporciona un prompt")
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Handle Ctrl+C
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sigChan
				cancel()
			}()

			ag.SetStreamCallback(func(content string) {
				fmt.Print(content)
			})

			ag.SetToolCallCallback(func(call llm.ToolCall) bool {
				fmt.Printf("\n🔧 Ejecutar herramienta: %s\n", call.Name)
				return true
			})

			ag.SetToolResultCallback(func(result string) {
				fmt.Printf("\n✅ Resultado: %s\n", result)
			})

			if err := ag.Run(ctx, args[0]); err != nil {
				return err
			}
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&cfgFile, "config", "c", "", "archivo de config")
	rootCmd.Flags().StringVarP(&provider, "provider", "p", "", "proveedor (openai, anthropic, gemini)")
	rootCmd.Flags().StringVarP(&model, "model", "m", "", "modelo")
	rootCmd.Flags().BoolVarP(&interactive, "interactive", "i", true, "modo interactivo")
}

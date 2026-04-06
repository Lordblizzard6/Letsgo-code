package agent

import (
"context"
"encoding/json"
"fmt"
"strings"
"sync"
"time"

"github.com/tuusuario/mycli/internal/llm"
"github.com/tuusuario/mycli/internal/mcp"
"github.com/tuusuario/mycli/internal/security"
"github.com/tuusuario/mycli/internal/tools"
)

// Agent representa el agente de IA principal
type Agent struct {
provider       llm.Provider
ctxManager     *ContextManager
toolRegistry   *tools.Registry
mcpManager     *mcp.Manager
securityPolicy *security.PolicyManager

maxIterations  int
maxTokens      int

onStream       func(string)
onToolCall     func(llm.ToolCall) bool
onToolResult   func(string)
onThinking     func(string)

mu             sync.Mutex
sessionID      string
startTime      time.Time
totalCost      float64
}

// AgentConfig configuración del agente
type AgentConfig struct {
MaxIterations  int
MaxTokens      int
SystemPrompt   string
RequireConfirm bool
AuditLogPath   string
MCPServers     []llm.MCPConfig
}

// NewAgent crea un nuevo agente
func NewAgent(
provider llm.Provider,
cfg AgentConfig,
) (*Agent, error) {
// Crear policy manager
policy := &llm.SecurityPolicy{
AllowedPaths:      []string{"./**", "/tmp/**"},
DeniedPaths:       []string{"/etc/**", "/root/**", "/proc/**", "/sys/**"},
MaxFileSize:       10 * 1024 * 1024, // 10MB
EnableAuditLog:    cfg.AuditLogPath != "",
RequireConfirmFor: []string{},
}

if cfg.RequireConfirm {
policy.RequireConfirmFor = append(policy.RequireConfirmFor, 
"write_file", "run_command", "delete_file")
}

policyManager, err := security.NewPolicyManager(policy, cfg.AuditLogPath)
if err != nil {
return nil, fmt.Errorf("create policy manager: %w", err)
}

// Crear MCP manager
mcpManager := mcp.NewManager()
for _, srvCfg := range cfg.MCPServers {
server := mcp.NewMCPServer(
srvCfg.ServerName,
srvCfg.Command,
srvCfg.Args,
srvCfg.Env,
time.Duration(srvCfg.Timeout)*time.Second,
)
mcpManager.AddServer(server)
}

// Conectar servidores MCP
if len(cfg.MCPServers) > 0 {
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := mcpManager.ConnectAll(ctx); err != nil {
// Log warning pero continuar
fmt.Printf("Warning: MCP connect error: %v\n", err)
}
}

// Crear tool registry
toolRegistry := tools.NewRegistry(policyManager, mcpManager)

// Crear context manager
ctxManager := NewContextManager(cfg.SystemPrompt, cfg.MaxTokens)

return &Agent{
provider:       provider,
ctxManager:     ctxManager,
toolRegistry:   toolRegistry,
mcpManager:     mcpManager,
securityPolicy: policyManager,
maxIterations:  cfg.MaxIterations,
maxTokens:      cfg.MaxTokens,
startTime:      time.Now(),
}, nil
}

// SetStreamCallback establece callback para streaming de texto
func (a *Agent) SetStreamCallback(fn func(string)) {
a.onStream = fn
}

// SetToolCallCallback establece callback para confirmación de herramientas
func (a *Agent) SetToolCallCallback(fn func(llm.ToolCall) bool) {
a.onToolCall = fn
}

// SetToolResultCallback establece callback para resultados de herramientas
func (a *Agent) SetToolResultCallback(fn func(string)) {
a.onToolResult = fn
}

// SetThinkingCallback establece callback para pensamientos internos
func (a *Agent) SetThinkingCallback(fn func(string)) {
a.onThinking = fn
}

// Run ejecuta el agente con un input
func (a *Agent) Run(ctx context.Context, input string) error {
a.mu.Lock()
a.sessionID = fmt.Sprintf("sess_%d", time.Now().UnixNano())
a.toolRegistry.SetSessionID(a.sessionID)
a.mu.Unlock()

// Añadir mensaje del usuario
a.ctxManager.AddMessage(llm.Message{
Role:    "user",
Content: input,
})

for iteration := 0; iteration < a.maxIterations; iteration++ {
select {
case <-ctx.Done():
return ctx.Err()
default:
}

messages := a.ctxManager.GetMessages()
availableTools := a.toolRegistry.GetDefinitions()

// Streaming
stream, err := a.provider.Stream(ctx, messages, availableTools)
if err != nil {
return fmt.Errorf("stream error: %w", err)
}

var fullResponse strings.Builder
var pendingCalls []llm.ToolCall

for chunk := range stream {
if chunk.Err != nil {
return fmt.Errorf("chunk error: %w", chunk.Err)
}

if chunk.Thought != "" && a.onThinking != nil {
a.onThinking(chunk.Thought)
}

if chunk.Content != "" {
fullResponse.WriteString(chunk.Content)
if a.onStream != nil {
a.onStream(chunk.Content)
}
}

if len(chunk.ToolCalls) > 0 {
pendingCalls = append(pendingCalls, chunk.ToolCalls...)
}

if chunk.Done {
break
}
}

// Guardar respuesta del asistente
a.ctxManager.AddMessage(llm.Message{
Role:      "assistant",
Content:   fullResponse.String(),
ToolCalls: pendingCalls,
})

// Si no hay tool calls, terminar
if len(pendingCalls) == 0 {
return nil
}

// Ejecutar herramientas
for _, call := range pendingCalls {
select {
case <-ctx.Done():
return ctx.Err()
default:
}

// Herramientas especiales
if call.Name == "think" {
// Solo registrar, no ejecutar
a.ctxManager.AddMessage(llm.Message{
Role:     "tool",
ToolID:   call.ID,
ToolName: call.Name,
Content:  "Thinking recorded",
})
continue
}

if call.Name == "ask_user" {
// Manejar especialmente en TUI
if a.onToolCall != nil {
if !a.onToolCall(call) {
a.ctxManager.AddMessage(llm.Message{
Role:     "tool",
ToolID:   call.ID,
ToolName: call.Name,
Content:  "User declined to answer",
})
continue
}
}
}

// Confirmación requerida
if a.securityPolicy.RequiresConfirmation(call.Name) {
if a.onToolCall != nil {
if !a.onToolCall(call) {
a.ctxManager.AddMessage(llm.Message{
Role:     "tool",
ToolID:   call.ID,
ToolName: call.Name,
Content:  "Tool execution cancelled by user",
})
continue
}
}
}

// Ejecutar herramienta
result, err := a.toolRegistry.Execute(ctx, call.Name, call.Args)
if err != nil {
result = fmt.Sprintf("Error executing tool: %v", err)
}

if a.onToolResult != nil {
a.onToolResult(result)
}

// Guardar resultado
a.ctxManager.AddMessage(llm.Message{
Role:     "tool",
ToolID:   call.ID,
ToolName: call.Name,
Content:  result,
})
}
}

return fmt.Errorf("max iterations (%d) reached", a.maxIterations)
}

// GetContext devuelve el manejador de contexto
func (a *Agent) GetContext() *ContextManager {
return a.ctxManager
}

// GetStats devuelve estadísticas de la sesión
func (a *Agent) GetStats() map[string]interface{} {
a.mu.Lock()
defer a.mu.Unlock()

return map[string]interface{}{
"session_id":     a.sessionID,
"duration":       time.Since(a.startTime).String(),
"message_count":  a.ctxManager.GetMessageCount(),
"token_count":    a.ctxManager.GetTokenCount(),
"total_cost":     a.totalCost,
"iterations":     a.maxIterations,
}
}

// Close libera recursos del agente
func (a *Agent) Close() error {
if a.mcpManager != nil {
return a.mcpManager.DisconnectAll()
}
if a.securityPolicy != nil {
return a.securityPolicy.Close()
}
return nil
}

// ReloadMCPTools recarga las herramientas MCP
func (a *Agent) ReloadMCPTools() {
a.toolRegistry.ReloadMCPTools()
}

// AddAttachment añade un archivo adjunto al contexto
func (a *Agent) AddAttachment(path string) error {
// Implementar lectura de archivo y añadir como attachment
return nil
}

package tools

import (
"context"
"encoding/json"
"fmt"
"sync"

"github.com/tuusuario/mycli/internal/llm"
"github.com/tuusuario/mycli/internal/mcp"
"github.com/tuusuario/mycli/internal/security"
)

// Registry gestiona todas las herramientas disponibles
type Registry struct {
tools         map[string]llm.Tool
mu            sync.RWMutex
policyManager *security.PolicyManager
mcpManager    *mcp.Manager
sessionID     string
}

// NewRegistry crea un nuevo registry de herramientas
func NewRegistry(policyManager *security.PolicyManager, mcpManager *mcp.Manager) *Registry {
r := &Registry{
tools:         make(map[string]llm.Tool),
policyManager: policyManager,
mcpManager:    mcpManager,
}

// Registrar herramientas por defecto
r.RegisterFileSystemTools()
r.RegisterShellTools()
r.RegisterGitTools()
r.RegisterSearchTools()
r.RegisterThinkTool()

// Añadir herramientas MCP si hay manager
if mcpManager != nil {
r.RegisterMCPTools()
}

return r
}

// Register registra una herramienta
func (r *Registry) Register(tool llm.Tool) {
r.mu.Lock()
defer r.mu.Unlock()

// Wrapper con seguridad
originalHandler := tool.Handler
tool.Handler = func(ctx context.Context, args json.RawMessage) (string, error) {
// Validar con policy manager
if r.policyManager != nil {
if allowed, err := r.policyManager.ValidateToolCall(ctx, tool, args); !allowed {
return "", err
}
}

// Ejecutar handler original
return originalHandler(ctx, args)
}

r.tools[tool.Name] = tool
}

// GetDefinitions devuelve todas las herramientas disponibles
func (r *Registry) GetDefinitions() []llm.Tool {
r.mu.RLock()
defer r.mu.RUnlock()

var out []llm.Tool
for _, t := range r.tools {
if !t.Hidden {
out = append(out, t)
}
}

// Añadir herramientas MCP
if r.mcpManager != nil {
out = append(out, r.mcpManager.GetAllTools()...)
}

return out
}

// Execute ejecuta una herramienta por nombre
func (r *Registry) Execute(ctx context.Context, name string, args json.RawMessage) (string, error) {
r.mu.RLock()
tool, exists := r.tools[name]
r.mu.RUnlock()

if !exists {
// Buscar en herramientas MCP
if r.mcpManager != nil {
mcpTools := r.mcpManager.GetAllTools()
for _, t := range mcpTools {
if t.Name == name {
return t.Handler(ctx, args)
}
}
}
return "", fmt.Errorf("tool not found: %s", name)
}

return tool.Handler(ctx, args)
}

// HasTool verifica si una herramienta existe
func (r *Registry) HasTool(name string) bool {
r.mu.RLock()
defer r.mu.RUnlock()
_, exists := r.tools[name]
return exists
}

// GetToolCategories devuelve las categorías de herramientas disponibles
func (r *Registry) GetToolCategories() []string {
r.mu.RLock()
defer r.mu.RUnlock()

categories := make(map[string]bool)
for _, t := range r.tools {
if t.Category != "" {
categories[t.Category] = true
}
}

var result []string
for cat := range categories {
result = append(result, cat)
}
return result
}

// GetToolsByCategory devuelve herramientas de una categoría
func (r *Registry) GetToolsByCategory(category string) []llm.Tool {
r.mu.RLock()
defer r.mu.RUnlock()

var result []llm.Tool
for _, t := range r.tools {
if t.Category == category && !t.Hidden {
result = append(result, t)
}
}
return result
}

// SetSessionID establece el ID de sesión para auditoría
func (r *Registry) SetSessionID(sessionID string) {
r.sessionID = sessionID
}

// RequireConfirmationFor marca herramientas que requieren confirmación
func (r *Registry) RequireConfirmationFor(toolNames ...string) {
if r.policyManager == nil {
return
}

policy := r.policyManager.GetPolicy()
policy.RequireConfirmFor = append(policy.RequireConfirmFor, toolNames...)
r.policyManager.UpdatePolicy(policy)
}

// HideTool oculta una herramienta del usuario
func (r *Registry) HideTool(name string) {
r.mu.Lock()
defer r.mu.Unlock()

if tool, exists := r.tools[name]; exists {
tool.Hidden = true
r.tools[name] = tool
}
}

// Unregister elimina una herramienta
func (r *Registry) Unregister(name string) {
r.mu.Lock()
defer r.mu.Unlock()
delete(r.tools, name)
}

// ReloadMCPTools recarga las herramientas MCP
func (r *Registry) ReloadMCPTools() {
r.mu.Lock()
defer r.mu.Unlock()

// Eliminar herramientas MCP anteriores
for name, tool := range r.tools {
if tool.Category == "mcp" {
delete(r.tools, name)
}
}

r.RegisterMCPTools()
}

// RegisterMCPTools registra herramientas desde servidores MCP
func (r *Registry) RegisterMCPTools() {
if r.mcpManager == nil {
return
}

mcpTools := r.mcpManager.GetAllTools()
for _, t := range mcpTools {
r.tools["mcp_"+t.Name] = t
}
}

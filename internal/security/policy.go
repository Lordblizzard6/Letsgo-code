package security

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/tuusuario/mycli/internal/llm"
)

// PolicyManager gestiona las políticas de seguridad
type PolicyManager struct {
	policy    *llm.SecurityPolicy
	mu        sync.RWMutex
	auditLog  *AuditLogger
	pathCache map[string]bool // Cache para paths validados
}

// AuditEntry entrada de log de auditoría
type AuditEntry struct {
	Timestamp   string `json:"timestamp"`
	Action      string `json:"action"`
	Tool        string `json:"tool,omitempty"`
	Path        string `json:"path,omitempty"`
	Command     string `json:"command,omitempty"`
	User        string `json:"user,omitempty"`
	Allowed     bool   `json:"allowed"`
	Reason      string `json:"reason,omitempty"`
	SessionID   string `json:"session_id"`
}

// NewPolicyManager crea un nuevo gestor de políticas
func NewPolicyManager(policy *llm.SecurityPolicy, auditLogPath string) (*PolicyManager, error) {
	pm := &PolicyManager{
		policy:    policy,
		pathCache: make(map[string]bool),
	}
	
	if policy.EnableAuditLog && auditLogPath != "" {
		var err error
		pm.auditLog, err = NewAuditLogger(auditLogPath)
		if err != nil {
			return nil, fmt.Errorf("create audit logger: %w", err)
		}
	}
	
	return pm, nil
}

// ValidatePath valida si un path es seguro para acceder
func (pm *PolicyManager) ValidatePath(path string) (bool, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	// Normalizar path
	cleanPath := filepath.Clean(path)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return false, fmt.Errorf("resolve path: %w", err)
	}
	
	// Check cache
	if allowed, exists := pm.pathCache[absPath]; exists {
		return allowed, nil
	}
	
	// Verificar path traversal
	if strings.Contains(cleanPath, "..") && !strings.HasPrefix(absPath, "/") {
		pm.logAudit("path_access", "", path, "", false, "path traversal detected")
		return false, fmt.Errorf("path traversal no permitido")
	}
	
	// Verificar denied paths primero (tienen prioridad)
	for _, pattern := range pm.policy.DeniedPaths {
		match, err := doublestar.Match(pattern, absPath)
		if err != nil {
			continue
		}
		if match {
			pm.logAudit("path_access", "", path, "", false, fmt.Sprintf("matches denied pattern: %s", pattern))
			pm.pathCache[absPath] = false
			return false, nil
		}
	}
	
	// Si hay allowed paths definidos, verificar que esté en la lista
	if len(pm.policy.AllowedPaths) > 0 {
		allowed := false
		for _, pattern := range pm.policy.AllowedPaths {
			match, err := doublestar.Match(pattern, absPath)
			if err != nil {
				continue
			}
			if match {
				allowed = true
				break
			}
		}
		
		if !allowed {
			pm.logAudit("path_access", "", path, "", false, "path not in allowed list")
			pm.pathCache[absPath] = false
			return false, nil
		}
	}
	
	pm.pathCache[absPath] = true
	pm.logAudit("path_access", "", path, "", true, "")
	return true, nil
}

// ValidateCommand valida si un comando es seguro para ejecutar
func (pm *PolicyManager) ValidateCommand(command string) (bool, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	// Normalizar comando
	cmd := strings.TrimSpace(command)
	
	// Lista negra de comandos peligrosos
	dangerousPatterns := []string{
		`rm\s+(-[rf]+\s+)?/`,         // rm en root
		`rm\s+-rf\s+\*`,              // rm -rf *
		`:\(\)\{\s*:\|:&\s*\}`,       // fork bomb
		`mkfs`,                       // formatear discos
		`dd\s+if=.*of=/dev/`,         // dd a dispositivos
		`chmod\s+[0-7]*\s+/`,         // chmod en root
		`chown\s+.*:/`,               // chown en root
		`sudo\s+su`,                  // sudo su
		`curl.*\|\s*(ba)?sh`,         // curl pipe to shell
		`wget.*\|\s*(ba)?sh`,         // wget pipe to shell
		`/dev/(zero|null|random)`,    // acceso a /dev
		`procfs`,                     // acceso a proc
		`sysfs`,                      // acceso a sys
	}
	
	for _, pattern := range dangerousPatterns {
		matched, _ := regexp.MatchString(pattern, cmd)
		if matched {
			pm.logAudit("command_execution", "", "", command, false, "dangerous command pattern detected")
			return false, fmt.Errorf("comando peligroso detectado")
		}
	}
	
	// Verificar denied commands
	for _, denied := range pm.policy.DeniedCommands {
		if strings.Contains(cmd, denied) {
			pm.logAudit("command_execution", "", "", command, false, fmt.Sprintf("matches denied command: %s", denied))
			return false, fmt.Errorf("comando denegado: %s", denied)
		}
	}
	
	// Si hay allowed commands, verificar whitelist
	if len(pm.policy.AllowedCommands) > 0 {
		allowed := false
		for _, allowedCmd := range pm.policy.AllowedCommands {
			if strings.HasPrefix(cmd, allowedCmd) {
				allowed = true
				break
			}
		}
		
		if !allowed {
			pm.logAudit("command_execution", "", "", command, false, "command not in allowed list")
			return false, fmt.Errorf("comando no está en la lista blanca")
		}
	}
	
	pm.logAudit("command_execution", "", "", command, true, "")
	return true, nil
}

// ValidateFileSize valida el tamaño de un archivo
func (pm *PolicyManager) ValidateFileSize(path string) (bool, error) {
	if pm.policy.MaxFileSize <= 0 {
		return true, nil
	}
	
	info, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("stat file: %w", err)
	}
	
	if info.Size() > pm.policy.MaxFileSize {
		pm.logAudit("file_size_check", "", path, "", false, fmt.Sprintf("file too large: %d bytes", info.Size()))
		return false, fmt.Errorf("archivo demasiado grande: %d bytes (max: %d)", info.Size(), pm.policy.MaxFileSize)
	}
	
	return true, nil
}

// RequiresConfirmation verifica si una herramienta requiere confirmación
func (pm *PolicyManager) RequiresConfirmation(toolName string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	for _, t := range pm.policy.RequireConfirmFor {
		if t == toolName || t == "*" {
			return true
		}
	}
	return false
}

// ValidateToolCall valida una llamada a herramienta
func (pm *PolicyManager) ValidateToolCall(ctx context.Context, tool llm.Tool, args json.RawMessage) (bool, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	// Extraer path si es herramienta de filesystem
	if tool.Category == "fs" {
		var params struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal(args, &params); err == nil && params.Path != "" {
			if allowed, err := pm.ValidatePath(params.Path); !allowed {
				return false, err
			}
			
			// Para write_file, validar tamaño
			if tool.Name == "write_file" {
				var writeParams struct {
					Path    string `json:"path"`
					Content string `json:"content"`
				}
				if err := json.Unmarshal(args, &writeParams); err == nil {
					if int64(len(writeParams.Content)) > pm.policy.MaxFileSize && pm.policy.MaxFileSize > 0 {
						return false, fmt.Errorf("contenido demasiado grande")
					}
				}
			}
		}
	}
	
	// Para shell, validar comando
	if tool.Category == "shell" {
		var params struct {
			Command string `json:"command"`
		}
		if err := json.Unmarshal(args, &params); err == nil && params.Command != "" {
			if allowed, err := pm.ValidateCommand(params.Command); !allowed {
				return false, err
			}
		}
	}
	
	return true, nil
}

func (pm *PolicyManager) logAudit(action, tool, path, command string, allowed bool, reason string) {
	if pm.auditLog == nil {
		return
	}
	
	entry := AuditEntry{
		Action:    action,
		Tool:      tool,
		Path:      path,
		Command:   command,
		Allowed:   allowed,
		Reason:    reason,
		SessionID: getSessionID(),
	}
	
	pm.auditLog.Log(entry)
}

// ClearCache limpia el cache de paths
func (pm *PolicyManager) ClearCache() {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.pathCache = make(map[string]bool)
}

// UpdatePolicy actualiza la política de seguridad
func (pm *PolicyManager) UpdatePolicy(policy *llm.SecurityPolicy) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.policy = policy
	pm.pathCache = make(map[string]bool) // Limpiar cache al cambiar política
}

// GetPolicy devuelve la política actual
func (pm *PolicyManager) GetPolicy() *llm.SecurityPolicy {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.policy
}

// Close cierra el policy manager y libera recursos
func (pm *PolicyManager) Close() error {
	if pm.auditLog != nil {
		return pm.auditLog.Close()
	}
	return nil
}

// Helpers
var sessionIDCounter int
var sessionIDMu sync.Mutex

func getSessionID() string {
	sessionIDMu.Lock()
	defer sessionIDMu.Unlock()
	sessionIDCounter++
	return fmt.Sprintf("sess_%d", sessionIDCounter)
}

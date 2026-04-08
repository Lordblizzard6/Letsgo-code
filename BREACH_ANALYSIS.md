# Análisis de Brechas - Claude Code Go vs Original

## Resumen de Implementación Actual

El port de Claude Code a Go tiene **paridad funcional 1:1** en las herramientas base, pero faltan componentes arquitectónicos avanzados del código original.

---

## ✅ IMPLEMENTADO (Paridad Completa)

### Herramientas Base (18 herramientas)
- [x] `ls`, `cat`, `glob`, `grep` - Operaciones de archivo
- [x] `read_file`, `edit`, `write_file` - Edición de archivos
- [x] `bash` - Ejecución de comandos con guardrails
- [x] `web_search`, `web_fetch` - Operaciones web
- [x] `todo_write`, `task_*` - Gestión de tareas
- [x] `agent`, `agent_get`, `agent_list` - Sistema de agentes
- [x] `brief`, `notebook_edit` - Contexto y notebooks
- [x] `enter_plan_mode`, `exit_plan_mode`, `plan_*` - Plan Mode
- [x] `skill` - Sistema de skills
- [x] `list_mcp_resources`, `read_mcp_resource`, `mcp_call` - MCP básico
- [x] `lsp` - LSP Tool
- [x] `ask_user` - Interacción con usuario

### APIs Soportadas
- [x] Anthropic (Claude)
- [x] OpenAI
- [x] Groq
- [x] **Ollama** (modelos locales) - *Nuevo*
- [x] **OpenRouter** - *Nuevo*

### Infraestructura
- [x] TUI con Bubble Tea
- [x] Streaming de respuestas
- [x] Sistema de permisos básico (guardrails)
- [x] Cost tracking
- [x] SQLite para historial
- [x] Prompt del sistema completo

---

## 🔴 P0 - CRÍTICOS (Arquitectura Core)

### 1. Sistema de Plugins
**Referencia:** `src/plugins/`, `src/commands/plugin/`

**Descripción:**
El código original tiene un sistema completo de plugins con:
- Carga dinámica de plugins
- Marketplace de plugins
- API para plugins (hooks, comandos, tools)
- Gestión de versiones
- Aislamiento de plugins

**Impacto:** ALTO - Los usuarios no pueden extender la funcionalidad

**Implementación Propuesta:**
- Sistema de plugins basado en WASM o shared libraries
- API de plugins con hooks para tool registration
- Comando `plugin install/remove/list`
- Configuración de plugins en `~/.claudego/plugins/`

---

### 2. MCP (Model Context Protocol) Completo
**Referencia:** `src/services/mcp/`

**Descripción:**
Aunque implementamos herramientas MCP básicas, el original tiene:
- Gestión de servidores MCP (start/stop/restart)
- Configuración de servidores MCP (JSON)
- Autenticación OAuth para MCP
- Recursos MCP con suscripción
- Tools MCP dinámicas
- Prompts MCP

**Archivos clave:**
- `src/services/mcp/client.js` - Cliente MCP
- `src/services/mcp/config.js` - Configuración
- `src/services/mcp/claudeai.js` - Integración Claude AI
- `src/commands/mcp/` - Comandos MCP

**Impacto:** ALTO - Integración limitada con contexto externo

**Implementación Propuesta:**
- Cliente MCP completo con JSON-RPC
- Gestión de servidores MCP (stdio/sse)
- Comandos: `mcp add`, `mcp list`, `mcp remove`
- Configuración en `~/.claudego/mcp.json`

---

### 3. LSP Manager
**Referencia:** `src/services/lsp/manager.js`

**Descripción:**
El original tiene un LSP Manager completo:
- Inicialización automática de servidores LSP
- Gestión de múltiples servidores (TypeScript, Python, Go, etc.)
- Comunicación con servidores LSP
- Funcionalidades: go-to-definition, hover, completions, diagnostics

**Impacto:** MEDIO - Falta inteligencia de código avanzada

**Implementación Propuesta:**
- Manager de procesos LSP
- Implementación del protocolo LSP (JSON-RPC)
- Integración con la tool `lsp`
- Detección automática de servidores LSP instalados

---

### 4. Sistema de Permisos Avanzado
**Referencia:** `src/utils/permissions/`

**Descripción:**
El sistema actual de guardrails es básico. El original tiene:
- **Permission Modes**: ask, auto, auto-bp (bypass), opt-out
- **Persistencia**: Configuración de permisos guardada
- **Granularidad**: Permisos por tool, por path, por command pattern
- **Contextos**: Diferentes permisos según contexto (git, docker, etc.)
- **UI de Permisos**: Dialogs interactivos para aprobar/rechazar

**Archivos clave:**
- `src/utils/permissions/PermissionMode.js`
- `src/utils/permissions/permissionSetup.js`
- `src/utils/permissions/autoModeState.js`

**Impacto:** ALTO - Seguridad y UX de permisos inferior

**Implementación Propuesta:**
- Sistema de modos de permiso (ask, auto, auto-bp)
- Persistencia en config
- UI en TUI para aprobación interactiva
- Granularidad por tool y por path

---

## 🟡 P1 - IMPORTANTES (Features)

### 5. Auto-updater
**Referencia:** `src/utils/autoUpdater.js`

- Detección de nuevas versiones
- Notificaciones de actualización
- Descarga automática
- Changelog

---

### 6. Teleport / Sesiones Remotas
**Referencia:** `src/remote/`, `src/commands/teleport/`

- Conexión a sesiones remotas
- Sincronización de estado
- Resumen de sesiones
- Reconexión

---

### 7. GitHub Integration
**Referencia:** `src/utils/github/`

- Autenticación GitHub
- Crear/ver PRs
- Issues
- Actions
- GitHub App integration

---

### 8. Chrome Integration
**Referencia:** `src/utils/claudeInChrome/`

- Integración con navegador Chrome
- Lectura de páginas web
- Interacción con DOM
- Captura de pantalla

---

### 9. Voice Input
**Referencia:** `src/voice/`

- Reconocimiento de voz
- Transcripción
- Comandos por voz

---

### 10. Dashboard Analytics
**Referencia:** `src/commands/insights.ts`

- Métricas de uso
- Costos históricos
- Estadísticas de sesiones
- Insights de productividad

---

## 🟢 P2 - NICE TO HAVE

### 11. REPL Mode
**Referencia:** `src/tools/REPLTool/`

- Modo interactivo avanzado
- Historia de comandos
- Autocompletado

---

### 12. Web Browser Tool
**Referencia:** `src/tools/WebBrowserTool/`

- Navegación real (no solo fetch)
- JavaScript execution
- Screenshots
- Form interaction

---

### 13. Tareas Agendadas (Cron)
**Referencia:** `src/tools/ScheduleCronTool/`

- Programación de tareas
- Ejecución recurrente
- Cron expressions

---

### 14. Notificaciones Push
**Referencia:** `src/tools/PushNotificationTool/`

- Notificaciones nativas
- Integración con SO

---

## 🔵 P3 - ENTERPRISE

### 15. Coordinator Mode
**Referencia:** `src/coordinator/`

- Modo swarm de agentes
- Coordinación multi-agente
- Distribución de tareas

---

### 16. Team/Workspace Management
**Referencia:** `src/utils/swarm/`, `src/commands/agents/`

- Gestión de equipos
- Colaboración en tiempo real
- Shared context

---

### 17. Settings Remotas
**Referencia:** `src/services/remoteManagedSettings/`

- Configuración gestionada
- Policy enforcement
- MDM integration

---

### 18. Policy Limits
**Referencia:** `src/services/policyLimits/`

- Límites de uso
- Cuotas
- Rate limiting

---

## 📋 Plan de Implementación Recomendado

### Fase 1: P0 Core (2-3 semanas)
1. Sistema de Permisos Avanzado
2. MCP Completo
3. LSP Manager

### Fase 2: P0 Extension (2 semanas)
4. Sistema de Plugins

### Fase 3: P1 Features (3-4 semanas)
5. GitHub Integration
6. Auto-updater
7. Dashboard Analytics

### Fase 4: P2+ (Opcional)
8. Web Browser Tool
9. Voice Input
10. REPL Mode

---

## Notas Técnicas

### Diferencias Arquitectónicas
- **Original:** TypeScript/React con Bun runtime, sistema de plugins dinámico
- **Go Port:** Go nativo, tools estáticas, TUI con Bubble Tea

### Ventajas del Port Go
- Binario único, sin dependencias
- Mejor performance
- Menor consumo de memoria
- Startup más rápido

### Limitaciones del Port Go
- Sin sistema de plugins dinámico (WASM puede ayudar)
- Menos ecosistema de extensiones
- UI más simple (TUI vs React Ink)

---

*Generado: Abril 2026*
*Versión del port: 0.1.0*

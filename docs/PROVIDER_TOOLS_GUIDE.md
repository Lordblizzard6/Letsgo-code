# Provider & Tools Compatibility Guide

## Resumen de Soporte por Proveedor

| Proveedor | Tool Support | Streaming | Function Calling | Notas |
|-----------|--------------|-----------|------------------|-------|
| **Anthropic** | ✅ Completo | ✅ | ✅ Nativo | Mejor soporte para tools complejas |
| **OpenAI** | ✅ Completo | ✅ | ✅ Nativo | GPT-4o, GPT-4o-mini soportan todas las tools |
| **Groq** | ⚠️ Parcial | ✅ | ✅ OpenAI-compatible | Llama 3.3 70B, GPT-OSS soportan tools básicas |
| **OpenRouter** | ⚠️ Parcial | ✅ | ⚠️ Depende del modelo | Varía según el modelo seleccionado |
| **Ollama** | ⚠️ Limitado | ✅ | ❌ No nativo | Depende completamente del modelo local |

---

## Tools Disponibles (35+ herramientas)

### 🗂️ File Operations
| Tool | Descripción | Anthropic | OpenAI | Groq | OpenRouter | Ollama |
|------|-------------|-----------|--------|------|------------|--------|
| `Ls` | Listar archivos en directorio | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `Cat` | Leer contenido de archivo | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `Glob` | Buscar archivos por patrón | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `Grep` | Buscar texto en archivos | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `Edit` | Editar archivos (find/replace) | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `WriteFile` | Escribir archivo completo | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `NotebookEdit` | Editar notebooks Jupyter | ✅ | ✅ | ✅ | ✅* | ⚠️ |

### 💻 Shell & Ejecución
| Tool | Descripción | Anthropic | OpenAI | Groq | OpenRouter | Ollama |
|------|-------------|-----------|--------|------|------------|--------|
| `Bash` | Ejecutar comandos bash | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `PowerShell` | Ejecutar comandos PowerShell | ✅ | ✅ | ✅ | ✅* | ⚠️ |

### 🌐 Web & Búsqueda
| Tool | Descripción | Anthropic | OpenAI | Groq | OpenRouter | Ollama |
|------|-------------|-----------|--------|------|------------|--------|
| `WebSearch` | Buscar en la web | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `WebFetch` | Obtener contenido de URL | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `WebBrowser` | Navegación web automatizada | ✅ | ✅ | ⚠️ | ⚠️ | ❌ |

### 📝 Gestión de Tareas
| Tool | Descripción | Anthropic | OpenAI | Groq | OpenRouter | Ollama |
|------|-------------|-----------|--------|------|------------|--------|
| `TodoWrite` | Crear lista de tareas | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `TaskCreate` | Crear tarea asíncrona | ✅ | ✅ | ⚠️ | ⚠️ | ❌ |
| `TaskGet` | Obtener estado de tarea | ✅ | ✅ | ⚠️ | ⚠️ | ❌ |
| `TaskUpdate` | Actualizar tarea | ✅ | ✅ | ⚠️ | ⚠️ | ❌ |
| `TaskList` | Listar tareas | ✅ | ✅ | ⚠️ | ⚠️ | ❌ |
| `TaskStop` | Detener tarea | ✅ | ✅ | ⚠️ | ⚠️ | ❌ |
| `TaskOutput` | Obtener output de tarea | ✅ | ✅ | ⚠️ | ⚠️ | ❌ |

### 🤖 Sistema de Agentes
| Tool | Descripción | Anthropic | OpenAI | Groq | OpenRouter | Ollama |
|------|-------------|-----------|--------|------|------------|--------|
| `Agent` | Crear sub-agente | ✅ | ✅ | ⚠️ | ⚠️ | ❌ |
| `AgentGet` | Obtener estado de agente | ✅ | ✅ | ⚠️ | ⚠️ | ❌ |
| `AgentList` | Listar agentes | ✅ | ✅ | ⚠️ | ⚠️ | ❌ |

### 📋 Plan Mode
| Tool | Descripción | Anthropic | OpenAI | Groq | OpenRouter | Ollama |
|------|-------------|-----------|--------|------|------------|--------|
| `EnterPlanMode` | Entrar en modo planificación | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `ExitPlanMode` | Salir de modo planificación | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `PlanStepAdd` | Agregar paso al plan | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `PlanShow` | Mostrar plan actual | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `PlanApprove` | Aprobar y ejecutar plan | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `PlanStepComplete` | Marcar paso completado | ✅ | ✅ | ✅ | ✅* | ⚠️ |

### 🔌 Integraciones
| Tool | Descripción | Anthropic | OpenAI | Groq | OpenRouter | Ollama |
|------|-------------|-----------|--------|------|------------|--------|
| `LSPTool` | Interactuar con LSP | ✅ | ✅ | ⚠️ | ⚠️ | ❌ |
| `ListMcpResources` | Listar recursos MCP | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `ReadMcpResource` | Leer recurso MCP | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `McpCall` | Llamar herramienta MCP | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `SkillTool` | Ejecutar skill | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `Workflow` | Ejecutar workflow | ✅ | ✅ | ✅ | ✅* | ⚠️ |

### 🔄 Contexto & Utilidades
| Tool | Descripción | Anthropic | OpenAI | Groq | OpenRouter | Ollama |
|------|-------------|-----------|--------|------|------------|--------|
| `Brief` | Resumir contexto | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `AskUser` | Preguntar al usuario | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `Sleep` | Pausar ejecución | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `ToolSearch` | Buscar herramientas disponibles | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `EnterWorktree` | Entrar a worktree git | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `ExitWorktree` | Salir de worktree git | ✅ | ✅ | ✅ | ✅* | ⚠️ |
| `ListWorktrees` | Listar worktrees git | ✅ | ✅ | ✅ | ✅* | ⚠️ |

---

## Detalles por Proveedor

### Anthropic (Claude) ✅
**Modelos recomendados:**
- `claude-sonnet-4-20250514` - Mejor relación calidad/velocidad
- `claude-3-opus-20240229` - Mayor capacidad de razonamiento

**Soporte de Tools:**
- ✅ Todas las 35+ herramientas soportadas
- ✅ Tool calling nativo con alta precisión
- ✅ Streaming de tool inputs en tiempo real
- ✅ Soporte para herramientas complejas (LSP, Agentes, Worktrees)

**Limitaciones:**
- Requiere API key de Anthropic
- Costo por token más alto que otros proveedores

---

### OpenAI (GPT-4o) ✅
**Modelos recomendados:**
- `gpt-4o` - Mejor rendimiento general
- `gpt-4o-mini` - Más rápido y económico

**Soporte de Tools:**
- ✅ Todas las herramientas soportadas
- ✅ Function calling nativo
- ✅ Streaming completo
- ✅ Compatible con formato OpenAI

**Limitaciones:**
- Requiere API key de OpenAI
- Algunas herramientas complejas pueden tener menor precisión que Claude

---

### Groq ⚠️
**Modelos disponibles:**
- `llama-3.3-70b-versatile` - Buen soporte de tools
- `llama-3.1-8b-instant` - Rápido pero limitado en tools complejas
- `openai/gpt-oss-120b` - Soporta tool básicas
- `openai/gpt-oss-20b` - Soporta tool básicas

**Soporte de Tools:**
- ✅ Tools básicas (file ops, bash, web)
- ⚠️ Agentes y tareas complejas pueden fallar
- ⚠️ Streaming de tool inputs puede ser inconsistente
- ✅ Formato OpenAI-compatible

**Limitaciones conocidas:**
- Rate limits estrictos (TPM limit)
- Algunos modelos no soportan function calling
- Herramientas complejas (LSP, MCP) pueden no funcionar correctamente

---

### OpenRouter ⚠️
**Modelos populares:**
- `anthropic/claude-sonnet-4` - Via OpenRouter (full support)
- `openai/gpt-4o` - Via OpenRouter (full support)
- `meta-llama/llama-3.3-70b-instruct` - Soporte variable
- `google/gemini-2.5-pro` - Soporte variable

**Soporte de Tools:**
- ⚠️ Depende completamente del modelo seleccionado
- ✅ Modelos Anthropic/OpenAI vía OpenRouter = full support
- ⚠️ Modelos Llama/Gemma = soporte básico
- ✅ Formato OpenAI-compatible

**Limitaciones:**
- No todos los modelos soportan function calling
- Latencia adicional por routing
- Algunos modelos gratuitos tienen límites estrictos

---

### Ollama ⚠️
**Modelos comunes:**
- `llama3.2` - Soporte de tools muy limitado
- `codellama` - Sin soporte nativo de tools
- `mistral` - Variable según versión

**Soporte de Tools:**
- ⚠️ La mayoría de modelos locales NO soportan tool calling nativo
- ⚠️ Herramientas ejecutadas en modo "texto" (menos confiable)
- ❌ Agentes, tareas asíncronas, y herramientas complejas no disponibles
- ⚠️ Requiere modelos específicos con soporte de function calling

**Limitaciones:**
- Performance depende del hardware local
- La mayoría de tools no funcionan correctamente
- Útil solo para chat básico sin herramientas

---

## Recomendaciones

### Para uso completo de tools (recomendado)
1. **Anthropic Claude Sonnet 4** - Mejor opción general
2. **OpenAI GPT-4o** - Alternativa excelente
3. **Anthropic Claude vía OpenRouter** - Si prefieres OpenRouter

### Para uso rápido/económico
1. **Groq Llama 3.3 70B** - Tools básicas funcionan bien
2. **OpenAI GPT-4o Mini** - Económico con buen soporte

### Para desarrollo offline
1. **Ollama** - Solo chat básico, sin tools

---

## Configuración de Environment Variables

```bash
# Anthropic (recomendado para tools complejas)
ANTHROPIC_API_KEY=sk-ant-xxx

# OpenAI (alternativa sólida)
OPENAI_API_KEY=sk-xxx

# Groq (rápido pero con limitaciones)
GROQ_API_KEY=gsk_xxx

# OpenRouter (acceso a múltiples modelos)
OPENROUTER_API_KEY=sk-or-xxx

# Ollama (local, no requiere API key)
# OLLAMA_HOST=http://localhost:11434 (opcional)
```

---

## Troubleshooting

### "Tool not found" o "Function not supported"
- **Causa:** El modelo no soporta function calling
- **Solución:** Cambiar a Anthropic Claude o OpenAI GPT-4o

### Errores 401/403 con tools
- **Causa:** API key incorrecta o sin permisos
- **Solución:** Verificar `GetAPIKeyForProvider()` detecta el proveedor correcto

### Tools funcionan parcialmente
- **Causa:** Modelo con soporte limitado (Groq/Ollama)
- **Solución:** Usar modelos recomendados en esta guía

### Streaming de tools no aparece
- **Causa:** Viewport no se actualiza durante streaming
- **Solución:** Verificar `updateViewportContent()` se llama en `deltaMsg`

---

## Historial de Cambios

- **v0.1.0** - Documentación inicial
- Agregado soporte detallado por proveedor
- Matriz de compatibilidad de 35+ herramientas
- Recomendaciones basadas en testing real

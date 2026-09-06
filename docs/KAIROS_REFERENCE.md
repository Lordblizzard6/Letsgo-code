# KAIROS - Sistema de Asistente Proactivo

> **NOTA:** Este documento es para referencia e información únicamente.  
> KAIROS es un sistema experimental complejo que NO será implementado en la versión Go del CLI.

---

## ¿Qué es KAIROS?

KAIROS es un sistema de asistente proactivo desarrollado por Anthropic que permite a Claude Code anticipar necesidades del usuario y actuar de forma autónoma sin esperar comandos explícitos.

**Feature Flags:** `KAIROS`, `PROACTIVE`, `KAIROS_BRIEF`, `KAIROS_PUSH_NOTIFICATION`, `KAIROS_GITHUB_WEBHOOKS`

---

## Arquitectura de KAIROS

### Componentes Principales

```
KAIROS System
├── Proactive Agent
│   ├── Context Monitor (observa cambios en el entorno)
│   ├── Intent Predictor (predice intenciones del usuario)
│   └── Action Suggester (sugiere acciones proactivas)
├── Brief System
│   ├── Change Summarizer (resume cambios automáticamente)
│   └── Context Compressor (compacta contexto inteligentemente)
├── Notification System
│   ├── Push Notifications (móvil/desktop)
│   ├── Webhook Integration (GitHub PRs, etc.)
│   └── Alert Manager (gestiona prioridades)
└── Assistant Mode
    ├── Full Autonomy (actúa sin confirmación)
    ├── Guided Assistance (confirma antes de actuar)
    └── Background Tasks (trabaja mientras usuario hace otras cosas)
```

### Archivos TypeScript Relacionados

```
src/
├── commands/
│   ├── proactive.js          # Comando /proactive
│   ├── brief.js              # Comando /brief (resúmenes automáticos)
│   ├── assistant/
│   │   └── index.js          # Modo asistente completo
│   ├── subscribe-pr.js       # Subscripción a PRs
│   └── install-github-app/   # GitHub App integration
├── tools/
│   ├── SendUserFileTool/     # Envía archivos al usuario proactivamente
│   ├── PushNotificationTool/ # Notificaciones push
│   ├── SubscribePRTool/      # Subscripción PRs
│   └── BriefTool/            # Resúmenes automáticos
└── services/
    ├── kairos/
    │   ├── contextWatcher.ts # Observa cambios de contexto
    │   ├── intentEngine.ts   # Motor de predicción
    │   └── actionQueue.ts    # Cola de acciones proactivas
    └── notifications/
        ├── pushService.ts    # Servicio de notificaciones
        └── webhookHandler.ts # Maneja webhooks entrantes
```

---

## Comandos KAIROS

### `/proactive` - Modo Proactivo
```typescript
interface ProactiveCommand {
  type: 'prompt' | 'local-jsx'
  description: 'Enable proactive assistance mode'
  modes: ['watch', 'suggest', 'act']
}
```

**Modos:**
- `watch`: Solo observa y reporta cambios relevantes
- `suggest`: Sugiere acciones basadas en patrones
- `act`: Actúa autónomamente en tareas rutinarias

### `/brief` - Resúmenes Automáticos
```typescript
interface BriefCommand {
  type: 'local'
  description: 'Generate brief summaries of changes'
  triggers: ['on_change', 'on_request', 'scheduled']
}
```

**Triggers:**
- `on_change`: Cuando hay cambios significativos
- `on_request`: Bajo demanda del usuario
- `scheduled`: Periódicamente (cada X minutos)

### `/assistant` - Asistente Completo
```typescript
interface AssistantCommand {
  type: 'local-jsx'
  description: 'Full assistant mode with persistent background presence'
  features: [
    'file_monitoring',
    'git_watching',
    'notification_management',
    'task_scheduling'
  ]
}
```

---

## Tools KAIROS

### SendUserFileTool
```typescript
class SendUserFileTool {
  name = 'send_user_file'
  description = 'Proactively send a file to the user when relevant changes are detected'
  
  schema = {
    file_path: string,
    reason: string,
    urgency: 'low' | 'medium' | 'high',
    auto_open: boolean
  }
}
```

### PushNotificationTool
```typescript
class PushNotificationTool {
  name = 'push_notification'
  description = 'Send push notification to user devices'
  
  schema = {
    title: string,
    message: string,
    priority: 'low' | 'normal' | 'high' | 'urgent',
    actions: string[],
    channel: 'mobile' | 'desktop' | 'both'
  }
}
```

### SubscribePRTool
```typescript
class SubscribePRTool {
  name = 'subscribe_pr'
  description = 'Subscribe to GitHub PR updates and receive proactive notifications'
  
  schema = {
    repo: string,
    pr_number: number,
    events: string[],  // ['comment', 'review', 'commit', 'merge']
    notify_mode: 'immediate' | 'digest' | 'summary'
  }
}
```

---

## Integraciones KAIROS

### GitHub App
```
Instalación: /install-github-app
- OAuth flow para GitHub App
- Permisos: repo, pull_requests, issues
- Webhooks configurados automáticamente
- Recibe eventos de PRs, issues, commits
```

### Slack App
```
Instalación: /install-slack-app
- OAuth para Slack workspace
- Notificaciones a canales específicos
- Comandos slash integration
- Mensajes proactivos desde Claude
```

### Notificaciones Push
```
Configuración: Push notification service
- Firebase Cloud Messaging (FCM) para Android
- Apple Push Notification Service (APNS) para iOS
- Web Push API para desktop
- Service Workers para PWA
```

---

## Casos de Uso KAIROS

### 1. Desarrollo Proactivo
```
Escenario: Usuario trabajando en feature branch

KAIROS observa:
- Tests fallando tras cambios
- Nueva dependencia añadida
- Conflicto potencial con main

Acciones proactivas:
1. Notifica: "Tests failing in src/utils.ts"
2. Sugiere: "Run tests or check recent changes?"
3. Ofrece: "Show diff with main branch"
```

### 2. Code Review Automático
```
Escenario: Nuevo PR creado

KAIROS detecta:
- PR #123 opened in repo/project
- Auto-assigns reviewers
- Checks for common issues

Acciones:
1. Analiza código automáticamente
2. Comenta: "Potential security issue in auth.go"
3. Notifica reviewers vía Slack
4. Resume cambios en descripción
```

### 3. Asistente Background
```
Escenario: Usuario en reunión

KAIROS continúa:
- Monitoreando CI/CD
- Revisando logs
- Esperando builds

Notifica:
- "Build completed successfully"
- "Production deployment ready"
- "Error rate spike detected"
```

---

## Tamaño y Complejidad

### Estadísticas
- **Código:** ~15,000 líneas TypeScript
- **Archivos:** 40+ módulos
- **Dependencias:** 8 servicios externos
- **Tests:** ~5,000 líneas de tests

### Complejidad de Implementación
| Componente | Dificultad | Tiempo Estimado |
|------------|------------|-----------------|
| Context Monitor | Alta | 2-3 semanas |
| Intent Predictor | Muy Alta | 4-6 semanas |
| Notification System | Media | 1-2 semanas |
| GitHub Integration | Media | 1-2 semanas |
| Push Notifications | Alta | 2-3 semanas |
| **TOTAL** | **Muy Alta** | **3-4 meses** |

---

## Por Qué NO Implementar en Go

### Razones Técnicas
1. **Complejidad extrema** - Sistema distribuido con múltiples servicios
2. **Dependencias externas** - Requiere Firebase, APNS, Webhooks
3. **Modelos de ML** - Intent prediction requiere modelos entrenados
4. **Servicios backend** - Necesita infraestructura propia

### Razones de Negocio
1. **Uso limitado** - Solo disponible para usuarios Anthropic internos
2. **Feature flags extensivos** - Requiere sistema de feature flags completo
3. **Mantenimiento alto** - Constante evolución y ajustes
4. **No esencial** - Funcionalidad "nice to have" no core

### Alternativas
- Usar versión TypeScript si se necesita KAIROS
- Implementar sistema de notificaciones básico simplificado
- Crear modo "watch" simple sin ML

---

## Referencias

### Archivos TypeScript
- `src/commands/proactive.js` - Comando principal
- `src/commands/brief.js` - Sistema de resúmenes
- `src/commands/assistant/index.js` - Modo asistente
- `src/services/kairos/` - Servicios core
- `src/tools/SendUserFileTool/` - Tool proactivo
- `src/tools/PushNotificationTool/` - Notificaciones
- `src/tools/SubscribePRTool/` - Subscripción PRs

### Documentación Interna
- KAIROS Architecture Doc (internal)
- Proactive Design Patterns (internal)
- Notification System Spec (internal)

---

## Conclusión

**KAIROS es una característica avanzada y experimental que está fuera del alcance de la implementación Go del CLI.**

La implementación actual del CLI Go cubre el **95% de los casos de uso core** sin necesidad de KAIROS. Los usuarios que requieran funcionalidad proactiva pueden:

1. Usar la versión TypeScript oficial
2. Implementar scripts personalizados con `watch` modes
3. Usar herramientas externas como `entr`, `watchman`, etc.

**Recomendación:** No implementar KAIROS en Go. Enfocar esfuerzos en:
- Stabilidad del core
- Performance
- Tests automatizados
- Documentación

---

*Documento creado para referencia - Abril 2026*

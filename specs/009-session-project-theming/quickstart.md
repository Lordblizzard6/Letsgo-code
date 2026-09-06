# Quickstart: Validación de Proyectos, Sesiones y Temas

Guía rápida para validar el flujo completo de proyectos, sesiones y persistencia de temas.

## Requisitos Previos

- Entorno de desarrollo Go (≥ 1.22)
- Node.js (≥ 18) y npm instalados
- Repositorio Letsgo-code clonado localmente

## Pasos de Verificación

### 1. Compilación del Sistema
```bash
# Compilar frontend
cd cmd/wails/frontend
npm run build
cd ../../..

# Compilar ejecutable completo
go build -o letsgo.exe ./cmd/wails
```

### 2. Flujo de Validación 1: Proyectos sin Conversaciones
1. Iniciar `letsgo.exe`.
2. En la cabecera superior, hacer clic en el selector de proyecto y pulsar **"Abrir proyecto"**.
3. Seleccionar una carpeta vacía o un proyecto nuevo en disco.
4. **Resultado esperado**:
   - La carpeta aparece en la barra lateral en la sección de proyectos/árbol con su nombre (`basename`).
   - El contador muestra `0` conversaciones.
   - El proyecto permanece listado y visible.

### 3. Flujo de Validación 2: Creación de Nueva Conversación desde el Proyecto
1. En el proyecto recién listado en la barra lateral, hacer clic en el botón `+` ("Nueva conversación").
2. **Resultado esperado**:
   - Se abre de inmediato una nueva pestaña de chat vinculada al proyecto.
   - El árbol lateral actualiza el contador a `1` y muestra la sesión con el formato `[YYYY-MM-DD] <Nombre>` y su insignia de tiempo `(0m)`.
   - Escribir un mensaje en el chat; el mensaje se envía y persiste en SQLite.

### 4. Flujo de Validación 3: Despeje del Espacio de Chat
1. Observar el área central de chat y compositor.
2. **Resultado esperado**:
   - No hay botones redundantes de "nueva conversación" en el espacio de conversación que confundan la carpeta de trabajo.

### 5. Flujo de Validación 4: Persistencia de Temas y Modal de Configuración
1. Hacer clic en el icono de engranaje (Ajustes / Configuración).
2. Cambiar entre **Goulm**, **Oscuro** y **Claro**.
3. **Resultado esperado**:
   - La ventana modal de ajustes adopta de inmediato los colores de fondo, bordes y textos del tema seleccionado.
4. Cerrar la ventana de configuración y salir de `letsgo.exe`.
5. Volver a ejecutar `letsgo.exe`.
6. **Resultado esperado**:
   - La aplicación inicia directamente con el último tema seleccionado (sin volver al valor por defecto).

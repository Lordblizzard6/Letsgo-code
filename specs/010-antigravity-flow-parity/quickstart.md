# Quickstart: Validación de Flujo Antigravity y Opencode

Este documento describe los escenarios de verificación para comprobar la correcta implementación de la paridad de flujo con Antigravity y Opencode tanto en la interfaz GUI como en Bubbletea TUI.

---

## 1. Validación de Escenarios (GUI)

### Escenario A: Selector de Modelos en el Compositor (US1)
1. Iniciar la aplicación GUI: `./letsgo.exe` o `go run ./cmd/wails`.
2. Observar el área inferior donde se redactan los mensajes (`#chat-input-pane`).
3. Verificar que dentro de la caja del compositor aparece el selector de modelo (`[🤖 <modelo_actual> ▾]`) y el selector de modo (`[🛠️ code ▾]`).
4. Hacer clic en el selector de modelo: comprobar que se despliega la lista flotante de modelos.
5. Seleccionar un modelo alternativo (ej. `openai/gpt-4o-mini` o `anthropic/claude-3-5-haiku`).
6. Redactar una consulta simple y pulsar Enviar: verificar que el mensaje se procesa con el modelo seleccionado sin tener que acudir a la cabecera superior.

### Escenario B: Rebobinado / Rollback de Mensajes (US2)
1. En la sesión activa, enviar un mensaje: `"Pregunta 1: Dime tres colores primarios"`.
2. Esperar la respuesta del asistente.
3. Enviar un segundo mensaje: `"Pregunta 2: ¿Cuál de ellos es más cálido?"`.
4. Esperar la respuesta del asistente.
5. Localizar el primer mensaje de usuario ("Pregunta 1") y hacer clic en el botón de acción **↩ Rebobinar**.
6. **Resultado esperado**:
   - Los mensajes posteriores a la Pregunta 1 desaparecen del historial.
   - El texto `"Pregunta 1: Dime tres colores primarios"` se carga de nuevo en el área de texto del compositor listo para ser editado.
   - Al reiniciar o cambiar de pestaña y volver, la base de datos refleja la poda y no muestra los mensajes eliminados.

### Escenario C: Panel Lateral de Git Diff (US3)
1. Abrir un proyecto que sea un repositorio git con cambios sin confirmar.
2. Pulsar el botón de Git Diff en la barra superior o presionar `Ctrl+D`.
3. **Resultado esperado**:
   - Se abre el panel lateral derecho mostrando la lista de archivos modificados.
   - Al seleccionar un archivo, se renderiza el diff con líneas en verde (adiciones) y rojo (eliminaciones).
   - El panel puede colapsarse o cerrarse en cualquier momento sin perder el hilo del chat.

### Escenario D: Reubicación de Ajustes al Pie del Sidebar (US4)
1. Mirar la parte inferior de la barra lateral izquierda donde se listan los proyectos.
2. Comprobar que en el pie (`#sidebar-footer`) se encuentra el botón **⚙ Ajustes**.
3. Hacer clic en el botón: debe abrir el modal de `SettingsView` con todos los controles de configuración funcionales y adaptados a la paleta cromática del tema activo.

---

## 2. Validación de Escenarios (Bubbletea TUI)

### Escenario E: Comandos `/rollback` y `/diff` en TUI (US5)
1. Iniciar la terminal: `go run ./main.go` o `./letsgo.exe` en modo terminal.
2. Enviar un mensaje de prueba a la IA y esperar respuesta.
3. Escribir `/rollback` y presionar Enter.
   - **Resultado**: Se descarta la última respuesta y el texto de la pregunta vuelve al cuadro de edición.
4. Escribir `/diff` y presionar Enter.
   - **Resultado**: El viewport principal muestra las diferencias de Git del proyecto actual.

---

## 3. Pruebas Automatizadas

```bash
# 1. Compilación y pruebas unitarias de frontend
cd cmd/wails/frontend
npm run build
npm test

# 2. Pruebas de integración y backend en Go
cd ../../..
go test ./...

# 3. Compilación final del binario
go build -o letsgo.exe ./cmd/wails
```

package tools

import (
"bufio"
"context"
"encoding/json"
"fmt"
"os"
"path/filepath"
"strings"

"github.com/bmatcuk/doublestar/v4"
"github.com/tuusuario/mycli/internal/llm"
)

// RegisterSearchTools registra herramientas de búsqueda de código
func (r *Registry) RegisterSearchTools() {
r.Register(llm.Tool{
Name:        "search_files",
Description: "Busca archivos por patrón glob. Ej: '*.go', '**/*.ts', 'src/**/*.py'",
Schema:      json.RawMessage(`{"type":"object","properties":{"pattern":{"type":"string","description":"Patrón glob para buscar archivos"},"path":{"type":"string","description":"Directorio base para la búsqueda"}},"required":["pattern"]}`),
Category:    "search",
Handler: func(ctx context.Context, args json.RawMessage) (string, error) {
var params struct {
Pattern string `json:"pattern"`
Path    string `json:"path,omitempty"`
}
if err := json.Unmarshal(args, &params); err != nil {
return "", err
}

searchPath := params.Path
if searchPath == "" {
searchPath = "."
}

var matches []string
err := doublestar.GlobWalk(os.DirFS(searchPath), params.Pattern, func(path string, d os.DirEntry) error {
if !d.IsDir() {
fullPath := filepath.Join(searchPath, path)
matches = append(matches, fullPath)
}
return nil
})

if err != nil {
return "", fmt.Errorf("search error: %w", err)
}

if len(matches) == 0 {
return "No se encontraron archivos coincidentes", nil
}

var result strings.Builder
result.WriteString(fmt.Sprintf("Encontrados %d archivos:\n", len(matches)))
for _, m := range matches[:min(len(matches), 100)] {
result.WriteString("  " + m + "\n")
}
if len(matches) > 100 {
result.WriteString(fmt.Sprintf("... y %d más", len(matches)-100))
}

return result.String(), nil
},
})

r.Register(llm.Tool{
Name:        "grep_search",
Description: "Busca contenido en archivos usando grep. Ideal para encontrar código específico.",
Schema:      json.RawMessage(`{"type":"object","properties":{"pattern":{"type":"string","description":"Patrón regex a buscar"},"path":{"type":"string","description":"Directorio base"},"include":{"type":"string","description":"Patrón de archivos a incluir (ej: '*.go')"},"exclude":{"type":"string","description":"Patrón de archivos a excluir"},"ignoreCase":{"type":"boolean","description":"Ignorar mayúsculas/minúsculas"}},"required":["pattern"]}`),
Category:    "search",
Handler: func(ctx context.Context, args json.RawMessage) (string, error) {
var params struct {
Pattern    string `json:"pattern"`
Path       string `json:"path,omitempty"`
Include    string `json:"include,omitempty"`
Exclude    string `json:"exclude,omitempty"`
IgnoreCase bool   `json:"ignoreCase,omitempty"`
}
if err := json.Unmarshal(args, &params); err != nil {
return "", err
}

searchPath := params.Path
if searchPath == "" {
searchPath = "."
}

var results []string
count := 0
maxResults := 200

err := filepath.WalkDir(searchPath, func(path string, d os.DirEntry, err error) error {
if err != nil || ctx.Err() != nil {
return err
}

if d.IsDir() {
// Skip common non-code directories
name := d.Name()
if name == "node_modules" || name == ".git" || name == "vendor" || 
   name == "__pycache__" || name == ".next" || name == "dist" ||
   name == "build" || name == "target" {
return filepath.SkipDir
}
return nil
}

// Check include pattern
if params.Include != "" {
match, _ := doublestar.Match(params.Include, d.Name())
if !match {
return nil
}
}

// Check exclude pattern
if params.Exclude != "" {
match, _ := doublestar.Match(params.Exclude, d.Name())
if match {
return nil
}
}

// Read file and search
file, err := os.Open(path)
if err != nil {
return nil
}
defer file.Close()

scanner := bufio.NewScanner(file)
lineNum := 0
for scanner.Scan() {
lineNum++
line := scanner.Text()

matched := false
if params.IgnoreCase {
matched = strings.Contains(strings.ToLower(line), strings.ToLower(params.Pattern))
} else {
matched = strings.Contains(line, params.Pattern)
}

if matched {
results = append(results, fmt.Sprintf("%s:%d: %s", path, lineNum, strings.TrimSpace(line)))
count++
if count >= maxResults {
return fmt.Errorf("max results reached")
}
}
}

return nil
})

if err != nil && err.Error() != "max results reached" {
return "", fmt.Errorf("search error: %w", err)
}

if len(results) == 0 {
return "No se encontraron coincidencias", nil
}

var result strings.Builder
result.WriteString(fmt.Sprintf("Coincidencias (%d):\n", len(results)))
for _, r := range results {
result.WriteString(r + "\n")
}
if err != nil && err.Error() == "max results reached" {
result.WriteString(fmt.Sprintf("\n... mostrando solo los primeros %d resultados", maxResults))
}

return result.String(), nil
},
})

r.Register(llm.Tool{
Name:        "read_file_range",
Description: "Lee un rango específico de líneas de un archivo. Útil para archivos grandes.",
Schema:      json.RawMessage(`{"type":"object","properties":{"path":{"type":"string","description":"Ruta del archivo"},"start":{"type":"integer","description":"Línea de inicio (1-based)"},"end":{"type":"integer","description":"Línea de fin (inclusive)"}},"required":["path","start","end"]}`),
Category:    "search",
Handler: func(ctx context.Context, args json.RawMessage) (string, error) {
var params struct {
Path  string `json:"path"`
Start int    `json:"start"`
End   int    `json:"end"`
}
if err := json.Unmarshal(args, &params); err != nil {
return "", err
}

if params.Start < 1 {
return "", fmt.Errorf("start debe ser >= 1")
}
if params.End < params.Start {
return "", fmt.Errorf("end debe ser >= start")
}

file, err := os.Open(params.Path)
if err != nil {
return "", err
}
defer file.Close()

var result strings.Builder
scanner := bufio.NewScanner(file)
lineNum := 0
for scanner.Scan() {
lineNum++
if lineNum < params.Start {
continue
}
if lineNum > params.End {
break
}
result.WriteString(fmt.Sprintf("%4d: %s\n", lineNum, scanner.Text()))
}

if result.Len() == 0 {
return "Rango fuera del archivo", nil
}

return result.String(), nil
},
})
}

func min(a, b int) int {
if a < b {
return a
}
return b
}

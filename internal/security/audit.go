package security

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuditLogger gestiona el log de auditoría
type AuditLogger struct {
	filePath string
	file     *os.File
	mu       sync.Mutex
	queue    []AuditEntry
	maxQueue int
}

// NewAuditLogger crea un nuevo logger de auditoría
func NewAuditLogger(path string) (*AuditLogger, error) {
	// Crear directorio si no existe
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	
	// Abrir archivo en modo append
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	
	return &AuditLogger{
		filePath: path,
		file:     file,
		queue:    make([]AuditEntry, 0),
		maxQueue: 100,
	}, nil
}

// Log añade una entrada al log
func (al *AuditLogger) Log(entry AuditEntry) {
	al.mu.Lock()
	defer al.mu.Unlock()
	
	// Añadir timestamp si no existe
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	
	al.queue = append(al.queue, entry)
	
	// Flush si queue está llena
	if len(al.queue) >= al.maxQueue {
		al.flush()
	}
}

// Flush escribe el queue al archivo
func (al *AuditLogger) Flush() error {
	al.mu.Lock()
	defer al.mu.Unlock()
	return al.flush()
}

func (al *AuditLogger) flush() error {
	if len(al.queue) == 0 {
		return nil
	}
	
	for _, entry := range al.queue {
		data, err := json.Marshal(entry)
		if err != nil {
			continue
		}
		
		if _, err := al.file.Write(append(data, '\n')); err != nil {
			return err
		}
	}
	
	al.queue = make([]AuditEntry, 0)
	return al.file.Sync()
}

// Close cierra el logger y hace flush final
func (al *AuditLogger) Close() error {
	al.mu.Lock()
	defer al.mu.Unlock()
	
	// Flush final
	if err := al.flush(); err != nil {
		return err
	}
	
	return al.file.Close()
}

// Rotate rota el archivo de log
func (al *AuditLogger) Rotate() error {
	al.mu.Lock()
	defer al.mu.Unlock()
	
	// Flush actual
	if err := al.flush(); err != nil {
		return err
	}
	
	// Cerrar archivo actual
	if err := al.file.Close(); err != nil {
		return err
	}
	
	// Rotar archivo
	timestamp := time.Now().Format("20060102_150405")
	rotatedPath := al.filePath + "." + timestamp
	if err := os.Rename(al.filePath, rotatedPath); err != nil {
		// Si falla, intentar reopen
		var err2 error
		al.file, err2 = os.OpenFile(al.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		return err2
	}
	
	// Abrir nuevo archivo
	var err error
	al.file, err = os.OpenFile(al.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	return err
}

// GetRecentEntries devuelve las últimas N entradas
func (al *AuditLogger) GetRecentEntries(n int) ([]AuditEntry, error) {
	al.mu.Lock()
	defer al.mu.Unlock()
	
	// Leer archivo
	data, err := os.ReadFile(al.filePath)
	if err != nil {
		return nil, err
	}
	
	lines := splitLines(data)
	if len(lines) == 0 {
		return []AuditEntry{}, nil
	}
	
	// Tomar últimas n líneas
	start := len(lines) - n
	if start < 0 {
		start = 0
	}
	
	entries := make([]AuditEntry, 0, n)
	for i := start; i < len(lines); i++ {
		var entry AuditEntry
		if err := json.Unmarshal([]byte(lines[i]), &entry); err != nil {
			continue
		}
		entries = append(entries, entry)
	}
	
	// Añadir queue actual
	queueStart := 0
	if len(al.queue) > n-len(entries) {
		queueStart = len(al.queue) - (n - len(entries))
	}
	entries = append(entries, al.queue[queueStart:]...)
	
	return entries, nil
}

// SearchEntries busca entradas por criterio
func (al *AuditLogger) SearchEntries(filter func(AuditEntry) bool) ([]AuditEntry, error) {
	al.mu.Lock()
	defer al.mu.Unlock()
	
	var results []AuditEntry
	
	// Buscar en archivo
	data, err := os.ReadFile(al.filePath)
	if err == nil {
		lines := splitLines(data)
		for _, line := range lines {
			var entry AuditEntry
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				continue
			}
			if filter(entry) {
				results = append(results, entry)
			}
		}
	}
	
	// Buscar en queue
	for _, entry := range al.queue {
		if filter(entry) {
			results = append(results, entry)
		}
	}
	
	return results, nil
}

func splitLines(data []byte) []string {
	var lines []string
	start := 0
	for i, b := range data {
		if b == '\n' {
			if i > start {
				lines = append(lines, string(data[start:i]))
			}
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, string(data[start:]))
	}
	return lines
}

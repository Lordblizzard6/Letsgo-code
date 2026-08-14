package tui

import (
	"strings"
	"testing"

	"github.com/user/go-claude-code/internal/db"
)

func initSearchDB(t *testing.T) string {
	t.Helper()
	if db.DB != nil {
		_ = db.DB.Close()
		db.DB = nil
	}
	t.Setenv("TEST_DB_PATH", t.TempDir()+"/history.db")
	t.Cleanup(func() {
		if db.DB != nil {
			_ = db.DB.Close()
			db.DB = nil
		}
	})
	if err := db.InitDB(); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	sid, err := db.CreateSession("test", "/tmp")
	if err != nil {
		t.Fatal(err)
	}
	return sid
}

func TestHandleSearchFound(t *testing.T) {
	sid := initSearchDB(t)
	_ = db.SaveMessage(sid, "user", "searchable term appears")
	res := handleSearch("searchable")
	if !strings.Contains(res, "Results for 'searchable'") {
		t.Fatalf("unexpected response: %s", res)
	}
}

func TestHandleSearchNotFound(t *testing.T) {
	sid := initSearchDB(t)
	_ = db.SaveMessage(sid, "user", "nothing useful")
	res := handleSearch("inexistente")
	if !strings.Contains(res, "No results") {
		t.Fatalf("unexpected response: %s", res)
	}
}

func TestHandleSearchWithLimit(t *testing.T) {
	sid := initSearchDB(t)
	_ = db.SaveMessage(sid, "user", "alpha")
	_ = db.SaveMessage(sid, "user", "alpha two")
	_ = db.SaveMessage(sid, "user", "alpha three")
	res := handleSearch("alpha")
	if !strings.Contains(res, "Results for 'alpha'") {
		t.Fatalf("unexpected response: %s", res)
	}
}

// TestSlashCommandsAndSuggestions (T024): the core slash commands respond and
// typing "/" surfaces autocomplete suggestions (FR-004, SC-004).
func TestSlashCommandsAndSuggestions(t *testing.T) {
	sid := initSearchDB(t)
	SetSlashCommandSessionID(sid)
	_ = db.SaveMessage(sid, "user", "mensaje de prueba")
	_ = db.SaveMessage(sid, "assistant", "respuesta")

	cases := []struct {
		cmd     string
		expects string
	}{
		{"/help", "Command Reference"},
		{"/model", "Current"},
		{"/cost", "cost"},
		{"/tokens", "Token"},
		{"/clear", "limpiada"},
		{"/quit", "Goodbye"},
	}
	for _, c := range cases {
		res, cont, quit := ProcessSlashCommand(c.cmd)
		if c.cmd == "/quit" {
			if cont {
				t.Fatalf("T024: /quit must not continue the chat")
			}
			if !quit {
				t.Fatalf("T024: /quit must request quit")
			}
		} else {
			if !cont {
				t.Fatalf("T024: %s must continue the chat", c.cmd)
			}
			if quit {
				t.Fatalf("T024: %s must not quit", c.cmd)
			}
		}
		if !strings.Contains(res, c.expects) {
			t.Fatalf("T024: %s response %q must contain %q", c.cmd, res, c.expects)
		}
	}

	// Suggestions appear when typing "/m".
	m := model{input: "/m", suggestions: []string{}, showSuggestions: false}
	m.updateSuggestions()
	if !m.showSuggestions || len(m.suggestions) == 0 {
		t.Fatal("T024: typing /m must surface autocomplete suggestions")
	}
	if !strings.Contains(strings.Join(m.suggestions, " "), "/model") {
		t.Fatalf("T024: suggestions must include /model, got %v", m.suggestions)
	}
}

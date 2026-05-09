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
	res := handleSearch("alpha limit=2")
	if !strings.Contains(res, "page 1/2") {
		t.Fatalf("expected pagination in response, got: %s", res)
	}
}

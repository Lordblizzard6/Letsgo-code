package db

import (
	"os"
	"testing"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	if DB != nil {
		_ = DB.Close()
		DB = nil
	}
	t.Setenv("TEST_DB_PATH", t.TempDir()+"/history.db")
	if err := InitDB(); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	t.Cleanup(func() {
		if DB != nil {
			_ = DB.Close()
			DB = nil
		}
		_ = os.Unsetenv("TEST_DB_PATH")
	})
}

func TestSearchMessagesFoundAndPagination(t *testing.T) {
	setupTestDB(t)
	sid, err := CreateSession("s", "/tmp")
	if err != nil { t.Fatal(err) }
	_ = SaveMessage(sid, "user", "hello alpha world")
	_ = SaveMessage(sid, "assistant", "beta alpha response")
	_ = SaveMessage(sid, "user", "gamma alpha data")

	results, total, err := SearchMessages(sid, "alpha", 2, 0)
	if err != nil { t.Fatal(err) }
	if total != 3 { t.Fatalf("want total 3 got %d", total) }
	if len(results) != 2 { t.Fatalf("want 2 results got %d", len(results)) }

	results2, total2, err := SearchMessages(sid, "alpha", 2, 2)
	if err != nil { t.Fatal(err) }
	if total2 != 3 || len(results2) != 1 {
		t.Fatalf("want total 3 and len 1 got total=%d len=%d", total2, len(results2))
	}
}

func TestSearchMessagesNoResults(t *testing.T) {
	setupTestDB(t)
	sid, _ := CreateSession("s", "/tmp")
	_ = SaveMessage(sid, "user", "hello world")

	results, total, err := SearchMessages(sid, "missing", 5, 0)
	if err != nil { t.Fatal(err) }
	if total != 0 || len(results) != 0 {
		t.Fatalf("expected 0 results, got total=%d len=%d", total, len(results))
	}
}

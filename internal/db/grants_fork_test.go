package db

import "testing"

func TestGrantCategoryCreatesUniqueGrant(t *testing.T) {
	setupTestDB(t)

	if err := GrantCategory("sess_1", "bash"); err != nil {
		t.Fatalf("GrantCategory: %v", err)
	}
	// Upsert same pair twice is idempotent.
	if err := GrantCategory("sess_1", "bash"); err != nil {
		t.Fatalf("GrantCategory second: %v", err)
	}

	grants, err := ListGrants("sess_1")
	if err != nil {
		t.Fatalf("ListGrants: %v", err)
	}
	if len(grants) != 1 {
		t.Fatalf("expected 1 grant, got %d", len(grants))
	}
	if grants[0].Category != "bash" {
		t.Fatalf("unexpected category %q", grants[0].Category)
	}

	ok, err := IsGranted("sess_1", "bash")
	if err != nil || !ok {
		t.Fatalf("IsGranted(bash) = %v,%v; want true", ok, err)
	}
}

func TestGrantDoesNotTrascendSessions(t *testing.T) {
	setupTestDB(t)

	if err := GrantCategory("sess_a", "file-edit"); err != nil {
		t.Fatal(err)
	}
	ok, _ := IsGranted("sess_b", "file-edit")
	if ok {
		t.Fatal("grant leaked to another session")
	}
}

func TestRevokeCategory(t *testing.T) {
	setupTestDB(t)
	if err := GrantCategory("sess_1", "bash"); err != nil {
		t.Fatal(err)
	}
	if err := RevokeCategory("sess_1", "bash"); err != nil {
		t.Fatal(err)
	}
	ok, _ := IsGranted("sess_1", "bash")
	if ok {
		t.Fatal("grant still active after revoke")
	}
}

func TestForkSessionCopiesUpToMessageID(t *testing.T) {
	setupTestDB(t)

	sid, err := CreateSession("original", "/tmp/proj")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		if err := SaveMessage(sid, "user", "msg-"+string(rune('a'+i))); err != nil {
			t.Fatal(err)
		}
	}
	history, err := GetHistory(sid)
	if err != nil || len(history) != 5 {
		t.Fatalf("history length %d, err %v", len(history), err)
	}
	// Fork at the 3rd message inclusive.
	target := history[2].ID
	forkID, err := ForkSession(sid, target)
	if err != nil {
		t.Fatal(err)
	}
	if forkID == sid {
		t.Fatal("fork produced the same session id")
	}

	forkHistory, _ := GetHistory(forkID)
	if len(forkHistory) != 3 {
		t.Fatalf("fork history = %d, want 3", len(forkHistory))
	}

	// Original must be untouched.
	orig, _ := GetHistory(sid)
	if len(orig) != 5 {
		t.Fatalf("original mutated: %d messages", len(orig))
	}
}

func TestForkSessionDoesNotCopyGrants(t *testing.T) {
	setupTestDB(t)

	sid, _ := CreateSession("src", "/tmp")
	if err := SaveMessage(sid, "user", "hello"); err != nil {
		t.Fatal(err)
	}
	_ = GrantCategory(sid, "bash")
	hist, _ := GetHistory(sid)

	forkID, err := ForkSession(sid, hist[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	ok, _ := IsGranted(forkID, "bash")
	if ok {
		t.Fatal("fork inherited a grant; grants must not trascend")
	}
}

func TestSaveMessageKindPersistsKind(t *testing.T) {
	setupTestDB(t)
	sid, err := CreateSession("k", "/tmp")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := SaveMessageKind(sid, "user", "steer text", "user:steer"); err != nil {
		t.Fatalf("SaveMessageKind: %v", err)
	}
	h, err := GetHistory(sid)
	if err != nil || len(h) != 1 {
		t.Fatalf("hist %v, err %v", h, err)
	}
	if h[0].Kind != "user:steer" {
		t.Fatalf("kind = %q, want user:steer", h[0].Kind)
	}
	if h[0].Content != "steer text" {
		t.Fatalf("content = %#v, want steer text", h[0].Content)
	}
}
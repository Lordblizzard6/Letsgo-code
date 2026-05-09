package db

import (
	"testing"
)

func TestInitDB(t *testing.T) {
	setupTestDB(t)
	if DB == nil {
		t.Fatal("expected DB to be initialized")
	}
}

func TestCreateSession(t *testing.T) {
	setupTestDB(t)
	
	sessionID, err := CreateSession("test-session", "/tmp/test")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	
	if sessionID == "" {
		t.Error("CreateSession returned empty sessionID")
	}
	
	// Verify session starts with "sess_"
	if len(sessionID) < 5 || sessionID[:5] != "sess_" {
		t.Errorf("Session ID should start with 'sess_', got: %s", sessionID)
	}
}

func TestSessionStructs(t *testing.T) {
	session := Session{
		ID:          "sess_123",
		Name:        "Test Session",
		ProjectPath: "/tmp/project",
		IsActive:    true,
	}
	
	if session.ID != "sess_123" {
		t.Errorf("Expected ID 'sess_123', got '%s'", session.ID)
	}
	
	if session.Name != "Test Session" {
		t.Errorf("Expected Name 'Test Session', got '%s'", session.Name)
	}
	
	if !session.IsActive {
		t.Error("Expected IsActive to be true")
	}
}

func TestMessageStruct(t *testing.T) {
	msg := Message{
		ID:      1,
		Role:    "user",
		Content: "Hello",
	}
	
	if msg.ID != 1 {
		t.Errorf("Expected ID 1, got %d", msg.ID)
	}
	
	if msg.Role != "user" {
		t.Errorf("Expected Role 'user', got '%s'", msg.Role)
	}
}

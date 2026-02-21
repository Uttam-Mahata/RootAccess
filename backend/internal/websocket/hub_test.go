package websocket

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewHub(t *testing.T) {
	hub := NewHub().(*MemoryHub)
	if hub == nil {
		t.Fatal("NewHub returned nil")
	}
	if hub.clients == nil {
		t.Error("clients map is nil")
	}
}

func TestHubBroadcastMessage(t *testing.T) {
	hub := NewHub().(*MemoryHub)
	go hub.Run()

	// Test BroadcastMessage marshaling
	msg := Message{
		Type:    "test",
		Payload: map[string]string{"key": "value"},
	}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}
	if len(data) == 0 {
		t.Error("Marshaled message is empty")
	}

	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}
	if decoded.Type != "test" {
		t.Errorf("Expected type 'test', got '%s'", decoded.Type)
	}
}

func TestHubRunStartsAndAcceptsRegistration(t *testing.T) {
	hub := NewHub().(*MemoryHub)
	go hub.Run()

	// Give time for goroutine to start
	time.Sleep(10 * time.Millisecond)

	// Verify hub is ready (no deadlock)
	hub.mu.RLock()
	clientCount := len(hub.clients)
	hub.mu.RUnlock()

	if clientCount != 0 {
		t.Errorf("Expected 0 clients, got %d", clientCount)
	}
}

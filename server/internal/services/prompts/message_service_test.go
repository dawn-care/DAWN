package prompts

import (
	"database/sql"
	"testing"

	"server/internal/models/prompts"
)

func TestMessageServiceCRUD(t *testing.T) {
	svc := NewMessageService(&sql.DB{})

	// Create
	msg := &prompts.MessagePrompt{Message: "Hello"}
	if err := svc.CreateMessage(msg); err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	// Get
	_, err := svc.GetMessageByID(1)
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}

	// Update
	if err := svc.UpdateMessage(&prompts.MessagePrompt{ID: 1, Message: "Updated"}); err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}

	// Delete
	if err := svc.DeleteMessage(1); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}
}

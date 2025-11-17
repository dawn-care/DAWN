package prompts_test

import (
	"database/sql"
	"testing"

	models "server/internal/models/modules/prompts"
	service "server/internal/services/modules/prompts"
)

type mockMessageDB struct{}

func (m *mockMessageDB) CreateMessage(message *models.MessagePrompt) error { return nil }
func (m *mockMessageDB) GetMessageByID(id int) (*models.MessagePrompt, error) {
	return &models.MessagePrompt{ID: id, Message: "Test"}, nil
}
func (m *mockMessageDB) UpdateMessage(message *models.MessagePrompt) error { return nil }
func (m *mockMessageDB) DeleteMessage(id int) error                        { return nil }

func TestMessageServiceCRUD(t *testing.T) {
	svc := service.NewMessageService(&sql.DB{})

	// Create
	msg := &models.MessagePrompt{Message: "Hello"}
	if err := svc.CreateMessage(msg); err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	// Get
	_, err := svc.GetMessageByID(1)
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}

	// Update
	if err := svc.UpdateMessage(&models.MessagePrompt{ID: 1, Message: "Updated"}); err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}

	// Delete
	if err := svc.DeleteMessage(1); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}
}

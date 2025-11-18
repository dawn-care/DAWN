package prompts

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	models "server/internal/models/prompts"
)

type mockMessageService struct{}

func (m *mockMessageService) CreateMessage(msg *models.MessagePrompt) error { return nil }
func (m *mockMessageService) GetMessageByID(id int) (*models.MessagePrompt, error) {
	return &models.MessagePrompt{ID: id, Message: "Test"}, nil
}
func (m *mockMessageService) UpdateMessage(msg *models.MessagePrompt) error { return nil }
func (m *mockMessageService) DeleteMessage(id int) error                    { return nil }

func TestMessageControllerHandlers(t *testing.T) {
	controller := NewMessageController(&mockMessageService{})

	// Create
	body := `{"message":"Hello"}`
	req := httptest.NewRequest(http.MethodPost, "/messages", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	controller.CreateMessageHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Get
	req = httptest.NewRequest(http.MethodGet, "/messages?id=1", nil)
	w = httptest.NewRecorder()
	controller.GetMessageByIDHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Update
	msg := models.MessagePrompt{ID: 1, Message: "Updated"}
	bodyBytes, _ := json.Marshal(msg)
	req = httptest.NewRequest(http.MethodPut, "/messages", bytes.NewBuffer(bodyBytes))
	w = httptest.NewRecorder()
	controller.UpdateMessageHandler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Delete
	req = httptest.NewRequest(http.MethodDelete, "/messages?id=1", nil)
	w = httptest.NewRecorder()
	controller.DeleteMessageHandler(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

package prompts

import (
	"encoding/json"
	"net/http"
	"strconv"

	models "server/internal/models/modules/prompts"
)

type MessageServiceInterface interface {
	CreateMessage(message *models.MessagePrompt) error
	GetMessageByID(id int) (*models.MessagePrompt, error)
	UpdateMessage(message *models.MessagePrompt) error
	DeleteMessage(id int) error
}

type MessageController struct {
	service MessageServiceInterface
}

func NewMessageController(service MessageServiceInterface) *MessageController {
	return &MessageController{service: service}
}

// Create a new message
func (c *MessageController) CreateMessageHandler(w http.ResponseWriter, r *http.Request) {
	var msg models.MessagePrompt
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.service.CreateMessage(&msg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(msg)
}

// Get a message by ID
func (c *MessageController) GetMessageByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "invalid message ID", http.StatusBadRequest)
		return
	}

	msg, err := c.service.GetMessageByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(msg)
}

// Update a message
func (c *MessageController) UpdateMessageHandler(w http.ResponseWriter, r *http.Request) {
	var msg models.MessagePrompt
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if msg.ID <= 0 {
		http.Error(w, "message ID is required", http.StatusBadRequest)
		return
	}

	if err := c.service.UpdateMessage(&msg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(msg)
}

// Delete a message by ID
func (c *MessageController) DeleteMessageHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "invalid message ID", http.StatusBadRequest)
		return
	}

	if err := c.service.DeleteMessage(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

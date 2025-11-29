package prompts

import (
	"encoding/json"
	"net/http"
	"strconv"

	"server/internal/models/prompts"
)

type MessageServiceInterface interface {
	CreateMessage(message *prompts.MessagePrompt) error
	GetMessageByID(id int) (*prompts.MessagePrompt, error)
	UpdateMessage(message *prompts.MessagePrompt) error
	DeleteMessage(id int) error
}

type MessageController struct {
	service MessageServiceInterface
}

func NewMessageController(service MessageServiceInterface) *MessageController {
	return &MessageController{service: service}
}

// CreateMessageHandler godoc
// @Summary Create a new message
// @Description Creates a new message prompt
// @Tags message
// @Accept json
// @Produce json
// @Param message body prompts.MessagePrompt true "Message prompt to create"
// @Success 200 {object} prompts.MessagePrompt
// @Failure 400 {string} string "invalid request body"
// @Failure 500 {string} string "internal server error"
// @Router /messages [post]
func (c *MessageController) CreateMessageHandler(w http.ResponseWriter, r *http.Request) {
	var msg prompts.MessagePrompt
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

// GetMessageByIDHandler godoc
// @Summary Get a message by ID
// @Description Retrieves a single message prompt by its ID
// @Tags message
// @Accept json
// @Produce json
// @Param id query int true "Message ID"
// @Success 200 {object} prompts.MessagePrompt
// @Failure 400 {string} string "invalid message ID"
// @Failure 500 {string} string "internal server error"
// @Router /messages [get]
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

// UpdateMessageHandler godoc
// @Summary Update a message
// @Description Updates an existing message prompt
// @Tags message
// @Accept json
// @Produce json
// @Param message body prompts.MessagePrompt true "Message prompt to update (must include ID)"
// @Success 200 {object} prompts.MessagePrompt
// @Failure 400 {string} string "invalid request body or missing message ID"
// @Failure 500 {string} string "internal server error"
// @Router /messages [put]
func (c *MessageController) UpdateMessageHandler(w http.ResponseWriter, r *http.Request) {
	var msg prompts.MessagePrompt
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

// DeleteMessageHandler godoc
// @Summary Delete a message
// @Description Deletes a single message prompt by ID
// @Tags message
// @Accept json
// @Produce json
// @Param id query int true "Message ID"
// @Success 204 {string} string "No Content"
// @Failure 400 {string} string "invalid message ID"
// @Failure 500 {string} string "internal server error"
// @Router /messages [delete]
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

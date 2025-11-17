package prompts

import (
	"database/sql"
	"errors"
	"fmt"

	models "server/internal/models/modules/prompts"
)

type MessageService struct {
	DB *sql.DB
}

func NewMessageService(db *sql.DB) *MessageService {
	return &MessageService{DB: db}
}

// Create a new message
func (s *MessageService) CreateMessage(message *models.MessagePrompt) error {
	if message == nil {
		return errors.New("message cannot be nil")
	}
	if message.Message == "" {
		return errors.New("message text is required")
	}
	return models.CreateMessage(s.DB, message)
}

// Get a message by ID
func (s *MessageService) GetMessageByID(id int) (*models.MessagePrompt, error) {
	if id <= 0 {
		return nil, errors.New("invalid message ID")
	}
	msg, err := models.GetMessageByID(s.DB, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	if msg == nil {
		return nil, fmt.Errorf("message not found with ID %d", id)
	}
	return msg, nil
}

// Update a message
func (s *MessageService) UpdateMessage(message *models.MessagePrompt) error {
	if message == nil {
		return errors.New("message cannot be nil")
	}
	if message.ID <= 0 {
		return errors.New("invalid message ID")
	}
	return models.UpdateMessagePrompt(s.DB, message)
}

// Delete a message by ID
func (s *MessageService) DeleteMessage(id int) error {
	if id <= 0 {
		return errors.New("invalid message ID")
	}
	return models.DeleteMessage(s.DB, id)
}

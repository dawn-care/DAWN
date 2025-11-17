package prompts

import (
	"database/sql"
	"errors"
	"fmt"

	"server/internal/models/modules/prompts"
)

type CaregiverLocationService struct {
	DB *sql.DB
}

func NewCaregiverLocationService(db *sql.DB) *CaregiverLocationService {
	return &CaregiverLocationService{DB: db}
}

// Create a single location
func (s *CaregiverLocationService) CreateCaregiverLocation(location *prompts.CaregiverLocationPrompt) error {
	if location == nil {
		return errors.New("location cannot be nil")
	}
	if location.Location == "" || location.Message == "" {
		return errors.New("location and message are required")
	}
	return prompts.CreateCaregiverLocation(s.DB, location)
}

// Get a specific location by ID
func (s *CaregiverLocationService) GetCaregiverLocationByID(id int) (*prompts.CaregiverLocationPrompt, error) {
	if id <= 0 {
		return nil, errors.New("invalid cargiver location ID")
	}
	loc, err := prompts.GetCaregiverLocationByID(s.DB, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get caregiver location: %w", err)
	}
	if loc == nil {
		return nil, fmt.Errorf("caregiver location with id %d not found", id)
	}
	return loc, nil
}

// Update a specific location
func (s *CaregiverLocationService) UpdateCaregiverLocation(location *prompts.CaregiverLocationPrompt) error {
	if location == nil {
		return errors.New("caregiver location cannot be nil")
	}
	if location.ID <= 0 {
		return errors.New("invalid caregiver location ID")
	}
	return prompts.UpdateCaregiverLocation(s.DB, location)
}

// Delete a specific location
func (s *CaregiverLocationService) DeleteCaregiverLocation(id int) error {
	if id <= 0 {
		return errors.New("invalid caregiver location ID")
	}
	return prompts.DeleteCaregiverLocation(s.DB, id)
}

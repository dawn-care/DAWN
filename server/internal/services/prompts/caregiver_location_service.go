package prompts

import (
	"database/sql"
	"errors"
	"fmt"

	prompts "server/internal/models/prompts"
)

type CaregiverLocationService struct {
	DB *sql.DB
}

func NewCaregiverLocationService(db *sql.DB) *CaregiverLocationService {
	return &CaregiverLocationService{DB: db}
}

// Create a new caregiver location
func (s *CaregiverLocationService) CreateCaregiverLocation(loc *prompts.CaregiverLocationPrompt) error {
	if loc == nil {
		return errors.New("caregiver location cannot be nil")
	}
	if loc.Location == "" {
		return errors.New("location is required")
	}

	return prompts.CreateCaregiverLocation(s.DB, loc)
}

// Retrieve caregiver location by ID
func (s *CaregiverLocationService) GetCaregiverLocationByID(id int) (*prompts.CaregiverLocationPrompt, error) {
	if id <= 0 {
		return nil, errors.New("invalid caregiver location ID")
	}

	loc, err := prompts.GetCaregiverLocationByID(s.DB, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve caregiver location: %w", err)
	}

	// loc may be nil if not found — let controller handle 404
	return loc, nil
}

// Update caregiver location
func (s *CaregiverLocationService) UpdateCaregiverLocation(loc *prompts.CaregiverLocationPrompt) error {
	if loc == nil {
		return errors.New("caregiver location cannot be nil")
	}
	if loc.ID <= 0 {
		return errors.New("invalid caregiver location ID")
	}

	return prompts.UpdateCaregiverLocation(s.DB, loc)
}

// Delete caregiver location by ID
func (s *CaregiverLocationService) DeleteCaregiverLocation(id int) error {
	if id <= 0 {
		return errors.New("invalid caregiver location ID")
	}

	return prompts.DeleteCaregiverLocation(s.DB, id)
}

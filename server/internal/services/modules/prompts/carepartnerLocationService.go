package prompts

import (
	"database/sql"
	"errors"
	"fmt"

	"server/internal/models/modules/prompts"
)

type CarepartnerLocationService struct {
	DB *sql.DB
}

func NewCarepartnerLocationService(db *sql.DB) *CarepartnerLocationService {
	return &CarepartnerLocationService{DB: db}
}

// Create a single location
func (s *CarepartnerLocationService) CreateCarepartnerLocation(location *prompts.CarepartnerLocationPrompt) error {
	if location == nil {
		return errors.New("location cannot be nil")
	}
	if location.Location == "" || location.Message == "" {
		return errors.New("location and message are required")
	}
	return prompts.CreateCarepartnerLocation(s.DB, location)
}

// Get a specific location by ID
func (s *CarepartnerLocationService) GetCarepartnerLocationByID(id int) (*prompts.CarepartnerLocationPrompt, error) {
	if id <= 0 {
		return nil, errors.New("invalid carepartner location ID")
	}
	loc, err := prompts.GetCarepartnerLocationByID(s.DB, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get carepartner location: %w", err)
	}
	if loc == nil {
		return nil, fmt.Errorf("carepartner location with id %d not found", id)
	}
	return loc, nil
}

// Update a specific location
func (s *CarepartnerLocationService) UpdateCarepartnerLocation(location *prompts.CarepartnerLocationPrompt) error {
	if location == nil {
		return errors.New("carepartner location cannot be nil")
	}
	if location.ID <= 0 {
		return errors.New("invalid carepartner location ID")
	}
	return prompts.UpdateCarepartnerLocation(s.DB, location)
}

// Delete a specific location
func (s *CarepartnerLocationService) DeleteCarepartnerLocation(id int) error {
	if id <= 0 {
		return errors.New("invalid carepartner location ID")
	}
	return prompts.DeleteCarepartnerLocation(s.DB, id)
}

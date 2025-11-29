package prompts

import (
	"database/sql"
	"errors"
	"fmt"

	prompts "server/internal/models/prompts"
)

type CarepartnerLocationService struct {
	DB *sql.DB
}

func NewCarepartnerLocationService(db *sql.DB) *CarepartnerLocationService {
	return &CarepartnerLocationService{DB: db}
}

// Create a carepartner location
func (s *CarepartnerLocationService) CreateCarepartnerLocation(loc *prompts.CarepartnerLocationPrompt) error {
	if loc == nil {
		return errors.New("carepartner location cannot be nil")
	}
	if loc.Location == "" {
		return errors.New("location is required")
	}

	return prompts.CreateCarepartnerLocation(s.DB, loc)
}

// Retrieve a carepartner location by ID
func (s *CarepartnerLocationService) GetCarepartnerLocationByID(id int) (*prompts.CarepartnerLocationPrompt, error) {
	if id <= 0 {
		return nil, errors.New("invalid carepartner location ID")
	}

	loc, err := prompts.GetCarepartnerLocationByID(s.DB, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve carepartner location: %w", err)
	}

	return loc, nil
}

// Update a carepartner location
func (s *CarepartnerLocationService) UpdateCarepartnerLocation(loc *prompts.CarepartnerLocationPrompt) error {
	if loc == nil {
		return errors.New("carepartner location cannot be nil")
	}
	if loc.ID <= 0 {
		return errors.New("invalid carepartner location ID")
	}

	return prompts.UpdateCarepartnerLocation(s.DB, loc)
}

// Delete a carepartner location
func (s *CarepartnerLocationService) DeleteCarepartnerLocation(id int) error {
	if id <= 0 {
		return errors.New("invalid carepartner location ID")
	}

	return prompts.DeleteCarepartnerLocation(s.DB, id)
}

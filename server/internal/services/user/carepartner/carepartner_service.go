package carepartner

import (
	"database/sql"
	"errors"

	"server/internal/models/screen_settings"
	"server/internal/models/user/carepartner"
)

type CarepartnerService struct {
	DB *sql.DB
}

func NewCarepartnerService(db *sql.DB) *CarepartnerService {
	return &CarepartnerService{DB: db}
}

// ─────────────────────────────────────────────
// CREATE
// ─────────────────────────────────────────────

// CreateCarepartner creates the carepartner and their default screen settings together
func (s *CarepartnerService) CreateCarepartner(c *carepartner.Carepartner) error {
	if c == nil {
		return errors.New("carepartner is nil")
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := carepartner.CreateCarepartner(tx, c); err != nil {
		return err
	}
	if err := screen_settings.CreateDefaultScreenSettings(tx, c.ID); err != nil {
		return err
	}

	return tx.Commit()
}

// ─────────────────────────────────────────────
// READ
// ─────────────────────────────────────────────

func (s *CarepartnerService) GetCarepartnerByID(id int) (*carepartner.Carepartner, error) {
	if id <= 0 {
		return nil, errors.New("invalid carepartner ID")
	}
	return carepartner.GetCarepartnerByID(s.DB, id)
}

func (s *CarepartnerService) GetAllCarepartners() ([]carepartner.Carepartner, error) {
	return carepartner.GetAllCarepartners(s.DB)
}

// ─────────────────────────────────────────────
// UPDATE
// ─────────────────────────────────────────────

func (s *CarepartnerService) UpdateCarepartner(c *carepartner.Carepartner) error {
	if c.ID <= 0 {
		return errors.New("invalid carepartner ID")
	}
	return carepartner.UpdateCarepartner(s.DB, c)
}

// ─────────────────────────────────────────────
// DELETE
// ─────────────────────────────────────────────

func (s *CarepartnerService) DeleteCarepartner(id int) error {
	if id <= 0 {
		return errors.New("invalid carepartner ID")
	}
	return carepartner.DeleteCarepartner(s.DB, id)
}

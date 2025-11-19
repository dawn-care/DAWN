package caregiver

import (
	"database/sql"
	"errors"
	"fmt"
	"server/internal/models/user/caregiver"
)

// CaregiverService defines all caregiver operations
type CaregiverService struct {
	DB *sql.DB
}

// NewCaregiverService constructs a new service instance
func NewCaregiverService(db *sql.DB) *CaregiverService {
	return &CaregiverService{DB: db}
}

// ─────────────────────────────────────────────
// CREATE
// ─────────────────────────────────────────────

func (s *CaregiverService) CreateCaregiver(c *caregiver.Caregiver) error {
	if c.Username == "" {
		return errors.New("username cannot be empty")
	}
	if c.CarereceiverID == 0 {
		return errors.New("carereceiver_id must match caregiver ID or be set")
	}

	return caregiver.CreateCaregiver(s.DB, c)
}

func (s *CaregiverService) CreateCaregiversBulk(list []caregiver.Caregiver) error {
	if len(list) == 0 {
		return errors.New("no caregivers provided")
	}

	for _, c := range list {
		if c.Username == "" {
			return errors.New("username cannot be empty in caregiver list")
		}
	}

	return caregiver.CreateCaregiversBulk(s.DB, list)
}

// ─────────────────────────────────────────────
// READ
// ─────────────────────────────────────────────

func (s *CaregiverService) GetCaregiverByID(id int) (*caregiver.Caregiver, error) {
	if id <= 0 {
		return nil, errors.New("invalid ID")
	}
	return caregiver.GetCaregiverByID(s.DB, id)
}

func (s *CaregiverService) GetCaregiverByCarereceiverID(crID int) (*caregiver.Caregiver, error) {
	if crID <= 0 {
		return nil, errors.New("invalid carereceiver ID")
	}
	return caregiver.GetCaregiverByCarereceiver(s.DB, crID)
}

func (s *CaregiverService) GetAllCaregivers() ([]caregiver.Caregiver, error) {
	return caregiver.GetCaregivers(s.DB)
}

func (s *CaregiverService) SearchCaregivers(keyword string) ([]caregiver.Caregiver, error) {
	if len(keyword) < 2 {
		return nil, errors.New("search keyword too short")
	}
	return caregiver.SearchCaregiversByUsername(s.DB, keyword)
}

// ─────────────────────────────────────────────
// UPDATE
// ─────────────────────────────────────────────

func (s *CaregiverService) UpdateCaregiver(c *caregiver.Caregiver) error {
	if c.ID <= 0 {
		return errors.New("invalid caregiver ID")
	}
	if c.Username == "" {
		return errors.New("username cannot be empty")
	}
	return caregiver.UpdateCaregiver(s.DB, c)
}

// ─────────────────────────────────────────────
// DELETE
// ─────────────────────────────────────────────

func (s *CaregiverService) DeleteCaregiver(id int) error {
	if id <= 0 {
		return errors.New("invalid ID")
	}

	// Optional: prevent deleting caregivers who still have modules
	modules, err := s.CountModulesForCaregiver(id)
	if err != nil {
		return err
	}
	if modules > 0 {
		return fmt.Errorf("cannot delete caregiver %d; they still have %d module(s)", id, modules)
	}

	return caregiver.DeleteCaregiver(s.DB, id)
}

// CountModulesForCaregiver checks dependency before deletion
func (s *CaregiverService) CountModulesForCaregiver(caregiverID int) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM "Module"
		WHERE caregiver_id = $1;
	`
	var count int
	err := s.DB.QueryRow(query, caregiverID).Scan(&count)
	return count, err
}

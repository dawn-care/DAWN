package caregiver

import (
	"database/sql"
	"errors"
	"fmt"
	"server/internal/models/user/caregiver"
	"server/internal/services/user"
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
	if c.Name == "" {
		return errors.New("name cannot be empty")
	}

	return caregiver.CreateCaregiver(s.DB, c)
}

func (s *CaregiverService) CreateCaregiversBulk(list []caregiver.Caregiver) error {
	if len(list) == 0 {
		return errors.New("no caregivers provided")
	}

	for _, c := range list {
		if c.Name == "" {
			return errors.New("name cannot be empty in caregiver list")
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

func (s *CaregiverService) GetCaregiverByCarepartnerID(crID int) (*caregiver.Caregiver, error) {
	if crID <= 0 {
		return nil, errors.New("invalid carepartner ID")
	}
	return caregiver.GetCaregiverByCarepartner(s.DB, crID)
}

func (s *CaregiverService) GetAllCaregivers() ([]caregiver.Caregiver, error) {
	return caregiver.GetCaregivers(s.DB)
}

func (s *CaregiverService) SearchCaregivers(keyword string) ([]caregiver.Caregiver, error) {
	if len(keyword) < 2 {
		return nil, errors.New("search keyword too short")
	}
	return caregiver.SearchCaregiversByName(s.DB, keyword)
}

// ─────────────────────────────────────────────
// UPDATE
// ─────────────────────────────────────────────

func (s *CaregiverService) UpdateCaregiver(c *caregiver.Caregiver) error {
	if c.ID <= 0 {
		return errors.New("invalid caregiver ID")
	}
	if c.Name == "" {
		return errors.New("name cannot be empty")
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

func (s *CaregiverService) FindOrCreateFromGoogle(googleUser *user.GoogleUser) (*caregiver.Caregiver, error) {
	// First try to find existing caregiver by Google ID
	existing, err := caregiver.GetCaregiverByGoogleID(s.DB, googleUser.ID)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	// if not found -> create new caregiver
	c := &caregiver.Caregiver{
		Name:     googleUser.Name,
		Role:     "caregiver",
		GoogleID: &googleUser.ID,
	}

	if err := caregiver.CreateCaregiver(s.DB, c); err != nil {
		return nil, err
	}

	return c, nil
}

package caregiver

import (
	"database/sql"
	"errors"
	"server/internal/models/user/caregiver"
	"server/internal/models/user/carepartner"
	"server/internal/models/user/device_pairing"
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

func (s *CaregiverService) GetCarepartnersByCaregiverID(caregiverID int) ([]caregiver.Caregiver, error) {
	if caregiverID <= 0 {
		return nil, errors.New("invalid caregiver ID")
	}

	return caregiver.GetCarepartnersByCaregiverID(s.DB, caregiverID)
}

// func (s *CaregiverService) GetCaregiverByCarepartnerID(crID int) (*caregiver.Caregiver, error) {
// 	if crID <= 0 {
// 		return nil, errors.New("invalid carepartner ID")
// 	}
// 	return caregiver.GetCaregiverByCarepartner(s.DB, crID)
// }

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

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Unpair the caregiver's carepartners first; the carepartner foreign key
	// blocks deleting a caregiver who is still paired
	if err := device_pairing.DeletePairingsForCaregiver(tx, id); err != nil {
		return err
	}
	if err := carepartner.UnpairCarepartnersFromCaregiver(tx, id); err != nil {
		return err
	}
	if err := caregiver.DeleteCaregiver(tx, id); err != nil {
		return err
	}

	return tx.Commit()
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

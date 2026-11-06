package caregiver

import (
	"database/sql"
	"errors"
	"server/internal/models/user/caregiver"
	"server/internal/models/user/carepartner"
	"server/internal/models/user/device_pairing"
	"server/internal/services/user"
	"strings"
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

// FindOrCreateFromGoogle returns the caregiver for a verified Google sign-in.
// Email is the key both sign-in methods share, so, in order:
//  1. A caregiver with this Google ID is the one. Their email is saved if they have none.
//  2. Otherwise, if Google verified the email, the caregiver with that email gets this Google
//     ID, unless they are already linked to a different Google account. An address can pass to
//     a new person, and that person must not take over the old account.
//  3. Otherwise a new caregiver is created.
func (s *CaregiverService) FindOrCreateFromGoogle(googleUser *user.GoogleUser) (*caregiver.Caregiver, error) {
	email := ""
	if googleUser.EmailVerified {
		email = NormalizeEmail(googleUser.Email)
	}

	existing, err := caregiver.GetCaregiverByGoogleID(s.DB, googleUser.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if existing.Email == nil && email != "" {
			saved, err := caregiver.SetEmail(s.DB, existing.ID, email)
			if err != nil {
				return nil, err
			}
			if saved {
				existing.Email = &email
			}
		}
		return existing, nil
	}

	// emailForNew stays empty when the address already belongs to someone else
	emailForNew := email
	if email != "" {
		byEmail, err := caregiver.GetCaregiverByEmail(s.DB, email)
		if err != nil {
			return nil, err
		}
		if byEmail != nil {
			emailForNew = ""
			if byEmail.GoogleID == nil {
				linked, err := caregiver.SetGoogleID(s.DB, byEmail.ID, googleUser.ID)
				if err != nil {
					return nil, err
				}
				if linked {
					byEmail.GoogleID = &googleUser.ID
					return byEmail, nil
				}
			}
		}
	}

	c := &caregiver.Caregiver{
		Name:     googleUser.Name,
		Role:     "caregiver",
		GoogleID: &googleUser.ID,
	}
	if emailForNew != "" {
		c.Email = &emailForNew
	}
	if err := caregiver.CreateCaregiver(s.DB, c); err != nil {
		return nil, err
	}

	return c, nil
}

// FindOrCreateFromEmail returns the caregiver for an email that has proven it receives mail
// (a correct sign-in code), creating one named after the part before the @ if there is none.
func (s *CaregiverService) FindOrCreateFromEmail(rawEmail string) (*caregiver.Caregiver, error) {
	email := NormalizeEmail(rawEmail)
	if email == "" {
		return nil, errors.New("email is required")
	}

	existing, err := caregiver.GetCaregiverByEmail(s.DB, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	name, _, _ := strings.Cut(email, "@")
	c := &caregiver.Caregiver{Name: name, Role: "caregiver", Email: &email}
	if err := caregiver.CreateCaregiver(s.DB, c); err != nil {
		// Two sign-ins for a new address at once: the other one created it first
		again, lookupErr := caregiver.GetCaregiverByEmail(s.DB, email)
		if lookupErr == nil && again != nil {
			return again, nil
		}
		return nil, err
	}
	return c, nil
}

// NormalizeEmail trims an email and makes it lowercase, the form stored and compared
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

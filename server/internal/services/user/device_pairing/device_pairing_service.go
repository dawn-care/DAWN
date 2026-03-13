package device_pairing

import (
	"database/sql"
	"errors"
	"math/rand"
	"time"

	"server/internal/models/user/device_pairing"
)

type DevicePairingService struct {
	DB *sql.DB
}

func NewDevicePairingService(db *sql.DB) *DevicePairingService {
	return &DevicePairingService{DB: db}
}

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// ─────────────────────────────────────────────
// CREATE
// ─────────────────────────────────────────────

func (s *DevicePairingService) CreatePairing(p *device_pairing.DevicePairing) error {

	if p.CarepartnerID <= 0 {
		return errors.New("invalid carepartner ID")
	}

	if p.PairingCode == "" {
		return errors.New("pairing code cannot be empty")
	}

	if p.ExpiresAt.Before(time.Now()) {
		return errors.New("expiration must be in the future")
	}

	return device_pairing.CreateDevicePairing(s.DB, p)
}

//
// READ
//

func (s *DevicePairingService) GetPairingByID(id int) (*device_pairing.DevicePairing, error) {

	if id <= 0 {
		return nil, errors.New("invalid ID")
	}

	return device_pairing.GetDevicePairingByID(s.DB, id)
}

func (s *DevicePairingService) GetPairingByCode(code string) (*device_pairing.DevicePairing, error) {

	if code == "" {
		return nil, errors.New("pairing code cannot be empty")
	}

	return device_pairing.GetDevicePairingByCode(s.DB, code)
}

func (s *DevicePairingService) GetAllPairings() ([]device_pairing.DevicePairing, error) {
	return device_pairing.GetDevicePairings(s.DB)
}

//
// UPDATE
//

func (s *DevicePairingService) UpdatePairing(p *device_pairing.DevicePairing) error {

	if p.ID <= 0 {
		return errors.New("invalid pairing ID")
	}

	return device_pairing.UpdateDevicePairing(s.DB, p)
}

//
// DELETE
//

func (s *DevicePairingService) DeletePairing(id int) error {

	if id <= 0 {
		return errors.New("invalid ID")
	}

	return device_pairing.DeleteDevicePairing(s.DB, id)
}

//
// Generate random code
//

func generateCode(length int) string {

	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

	code := make([]byte, length)

	for i := range code {
		code[i] = charset[rng.Intn(len(charset))]
	}

	return string(code)
}

//
// Generate and store pairing code
//

func (s *DevicePairingService) GeneratePairingCode(carepartnerID int) (*device_pairing.DevicePairing, error) {

	if carepartnerID <= 0 {
		return nil, errors.New("invalid carepartner ID")
	}

	const maxAttempts = 5

	for i := 0; i < maxAttempts; i++ {

		code := generateCode(6)

		pairing := &device_pairing.DevicePairing{
			CarepartnerID: carepartnerID,
			PairingCode:   code,
			ExpiresAt:     time.Now().Add(10 * time.Minute),
			Paired:        false,
		}

		err := device_pairing.CreateDevicePairing(s.DB, pairing)

		if err == nil {
			return pairing, nil
		}

		// If unique constraint collision happens, retry
	}

	return nil, errors.New("could not generate unique pairing code")
}

//
// Verify pairing code
//

func (s *DevicePairingService) VerifyPairingCode(code string, caregiverID int) error {

	if code == "" {
		return errors.New("pairing code cannot be empty")
	}

	if caregiverID <= 0 {
		return errors.New("invalid caregiver ID")
	}

	pairing, err := device_pairing.GetDevicePairingByCode(s.DB, code)
	if err != nil {
		return err
	}

	if pairing == nil {
		return errors.New("pairing code not found")
	}

	if pairing.Paired {
		return errors.New("pairing code already used")
	}

	if time.Now().After(pairing.ExpiresAt) {
		return errors.New("pairing code expired")
	}

	// Link caregiver to carepartner
	query := `
		UPDATE "Caregiver"
		SET carepartner_id = $1
		WHERE id = $2;
	`

	_, err = s.DB.Exec(query, pairing.CarepartnerID, caregiverID)
	if err != nil {
		return err
	}

	// Mark pairing as used
	pairing.Paired = true

	return device_pairing.UpdateDevicePairing(s.DB, pairing)
}

//
// Check if device is paired
//

func (s *DevicePairingService) IsDevicePaired(carepartnerID int) (bool, error) {

	if carepartnerID <= 0 {
		return false, errors.New("invalid carepartner ID")
	}

	query := `
		SELECT paired
		FROM "DevicePairing"
		WHERE carepartner_id = $1
		ORDER BY id DESC
		LIMIT 1;
	`

	var paired bool

	err := s.DB.QueryRow(query, carepartnerID).Scan(&paired)
	if err != nil {

		if err == sql.ErrNoRows {
			return false, nil
		}

		return false, err
	}

	return paired, nil
}

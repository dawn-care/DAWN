package module

import (
	"database/sql"
	"fmt"
)

// References media (metadata_id) and stores caregiver/carepartner
type Module struct {
	ID            int    `json:"id"`
	ModuleType    string `json:"module_type"`
	CaregiverID   int    `json:"caregiver_id"`
	CarepartnerID int    `json:"carepartner_id"`
	MetadataID    int    `json:"metadata_id"`
}

// Insert a new module
func CreateModule(db *sql.DB, m *Module) error {
	query := `
		INSERT INTO "Module" (module_type, caregiver_id, carepartner_id, metadata_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`
	return db.QueryRow(query, m.ModuleType, m.CaregiverID, m.CarepartnerID, m.MetadataID).Scan(&m.ID)
}

// Retrieve a module by its ID
func GetModuleByID(db *sql.DB, id int) (*Module, error) {
	query := `
		SELECT id, module_type, caregiver_id, carepartner_id, metadata_id
		FROM "Module"
		WHERE id = $1;
	`
	mod := &Module{}
	err := db.QueryRow(query, id).Scan(&mod.ID, &mod.ModuleType, &mod.CaregiverID, &mod.CarepartnerID, &mod.MetadataID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return mod, nil
}

// Find the caregiver and carepartner IDs for a given video by metadata_id
func GetVideoOwners(db *sql.DB, videoID int) (carepartnerID int, caregiverID int, err error) {
	query := `
		SELECT carepartner_id, caregiver_id
		FROM "Module"
		WHERE metadata_id = $1;
	`
	err = db.QueryRow(query, videoID).Scan(&carepartnerID, &caregiverID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, fmt.Errorf("no module found for video id %d", videoID)
		}
		return 0, 0, err
	}
	return carepartnerID, caregiverID, nil
}

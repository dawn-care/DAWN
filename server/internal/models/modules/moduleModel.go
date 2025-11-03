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

// Get all metadata IDs by caregiver ID, carepartner ID, and module type
func GetMetadataByUsersAndType(db *sql.DB, caregiverID int, carepartnerID int, moduleType string) ([]int, error) {
	query := `
		SELECT metadata_id
		FROM "Module"
		WHERE caregiver_id = $1 AND carepartner_id = $2 AND module_type = $3;
	`

	rows, err := db.Query(query, caregiverID, carepartnerID, moduleType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metadataIDs []int
	for rows.Next() {
		var metadataID int
		if err := rows.Scan(&metadataID); err != nil {
			return nil, err
		}
		metadataIDs = append(metadataIDs, metadataID)
	}

	if len(metadataIDs) == 0 {
		return nil, fmt.Errorf("no metadata found for caregiver %d, carepartner %d, and module type %s", caregiverID, carepartnerID, moduleType)
	}

	return metadataIDs, nil
}

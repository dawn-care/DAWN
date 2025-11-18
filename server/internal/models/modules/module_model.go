package modules

import (
	"database/sql"
	"fmt"
)

/*
All media, prompts, and widgets are each considered a Module, which has an associated caregiver, carepartner, and metadata
ID - the ID of the module entry in the database
ModuleType - the type of module it is (Media, Prompt, Widget, or possible future additions)
CaregiverID - the caregiver responsible for the module
CarepartnerID - the carepartner who will be able to view this module
MetadataID - the metadata ID of the module
*/
type Module struct {
	ID            int    `json:"id"`
	ModuleType    string `json:"module_type"`
	CaregiverID   int    `json:"caregiver_id"`
	CarepartnerID int    `json:"carepartner_id"`
	MetadataID    int    `json:"metadata_id"`
}

// CREATE OPERATION

/*
Create a new module
db - instance of the database
m - the module to upload
*/
func CreateModule(db *sql.DB, m *Module) error {
	query := `
		INSERT INTO "Module" (module_type, caregiver_id, carepartner_id, metadata_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`
	return db.QueryRow(query, m.ModuleType, m.CaregiverID, m.CarepartnerID, m.MetadataID).Scan(&m.ID)
}

// RETRIEVE OPERATIONS

/*
Retrieve a module by its ID
db - instance of the database
id - the ID of the module in the database
*/
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

/*
Get all metadata IDs by caregiver ID, carepartner ID, and module type
db - instance of the database
caregiverID - ID of the caregiver responsible for the module
carepartnerID - the ID of the carepartner who receives the module
moduleType - the type of module it is (Media, Prompt, Widget, etc.)
*/
func GetMetadataByCaregiverCarepartnerAndType(db *sql.DB, caregiverID int, carepartnerID int, moduleType string) ([]int, error) {
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

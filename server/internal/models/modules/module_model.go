package modules

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

// Insert a new module (DB)
func CreateModule(db *sql.DB, m *Module) error {
	return CreateModuleTx(db, m)
}

// Insert a new module (Tx)
func CreateModuleTx(exec interface {
	QueryRow(query string, args ...any) *sql.Row
}, m *Module) error {
	query := `
		INSERT INTO "Module" (module_type, caregiver_id, carepartner_id, metadata_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`
	return exec.QueryRow(query, m.ModuleType, m.CaregiverID, m.CarepartnerID, m.MetadataID).Scan(&m.ID)
}

// Retrieve a module by its ID (DB)
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

// Returns all modules for a given caregiver and carepartner
func GetModulesByCaregiverAndCarepartner(db *sql.DB, caregiverID, carepartnerID int) ([]Module, error) {
	query := `
		SELECT id, module_type, caregiver_id, carepartner_id, metadata_id
		FROM "Module"
		WHERE caregiver_id = $1 AND carepartner_id = $2;
	`

	rows, err := db.Query(query, caregiverID, carepartnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query modules: %w", err)
	}
	defer rows.Close()

	var modulesList []Module
	for rows.Next() {
		var m Module
		if err := rows.Scan(&m.ID, &m.ModuleType, &m.CaregiverID, &m.CarepartnerID, &m.MetadataID); err != nil {
			return nil, err
		}
		modulesList = append(modulesList, m)
	}

	return modulesList, nil
}

// Get all metadata IDs for a caregiver and carepartner (any module type)
func  GetAllMetadataByCaregiverCarepartner(db *sql.DB, caregiverID, carepartnerID int) ([]int, error) {
	query := `
		SELECT metadata_id
		FROM "Module"
		WHERE caregiver_id = $1 AND carepartner_id = $2;
	`

	rows, err := db.Query(query, caregiverID, carepartnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query all metadata: %w", err)
	}
	defer rows.Close()

	var metadataIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		metadataIDs = append(metadataIDs, id)
	}

	return metadataIDs, nil
}

// Get all metadata IDs by caregiver ID, carepartner ID, and module type
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

// -------------------------------------------------------------
// IGNORE - was testing
// -------------------------------------------------------------

// func GetPhotoMediaByIDs(db *sql.DB, ids []int) ([]map[string]interface{}, error) {
// 	return fetchChildRecords(db, "PhotoMedia", ids)
// }

// func GetVideoMediaByIDs(db *sql.DB, ids []int) ([]map[string]interface{}, error) {
// 	return fetchChildRecords(db, "VideoMedia", ids)
// }

// func GetAudioMediaByIDs(db *sql.DB, ids []int) ([]map[string]interface{}, error) {
// 	return fetchChildRecords(db, "AudioMedia", ids)
// }

// func GetWeatherWidgetsByIDs(db *sql.DB, ids []int) ([]map[string]interface{}, error) {
// 	return fetchChildRecords(db, "WeatherWidget", ids)
// }

// func GetNewsWidgetsByIDs(db *sql.DB, ids []int) ([]map[string]interface{}, error) {
// 	return fetchChildRecords(db, "NewsWidget", ids)
// }

// func GetCarepartnerLocationPromptsByIDs(db *sql.DB, ids []int) ([]map[string]interface{}, error) {
// 	return fetchChildRecords(db, "CarepartnerLocationPrompt", ids)
// }

// func GetCaregiverLocationPromptsByIDs(db *sql.DB, ids []int) ([]map[string]interface{}, error) {
// 	return fetchChildRecords(db, "CaregiverLocationPrompt", ids)
// }

// func GetMessagePromptsByIDs(db *sql.DB, ids []int) ([]map[string]interface{}, error) {
// 	return fetchChildRecords(db, "MessagePrompt", ids)
// }

// func GetEventPromptsByIDs(db *sql.DB, ids []int) ([]map[string]interface{}, error) {
// 	return fetchChildRecords(db, "EventPrompt", ids)
// }

// // -------------------------------------------------------------
// // Generic function to fetch child records from any table by IDs
// // -------------------------------------------------------------
// func fetchChildRecords(db *sql.DB, table string, ids []int) ([]map[string]interface{}, error) {
// 	if len(ids) == 0 {
// 		return nil, nil
// 	}

// 	query := fmt.Sprintf(`SELECT * FROM "%s" WHERE id = ANY($1)`, table)

// 	rows, err := db.Query(query, pq.Array(ids))
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	cols, err := rows.Columns()
// 	if err != nil {
// 		return nil, err
// 	}

// 	var results []map[string]interface{}
// 	for rows.Next() {
// 		columns := make([]interface{}, len(cols))
// 		columnPointers := make([]interface{}, len(cols))
// 		for i := range columns {
// 			columnPointers[i] = &columns[i]
// 		}

// 		if err := rows.Scan(columnPointers...); err != nil {
// 			return nil, err
// 		}

// 		m := make(map[string]interface{})
// 		for i, colName := range cols {
// 			m[colName] = columns[i]
// 		}
// 		results = append(results, m)
// 	}

// 	return results, nil
// }

// // DeleteModule deletes a module by its ID
// func DeleteModule(db *sql.DB, id int) error {
// 	if id <= 0 {
// 		return fmt.Errorf("invalid module ID: %d", id)
// 	}

// 	res, err := db.Exec(`DELETE FROM "Module" WHERE id=$1`, id)
// 	if err != nil {
// 		return fmt.Errorf("failed to delete module: %w", err)
// 	}

// 	rows, err := res.RowsAffected()
// 	if err != nil {
// 		return fmt.Errorf("failed to check affected rows: %w", err)
// 	}
// 	if rows == 0 {
// 		return fmt.Errorf("module with ID %d not found", id)
// 	}
// 	return nil
// }

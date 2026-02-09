package modules

import (
	"database/sql"
	"errors"
	"fmt"

	modules "server/internal/models/modules"
)

type ModuleService struct {
	DB *sql.DB
}

func NewModuleService(db *sql.DB) *ModuleService {
	return &ModuleService{DB: db}
}

// Create a new module with an existing metadata_id
func (s *ModuleService) CreateModule(moduleType string, caregiverID, carepartnerID, metadataID int) (*modules.Module, error) {
	if moduleType == "" || caregiverID <= 0 || carepartnerID <= 0 || metadataID <= 0 {
		return nil, errors.New("all parameters are required and must be valid")
	}

	mod := &modules.Module{
		ModuleType:    moduleType,
		CaregiverID:   caregiverID,
		CarepartnerID: carepartnerID,
		MetadataID:    metadataID,
	}

	if err := modules.CreateModule(s.DB, mod); err != nil {
		return nil, fmt.Errorf("failed to create module: %w", err)
	}

	return mod, nil
}

// Get a module by ID
func (s *ModuleService) GetModuleByID(id int) (*modules.Module, error) {
	if id <= 0 {
		return nil, errors.New("invalid module ID")
	}

	mod, err := modules.GetModuleByID(s.DB, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get module: %w", err)
	}
	if mod == nil {
		return nil, fmt.Errorf("module with id %d not found", id)
	}
	return mod, nil
}

// Get all metadata IDs for a given caregiver, carepartner, and module type
func (s *ModuleService) GetMetadataByCaregiverCarepartnerAndType(caregiverID, carepartnerID int, moduleType string) ([]int, error) {
	if caregiverID <= 0 || carepartnerID <= 0 {
		return nil, errors.New("invalid caregiver or carepartner ID")
	}
	if moduleType == "" {
		return nil, errors.New("module type is required")
	}

	metadataIDs, err := modules.GetMetadataByCaregiverCarepartnerAndType(s.DB, caregiverID, carepartnerID, moduleType)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}
	return metadataIDs, nil
}

// Get all metadata IDs for a caregiver and carepartner (any module type)
func (s *ModuleService) GetAllMetadataByCaregiverCarepartner(caregiverID, carepartnerID int) ([]int, error) {
	if caregiverID <= 0 || carepartnerID <= 0 {
		return nil, errors.New("invalid caregiver or carepartner ID")
	}

	query := `
		SELECT metadata_id
		FROM "Module"
		WHERE caregiver_id = $1 AND carepartner_id = $2;
	`

	rows, err := s.DB.Query(query, caregiverID, carepartnerID)
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

// DeleteModule deletes a module by its ID
func (s *ModuleService) DeleteModule(id int) error {
	return modules.DeleteModule(s.DB, id)
}

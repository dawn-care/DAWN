package module

import (
	"database/sql"
	"errors"
	"fmt"

	modules "server/internal/models/modules"
	"server/internal/models/modules/media"
)

type ModuleService struct {
	DB *sql.DB
}

func NewModuleService(db *sql.DB) *ModuleService {
	return &ModuleService{DB: db}
}

// Create a new module and the corresponding media
func (s *ModuleService) CreateModuleWithMedia(moduleType string, caregiverID, carepartnerID int, uri, filename, extension string) (*modules.Module, error) {
	if moduleType == "" || caregiverID <= 0 || carepartnerID <= 0 || uri == "" || filename == "" || extension == "" {
		return nil, errors.New("all parameters are required and must be valid")
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}

	var metadataID int

	switch moduleType {
	case "audio":
		audio := &media.AudioMedia{
			URI:       uri,
			Filename:  filename,
			Extension: extension,
		}
		if err := media.CreateAudioTx(tx, audio); err != nil {
			tx.Rollback()
			return nil, err
		}
		metadataID = audio.ID
	case "video":
		video := &media.VideoMedia{
			URI:       uri,
			Filename:  filename,
			Extension: extension,
		}
		if err := media.CreateVideoTx(tx, video); err != nil {
			tx.Rollback()
			return nil, err
		}
		metadataID = video.ID
	case "photo":
		photo := &media.PhotoMedia{
			URI:       uri,
			Filename:  filename,
			Extension: extension,
		}
		if err := media.CreatePhotoTx(tx, photo); err != nil {
			tx.Rollback()
			return nil, err
		}
		metadataID = photo.ID
	default:
		tx.Rollback()
		return nil, fmt.Errorf("unsupported module type: %s", moduleType)
	}

	mod := &modules.Module{
		ModuleType:    moduleType,
		CaregiverID:   caregiverID,
		CarepartnerID: carepartnerID,
		MetadataID:    metadataID,
	}

	if err := modules.CreateModuleTx(tx, mod); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
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

package family

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"server/internal/models/family"
)

// Longest name, relationship and guidance note accepted, in characters
const (
	MaxNameLength         = 100
	MaxRelationshipLength = 60
	MaxGuidanceLength     = 1000
)

var (
	ErrInvalidCarepartner  = errors.New("invalid carepartner_id")
	ErrInvalidName         = fmt.Errorf("name must be 1 to %d characters", MaxNameLength)
	ErrInvalidRelationship = fmt.Errorf("relationship must be at most %d characters", MaxRelationshipLength)
	ErrInvalidGuidance     = fmt.Errorf("guidance_note must be at most %d characters", MaxGuidanceLength)
)

type FamilyService struct {
	DB *sql.DB
}

func NewFamilyService(db *sql.DB) *FamilyService {
	return &FamilyService{DB: db}
}

// CreateInput is a new family member; GuidanceNote may be nil
type CreateInput struct {
	CarepartnerID int
	Name          string
	Relationship  string
	GuidanceNote  *string
}

func tooLong(s string, max int) bool {
	return len([]rune(s)) > max
}

// List a carepartner's family members in the order they were added
func (s *FamilyService) ListFamilyMembers(carepartnerID int) ([]family.FamilyMember, error) {
	if carepartnerID <= 0 {
		return nil, ErrInvalidCarepartner
	}
	members, err := family.ListFamilyMembers(s.DB, carepartnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list family members: %w", err)
	}
	return members, nil
}

// Create a family member. Text is trimmed; an empty guidance note is stored as none.
func (s *FamilyService) CreateFamilyMember(in CreateInput) (*family.FamilyMember, error) {
	if in.CarepartnerID <= 0 {
		return nil, ErrInvalidCarepartner
	}
	name := strings.TrimSpace(in.Name)
	if name == "" || tooLong(name, MaxNameLength) {
		return nil, ErrInvalidName
	}
	relationship := strings.TrimSpace(in.Relationship)
	if tooLong(relationship, MaxRelationshipLength) {
		return nil, ErrInvalidRelationship
	}
	var guidance *string
	if in.GuidanceNote != nil {
		if note := strings.TrimSpace(*in.GuidanceNote); note != "" {
			if tooLong(note, MaxGuidanceLength) {
				return nil, ErrInvalidGuidance
			}
			guidance = &note
		}
	}

	created, err := family.CreateFamilyMember(s.DB, family.FamilyMember{
		CarepartnerID: in.CarepartnerID,
		Name:          name,
		Relationship:  relationship,
		GuidanceNote:  guidance,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create family member: %w", err)
	}
	if created == nil {
		return nil, ErrInvalidCarepartner
	}
	return created, nil
}

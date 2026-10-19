package family

import (
	"database/sql"
	"time"

	"server/internal/db"
)

/*
FamilyMember is someone in a carepartner's family. Members are folders for grouping
media, not accounts.
Relationship - how they are related, e.g. "Daughter"; may be empty
GuidanceNote - what staff should say or avoid when the carepartner asks for them; nil if none
*/
type FamilyMember struct {
	ID            int       `json:"id"`
	CarepartnerID int       `json:"carepartner_id"`
	Name          string    `json:"name"`
	Relationship  string    `json:"relationship"`
	GuidanceNote  *string   `json:"guidance_note"`
	CreatedAt     time.Time `json:"created_at"`
}

const columns = `id, carepartner_id, name, relationship, guidance_note, created_at`

func scanFamilyMember(row db.Scanner) (*FamilyMember, error) {
	f := &FamilyMember{}
	if err := row.Scan(&f.ID, &f.CarepartnerID, &f.Name, &f.Relationship, &f.GuidanceNote, &f.CreatedAt); err != nil {
		return nil, err
	}
	return f, nil
}

// List a carepartner's family members in the order they were added
func ListFamilyMembers(q db.Querier, carepartnerID int) ([]FamilyMember, error) {
	rows, err := q.Query(`SELECT `+columns+` FROM "FamilyMember" WHERE carepartner_id = $1 ORDER BY id;`, carepartnerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := []FamilyMember{}
	for rows.Next() {
		f, err := scanFamilyMember(rows)
		if err != nil {
			return nil, err
		}
		members = append(members, *f)
	}
	return members, rows.Err()
}

// Create a family member; returns nil if the carepartner doesn't exist
func CreateFamilyMember(q db.QueryRower, f FamilyMember) (*FamilyMember, error) {
	created, err := scanFamilyMember(q.QueryRow(`
		INSERT INTO "FamilyMember" (carepartner_id, name, relationship, guidance_note)
		SELECT c.id, $2, $3, $4 FROM "Carepartner" c WHERE c.id = $1
		RETURNING `+columns+`;
	`, f.CarepartnerID, f.Name, f.Relationship, f.GuidanceNote))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return created, err
}

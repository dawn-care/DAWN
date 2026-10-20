package media

import (
	"database/sql"
	"slices"

	"github.com/lib/pq"

	"server/internal/db"
)

// Situations staff choose from when looking for media, in the order the apps show
// them. "Asking for someone" isn't one: staff reach that media through the family
// member they are asked about.
var Situations = []string{
	"wants_to_go_home",
	"doesnt_know_where",
	"anxious_or_upset",
	"starting_the_day",
	"winding_down",
}

// ValidSituation reports whether s is one of Situations
func ValidSituation(s string) bool {
	return slices.Contains(Situations, s)
}

/*
LoadTags fills in Situations and FamilyMemberIDs on each item with one query.
Situations come back in the order of Situations, family members by ID.
*/
func LoadTags(q db.Querier, items []Media) error {
	if len(items) == 0 {
		return nil
	}
	ids := make([]int64, len(items))
	for i, m := range items {
		ids[i] = int64(m.ID)
	}
	rows, err := q.Query(`
		SELECT m.id,
		       ARRAY(SELECT s.situation FROM "MediaSituation" s WHERE s.media_id = m.id),
		       ARRAY(SELECT f.family_member_id FROM "MediaFamilyMember" f
		             WHERE f.media_id = m.id ORDER BY f.family_member_id)
		FROM unnest($1::int[]) AS m(id);
	`, pq.Array(ids))
	if err != nil {
		return err
	}
	defer rows.Close()

	type tags struct {
		situations []string
		family     []int
	}
	byID := map[int]tags{}
	for rows.Next() {
		var id int
		var situations []string
		var family []int64
		if err := rows.Scan(&id, pq.Array(&situations), pq.Array(&family)); err != nil {
			return err
		}
		t := tags{situations: []string{}, family: make([]int, len(family))}
		// Keep the apps' order, whatever order the rows were stored in
		for _, s := range Situations {
			if slices.Contains(situations, s) {
				t.situations = append(t.situations, s)
			}
		}
		for i, f := range family {
			t.family[i] = int(f)
		}
		byID[id] = t
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for i := range items {
		t, ok := byID[items[i].ID]
		if !ok {
			t = tags{situations: []string{}, family: []int{}}
		}
		items[i].Situations = &t.situations
		items[i].FamilyMemberIDs = &t.family
	}
	return nil
}

/*
SetTagsTx replaces a media item's situations and family members inside tx.
situations must already be valid. Returns false, and changes nothing the caller
should commit, if any family member is not one of carepartnerID's.
*/
func SetTagsTx(tx *sql.Tx, mediaID, carepartnerID int, situations []string, familyMemberIDs []int) (bool, error) {
	if _, err := tx.Exec(`DELETE FROM "MediaSituation" WHERE media_id = $1;`, mediaID); err != nil {
		return false, err
	}
	if _, err := tx.Exec(`DELETE FROM "MediaFamilyMember" WHERE media_id = $1;`, mediaID); err != nil {
		return false, err
	}
	if len(situations) > 0 {
		if _, err := tx.Exec(
			`INSERT INTO "MediaSituation" (media_id, situation) SELECT $1, unnest($2::text[]);`,
			mediaID, pq.Array(situations),
		); err != nil {
			return false, err
		}
	}
	if len(familyMemberIDs) == 0 {
		return true, nil
	}

	ids := make([]int64, len(familyMemberIDs))
	for i, id := range familyMemberIDs {
		ids[i] = int64(id)
	}
	// Only the carepartner's own family members are linked; a shortfall means one wasn't
	res, err := tx.Exec(`
		INSERT INTO "MediaFamilyMember" (media_id, family_member_id)
		SELECT $1, f.id FROM "FamilyMember" f
		WHERE f.id = ANY($2::int[]) AND f.carepartner_id = $3;
	`, mediaID, pq.Array(ids), carepartnerID)
	if err != nil {
		return false, err
	}
	linked, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return linked == int64(len(familyMemberIDs)), nil
}

// CarepartnerOfMediaTx locks a media item inside tx and returns its carepartner, or
// false if there is no such item
func CarepartnerOfMediaTx(tx *sql.Tx, mediaID int) (int, bool, error) {
	var carepartnerID int
	err := tx.QueryRow(`SELECT carepartner_id FROM "Media" WHERE id = $1 FOR UPDATE;`, mediaID).Scan(&carepartnerID)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	return carepartnerID, err == nil, err
}

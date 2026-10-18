package family

import (
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"server/internal/utils"
)

var familyColumns = []string{"id", "carepartner_id", "name", "relationship", "guidance_note", "created_at"}

func newTestService(t *testing.T) (*FamilyService, sqlmock.Sqlmock) {
	db, mock := utils.SetupMockDB(t)
	t.Cleanup(func() { db.Close() })
	return NewFamilyService(db), mock
}

func TestListFamilyMembers(t *testing.T) {
	svc, mock := newTestService(t)
	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(`FROM "FamilyMember" WHERE carepartner_id = $1 ORDER BY id;`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(familyColumns).
			AddRow(1, 1, "Elena", "Daughter", nil, now).
			AddRow(4, 1, "Daniel", "Husband", "Please don't tell them again.", now))

	members, err := svc.ListFamilyMembers(1)

	require.NoError(t, err)
	require.Len(t, members, 2)
	assert.Equal(t, "Elena", members[0].Name)
	assert.Nil(t, members[0].GuidanceNote)
	require.NotNil(t, members[1].GuidanceNote)
	assert.Equal(t, "Please don't tell them again.", *members[1].GuidanceNote)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListFamilyMembers_EmptyAndErrors(t *testing.T) {
	svc, mock := newTestService(t)
	mock.ExpectQuery(`FROM "FamilyMember"`).WithArgs(2).WillReturnRows(sqlmock.NewRows(familyColumns))
	members, err := svc.ListFamilyMembers(2)
	require.NoError(t, err)
	assert.NotNil(t, members, "none is an empty list, not nil")
	assert.Empty(t, members)

	_, err = svc.ListFamilyMembers(0)
	assert.ErrorIs(t, err, ErrInvalidCarepartner)

	cause := errors.New("database down")
	mock.ExpectQuery(`FROM "FamilyMember"`).WillReturnError(cause)
	_, err = svc.ListFamilyMembers(1)
	assert.ErrorIs(t, err, cause)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateFamilyMember(t *testing.T) {
	svc, mock := newTestService(t)
	note := "  Talk about the garden.  "
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "FamilyMember"`)).
		WithArgs(1, "Elena", "Daughter", "Talk about the garden.").
		WillReturnRows(sqlmock.NewRows(familyColumns).
			AddRow(9, 1, "Elena", "Daughter", "Talk about the garden.", time.Now()))

	created, err := svc.CreateFamilyMember(CreateInput{
		CarepartnerID: 1, Name: " Elena ", Relationship: " Daughter ", GuidanceNote: &note,
	})

	require.NoError(t, err)
	assert.Equal(t, 9, created.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateFamilyMember_BlankNoteIsNone(t *testing.T) {
	svc, mock := newTestService(t)
	blank := "   "
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "FamilyMember"`)).
		WithArgs(1, "Marco", "", nil).
		WillReturnRows(sqlmock.NewRows(familyColumns).AddRow(10, 1, "Marco", "", nil, time.Now()))

	created, err := svc.CreateFamilyMember(CreateInput{CarepartnerID: 1, Name: "Marco", GuidanceNote: &blank})

	require.NoError(t, err)
	assert.Nil(t, created.GuidanceNote)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateFamilyMember_Invalid(t *testing.T) {
	svc, mock := newTestService(t)
	long := strings.Repeat("a", MaxGuidanceLength+1)
	cases := []struct {
		in   CreateInput
		want error
	}{
		{CreateInput{CarepartnerID: 0, Name: "Elena"}, ErrInvalidCarepartner},
		{CreateInput{CarepartnerID: 1, Name: "  "}, ErrInvalidName},
		{CreateInput{CarepartnerID: 1, Name: strings.Repeat("a", MaxNameLength+1)}, ErrInvalidName},
		{CreateInput{CarepartnerID: 1, Name: "Elena", Relationship: strings.Repeat("a", MaxRelationshipLength+1)}, ErrInvalidRelationship},
		{CreateInput{CarepartnerID: 1, Name: "Elena", GuidanceNote: &long}, ErrInvalidGuidance},
	}
	for _, tc := range cases {
		_, err := svc.CreateFamilyMember(tc.in)
		assert.ErrorIs(t, err, tc.want)
	}

	// The carepartner doesn't exist, so nothing is inserted
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "FamilyMember"`)).WillReturnRows(sqlmock.NewRows(familyColumns))
	_, err := svc.CreateFamilyMember(CreateInput{CarepartnerID: 99, Name: "Elena"})
	assert.ErrorIs(t, err, ErrInvalidCarepartner)
	assert.NoError(t, mock.ExpectationsWereMet())
}

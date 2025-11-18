package prompts

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// CREATE

func TestCreateEventList(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	listID := 5
	now := time.Now()

	eventList := []*EventPrompt{
		{List_ID: listID, Time: now, Text: "Take Meds"},
		{List_ID: listID, Time: now, Text: "Grandkids coming to visit"},
		{List_ID: listID, Time: now, Text: "Lunch"},
	}

	mock.ExpectBegin()

	prep := mock.ExpectPrepare(regexp.QuoteMeta(`
		INSERT INTO "EventPrompt" (list_id, time, text)
		VALUES ($1, $2, $3)
		RETURNING id
	`))

	for i, ev := range eventList {
		prep.ExpectQuery().
			WithArgs(ev.List_ID, ev.Time, ev.Text).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(i + 1))
	}

	mock.ExpectCommit()

	err := CreateEventList(db, eventList)
	assert.NoError(t, err)
	assert.Equal(t, 1, eventList[0].ID)
	assert.Equal(t, 2, eventList[1].ID)
	assert.Equal(t, 3, eventList[2].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// RETRIEVE

func TestGetEventListByListID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	listID := 5
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "list_id", "time", "text"}).
		AddRow(1, listID, now, "Take Meds").
		AddRow(2, listID, now, "Grandkids coming to visit").
		AddRow(3, listID, now, "Lunch")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT *
		FROM "EventPrompt"
		WHERE list_id = $1;
	`)).
		WithArgs(listID).
		WillReturnRows(rows)

	got, err := GetEventListByListID(db, &listID)
	assert.NoError(t, err)
	assert.Len(t, got, 3)
	assert.Equal(t, "Take Meds", got[0].Text)
	assert.Equal(t, "Grandkids coming to visit", got[1].Text)
	assert.Equal(t, "Lunch", got[2].Text)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// UPDATE

func TestUpdateEvent(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	ev := &EventPrompt{
		ID:      1,
		List_ID: 123456789,
		Time:    time.Now(),
		Text:    "Don't take meds",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "EventPrompt"
		SET list_id = $1, time = $2, text = $3
		WHERE id = $4;
	`)).
		WithArgs(ev.List_ID, sqlmock.AnyArg(), ev.Text, ev.ID). // <- AnyArg() for time
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := UpdateEvent(db, ev)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// DELETE

func TestDeleteEventList(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	listID := 5

	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM "EventPrompt"
		WHERE list_id = $1;
	`)).
		WithArgs(listID).
		WillReturnResult(sqlmock.NewResult(1, 3))

	err := DeleteEventList(db, &listID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

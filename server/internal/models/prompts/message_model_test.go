package prompts

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// CREATE

func TestCreateMessage(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	message := &MessagePrompt{
		Message: "We love you!!",
		URI:     "C:\\image1.png",
	}

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO "MessagePrompt" (message, image_uri)
		VALUES ($1, $2)
		RETURNING id;
	`)).
		WithArgs(message.Message, message.URI).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err := CreateMessage(db, message)
	assert.NoError(t, err)
	assert.Equal(t, 1, message.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// READ

func TestReadMessageByID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	// Column names must match the function: image_uri
	rows := sqlmock.NewRows([]string{"id", "message", "image_uri"}).
		AddRow(1, "We love you!!", "C:\\image1.png")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, message, image_uri
		FROM "MessagePrompt"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnRows(rows)

	got, err := GetMessageByID(db, 1)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "We love you!!", got.Message)
	assert.Equal(t, "C:\\image1.png", got.URI)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// UPDATE

func TestUpdateMessage(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	message := &MessagePrompt{
		ID:      1,
		Message: "Don't forget to drink water!",
		URI:     "C:\\image1.png",
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE "MessagePrompt"
		SET message = $1, image_uri = $2
		WHERE id = $3;
	`)).
		WithArgs(message.Message, message.URI, message.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := UpdateMessagePrompt(db, message)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

// DELETE

func TestDeleteMessage(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM "MessagePrompt"
		WHERE id = $1;
	`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := DeleteMessage(db, 1)
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

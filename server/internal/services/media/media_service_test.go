package media

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	modelsMedia "server/internal/models/media"
	"server/internal/utils"
)

// fakeStore records what the service puts in and removes from blob storage
type fakeStore struct {
	putKey, putType string
	putData         []byte
	putSize         int64
	putErr          error
	removed         []string
	removeErr       error
	objects         map[string][]byte
}

func (f *fakeStore) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	if f.putErr != nil {
		return f.putErr
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	f.putKey, f.putType, f.putData, f.putSize = key, contentType, data, size
	return nil
}

type nopSeekCloser struct{ *bytes.Reader }

func (nopSeekCloser) Close() error { return nil }

func (f *fakeStore) Get(ctx context.Context, key string) (io.ReadSeekCloser, error) {
	data, ok := f.objects[key]
	if !ok {
		return nil, errors.New("no such object")
	}
	return nopSeekCloser{bytes.NewReader(data)}, nil
}

func (f *fakeStore) Remove(ctx context.Context, key string) error {
	f.removed = append(f.removed, key)
	return f.removeErr
}

func newTestService(t *testing.T, seconds float64, probeErr error) (*MediaService, sqlmock.Sqlmock, *fakeStore, *int) {
	db, mock := utils.SetupMockDB(t)
	t.Cleanup(func() { db.Close() })
	store := &fakeStore{}
	probes := 0
	svc := &MediaService{
		DB:    db,
		Store: store,
		Probe: func(ctx context.Context, path string) (float64, error) {
			probes++
			return seconds, probeErr
		},
	}
	return svc, mock, store, &probes
}

var mediaColumns = []string{
	"id", "carepartner_id", "kind", "display_name", "object_key", "mime_type",
	"size_bytes", "duration_seconds", "created_at", "on_screen", "audio_background", "scheduled",
}

var insertMedia = regexp.QuoteMeta(`INSERT INTO "Media"`)

// ─────────────────────────────────────────────
// Upload
// ─────────────────────────────────────────────

func TestUpload_Photo(t *testing.T) {
	svc, mock, store, probes := newTestService(t, 0, nil)
	created := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(insertMedia).
		WithArgs(1, "photo", "Beach", sqlmock.AnyArg(), "image/png", int64(5), nil).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(7, created))

	m, err := svc.Upload(context.Background(), UploadInput{
		CarepartnerID: 1, Kind: "photo", DisplayName: "  Beach ", Extension: ".PNG",
		MimeType: "image/png", File: strings.NewReader("hello"),
	})

	require.NoError(t, err)
	assert.Equal(t, 7, m.ID)
	assert.Equal(t, "Beach", m.DisplayName)
	assert.Nil(t, m.DurationSeconds)
	assert.Equal(t, []string{}, m.Usage)
	assert.Equal(t, "/api/media/7/file", m.URL)
	assert.Equal(t, 0, *probes, "photos are not probed")
	assert.Regexp(t, `^carepartners/1/photo/[0-9a-f-]{36}\.png$`, store.putKey)
	assert.Equal(t, store.putKey, m.ObjectKey)
	assert.Equal(t, []byte("hello"), store.putData)
	assert.Equal(t, int64(5), store.putSize)
	assert.Equal(t, "image/png", store.putType)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpload_VideoDurationRounded(t *testing.T) {
	svc, mock, _, probes := newTestService(t, 600.4, nil)

	mock.ExpectQuery(insertMedia).
		WithArgs(1, "video", "Clip", sqlmock.AnyArg(), "video/mp4", int64(3), 600).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(8, time.Now()))

	m, err := svc.Upload(context.Background(), UploadInput{
		CarepartnerID: 1, Kind: "video", DisplayName: "Clip", Extension: ".mp4",
		MimeType: "video/mp4", File: strings.NewReader("abc"),
	})

	require.NoError(t, err)
	require.NotNil(t, m.DurationSeconds)
	assert.Equal(t, 600, *m.DurationSeconds)
	assert.Equal(t, 1, *probes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpload_TooLong(t *testing.T) {
	svc, mock, store, _ := newTestService(t, 600.6, nil)

	_, err := svc.Upload(context.Background(), UploadInput{
		CarepartnerID: 1, Kind: "audio", DisplayName: "Song", MimeType: "audio/mpeg", File: strings.NewReader("abc"),
	})

	assert.ErrorIs(t, err, ErrTooLong)
	assert.Empty(t, store.putKey, "nothing is stored")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpload_Unreadable(t *testing.T) {
	svc, _, store, _ := newTestService(t, 0, errors.New("ffprobe failed"))

	_, err := svc.Upload(context.Background(), UploadInput{
		CarepartnerID: 1, Kind: "video", DisplayName: "Clip", MimeType: "video/mp4", File: strings.NewReader("abc"),
	})

	assert.ErrorIs(t, err, ErrUnreadableMedia)
	assert.Empty(t, store.putKey)
}

func TestUpload_DBFailureRemovesFile(t *testing.T) {
	svc, mock, store, _ := newTestService(t, 0, nil)
	mock.ExpectQuery(insertMedia).WillReturnError(errors.New("db down"))

	_, err := svc.Upload(context.Background(), UploadInput{
		CarepartnerID: 1, Kind: "photo", DisplayName: "Beach", MimeType: "image/png", File: strings.NewReader("x"),
	})

	assert.Error(t, err)
	assert.Equal(t, []string{store.putKey}, store.removed)
}

func TestUpload_InvalidInput(t *testing.T) {
	svc, _, _, _ := newTestService(t, 0, nil)
	ctx := context.Background()

	_, err := svc.Upload(ctx, UploadInput{CarepartnerID: 0, Kind: "photo", DisplayName: "x", File: strings.NewReader("")})
	assert.ErrorIs(t, err, ErrInvalidCarepartner)

	_, err = svc.Upload(ctx, UploadInput{CarepartnerID: 1, Kind: "gif", DisplayName: "x", File: strings.NewReader("")})
	assert.ErrorIs(t, err, ErrInvalidKind)

	_, err = svc.Upload(ctx, UploadInput{CarepartnerID: 1, Kind: "photo", DisplayName: "   ", File: strings.NewReader("")})
	assert.ErrorIs(t, err, ErrInvalidName)

	_, err = svc.Upload(ctx, UploadInput{CarepartnerID: 1, Kind: "photo", DisplayName: strings.Repeat("a", 201), File: strings.NewReader("")})
	assert.ErrorIs(t, err, ErrInvalidName)
}

// ─────────────────────────────────────────────
// Read
// ─────────────────────────────────────────────

func TestListMedia_Usage(t *testing.T) {
	svc, mock, _, _ := newTestService(t, 0, nil)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`FROM "Media" m`)).
		WithArgs(1, "").
		WillReturnRows(sqlmock.NewRows(mediaColumns).
			AddRow(1, 1, "photo", "Beach", "k1", "image/png", 10, nil, now, true, true, true).
			AddRow(2, 1, "audio", "Hymn", "k2", "audio/mpeg", 10, 214, now, false, false, true).
			AddRow(3, 1, "photo", "Dog", "k3", "image/jpeg", 10, nil, now, false, false, false))

	items, err := svc.ListMedia(1, "")

	require.NoError(t, err)
	require.Len(t, items, 3)
	assert.Equal(t, []string{"on_screen", "audio_background", "scheduled"}, items[0].Usage)
	assert.Equal(t, []string{"scheduled"}, items[1].Usage)
	assert.Equal(t, 214, *items[1].DurationSeconds)
	assert.Equal(t, []string{}, items[2].Usage, "saved items have an empty usage list")
	assert.Equal(t, "/api/media/3/file", items[2].URL)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListMedia_ByKind(t *testing.T) {
	svc, mock, _, _ := newTestService(t, 0, nil)
	mock.ExpectQuery(regexp.QuoteMeta(`FROM "Media" m`)).
		WithArgs(1, "video").
		WillReturnRows(sqlmock.NewRows(mediaColumns))

	items, err := svc.ListMedia(1, "video")

	require.NoError(t, err)
	assert.Equal(t, []modelsMedia.Media{}, items, "no rows is an empty list, not nil")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListMedia_InvalidInput(t *testing.T) {
	svc, _, _, _ := newTestService(t, 0, nil)
	_, err := svc.ListMedia(0, "")
	assert.ErrorIs(t, err, ErrInvalidCarepartner)
	_, err = svc.ListMedia(1, "gif")
	assert.ErrorIs(t, err, ErrInvalidKind)
}

func TestGetMedia_NotFound(t *testing.T) {
	svc, mock, _, _ := newTestService(t, 0, nil)
	mock.ExpectQuery(regexp.QuoteMeta(`WHERE m.id = $1`)).WithArgs(9).WillReturnRows(sqlmock.NewRows(mediaColumns))

	_, err := svc.GetMedia(9)
	assert.ErrorIs(t, err, ErrNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestOpenFile(t *testing.T) {
	svc, mock, store, _ := newTestService(t, 0, nil)
	store.objects = map[string][]byte{"k1": []byte("filedata")}
	mock.ExpectQuery(regexp.QuoteMeta(`WHERE m.id = $1`)).WithArgs(1).
		WillReturnRows(sqlmock.NewRows(mediaColumns).
			AddRow(1, 1, "photo", "Beach", "k1", "image/png", 8, nil, time.Now(), false, false, false))

	m, file, err := svc.OpenFile(context.Background(), 1)

	require.NoError(t, err)
	defer file.Close()
	data, _ := io.ReadAll(file)
	assert.Equal(t, "image/png", m.MimeType)
	assert.Equal(t, []byte("filedata"), data)
}

// ─────────────────────────────────────────────
// Rename / Delete
// ─────────────────────────────────────────────

func TestRename(t *testing.T) {
	svc, mock, _, _ := newTestService(t, 0, nil)
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "Media" SET display_name = $1 WHERE id = $2;`)).
		WithArgs("New name", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`WHERE m.id = $1`)).WithArgs(1).
		WillReturnRows(sqlmock.NewRows(mediaColumns).
			AddRow(1, 1, "photo", "New name", "k1", "image/png", 8, nil, time.Now(), false, false, false))

	m, err := svc.Rename(1, " New name ")

	require.NoError(t, err)
	assert.Equal(t, "New name", m.DisplayName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRename_NotFoundAndInvalid(t *testing.T) {
	svc, mock, _, _ := newTestService(t, 0, nil)
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "Media"`)).WillReturnResult(sqlmock.NewResult(0, 0))

	_, err := svc.Rename(5, "Name")
	assert.ErrorIs(t, err, ErrNotFound)

	_, err = svc.Rename(5, "")
	assert.ErrorIs(t, err, ErrInvalidName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete(t *testing.T) {
	svc, mock, store, _ := newTestService(t, 0, nil)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT object_key FROM "Media" WHERE id = $1 FOR UPDATE;`)).
		WithArgs(4).WillReturnRows(sqlmock.NewRows([]string{"object_key"}).AddRow("k4"))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "ScreenWidget" WHERE type = 'photo' AND photo_media_id = $1;`)).
		WithArgs(4).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "Media" WHERE id = $1;`)).
		WithArgs(4).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := svc.Delete(context.Background(), 4)

	require.NoError(t, err)
	assert.Equal(t, []string{"k4"}, store.removed, "file removed after commit")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete_NotFound(t *testing.T) {
	svc, mock, store, _ := newTestService(t, 0, nil)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT object_key`)).WithArgs(4).WillReturnRows(sqlmock.NewRows([]string{"object_key"}))
	mock.ExpectRollback()

	err := svc.Delete(context.Background(), 4)

	assert.ErrorIs(t, err, ErrNotFound)
	assert.Empty(t, store.removed)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete_FileRemovalFailureStillSucceeds(t *testing.T) {
	svc, mock, store, _ := newTestService(t, 0, nil)
	store.removeErr = errors.New("minio down")
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT object_key`)).WithArgs(4).WillReturnRows(sqlmock.NewRows([]string{"object_key"}).AddRow("k4"))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "ScreenWidget"`)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "Media"`)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	assert.NoError(t, svc.Delete(context.Background(), 4))
}

// ─────────────────────────────────────────────
// ffprobe (runs only where ffprobe is installed)
// ─────────────────────────────────────────────

func TestFFProbeDuration(t *testing.T) {
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not installed")
	}
	ctx := context.Background()

	for _, name := range []string{"test_video.mp4", "test_audio.mp3"} {
		seconds, err := FFProbeDuration(ctx, "../../utils/testFiles/"+name)
		assert.NoError(t, err, name)
		assert.Greater(t, seconds, 0.0, name)
	}

	_, err := FFProbeDuration(ctx, "../../utils/testFiles/test_photo.png")
	assert.Error(t, err, "a still image has no duration")
}

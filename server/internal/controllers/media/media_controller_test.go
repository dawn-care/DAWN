package media

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	models "server/internal/models/media"
	mediaSvc "server/internal/services/media"
)

// mockMediaService implements MediaServiceInterface and records upload input
type mockMediaService struct {
	err      error
	media    *models.Media
	list     []models.Media
	file     []byte
	uploaded *mediaSvc.UploadInput
	renamed  string
	deleted  int
}

func (m *mockMediaService) Upload(ctx context.Context, in mediaSvc.UploadInput) (*models.Media, error) {
	m.uploaded = &in
	return m.media, m.err
}

func (m *mockMediaService) GetMedia(id int) (*models.Media, error) { return m.media, m.err }

func (m *mockMediaService) ListMedia(carepartnerID int, kind string) ([]models.Media, error) {
	return m.list, m.err
}

type readSeekNopCloser struct{ *bytes.Reader }

func (readSeekNopCloser) Close() error { return nil }

func (m *mockMediaService) OpenFile(ctx context.Context, id int) (*models.Media, io.ReadSeekCloser, error) {
	if m.err != nil {
		return nil, nil, m.err
	}
	return m.media, readSeekNopCloser{bytes.NewReader(m.file)}, nil
}

func (m *mockMediaService) Rename(id int, displayName string) (*models.Media, error) {
	m.renamed = displayName
	return m.media, m.err
}

func (m *mockMediaService) Delete(ctx context.Context, id int) error {
	m.deleted = id
	return m.err
}

// withID adds a chi {id} URL parameter to a request
func withID(r *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func uploadRequest(t *testing.T, fields map[string]string, filename string, content []byte) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for k, v := range fields {
		require.NoError(t, writer.WriteField(k, v))
	}
	if filename != "" {
		part, err := writer.CreateFormFile("file", filename)
		require.NoError(t, err)
		_, _ = part.Write(content)
	}
	require.NoError(t, writer.Close())
	req := httptest.NewRequest(http.MethodPost, "/api/media", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func decodeError(t *testing.T, w *httptest.ResponseRecorder) MediaError {
	var e MediaError
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &e))
	return e
}

// ─────────────────────────────────────────────
// Upload
// ─────────────────────────────────────────────

func TestUploadMediaHandler_Photo(t *testing.T) {
	png, err := os.ReadFile("../../utils/testFiles/test_photo.png")
	require.NoError(t, err)
	svc := &mockMediaService{media: &models.Media{ID: 3, Kind: "photo", Usage: []string{}}}
	ctrl := NewMediaController(svc)

	w := httptest.NewRecorder()
	ctrl.UploadMediaHandler(w, uploadRequest(t, map[string]string{"kind": "photo", "carepartner_id": "1"}, "Beach Day.png", png))

	assert.Equal(t, http.StatusCreated, w.Code, w.Body.String())
	require.NotNil(t, svc.uploaded)
	assert.Equal(t, 1, svc.uploaded.CarepartnerID)
	assert.Equal(t, "photo", svc.uploaded.Kind)
	assert.Equal(t, "image/png", svc.uploaded.MimeType)
	assert.Equal(t, ".png", svc.uploaded.Extension)
	assert.Equal(t, "Beach Day", svc.uploaded.DisplayName, "defaults to the file name without extension")
	data, _ := io.ReadAll(svc.uploaded.File)
	assert.Equal(t, png, data, "service gets the whole file from the start")
}

func TestUploadMediaHandler_WrongTypeForKind(t *testing.T) {
	png, _ := os.ReadFile("../../utils/testFiles/test_photo.png")
	svc := &mockMediaService{}
	w := httptest.NewRecorder()
	NewMediaController(svc).UploadMediaHandler(w, uploadRequest(t, map[string]string{"kind": "video", "carepartner_id": "1"}, "a.png", png))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, CodeInvalidFileType, decodeError(t, w).Code)
	assert.Nil(t, svc.uploaded)
}

func TestUploadMediaHandler_BadInput(t *testing.T) {
	png, _ := os.ReadFile("../../utils/testFiles/test_photo.png")
	cases := []struct {
		name   string
		fields map[string]string
		file   string
		code   string
	}{
		{"bad kind", map[string]string{"kind": "gif", "carepartner_id": "1"}, "a.png", CodeInvalidKind},
		{"no carepartner", map[string]string{"kind": "photo"}, "a.png", CodeInvalidRequest},
		{"no file", map[string]string{"kind": "photo", "carepartner_id": "1"}, "", CodeInvalidRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			NewMediaController(&mockMediaService{}).UploadMediaHandler(w, uploadRequest(t, tc.fields, tc.file, png))
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Equal(t, tc.code, decodeError(t, w).Code)
		})
	}
}

func TestUploadMediaHandler_ServiceErrors(t *testing.T) {
	mp3, err := os.ReadFile("../../utils/testFiles/test_audio.mp3")
	require.NoError(t, err)
	cases := []struct {
		err    error
		status int
		code   string
	}{
		{mediaSvc.ErrTooLong, http.StatusUnprocessableEntity, CodeMediaTooLong},
		{mediaSvc.ErrUnreadableMedia, http.StatusUnprocessableEntity, CodeUnreadableMedia},
		{mediaSvc.ErrInvalidName, http.StatusBadRequest, CodeInvalidName},
	}
	for _, tc := range cases {
		w := httptest.NewRecorder()
		NewMediaController(&mockMediaService{err: tc.err}).UploadMediaHandler(w,
			uploadRequest(t, map[string]string{"kind": "audio", "carepartner_id": "1"}, "song.mp3", mp3))
		assert.Equal(t, tc.status, w.Code)
		assert.Equal(t, tc.code, decodeError(t, w).Code)
	}
}

func TestDetectMimeType(t *testing.T) {
	wav := append([]byte("RIFF\x00\x00\x00\x00WAVEfmt "), make([]byte, 20)...)
	mp4 := append([]byte("\x00\x00\x00\x18ftypmp42"), make([]byte, 20)...)
	png, _ := os.ReadFile("../../utils/testFiles/test_photo.png")
	mp3, _ := os.ReadFile("../../utils/testFiles/test_audio.mp3")

	assert.Equal(t, "image/png", detectMimeType("photo", png, ".png"))
	assert.Equal(t, "audio/mpeg", detectMimeType("audio", mp3, ".mp3"), "untagged MP3 (no ID3 header)")
	assert.Equal(t, "audio/mpeg", detectMimeType("audio", append([]byte("ID3\x04"), make([]byte, 20)...), ".mp3"))
	assert.Equal(t, "", detectMimeType("audio", mp3, ".wav"), "frame sniffing only applies to .mp3")
	assert.Equal(t, "audio/wav", detectMimeType("audio", wav, ".wav"), "Go sniffs WAV as audio/wave")
	assert.Equal(t, "video/mp4", detectMimeType("video", mp4, ".mp4"))
	assert.Equal(t, "audio/mp4", detectMimeType("audio", mp4, ".m4a"))
	assert.Equal(t, "", detectMimeType("audio", mp4, ".mp4"), "an .mp4 is a video, not audio")
	assert.Equal(t, "video/quicktime", detectMimeType("video", []byte("\x00\x00\x00\x14ftypqt  "), ".mov"))
	assert.Equal(t, "", detectMimeType("photo", mp3, ".png"), "the extension can't make an mp3 a photo")
	assert.Equal(t, "", detectMimeType("gif", png, ".png"))
}

// ─────────────────────────────────────────────
// Read
// ─────────────────────────────────────────────

func TestListMediaHandler(t *testing.T) {
	svc := &mockMediaService{list: []models.Media{{ID: 1, Kind: "photo", Usage: []string{"on_screen"}}}}
	w := httptest.NewRecorder()
	NewMediaController(svc).ListMediaHandler(w, httptest.NewRequest(http.MethodGet, "/api/media?carepartner_id=1&kind=photo", nil))

	assert.Equal(t, http.StatusOK, w.Code)
	var items []models.Media
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &items))
	assert.Equal(t, []string{"on_screen"}, items[0].Usage)
	assert.NotContains(t, w.Body.String(), "object_key", "storage keys are never sent")
}

func TestListMediaHandler_Errors(t *testing.T) {
	w := httptest.NewRecorder()
	NewMediaController(&mockMediaService{}).ListMediaHandler(w, httptest.NewRequest(http.MethodGet, "/api/media", nil))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	NewMediaController(&mockMediaService{err: mediaSvc.ErrInvalidKind}).ListMediaHandler(w, httptest.NewRequest(http.MethodGet, "/api/media?carepartner_id=1&kind=gif", nil))
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, CodeInvalidKind, decodeError(t, w).Code)
}

func TestGetMediaHandler_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	NewMediaController(&mockMediaService{err: mediaSvc.ErrNotFound}).GetMediaHandler(w, withID(httptest.NewRequest(http.MethodGet, "/api/media/9", nil), "9"))
	assert.Equal(t, http.StatusNotFound, w.Code)

	w = httptest.NewRecorder()
	NewMediaController(&mockMediaService{}).GetMediaHandler(w, withID(httptest.NewRequest(http.MethodGet, "/api/media/abc", nil), "abc"))
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetMediaFileHandler_Range(t *testing.T) {
	svc := &mockMediaService{
		media: &models.Media{ID: 1, MimeType: "video/mp4", CreatedAt: time.Now()},
		file:  []byte("0123456789"),
	}
	req := withID(httptest.NewRequest(http.MethodGet, "/api/media/1/file", nil), "1")
	req.Header.Set("Range", "bytes=2-5")
	w := httptest.NewRecorder()

	NewMediaController(svc).GetMediaFileHandler(w, req)

	assert.Equal(t, http.StatusPartialContent, w.Code, "video seeking needs Range support")
	assert.Equal(t, "2345", w.Body.String())
	assert.Equal(t, "video/mp4", w.Header().Get("Content-Type"))
}

// ─────────────────────────────────────────────
// Rename / Delete
// ─────────────────────────────────────────────

func TestRenameMediaHandler(t *testing.T) {
	svc := &mockMediaService{media: &models.Media{ID: 1, DisplayName: "New"}}
	req := withID(httptest.NewRequest(http.MethodPatch, "/api/media/1", strings.NewReader(`{"display_name":"New"}`)), "1")
	w := httptest.NewRecorder()

	NewMediaController(svc).RenameMediaHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "New", svc.renamed)
}

func TestRenameMediaHandler_Errors(t *testing.T) {
	w := httptest.NewRecorder()
	NewMediaController(&mockMediaService{}).RenameMediaHandler(w,
		withID(httptest.NewRequest(http.MethodPatch, "/api/media/1", strings.NewReader(`not json`)), "1"))
	assert.Equal(t, http.StatusBadRequest, w.Code)

	w = httptest.NewRecorder()
	NewMediaController(&mockMediaService{err: mediaSvc.ErrNotFound}).RenameMediaHandler(w,
		withID(httptest.NewRequest(http.MethodPatch, "/api/media/1", strings.NewReader(`{"display_name":"x"}`)), "1"))
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteMediaHandler(t *testing.T) {
	svc := &mockMediaService{}
	w := httptest.NewRecorder()
	NewMediaController(svc).DeleteMediaHandler(w, withID(httptest.NewRequest(http.MethodDelete, "/api/media/4", nil), "4"))

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, 4, svc.deleted)

	w = httptest.NewRecorder()
	NewMediaController(&mockMediaService{err: mediaSvc.ErrNotFound}).DeleteMediaHandler(w, withID(httptest.NewRequest(http.MethodDelete, "/api/media/4", nil), "4"))
	assert.Equal(t, http.StatusNotFound, w.Code)
}

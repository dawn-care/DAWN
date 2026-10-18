package media

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	models "server/internal/models/media"
	mediaSvc "server/internal/services/media"
)

// octetStream is what Go's sniffer returns when it can't identify a file
const octetStream = "application/octet-stream"

// MAX_MEDIA_UPLOAD_MB caps a single upload; a 10 minute phone video can approach 1 GB
const MAX_MEDIA_UPLOAD_MB = 1024

// Error codes returned in the "code" field of error responses
const (
	CodeInvalidRequest   = "INVALID_REQUEST"
	CodeInvalidKind      = "INVALID_KIND"
	CodeInvalidFileType  = "INVALID_FILE_TYPE"
	CodeInvalidName      = "INVALID_NAME"
	CodeFileTooLarge     = "FILE_TOO_LARGE"
	CodeMediaTooLong     = "MEDIA_TOO_LONG"
	CodeUnreadableMedia  = "UNREADABLE_MEDIA"
	CodeNotFound         = "NOT_FOUND"
	CodeInvalidSituation = "INVALID_SITUATION"
	CodeInvalidFamily    = "INVALID_FAMILY_MEMBER"
	CodeInternal         = "INTERNAL_ERROR"
)

// ---- Interface ----
type MediaServiceInterface interface {
	Upload(ctx context.Context, in mediaSvc.UploadInput) (*models.Media, error)
	GetMedia(id int) (*models.Media, error)
	ListMedia(carepartnerID int, filter models.ListFilter) ([]models.Media, error)
	OpenFile(ctx context.Context, id int) (*models.Media, io.ReadSeekCloser, error)
	Rename(id int, displayName string) (*models.Media, error)
	SetTags(id int, tags mediaSvc.Tags) (*models.Media, error)
	Delete(ctx context.Context, id int) error
}

// ---- Controller ----
type MediaController struct {
	Services MediaServiceInterface
}

func NewMediaController(s MediaServiceInterface) *MediaController {
	return &MediaController{Services: s}
}

// MediaError is the body of every error response from the media endpoints
type MediaError struct {
	Code  string `json:"code"`
	Error string `json:"error"`
}

func writeMediaJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeMediaError(w http.ResponseWriter, status int, code, msg string) {
	writeMediaJSON(w, status, MediaError{Code: code, Error: msg})
}

// writeServiceError maps service errors to status codes; anything unexpected is a 500
func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, mediaSvc.ErrNotFound):
		writeMediaError(w, http.StatusNotFound, CodeNotFound, err.Error())
	case errors.Is(err, mediaSvc.ErrInvalidKind):
		writeMediaError(w, http.StatusBadRequest, CodeInvalidKind, err.Error())
	case errors.Is(err, mediaSvc.ErrInvalidName):
		writeMediaError(w, http.StatusBadRequest, CodeInvalidName, err.Error())
	case errors.Is(err, mediaSvc.ErrInvalidCarepartner):
		writeMediaError(w, http.StatusBadRequest, CodeInvalidRequest, err.Error())
	case errors.Is(err, mediaSvc.ErrTooLong):
		writeMediaError(w, http.StatusUnprocessableEntity, CodeMediaTooLong, err.Error())
	case errors.Is(err, mediaSvc.ErrUnreadableMedia):
		writeMediaError(w, http.StatusUnprocessableEntity, CodeUnreadableMedia, err.Error())
	case errors.Is(err, mediaSvc.ErrInvalidSituation):
		writeMediaError(w, http.StatusBadRequest, CodeInvalidSituation, err.Error())
	case errors.Is(err, mediaSvc.ErrInvalidFamily):
		writeMediaError(w, http.StatusBadRequest, CodeInvalidFamily, err.Error())
	default:
		writeMediaError(w, http.StatusInternalServerError, CodeInternal, err.Error())
	}
}

func mediaIDParam(r *http.Request) (int, bool) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	return id, err == nil && id > 0
}

// parseIDs reads repeated form or query values as positive IDs
func parseIDs(values []string) ([]int, bool) {
	ids := make([]int, 0, len(values))
	for _, v := range values {
		id, err := strconv.Atoi(v)
		if err != nil || id <= 0 {
			return nil, false
		}
		ids = append(ids, id)
	}
	return ids, true
}

/*
detectMimeType checks the sniffed type of a file against its kind, using the
existing VALID_*_TYPES lists. Go's sniffer names a few formats differently from
those lists, or can't identify them from the bytes alone, so those are mapped here.
Returns "" if the file isn't an accepted type for that kind.
*/
func detectMimeType(kind string, head []byte, ext string) string {
	sniffed := http.DetectContentType(head)
	ext = strings.ToLower(ext)

	switch kind {
	case models.KindPhoto:
		if VALID_IMAGE_TYPES[sniffed] {
			return sniffed
		}
		if sniffed == octetStream && (ext == ".heic" || ext == ".heif") {
			return "image/heic"
		}
	case models.KindVideo:
		if VALID_VIDEO_TYPES[sniffed] {
			return sniffed
		}
		if sniffed == "application/ogg" {
			return "video/ogg"
		}
		if sniffed == octetStream && ext == ".mov" {
			return "video/quicktime"
		}
	case models.KindAudio:
		if VALID_AUDIO_TYPES[sniffed] {
			return sniffed
		}
		switch {
		case sniffed == "audio/wave":
			return "audio/wav"
		case sniffed == "application/ogg":
			return "audio/ogg"
		case sniffed == "video/webm":
			return "audio/webm"
		case sniffed == "video/mp4" && ext == ".m4a":
			return "audio/mp4"
		case sniffed == octetStream && ext == ".mp3" && isMPEGFrame(head):
			// Go only sniffs MP3s that start with an ID3 tag; untagged ones start with a frame
			return "audio/mpeg"
		}
	}
	return ""
}

// isMPEGFrame reports whether head starts with an MPEG audio frame sync (11 set bits)
func isMPEGFrame(head []byte) bool {
	return len(head) >= 2 && head[0] == 0xFF && head[1]&0xE0 == 0xE0
}

// UploadMediaHandler godoc
// @Summary Upload a photo, video or audio file
// @Description Stores the file and creates its media item. Video and audio over 10 minutes are rejected (MEDIA_TOO_LONG); their length is read by the server.
// @Tags media
// @Accept mpfd
// @Produce json
// @Param file formData file true "Media file"
// @Param kind formData string true "photo, video or audio"
// @Param carepartner_id formData int true "Carepartner ID"
// @Param display_name formData string false "Name shown to the caregiver; defaults to the file name"
// @Param situations formData []string false "Situations staff find it under; repeat the field for each. None means every situation" collectionFormat(multi)
// @Param family_member_ids formData []int false "Family members it is about; repeat the field for each" collectionFormat(multi)
// @Success 201 {object} media.Media
// @Failure 400 {object} media.MediaError
// @Failure 413 {object} media.MediaError
// @Failure 422 {object} media.MediaError
// @Failure 500 {object} media.MediaError
// @Router /media [post]
func (c *MediaController) UploadMediaHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MAX_MEDIA_UPLOAD_MB<<20)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeMediaError(w, http.StatusRequestEntityTooLarge, CodeFileTooLarge, "file is larger than "+strconv.Itoa(MAX_MEDIA_UPLOAD_MB)+" MB")
			return
		}
		writeMediaError(w, http.StatusBadRequest, CodeInvalidRequest, "failed to parse multipart form")
		return
	}
	defer r.MultipartForm.RemoveAll()

	kind := r.FormValue("kind")
	carepartnerID, err := strconv.Atoi(r.FormValue("carepartner_id"))
	if err != nil || carepartnerID <= 0 {
		writeMediaError(w, http.StatusBadRequest, CodeInvalidRequest, "invalid carepartner_id")
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		writeMediaError(w, http.StatusBadRequest, CodeInvalidRequest, "file is required")
		return
	}
	defer file.Close()

	// Read the first 512 bytes for type detection, then rewind
	head := make([]byte, 512)
	n, err := io.ReadFull(file, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		writeMediaError(w, http.StatusBadRequest, CodeInvalidRequest, "unable to read file content")
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		writeMediaError(w, http.StatusInternalServerError, CodeInternal, "failed to reset file pointer")
		return
	}

	ext := filepath.Ext(handler.Filename)
	mimeType := detectMimeType(kind, head[:n], ext)
	if mimeType == "" {
		if kind != models.KindPhoto && kind != models.KindVideo && kind != models.KindAudio {
			writeMediaError(w, http.StatusBadRequest, CodeInvalidKind, mediaSvc.ErrInvalidKind.Error())
			return
		}
		writeMediaError(w, http.StatusBadRequest, CodeInvalidFileType, "file is not a supported "+kind+" type")
		return
	}

	displayName := r.FormValue("display_name")
	if strings.TrimSpace(displayName) == "" {
		displayName = strings.TrimSuffix(handler.Filename, ext)
	}

	familyMemberIDs, ok := parseIDs(r.MultipartForm.Value["family_member_ids"])
	if !ok {
		writeMediaError(w, http.StatusBadRequest, CodeInvalidFamily, mediaSvc.ErrInvalidFamily.Error())
		return
	}

	m, err := c.Services.Upload(r.Context(), mediaSvc.UploadInput{
		CarepartnerID: carepartnerID,
		Kind:          kind,
		DisplayName:   displayName,
		Extension:     ext,
		MimeType:      mimeType,
		File:          file,
		Tags: mediaSvc.Tags{
			Situations:      r.MultipartForm.Value["situations"],
			FamilyMemberIDs: familyMemberIDs,
		},
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeMediaJSON(w, http.StatusCreated, m)
}

// ListMediaHandler godoc
// @Summary List a carepartner's media
// @Description Newest first. Each item's usage lists where it is used (on_screen, audio_background, scheduled); an empty list means it is only saved.
// @Tags media
// @Produce json
// @Param carepartner_id query int true "Carepartner ID"
// @Param kind query string false "photo, video or audio; omit for all"
// @Param situation query string false "Only media under this situation"
// @Param family_member_id query int false "Only media about this family member"
// @Success 200 {array} media.Media
// @Failure 400 {object} media.MediaError
// @Failure 500 {object} media.MediaError
// @Router /media [get]
func (c *MediaController) ListMediaHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	carepartnerID, err := strconv.Atoi(query.Get("carepartner_id"))
	if err != nil || carepartnerID <= 0 {
		writeMediaError(w, http.StatusBadRequest, CodeInvalidRequest, "invalid carepartner_id")
		return
	}

	filter := models.ListFilter{Kind: query.Get("kind"), Situation: query.Get("situation")}
	if raw := query.Get("family_member_id"); raw != "" {
		ids, ok := parseIDs([]string{raw})
		if !ok {
			writeMediaError(w, http.StatusBadRequest, CodeInvalidFamily, mediaSvc.ErrInvalidFamily.Error())
			return
		}
		filter.FamilyMemberID = ids[0]
	}

	items, err := c.Services.ListMedia(carepartnerID, filter)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeMediaJSON(w, http.StatusOK, items)
}

// GetMediaHandler godoc
// @Summary Get one media item
// @Tags media
// @Produce json
// @Param id path int true "Media ID"
// @Success 200 {object} media.Media
// @Failure 404 {object} media.MediaError
// @Failure 500 {object} media.MediaError
// @Router /media/{id} [get]
func (c *MediaController) GetMediaHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := mediaIDParam(r)
	if !ok {
		writeMediaError(w, http.StatusNotFound, CodeNotFound, mediaSvc.ErrNotFound.Error())
		return
	}

	m, err := c.Services.GetMedia(id)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeMediaJSON(w, http.StatusOK, m)
}

// GetMediaFileHandler godoc
// @Summary Download a media item's file
// @Description Supports Range requests, so video and audio can be seeked.
// @Tags media
// @Produce application/octet-stream
// @Param id path int true "Media ID"
// @Success 200 {file} file "Media file"
// @Failure 404 {object} media.MediaError
// @Failure 500 {object} media.MediaError
// @Router /media/{id}/file [get]
func (c *MediaController) GetMediaFileHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := mediaIDParam(r)
	if !ok {
		writeMediaError(w, http.StatusNotFound, CodeNotFound, mediaSvc.ErrNotFound.Error())
		return
	}

	m, file, err := c.Services.OpenFile(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", m.MimeType)
	w.Header().Set("Content-Disposition", "inline")
	http.ServeContent(w, r, "", m.CreatedAt, file)
}

// RenameMediaRequest is the body of a rename
type RenameMediaRequest struct {
	DisplayName string `json:"display_name"`
}

// RenameMediaHandler godoc
// @Summary Rename a media item
// @Description Changes only the name the caregiver sees.
// @Tags media
// @Accept json
// @Produce json
// @Param id path int true "Media ID"
// @Param body body media.RenameMediaRequest true "New name"
// @Success 200 {object} media.Media
// @Failure 400 {object} media.MediaError
// @Failure 404 {object} media.MediaError
// @Failure 500 {object} media.MediaError
// @Router /media/{id} [patch]
func (c *MediaController) RenameMediaHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := mediaIDParam(r)
	if !ok {
		writeMediaError(w, http.StatusNotFound, CodeNotFound, mediaSvc.ErrNotFound.Error())
		return
	}

	var req RenameMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeMediaError(w, http.StatusBadRequest, CodeInvalidRequest, "invalid request body")
		return
	}

	m, err := c.Services.Rename(id, req.DisplayName)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeMediaJSON(w, http.StatusOK, m)
}

// SetMediaTagsRequest replaces both lists; an empty situations list means every situation
type SetMediaTagsRequest struct {
	Situations      []string `json:"situations"`
	FamilyMemberIDs []int    `json:"family_member_ids"`
}

// SetMediaTagsHandler godoc
// @Summary Set how staff find a media item
// @Description Replaces its situations and family members. No situations means every situation; family members must be the carepartner's.
// @Tags media
// @Accept json
// @Produce json
// @Param id path int true "Media ID"
// @Param body body media.SetMediaTagsRequest true "Situations and family members"
// @Success 200 {object} media.Media
// @Failure 400 {object} media.MediaError
// @Failure 404 {object} media.MediaError
// @Failure 500 {object} media.MediaError
// @Router /media/{id}/tags [put]
func (c *MediaController) SetMediaTagsHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := mediaIDParam(r)
	if !ok {
		writeMediaError(w, http.StatusNotFound, CodeNotFound, mediaSvc.ErrNotFound.Error())
		return
	}

	var req SetMediaTagsRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeMediaError(w, http.StatusBadRequest, CodeInvalidRequest, "invalid request body")
		return
	}

	m, err := c.Services.SetTags(id, mediaSvc.Tags{Situations: req.Situations, FamilyMemberIDs: req.FamilyMemberIDs})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeMediaJSON(w, http.StatusOK, m)
}

// DeleteMediaHandler godoc
// @Summary Delete a media item and its file
// @Description Also deletes photo widgets showing it and scheduled items that play it, and clears it from message widgets and audio backgrounds.
// @Tags media
// @Param id path int true "Media ID"
// @Success 204 {string} string "No Content"
// @Failure 404 {object} media.MediaError
// @Failure 500 {object} media.MediaError
// @Router /media/{id} [delete]
func (c *MediaController) DeleteMediaHandler(w http.ResponseWriter, r *http.Request) {
	id, ok := mediaIDParam(r)
	if !ok {
		writeMediaError(w, http.StatusNotFound, CodeNotFound, mediaSvc.ErrNotFound.Error())
		return
	}

	if err := c.Services.Delete(r.Context(), id); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

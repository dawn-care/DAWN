package family

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	model "server/internal/models/family"
	svc "server/internal/services/family"
)

// Error codes returned in the "code" field of error responses
const (
	CodeInvalidRequest      = "INVALID_REQUEST"
	CodeInvalidName         = "INVALID_NAME"
	CodeInvalidRelationship = "INVALID_RELATIONSHIP"
	CodeInvalidGuidance     = "INVALID_GUIDANCE_NOTE"
	CodeInternal            = "INTERNAL_ERROR"
)

// ---- Interface ----
type FamilyServiceInterface interface {
	ListFamilyMembers(carepartnerID int) ([]model.FamilyMember, error)
	CreateFamilyMember(in svc.CreateInput) (*model.FamilyMember, error)
}

// ---- Controller ----
type FamilyController struct {
	Service FamilyServiceInterface
}

func NewFamilyController(s FamilyServiceInterface) *FamilyController {
	return &FamilyController{Service: s}
}

// FamilyError is the body of every error response from these endpoints
type FamilyError struct {
	Code  string `json:"code"`
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, FamilyError{Code: code, Error: msg})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, svc.ErrInvalidCarepartner):
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, err.Error())
	case errors.Is(err, svc.ErrInvalidName):
		writeError(w, http.StatusBadRequest, CodeInvalidName, err.Error())
	case errors.Is(err, svc.ErrInvalidRelationship):
		writeError(w, http.StatusBadRequest, CodeInvalidRelationship, err.Error())
	case errors.Is(err, svc.ErrInvalidGuidance):
		writeError(w, http.StatusBadRequest, CodeInvalidGuidance, err.Error())
	default:
		// Never expose SQL errors
		writeError(w, http.StatusInternalServerError, CodeInternal, "something went wrong")
	}
}

// CreateFamilyMemberRequest is the body of POST /family-members
type CreateFamilyMemberRequest struct {
	CarepartnerID int     `json:"carepartner_id"`
	Name          string  `json:"name"`
	Relationship  string  `json:"relationship"`
	GuidanceNote  *string `json:"guidance_note"`
}

// ListFamilyMembersHandler godoc
// @Summary List a carepartner's family members
// @Description In the order they were added. Family members are folders for grouping media, not accounts.
// @Tags family
// @Produce json
// @Param carepartner_id query int true "Carepartner ID"
// @Success 200 {array} family.FamilyMember
// @Failure 400 {object} family.FamilyError
// @Failure 500 {object} family.FamilyError
// @Router /family-members [get]
func (c *FamilyController) ListFamilyMembersHandler(w http.ResponseWriter, r *http.Request) {
	carepartnerID, err := strconv.Atoi(r.URL.Query().Get("carepartner_id"))
	if err != nil || carepartnerID <= 0 {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, svc.ErrInvalidCarepartner.Error())
		return
	}
	members, err := c.Service.ListFamilyMembers(carepartnerID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, members)
}

// CreateFamilyMemberHandler godoc
// @Summary Add a family member
// @Description Name is required; relationship and guidance_note are optional.
// @Tags family
// @Accept json
// @Produce json
// @Param body body family.CreateFamilyMemberRequest true "Family member"
// @Success 201 {object} family.FamilyMember
// @Failure 400 {object} family.FamilyError
// @Failure 500 {object} family.FamilyError
// @Router /family-members [post]
func (c *FamilyController) CreateFamilyMemberHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var req CreateFamilyMemberRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, CodeInvalidRequest, "invalid request body")
		return
	}
	created, err := c.Service.CreateFamilyMember(svc.CreateInput{
		CarepartnerID: req.CarepartnerID,
		Name:          req.Name,
		Relationship:  req.Relationship,
		GuidanceNote:  req.GuidanceNote,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

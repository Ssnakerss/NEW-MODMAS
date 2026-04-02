package field

import (
	"encoding/json"
	"net/http"

	"github.com/Ssnakerss/modmas/internal/types"
	"github.com/Ssnakerss/modmas/pkg/response"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ─── Create ───────────────────────────────────────────────────────────────────

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")

	var input types.CreateFieldInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if input.Name == "" {
		response.BadRequest(w, "field name is required")
		return
	}

	if !ValidFieldType(input.FieldType) {
		response.BadRequest(w, "invalid field type: "+input.FieldType)
		return
	}

	if input.Options != nil {
		if err := ValidateOptions(input.FieldType, input.Options); err != nil {
			response.BadRequest(w, err.Error())
			return
		}
	}

	field, err := h.service.Create(r.Context(), spreadsheetID, input)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.Created(w, field)
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	fieldID := chi.URLParam(r, "fieldId")

	var input types.UpdateFieldInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if input.FieldType != nil && !ValidFieldType(*input.FieldType) {
		response.BadRequest(w, "invalid field type: "+*input.FieldType)
		return
	}

	if input.FieldType != nil && input.Options != nil {
		if err := ValidateOptions(*input.FieldType, input.Options); err != nil {
			response.BadRequest(w, err.Error())
			return
		}
	}

	field, err := h.service.Update(r.Context(), fieldID, input)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, field)
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	fieldID := chi.URLParam(r, "fieldId")

	if err := h.service.Delete(r.Context(), fieldID); err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.NoContent(w)
}

// ─── List ─────────────────────────────────────────────────────────────────────

func (h *Handler) ListBySpreadsheet(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")

	fields, err := h.service.ListBySpreadsheet(r.Context(), spreadsheetID)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, fields)
}

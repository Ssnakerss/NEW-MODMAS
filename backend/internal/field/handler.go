package field

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Ssnakerss/modmas/internal/types"
	"github.com/Ssnakerss/modmas/pkg/response"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// ─── Create ───────────────────────────────────────────────────────────────────

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")

	var input types.CreateFieldInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("failed to decode request body", "error", err, "handler", "field.Create")
		response.BadRequest(w, "invalid request body")
		return
	}

	if input.Name == "" {
		h.logger.Error("validation failed", "error", "field name is required", "handler", "field.Create")
		response.BadRequest(w, "field name is required")
		return
	}

	if !ValidFieldType(input.FieldType) {
		h.logger.Error("validation failed", "error", "invalid field type: "+input.FieldType, "handler", "field.Create")
		response.BadRequest(w, "invalid field type: "+input.FieldType)
		return
	}

	if input.Options != nil {
		if err := ValidateOptions(input.FieldType, input.Options); err != nil {
			h.logger.Error("validation failed", "error", err, "handler", "field.Create")
			response.BadRequest(w, err.Error())
			return
		}
	}

	field, err := h.service.Create(r.Context(), spreadsheetID, input)
	if err != nil {
		h.logger.Error("failed to create field", "error", err, "handler", "field.Create", "spreadsheetId", spreadsheetID)
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
		h.logger.Error("failed to decode request body", "error", err, "handler", "field.Update")
		response.BadRequest(w, "invalid request body")
		return
	}

	if input.FieldType != nil && !ValidFieldType(*input.FieldType) {
		h.logger.Error("validation failed", "error", "invalid field type: "+*input.FieldType, "handler", "field.Update")
		response.BadRequest(w, "invalid field type: "+*input.FieldType)
		return
	}

	if input.FieldType != nil && input.Options != nil {
		if err := ValidateOptions(*input.FieldType, input.Options); err != nil {
			h.logger.Error("validation failed", "error", err, "handler", "field.Update")
			response.BadRequest(w, err.Error())
			return
		}
	}

	field, err := h.service.Update(r.Context(), fieldID, input)
	if err != nil {
		h.logger.Error("failed to update field", "error", err, "handler", "field.Update", "fieldId", fieldID)
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, field)
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	fieldID := chi.URLParam(r, "fieldId")

	if err := h.service.Delete(r.Context(), fieldID); err != nil {
		h.logger.Error("failed to delete field", "error", err, "handler", "field.Delete", "fieldId", fieldID)
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
		h.logger.Error("failed to list fields by spreadsheet", "error", err, "handler", "field.ListBySpreadsheet", "spreadsheetId", spreadsheetID)
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, fields)
}

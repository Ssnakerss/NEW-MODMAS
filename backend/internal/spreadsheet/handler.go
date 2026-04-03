package spreadsheet

import (
	"encoding/json"
	"fmt"
	"net/http"

	"log/slog"

	"github.com/Ssnakerss/modmas/internal/middleware"
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
	userID := middleware.GetUserID(r.Context())

	var input types.CreateSpreadsheetInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("failed to decode request body", "error", err, "handler", "spreadsheet.Create")
		response.BadRequest(w, "invalid request body")
		return
	}

	if input.Name == "" {
		h.logger.Error("validation failed", "error", "name is required", "handler", "spreadsheet.Create")
		response.BadRequest(w, "name is required")
		return
	}

	if input.WorkspaceID == "" {
		h.logger.Error("validation failed", "error", "workspace_id is required", "handler", "spreadsheet.Create")
		response.BadRequest(w, "workspace_id is required")
		return
	}

	for i, f := range input.Fields {
		if f.Name == "" {
			h.logger.Error("validation failed", "error", fmt.Sprintf("field[%d]: name is required", i), "handler", "spreadsheet.Create")
			response.BadRequest(w, fmt.Sprintf("field[%d]: name is required", i))
			return
		}
		if !isValidFieldType(f.FieldType) {
			h.logger.Error("validation failed", "error", fmt.Sprintf("field[%d]: invalid field type '%s'", i, f.FieldType), "handler", "spreadsheet.Create")
			response.BadRequest(w, fmt.Sprintf("field[%d]: invalid field type '%s'", i, f.FieldType))
			return
		}
		if f.Options != nil {
			if err := validateFieldOptions(f.FieldType, f.Options); err != nil {
				h.logger.Error("validation failed", "error", err, "handler", "spreadsheet.Create")
				response.BadRequest(w, fmt.Sprintf("field[%d]: %s", i, err.Error()))
				return
			}
		}
	}

	result, err := h.service.Create(r.Context(), input, userID)
	if err != nil {
		h.logger.Error("failed to create spreadsheet", "error", err, "handler", "spreadsheet.Create")
		response.InternalError(w, err.Error())
		return
	}

	response.Created(w, result)

}

// ─── Get ──────────────────────────────────────────────────────────────────────

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := h.service.GetWithFields(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get spreadsheet", "error", err, "handler", "spreadsheet.Get", "id", id)
		response.NotFound(w, "spreadsheet not found")
		return
	}

	response.OK(w, result)
}

// ─── ListByWorkspace ──────────────────────────────────────────────────────────

func (h *Handler) ListByWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "workspaceId")

	result, err := h.service.ListByWorkspace(r.Context(), workspaceID)
	if err != nil {
		h.logger.Error("failed to list spreadsheets by workspace", "error", err, "handler", "spreadsheet.ListByWorkspace", "workspaceId", workspaceID)
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, result)
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("failed to decode request body", "error", err, "handler", "spreadsheet.Update")
		response.BadRequest(w, "invalid request body")
		return
	}

	if input.Name == "" {
		h.logger.Error("validation failed", "error", "name is required", "handler", "spreadsheet.Update")
		response.BadRequest(w, "name is required")
		return
	}

	result, err := h.service.Update(r.Context(), id, input.Name, input.Description)
	if err != nil {
		h.logger.Error("failed to update spreadsheet", "error", err, "handler", "spreadsheet.Update", "id", id)
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, result)
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.Delete(r.Context(), id); err != nil {
		h.logger.Error("failed to delete spreadsheet", "error", err, "handler", "spreadsheet.Delete", "id", id)
		response.InternalError(w, err.Error())
		return
	}

	response.NoContent(w)
}

// ─── Вспомогательные функции ──────────────────────────────────────────────────

func isValidFieldType(fieldType string) bool {
	switch fieldType {
	case "text", "integer", "decimal", "boolean",
		"date", "datetime", "select", "multi_select",
		"email", "url", "phone", "attachment":
		return true
	}
	return false
}

func validateFieldOptions(fieldType string, options map[string]interface{}) error {
	switch fieldType {
	case "select", "multi_select":
		choices, ok := options["choices"]
		if !ok {
			return fmt.Errorf("select field must have 'choices' option")
		}
		arr, ok := choices.([]interface{})
		if !ok || len(arr) == 0 {
			return fmt.Errorf("'choices' must be a non-empty array")
		}
		for i, c := range arr {
			choice, ok := c.(map[string]interface{})
			if !ok {
				return fmt.Errorf("choice[%d] must be an object", i)
			}
			if _, ok := choice["value"].(string); !ok {
				return fmt.Errorf("choice[%d] must have a 'value' string field", i)
			}
			if _, ok := choice["label"].(string); !ok {
				return fmt.Errorf("choice[%d] must have a 'label' string field", i)
			}
		}
	}
	return nil
}

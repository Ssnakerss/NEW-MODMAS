package row

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Ssnakerss/modmas/internal/middleware"
	permissionPkg "github.com/Ssnakerss/modmas/internal/permission"
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

type queryRequest struct {
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
	Filters []FilterCondition `json:"filters"`
	Sorts   []SortCondition   `json:"sorts"`
}

func (h *Handler) Query(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	var req queryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("failed to decode request body, using defaults", "error", err, "handler", "row.Query")
		req = queryRequest{Limit: 50}
	}

	result, err := h.service.Query(r.Context(), userID, spreadsheetID, QueryInput{
		Limit:   req.Limit,
		Offset:  req.Offset,
		Filters: req.Filters,
		Sorts:   req.Sorts,
	})
	if err != nil {
		if permissionPkg.IsForbidden(err) {
			h.logger.Error("permission denied", "error", err, "handler", "row.Query", "userId", userID, "spreadsheetId", spreadsheetID)
			response.Forbidden(w, err.Error())
			return
		}
		h.logger.Error("failed to query rows", "error", err, "handler", "row.Query", "userId", userID, "spreadsheetId", spreadsheetID)
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, result)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	var data RowData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.logger.Warn("failed to decode request body, using empty data", "error", err, "handler", "row.Create")
		data = RowData{}
	}

	row, err := h.service.Create(r.Context(), userID, spreadsheetID, data)
	if err != nil {
		if permissionPkg.IsForbidden(err) {
			h.logger.Error("permission denied", "error", err, "handler", "row.Create", "userId", userID, "spreadsheetId", spreadsheetID)
			response.Forbidden(w, err.Error())
			return
		}
		h.logger.Error("failed to create row", "error", err, "handler", "row.Create", "userId", userID, "spreadsheetId", spreadsheetID)
		response.BadRequest(w, err.Error())
		return
	}

	response.Created(w, row)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	rowID := chi.URLParam(r, "rowId")
	userID := middleware.GetUserID(r.Context())

	var data RowData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		h.logger.Error("failed to decode request body", "error", err, "handler", "row.Update")
		response.BadRequest(w, "invalid request body")
		return
	}

	row, err := h.service.Update(r.Context(), userID, spreadsheetID, rowID, data)
	if err != nil {
		if permissionPkg.IsForbidden(err) {
			h.logger.Error("permission denied", "error", err, "handler", "row.Update", "userId", userID, "spreadsheetId", spreadsheetID, "rowId", rowID)
			response.Forbidden(w, err.Error())
			return
		}
		h.logger.Error("failed to update row", "error", err, "handler", "row.Update", "userId", userID, "spreadsheetId", spreadsheetID, "rowId", rowID)
		response.BadRequest(w, err.Error())
		return
	}

	response.OK(w, row)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	rowID := chi.URLParam(r, "rowId")
	userID := middleware.GetUserID(r.Context())

	if err := h.service.Delete(r.Context(), userID, spreadsheetID, rowID); err != nil {
		if permissionPkg.IsForbidden(err) {
			h.logger.Error("permission denied", "error", err, "handler", "row.Delete", "userId", userID, "spreadsheetId", spreadsheetID, "rowId", rowID)
			response.Forbidden(w, err.Error())
			return
		}
		h.logger.Error("failed to delete row", "error", err, "handler", "row.Delete", "userId", userID, "spreadsheetId", spreadsheetID, "rowId", rowID)
		response.InternalError(w, err.Error())
		return
	}

	response.NoContent(w)
}

func (h *Handler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	var input BulkDeleteInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Error("failed to decode request body", "error", err, "handler", "row.BulkDelete")
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.service.BulkDelete(r.Context(), userID, spreadsheetID, input); err != nil {
		if permissionPkg.IsForbidden(err) {
			h.logger.Error("permission denied", "error", err, "handler", "row.BulkDelete", "userId", userID, "spreadsheetId", spreadsheetID)
			response.Forbidden(w, err.Error())
			return
		}
		h.logger.Error("failed to bulk delete rows", "error", err, "handler", "row.BulkDelete", "userId", userID, "spreadsheetId", spreadsheetID)
		response.BadRequest(w, err.Error())
		return
	}

	response.NoContent(w)
}

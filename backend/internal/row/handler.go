package row

import (
	"encoding/json"
	"net/http"

	"github.com/Ssnakerss/modmas/internal/middleware"
	permissionPkg "github.com/Ssnakerss/modmas/internal/permission"
	"github.com/Ssnakerss/modmas/pkg/response"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
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
			response.Forbidden(w, err.Error())
			return
		}
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
		data = RowData{}
	}

	row, err := h.service.Create(r.Context(), userID, spreadsheetID, data)
	if err != nil {
		if permissionPkg.IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
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
		response.BadRequest(w, "invalid request body")
		return
	}

	row, err := h.service.Update(r.Context(), userID, spreadsheetID, rowID, data)
	if err != nil {
		if permissionPkg.IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
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
			response.Forbidden(w, err.Error())
			return
		}
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
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.service.BulkDelete(r.Context(), userID, spreadsheetID, input); err != nil {
		if permissionPkg.IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.BadRequest(w, err.Error())
		return
	}

	response.NoContent(w)
}

package permission

import (
	"encoding/json"
	"net/http"

	"github.com/Ssnakerss/modmas/internal/middleware"
	"github.com/Ssnakerss/modmas/pkg/response"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ─── Spreadsheet Access ───────────────────────────────────────────────────────

func (h *Handler) GetSpreadsheetAccess(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	accesses, err := h.service.GetSpreadsheetAccess(r.Context(), userID, spreadsheetID)
	if err != nil {
		if IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, accesses)
}

func (h *Handler) UpsertSpreadsheetAccess(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	var input UpsertSpreadsheetAccessInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.service.UpsertSpreadsheetAccess(r.Context(), userID, spreadsheetID, input); err != nil {
		if IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.BadRequest(w, err.Error())
		return
	}

	response.NoContent(w)
}

func (h *Handler) RemoveSpreadsheetAccess(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	principalID := chi.URLParam(r, "principalId")
	userID := middleware.GetUserID(r.Context())

	if err := h.service.RemoveSpreadsheetAccess(r.Context(), userID, spreadsheetID, principalID); err != nil {
		if IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.InternalError(w, err.Error())
		return
	}

	response.NoContent(w)
}

// ─── Field Access ─────────────────────────────────────────────────────────────

func (h *Handler) GetFieldAccess(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	accesses, err := h.service.GetFieldAccess(r.Context(), userID, spreadsheetID)
	if err != nil {
		if IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, accesses)
}

func (h *Handler) UpsertFieldAccess(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	fieldID := chi.URLParam(r, "fieldId")
	userID := middleware.GetUserID(r.Context())

	var input UpsertFieldAccessInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	input.FieldID = fieldID

	if err := h.service.UpsertFieldAccess(r.Context(), userID, spreadsheetID, input); err != nil {
		if IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.BadRequest(w, err.Error())
		return
	}

	response.NoContent(w)
}

// ─── Row Rules ────────────────────────────────────────────────────────────────

func (h *Handler) GetRowRules(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	rules, err := h.service.GetRowRules(r.Context(), userID, spreadsheetID)
	if err != nil {
		if IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, rules)
}

func (h *Handler) UpsertRowRule(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	var input UpsertRowRuleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	rule, err := h.service.UpsertRowRule(r.Context(), userID, spreadsheetID, input)
	if err != nil {
		if IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.BadRequest(w, err.Error())
		return
	}

	response.OK(w, rule)
}

func (h *Handler) DeleteRowRule(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	ruleID := chi.URLParam(r, "ruleId")
	userID := middleware.GetUserID(r.Context())

	if err := h.service.DeleteRowRule(r.Context(), userID, spreadsheetID, ruleID); err != nil {
		if IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.InternalError(w, err.Error())
		return
	}

	response.NoContent(w)
}

// ─── My Permissions ───────────────────────────────────────────────────────────

func (h *Handler) GetMyPermissions(w http.ResponseWriter, r *http.Request) {
	spreadsheetID := chi.URLParam(r, "id")
	userID := middleware.GetUserID(r.Context())

	perms, err := h.service.GetMyPermissions(r.Context(), userID, spreadsheetID)
	if err != nil {
		if IsForbidden(err) {
			response.Forbidden(w, err.Error())
			return
		}
		response.InternalError(w, err.Error())
		return
	}

	response.OK(w, perms)
}

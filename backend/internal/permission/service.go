package permission

import (
	"context"
	"fmt"

	"github.com/Ssnakerss/modmas/internal/types"
	workspacePkg "github.com/Ssnakerss/modmas/internal/workspace"
)

type Service struct {
	repo     *Repository
	wsRepo   *workspacePkg.Repository
	enforcer *Enforcer
}

func NewService(
	repo *Repository,
	wsRepo *workspacePkg.Repository,
	enforcer *Enforcer,
) *Service {
	return &Service{
		repo:     repo,
		wsRepo:   wsRepo,
		enforcer: enforcer,
	}
}

// ─── Spreadsheet Access ───────────────────────────────────────────────────────

type UpsertSpreadsheetAccessInput struct {
	PrincipalID   string `json:"principal_id"`
	PrincipalType string `json:"principal_type"`
	CanView       bool   `json:"can_view"`
	CanInsert     bool   `json:"can_insert"`
	CanEdit       bool   `json:"can_edit"`
	CanDelete     bool   `json:"can_delete"`
	CanManage     bool   `json:"can_manage"`
}

type RemoveSpreadsheetAccessInput struct {
	PrincipalID string `json:"principal_id"`
}

// GetSpreadsheetAccess возвращает все права доступа к таблице.
// Только пользователи с правом can_manage могут просматривать список прав.
func (s *Service) GetSpreadsheetAccess(ctx context.Context, requesterID, spreadsheetID string) ([]*types.SpreadsheetAccess, error) {
	if err := s.enforcer.RequireManage(ctx, requesterID, spreadsheetID); err != nil {
		return nil, err
	}

	accesses, err := s.repo.GetSpreadsheetAccess(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("get spreadsheet access: %w", err)
	}

	return accesses, nil
}

// UpsertSpreadsheetAccess создаёт или обновляет права доступа к таблице.
// Только пользователи с правом can_manage могут изменять права.
func (s *Service) UpsertSpreadsheetAccess(
	ctx context.Context,
	requesterID, spreadsheetID string,
	input UpsertSpreadsheetAccessInput,
) error {
	if err := s.enforcer.RequireManage(ctx, requesterID, spreadsheetID); err != nil {
		return err
	}

	if input.PrincipalID == "" {
		return fmt.Errorf("principal_id is required")
	}

	if input.PrincipalType == "" {
		return fmt.Errorf("principal_type is required")
	}

	if input.PrincipalType != "user" && input.PrincipalType != "workspace_role" {
		return fmt.Errorf("principal_type must be 'user' or 'workspace_role'")
	}

	// Нельзя изменить права владельца workspace
	isOwner, err := s.isWorkspaceOwner(ctx, input.PrincipalID, spreadsheetID)
	if err != nil {
		return fmt.Errorf("check owner: %w", err)
	}
	if isOwner {
		return fmt.Errorf("cannot modify owner permissions")
	}

	access := &types.SpreadsheetAccess{
		SpreadsheetID: spreadsheetID,
		PrincipalID:   input.PrincipalID,
		PrincipalType: input.PrincipalType,
		CanView:       input.CanView,
		CanInsert:     input.CanInsert,
		CanEdit:       input.CanEdit,
		CanDelete:     input.CanDelete,
		CanManage:     input.CanManage,
	}

	if err := s.repo.UpsertSpreadsheetAccess(ctx, access); err != nil {
		return fmt.Errorf("upsert spreadsheet access: %w", err)
	}

	return nil
}

// RemoveSpreadsheetAccess удаляет права доступа пользователя к таблице.
// Только пользователи с правом can_manage могут удалять права.
func (s *Service) RemoveSpreadsheetAccess(
	ctx context.Context,
	requesterID, spreadsheetID, principalID string,
) error {
	if err := s.enforcer.RequireManage(ctx, requesterID, spreadsheetID); err != nil {
		return err
	}

	// Нельзя удалить права владельца workspace
	isOwner, err := s.isWorkspaceOwner(ctx, principalID, spreadsheetID)
	if err != nil {
		return fmt.Errorf("check owner: %w", err)
	}
	if isOwner {
		return fmt.Errorf("cannot remove owner permissions")
	}

	if err := s.repo.RemoveSpreadsheetAccess(ctx, spreadsheetID, principalID); err != nil {
		return fmt.Errorf("remove spreadsheet access: %w", err)
	}

	return nil
}

// ─── Field Access ─────────────────────────────────────────────────────────────

type UpsertFieldAccessInput struct {
	FieldID       string `json:"field_id"`
	PrincipalID   string `json:"principal_id"`
	PrincipalType string `json:"principal_type"`
	CanView       bool   `json:"can_view"`
	CanEdit       bool   `json:"can_edit"`
}

// GetFieldAccess возвращает права доступа ко всем полям таблицы.
func (s *Service) GetFieldAccess(ctx context.Context, requesterID, spreadsheetID string) ([]*types.FieldAccess, error) {
	if err := s.enforcer.RequireManage(ctx, requesterID, spreadsheetID); err != nil {
		return nil, err
	}

	accesses, err := s.repo.GetFieldAccess(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("get field access: %w", err)
	}

	return accesses, nil
}

// UpsertFieldAccess создаёт или обновляет права доступа к полю.
func (s *Service) UpsertFieldAccess(
	ctx context.Context,
	requesterID, spreadsheetID string,
	input UpsertFieldAccessInput,
) error {
	if err := s.enforcer.RequireManage(ctx, requesterID, spreadsheetID); err != nil {
		return err
	}

	if input.FieldID == "" {
		return fmt.Errorf("field_id is required")
	}

	if input.PrincipalID == "" {
		return fmt.Errorf("principal_id is required")
	}

	if input.PrincipalType == "" {
		return fmt.Errorf("principal_type is required")
	}

	if input.PrincipalType != "user" && input.PrincipalType != "workspace_role" {
		return fmt.Errorf("principal_type must be 'user' or 'workspace_role'")
	}

	fa := &types.FieldAccess{
		FieldID:       input.FieldID,
		PrincipalID:   input.PrincipalID,
		PrincipalType: input.PrincipalType,
		CanView:       input.CanView,
		CanEdit:       input.CanEdit,
	}

	if err := s.repo.UpsertFieldAccess(ctx, fa); err != nil {
		return fmt.Errorf("upsert field access: %w", err)
	}

	return nil
}

// ─── Row Access Rules ─────────────────────────────────────────────────────────

type UpsertRowRuleInput struct {
	PrincipalID   string                 `json:"principal_id"`
	PrincipalType string                 `json:"principal_type"`
	Condition     map[string]interface{} `json:"condition"`
	CanView       bool                   `json:"can_view"`
	CanEdit       bool                   `json:"can_edit"`
}

type RowAccessRule struct {
	ID            string                 `json:"id"`
	SpreadsheetID string                 `json:"spreadsheet_id"`
	PrincipalID   string                 `json:"principal_id"`
	PrincipalType string                 `json:"principal_type"`
	Condition     map[string]interface{} `json:"condition"`
	CanView       bool                   `json:"can_view"`
	CanEdit       bool                   `json:"can_edit"`
}

// GetRowRules возвращает все row-level правила для таблицы.
func (s *Service) GetRowRules(ctx context.Context, requesterID, spreadsheetID string) ([]*RowAccessRule, error) {
	if err := s.enforcer.RequireManage(ctx, requesterID, spreadsheetID); err != nil {
		return nil, err
	}

	rules, err := s.repo.GetRowRules(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("get row rules: %w", err)
	}

	return rules, nil
}

// UpsertRowRule создаёт или обновляет row-level правило.
func (s *Service) UpsertRowRule(
	ctx context.Context,
	requesterID, spreadsheetID string,
	input UpsertRowRuleInput,
) (*RowAccessRule, error) {
	if err := s.enforcer.RequireManage(ctx, requesterID, spreadsheetID); err != nil {
		return nil, err
	}

	if input.PrincipalID == "" {
		return nil, fmt.Errorf("principal_id is required")
	}

	if input.PrincipalType != "user" && input.PrincipalType != "workspace_role" {
		return nil, fmt.Errorf("principal_type must be 'user' or 'workspace_role'")
	}

	if len(input.Condition) == 0 {
		return nil, fmt.Errorf("condition is required")
	}

	// Валидируем condition
	if err := validateRowCondition(input.Condition); err != nil {
		return nil, fmt.Errorf("invalid condition: %w", err)
	}

	rule := &RowAccessRule{
		SpreadsheetID: spreadsheetID,
		PrincipalID:   input.PrincipalID,
		PrincipalType: input.PrincipalType,
		Condition:     input.Condition,
		CanView:       input.CanView,
		CanEdit:       input.CanEdit,
	}

	created, err := s.repo.UpsertRowRule(ctx, rule)
	if err != nil {
		return nil, fmt.Errorf("upsert row rule: %w", err)
	}

	return created, nil
}

// DeleteRowRule удаляет row-level правило по ID.
func (s *Service) DeleteRowRule(ctx context.Context, requesterID, spreadsheetID, ruleID string) error {
	if err := s.enforcer.RequireManage(ctx, requesterID, spreadsheetID); err != nil {
		return err
	}

	if err := s.repo.DeleteRowRule(ctx, ruleID); err != nil {
		return fmt.Errorf("delete row rule: %w", err)
	}

	return nil
}

// ─── My Permissions ───────────────────────────────────────────────────────────

type MyPermissions struct {
	Spreadsheet *SpreadsheetPerms      `json:"spreadsheet"`
	Fields      map[string]*FieldPerms `json:"fields"`
}

// GetMyPermissions возвращает все права текущего пользователя на таблицу.
// Используется фронтендом для отображения доступных действий.
func (s *Service) GetMyPermissions(ctx context.Context, userID, spreadsheetID string) (*MyPermissions, error) {
	spreadPerms, err := s.enforcer.GetSpreadsheetPerms(ctx, userID, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("get spreadsheet perms: %w", err)
	}

	// Если нет права на просмотр — возвращаем только это
	if !spreadPerms.CanView {
		return &MyPermissions{
			Spreadsheet: spreadPerms,
			Fields:      map[string]*FieldPerms{},
		}, nil
	}

	// Загружаем права по полям
	visibleFields, err := s.enforcer.VisibleFields(ctx, userID, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("get visible fields: %w", err)
	}

	fieldPerms := make(map[string]*FieldPerms, len(visibleFields))
	for _, f := range visibleFields {
		fp, err := s.enforcer.GetFieldPerms(ctx, userID, f.ID)
		if err != nil {
			continue
		}
		fieldPerms[f.ID] = fp
	}

	return &MyPermissions{
		Spreadsheet: spreadPerms,
		Fields:      fieldPerms,
	}, nil
}

// ─── Вспомогательные методы ───────────────────────────────────────────────────

// isWorkspaceOwner проверяет, является ли пользователь владельцем workspace таблицы
func (s *Service) isWorkspaceOwner(ctx context.Context, userID, spreadsheetID string) (bool, error) {
	return s.enforcer.isWorkspaceOwner(ctx, userID, spreadsheetID)
}

// validateRowCondition проверяет корректность JSON-условия row-level правила
func validateRowCondition(condition map[string]interface{}) error {
	op, ok := condition["op"].(string)
	if !ok || op == "" {
		return fmt.Errorf("condition must have 'op' field")
	}

	validOps := map[string]bool{
		"eq":              true,
		"neq":             true,
		"contains":        true,
		"eq_current_user": true,
	}

	if !validOps[op] {
		return fmt.Errorf("unknown op %q, valid ops: eq, neq, contains, eq_current_user", op)
	}

	if _, ok := condition["column_name"].(string); !ok {
		return fmt.Errorf("condition must have 'column_name' field")
	}

	// Для операций с явным значением требуем поле value
	if op == "eq" || op == "neq" || op == "contains" {
		if _, ok := condition["value"]; !ok {
			return fmt.Errorf("condition with op %q must have 'value' field", op)
		}
	}

	return nil
}

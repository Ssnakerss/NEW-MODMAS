package permission

import (
	"context"
	"encoding/json"
	"fmt"

	fieldPkg "github.com/Ssnakerss/modmas/internal/field"
	"github.com/Ssnakerss/modmas/internal/types"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Enforcer проверяет права доступа пользователя к объектам системы
type Enforcer struct {
	pool      *pgxpool.Pool
	permRepo  *Repository
	fieldRepo *fieldPkg.Repository
}

func NewEnforcer(pool *pgxpool.Pool, permRepo *Repository, fieldRepo *fieldPkg.Repository) *Enforcer {
	return &Enforcer{
		pool:      pool,
		permRepo:  permRepo,
		fieldRepo: fieldRepo,
	}
}

// ─── Права на таблицу ─────────────────────────────────────────────────────────

// SpreadsheetPerms содержит все права пользователя на конкретную таблицу
type SpreadsheetPerms struct {
	CanView   bool
	CanInsert bool
	CanEdit   bool
	CanDelete bool
	CanManage bool
}

// GetSpreadsheetPerms возвращает права пользователя на таблицу.
// Владелец workspace получает полные права автоматически.
func (e *Enforcer) GetSpreadsheetPerms(ctx context.Context, userID, spreadsheetID string) (*SpreadsheetPerms, error) {
	isOwner, err := e.isWorkspaceOwner(ctx, userID, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("check workspace owner: %w", err)
	}
	if isOwner {
		return &SpreadsheetPerms{
			CanView:   true,
			CanInsert: true,
			CanEdit:   true,
			CanDelete: true,
			CanManage: true,
		}, nil
	}

	access, err := e.permRepo.CheckAccess(ctx, spreadsheetID, userID)
	if err != nil {
		// Нет записи — нет доступа
		return &SpreadsheetPerms{}, nil
	}

	return &SpreadsheetPerms{
		CanView:   access.CanView,
		CanInsert: access.CanInsert,
		CanEdit:   access.CanEdit,
		CanDelete: access.CanDelete,
		CanManage: access.CanManage,
	}, nil
}

// RequireView возвращает ошибку если пользователь не может просматривать таблицу
func (e *Enforcer) RequireView(ctx context.Context, userID, spreadsheetID string) error {
	perms, err := e.GetSpreadsheetPerms(ctx, userID, spreadsheetID)
	if err != nil {
		return err
	}
	if !perms.CanView {
		return ErrForbidden("view", "spreadsheet", spreadsheetID)
	}
	return nil
}

// RequireInsert возвращает ошибку если пользователь не может добавлять строки
func (e *Enforcer) RequireInsert(ctx context.Context, userID, spreadsheetID string) error {
	perms, err := e.GetSpreadsheetPerms(ctx, userID, spreadsheetID)
	if err != nil {
		return err
	}
	if !perms.CanInsert {
		return ErrForbidden("insert", "spreadsheet", spreadsheetID)
	}
	return nil
}

// RequireEdit возвращает ошибку если пользователь не может редактировать строки
func (e *Enforcer) RequireEdit(ctx context.Context, userID, spreadsheetID string) error {
	perms, err := e.GetSpreadsheetPerms(ctx, userID, spreadsheetID)
	if err != nil {
		return err
	}
	if !perms.CanEdit {
		return ErrForbidden("edit", "spreadsheet", spreadsheetID)
	}
	return nil
}

// RequireDelete возвращает ошибку если пользователь не может удалять строки
func (e *Enforcer) RequireDelete(ctx context.Context, userID, spreadsheetID string) error {
	perms, err := e.GetSpreadsheetPerms(ctx, userID, spreadsheetID)
	if err != nil {
		return err
	}
	if !perms.CanDelete {
		return ErrForbidden("delete", "spreadsheet", spreadsheetID)
	}
	return nil
}

// RequireManage возвращает ошибку если пользователь не может управлять структурой таблицы
func (e *Enforcer) RequireManage(ctx context.Context, userID, spreadsheetID string) error {
	perms, err := e.GetSpreadsheetPerms(ctx, userID, spreadsheetID)
	if err != nil {
		return err
	}
	if !perms.CanManage {
		return ErrForbidden("manage", "spreadsheet", spreadsheetID)
	}
	return nil
}

// ─── Права на поля ────────────────────────────────────────────────────────────

// FieldPerms содержит права пользователя на конкретное поле
type FieldPerms struct {
	CanView bool
	CanEdit bool
}

// GetFieldPerms возвращает права пользователя на конкретное поле.
// Если явных ограничений нет — по умолчанию разрешено всё.
func (e *Enforcer) GetFieldPerms(ctx context.Context, userID, fieldID string) (*FieldPerms, error) {
	var canView, canEdit bool

	err := e.pool.QueryRow(ctx, `
		SELECT
			COALESCE(fa.can_view, true),
			COALESCE(fa.can_edit, true)
		FROM meta.fields f
		LEFT JOIN meta.field_access fa
			ON fa.field_id = f.id
			AND fa.principal_id = $2
		WHERE f.id = $1
	`, fieldID, userID).Scan(&canView, &canEdit)
	if err != nil {
		return nil, fmt.Errorf("get field perms: %w", err)
	}

	return &FieldPerms{CanView: canView, CanEdit: canEdit}, nil
}

// VisibleFields возвращает только те поля, которые пользователь может видеть
func (e *Enforcer) VisibleFields(ctx context.Context, userID, spreadsheetID string) ([]*types.Field, error) {
	fields, err := e.fieldRepo.ListBySpreadsheet(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("list fields: %w", err)
	}

	hiddenFields, err := e.getHiddenFieldIDs(ctx, userID, spreadsheetID)
	if err != nil {
		return nil, err
	}

	if len(hiddenFields) == 0 {
		return fields, nil
	}

	visible := make([]*types.Field, 0, len(fields))
	for _, f := range fields {
		if !hiddenFields[f.ID] {
			visible = append(visible, f)
		}
	}
	return visible, nil
}

// EditableFields возвращает только те поля, которые пользователь может редактировать
func (e *Enforcer) EditableFields(ctx context.Context, userID, spreadsheetID string) ([]*types.Field, error) {
	fields, err := e.fieldRepo.ListBySpreadsheet(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("list fields: %w", err)
	}

	readonlyFields, err := e.getReadonlyFieldIDs(ctx, userID, spreadsheetID)
	if err != nil {
		return nil, err
	}

	if len(readonlyFields) == 0 {
		return fields, nil
	}

	editable := make([]*types.Field, 0, len(fields))
	for _, f := range fields {
		if !readonlyFields[f.ID] {
			editable = append(editable, f)
		}
	}
	return editable, nil
}

// ─── Row-level фильтрация ─────────────────────────────────────────────────────

// RowFilter описывает SQL-условие для фильтрации строк
type RowFilter struct {
	SQL  string
	Args []interface{}
}

// GetRowFilters возвращает SQL-условия для фильтрации строк
// на основе row_access_rules для данного пользователя
func (e *Enforcer) GetRowFilters(ctx context.Context, userID, spreadsheetID string) ([]RowFilter, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT condition
		FROM meta.row_access_rules
		WHERE spreadsheet_id = $1
		  AND principal_id = $2
		  AND can_view = true
	`, spreadsheetID, userID)
	if err != nil {
		return nil, fmt.Errorf("get row rules: %w", err)
	}
	defer rows.Close()

	var filters []RowFilter
	for rows.Next() {
		var condJSON []byte
		if err := rows.Scan(&condJSON); err != nil {
			return nil, err
		}

		var cond map[string]interface{}
		if err := json.Unmarshal(condJSON, &cond); err != nil {
			continue
		}

		filter, err := e.buildRowFilter(cond, userID)
		if err != nil {
			continue
		}
		filters = append(filters, filter)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return filters, nil
}

// buildRowFilter преобразует JSON-условие правила в SQL-фрагмент
func (e *Enforcer) buildRowFilter(condition map[string]interface{}, userID string) (RowFilter, error) {
	colName, ok := condition["column_name"].(string)
	if !ok || colName == "" {
		return RowFilter{}, fmt.Errorf("missing column_name in condition")
	}

	op, ok := condition["op"].(string)
	if !ok || op == "" {
		return RowFilter{}, fmt.Errorf("missing op in condition")
	}

	quotedCol := fmt.Sprintf(`"%s"`, colName)

	switch op {
	case "eq_current_user":
		return RowFilter{
			SQL:  fmt.Sprintf(`%s = $1`, quotedCol),
			Args: []interface{}{userID},
		}, nil

	case "eq":
		val, _ := condition["value"].(string)
		return RowFilter{
			SQL:  fmt.Sprintf(`%s = $1`, quotedCol),
			Args: []interface{}{val},
		}, nil

	case "neq":
		val, _ := condition["value"].(string)
		return RowFilter{
			SQL:  fmt.Sprintf(`%s != $1`, quotedCol),
			Args: []interface{}{val},
		}, nil

	case "contains":
		val, _ := condition["value"].(string)
		return RowFilter{
			SQL:  fmt.Sprintf(`%s ILIKE $1`, quotedCol),
			Args: []interface{}{"%" + val + "%"},
		}, nil

	default:
		return RowFilter{}, fmt.Errorf("unknown op: %s", op)
	}
}

// ─── Вспомогательные методы ───────────────────────────────────────────────────

// isWorkspaceOwner проверяет, является ли пользователь владельцем workspace
func (e *Enforcer) isWorkspaceOwner(ctx context.Context, userID, spreadsheetID string) (bool, error) {
	var isOwner bool
	err := e.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM auth.workspaces w
			JOIN meta.spreadsheets s ON s.workspace_id = w.id
			WHERE s.id = $1
			  AND w.owner_id = $2
		)
	`, spreadsheetID, userID).Scan(&isOwner)
	if err != nil {
		return false, err
	}
	return isOwner, nil
}

// getHiddenFieldIDs возвращает set field_id скрытых для пользователя полей
func (e *Enforcer) getHiddenFieldIDs(ctx context.Context, userID, spreadsheetID string) (map[string]bool, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT fa.field_id
		FROM meta.field_access fa
		JOIN meta.fields f ON f.id = fa.field_id
		WHERE f.spreadsheet_id = $1
		  AND fa.principal_id = $2
		  AND fa.can_view = false
	`, spreadsheetID, userID)
	if err != nil {
		return nil, fmt.Errorf("get hidden fields: %w", err)
	}
	defer rows.Close()

	hidden := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		hidden[id] = true
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return hidden, nil
}

// getReadonlyFieldIDs возвращает set field_id полей недоступных для редактирования
func (e *Enforcer) getReadonlyFieldIDs(ctx context.Context, userID, spreadsheetID string) (map[string]bool, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT fa.field_id
		FROM meta.field_access fa
		JOIN meta.fields f ON f.id = fa.field_id
		WHERE f.spreadsheet_id = $1
		  AND fa.principal_id = $2
		  AND fa.can_edit = false
	`, spreadsheetID, userID)
	if err != nil {
		return nil, fmt.Errorf("get readonly fields: %w", err)
	}
	defer rows.Close()

	readonly := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		readonly[id] = true
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return readonly, nil
}

// ─── Ошибки ───────────────────────────────────────────────────────────────────

// AccessError описывает ошибку доступа
type AccessError struct {
	Action     string
	Resource   string
	ResourceID string
}

func (e *AccessError) Error() string {
	return fmt.Sprintf("forbidden: cannot %s %s %s", e.Action, e.Resource, e.ResourceID)
}

// ErrForbidden создаёт ошибку запрета доступа
func ErrForbidden(action, resource, resourceID string) error {
	return &AccessError{
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
	}
}

// IsForbidden проверяет, является ли ошибка ошибкой доступа
func IsForbidden(err error) bool {
	_, ok := err.(*AccessError)
	return ok
}

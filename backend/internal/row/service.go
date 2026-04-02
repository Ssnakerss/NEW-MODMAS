package row

import (
	"context"
	"fmt"
	"strings"

	fieldPkg "github.com/Ssnakerss/modmas/internal/field"
	permissionPkg "github.com/Ssnakerss/modmas/internal/permission"
	spreadsheetPkg "github.com/Ssnakerss/modmas/internal/spreadsheet"
	"github.com/Ssnakerss/modmas/internal/types"
)

type Service struct {
	repo       *Repository
	spreadRepo *spreadsheetPkg.Repository
	fieldRepo  *fieldPkg.Repository
	enforcer   *permissionPkg.Enforcer
}

func NewService(
	repo *Repository,
	spreadRepo *spreadsheetPkg.Repository,
	fieldRepo *fieldPkg.Repository,
	enforcer *permissionPkg.Enforcer,
) *Service {
	return &Service{
		repo:       repo,
		spreadRepo: spreadRepo,
		fieldRepo:  fieldRepo,
		enforcer:   enforcer,
	}
}

// ─── Query ────────────────────────────────────────────────────────────────────

type QueryInput struct {
	Limit   int
	Offset  int
	Filters []FilterCondition
	Sorts   []SortCondition
}

type QueryResult struct {
	Data   []RowData `json:"data"`
	Total  int       `json:"total"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
}

func (s *Service) Query(ctx context.Context, userID, spreadsheetID string, input QueryInput) (*QueryResult, error) {
	// Проверяем право на просмотр
	if err := s.enforcer.RequireView(ctx, userID, spreadsheetID); err != nil {
		return nil, err
	}

	spread, err := s.spreadRepo.GetByID(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("spreadsheet not found: %w", err)
	}

	// Получаем только видимые поля
	visibleFields, err := s.enforcer.VisibleFields(ctx, userID, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("get visible fields: %w", err)
	}

	// Фильтруем входящие фильтры — оставляем только по видимым полям
	visibleFieldIDs := make(map[string]bool, len(visibleFields))
	for _, f := range visibleFields {
		visibleFieldIDs[f.ID] = true
	}

	allowedFilters := make([]FilterCondition, 0, len(input.Filters))
	for _, filter := range input.Filters {
		if visibleFieldIDs[filter.FieldID] {
			allowedFilters = append(allowedFilters, filter)
		}
	}

	allowedSorts := make([]SortCondition, 0, len(input.Sorts))
	for _, sort := range input.Sorts {
		if visibleFieldIDs[sort.FieldID] {
			allowedSorts = append(allowedSorts, sort)
		}
	}

	// Добавляем row-level фильтры из enforcer
	rowFilters, err := s.enforcer.GetRowFilters(ctx, userID, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("get row filters: %w", err)
	}

	if input.Limit <= 0 {
		input.Limit = 50
	}

	params := QueryParams{
		Limit:      input.Limit,
		Offset:     input.Offset,
		Filters:    allowedFilters,
		Sorts:      allowedSorts,
		RowFilters: rowFilters,
	}

	rows, total, err := s.repo.List(ctx, spread.DBSchema, spread.TableName, visibleFields, params)
	if err != nil {
		return nil, fmt.Errorf("list rows: %w", err)
	}

	// Скрываем скрытые поля из результата
	sanitized := s.sanitizeRows(rows, visibleFields)

	return &QueryResult{
		Data:   sanitized,
		Total:  total,
		Limit:  input.Limit,
		Offset: input.Offset,
	}, nil
}

// ─── Create ───────────────────────────────────────────────────────────────────

func (s *Service) Create(ctx context.Context, userID, spreadsheetID string, data RowData) (RowData, error) {
	if err := s.enforcer.RequireInsert(ctx, userID, spreadsheetID); err != nil {
		return nil, err
	}

	spread, err := s.spreadRepo.GetByID(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("spreadsheet not found: %w", err)
	}

	// Получаем редактируемые поля
	editableFields, err := s.enforcer.EditableFields(ctx, userID, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("get editable fields: %w", err)
	}

	// Валидируем и фильтруем входящие данные
	cleanData, err := s.validateAndFilterData(ctx, spreadsheetID, data, editableFields)
	if err != nil {
		return nil, fmt.Errorf("validate data: %w", err)
	}

	row, err := s.repo.Create(ctx, spread.DBSchema, spread.TableName, cleanData, userID)
	if err != nil {
		return nil, fmt.Errorf("create row: %w", err)
	}

	return row, nil
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (s *Service) Update(ctx context.Context, userID, spreadsheetID, rowID string, data RowData) (RowData, error) {
	if err := s.enforcer.RequireEdit(ctx, userID, spreadsheetID); err != nil {
		return nil, err
	}

	spread, err := s.spreadRepo.GetByID(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("spreadsheet not found: %w", err)
	}

	// Получаем редактируемые поля
	editableFields, err := s.enforcer.EditableFields(ctx, userID, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("get editable fields: %w", err)
	}

	// Удаляем системные поля
	s.stripSystemFields(data)

	// Валидируем и фильтруем
	cleanData, err := s.validateAndFilterData(ctx, spreadsheetID, data, editableFields)
	if err != nil {
		return nil, fmt.Errorf("validate data: %w", err)
	}

	if len(cleanData) == 0 {
		return nil, fmt.Errorf("no editable fields in request")
	}

	row, err := s.repo.Update(ctx, spread.DBSchema, spread.TableName, rowID, cleanData)
	if err != nil {
		return nil, fmt.Errorf("update row: %w", err)
	}

	return row, nil
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (s *Service) Delete(ctx context.Context, userID, spreadsheetID, rowID string) error {
	if err := s.enforcer.RequireDelete(ctx, userID, spreadsheetID); err != nil {
		return err
	}

	spread, err := s.spreadRepo.GetByID(ctx, spreadsheetID)
	if err != nil {
		return fmt.Errorf("spreadsheet not found: %w", err)
	}

	if err := s.repo.Delete(ctx, spread.DBSchema, spread.TableName, rowID); err != nil {
		return fmt.Errorf("delete row: %w", err)
	}

	return nil
}

// ─── BulkDelete ───────────────────────────────────────────────────────────────

type BulkDeleteInput struct {
	RowIDs []string `json:"row_ids"`
}

func (s *Service) BulkDelete(ctx context.Context, userID, spreadsheetID string, input BulkDeleteInput) error {
	if err := s.enforcer.RequireDelete(ctx, userID, spreadsheetID); err != nil {
		return err
	}

	if len(input.RowIDs) == 0 {
		return fmt.Errorf("row_ids is required")
	}

	if len(input.RowIDs) > 500 {
		return fmt.Errorf("cannot delete more than 500 rows at once")
	}

	spread, err := s.spreadRepo.GetByID(ctx, spreadsheetID)
	if err != nil {
		return fmt.Errorf("spreadsheet not found: %w", err)
	}

	if err := s.repo.BulkDelete(ctx, spread.DBSchema, spread.TableName, input.RowIDs); err != nil {
		return fmt.Errorf("bulk delete rows: %w", err)
	}

	return nil
}

// ─── Вспомогательные методы ───────────────────────────────────────────────────

// validateAndFilterData проверяет данные на соответствие полям и правам доступа
func (s *Service) validateAndFilterData(
	ctx context.Context,
	spreadsheetID string,
	data RowData,
	editableFields []*types.Field,
) (RowData, error) {
	// Строим карту column_name → field для быстрого поиска
	editableByColumn := make(map[string]*types.Field, len(editableFields))
	for _, f := range editableFields {
		editableByColumn[f.ColumnName] = f
	}

	// Получаем все поля для валидации обязательных
	allFields, err := s.fieldRepo.ListBySpreadsheet(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("list fields: %w", err)
	}

	clean := make(RowData, len(data))

	// Фильтруем: оставляем только редактируемые поля
	for colName, value := range data {
		f, ok := editableByColumn[colName]
		if !ok {
			// Колонка не найдена среди редактируемых — пропускаем
			continue
		}

		// Валидируем значение
		if err := validateFieldValue(f, value); err != nil {
			return nil, fmt.Errorf("field %q: %w", f.Name, err)
		}

		clean[colName] = value
	}

	// Проверяем обязательные поля при создании (если данные не пустые)
	for _, f := range allFields {
		if !f.IsRequired {
			continue
		}
		val, provided := data[f.ColumnName]
		if !provided {
			continue
		}
		if isEmptyValue(val) {
			return nil, fmt.Errorf("field %q is required", f.Name)
		}
	}

	return clean, nil
}

// validateFieldValue проверяет значение поля на соответствие его типу
func validateFieldValue(f *types.Field, value interface{}) error {
	if value == nil {
		if f.IsRequired {
			return fmt.Errorf("required field cannot be null")
		}
		return nil
	}

	switch f.FieldType {
	case "integer":
		switch v := value.(type) {
		case float64:
			if v != float64(int64(v)) {
				return fmt.Errorf("expected integer, got decimal")
			}
		case int, int32, int64:
			// ok
		default:
			return fmt.Errorf("expected integer, got %T", value)
		}

	case "decimal":
		switch value.(type) {
		case float64, float32, int, int32, int64:
			// ok
		default:
			return fmt.Errorf("expected number, got %T", value)
		}

	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}

	case "select":
		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string for select, got %T", value)
		}
		if f.Options != nil {
			if err := validateSelectChoice(f, strVal); err != nil {
				return err
			}
		}

	case "multi_select":
		arr, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("expected array for multi_select, got %T", value)
		}
		if f.Options != nil {
			for _, item := range arr {
				strItem, ok := item.(string)
				if !ok {
					return fmt.Errorf("multi_select values must be strings")
				}
				if err := validateSelectChoice(f, strItem); err != nil {
					return err
				}
			}
		}

	case "email":
		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string for email")
		}
		if strVal != "" && !strings.Contains(strVal, "@") {
			return fmt.Errorf("invalid email format")
		}

	case "url":
		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string for url")
		}
		if strVal != "" && !strings.HasPrefix(strVal, "http://") && !strings.HasPrefix(strVal, "https://") {
			return fmt.Errorf("url must start with http:// or https://")
		}
	}

	return nil
}

// validateSelectChoice проверяет, что значение входит в список допустимых вариантов
func validateSelectChoice(f *types.Field, value string) error {
	choices, ok := f.Options["choices"].([]interface{})
	if !ok {
		return nil
	}

	for _, c := range choices {
		choice, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		if v, ok := choice["value"].(string); ok && v == value {
			return nil
		}
	}

	return fmt.Errorf("value %q is not in allowed choices", value)
}

// stripSystemFields удаляет системные поля из данных запроса
func (s *Service) stripSystemFields(data RowData) {
	delete(data, "_id")
	delete(data, "_created_by")
	delete(data, "_created_at")
	delete(data, "_updated_at")
	delete(data, "_position")
}

// sanitizeRows оставляет в строках только видимые поля плюс системные
func (s *Service) sanitizeRows(rows []RowData, visibleFields []*types.Field) []RowData {
	if len(visibleFields) == 0 {
		return rows
	}

	// Строим set видимых column_name
	visibleCols := make(map[string]bool, len(visibleFields)+5)
	for _, f := range visibleFields {
		visibleCols[f.ColumnName] = true
	}

	// Системные поля всегда видны
	systemCols := []string{"_id", "_created_by", "_created_at", "_updated_at", "_position"}
	for _, col := range systemCols {
		visibleCols[col] = true
	}

	sanitized := make([]RowData, 0, len(rows))
	for _, row := range rows {
		clean := make(RowData, len(visibleCols))
		for col, val := range row {
			if visibleCols[col] {
				clean[col] = val
			}
		}
		sanitized = append(sanitized, clean)
	}

	return sanitized
}

// isEmptyValue проверяет, является ли значение пустым
func isEmptyValue(value interface{}) bool {
	if value == nil {
		return true
	}
	if str, ok := value.(string); ok {
		return strings.TrimSpace(str) == ""
	}
	return false
}

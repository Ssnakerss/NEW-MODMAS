package row

import (
	"context"
	"fmt"
	"strings"

	permissionPkg "github.com/Ssnakerss/modmas/internal/permission"
	"github.com/Ssnakerss/modmas/internal/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Добавить в структуру QueryParams поле RowFilters:

type QueryParams struct {
	Limit      int
	Offset     int
	Filters    []FilterCondition
	Sorts      []SortCondition
	RowFilters []permissionPkg.RowFilter // фильтры row-level security
}
type FilterCondition struct {
	FieldID string `json:"field_id"`
	Op      string `json:"op"`
	Value   string `json:"value"`
}

type SortCondition struct {
	FieldID   string `json:"field_id"`
	Direction string `json:"direction"`
}

type RowData map[string]interface{}

func (r *Repository) List(
	ctx context.Context,
	schema, table string,
	fields []*types.Field,
	params QueryParams,
) ([]RowData, int, error) {
	if params.Limit <= 0 {
		params.Limit = 50
	}
	if params.Limit > 500 {
		params.Limit = 500
	}

	// Карта field_id → column_name
	fieldMap := make(map[string]string, len(fields))
	for _, f := range fields {
		fieldMap[f.ID] = f.ColumnName
	}

	args := []interface{}{}
	argIdx := 1

	// WHERE
	whereClauses := []string{}
	for _, filter := range params.Filters {
		colName, ok := fieldMap[filter.FieldID]
		if !ok {
			continue
		}
		col := fmt.Sprintf(`"%s"`, colName)

		switch filter.Op {
		case "eq":
			whereClauses = append(whereClauses, fmt.Sprintf(`%s = $%d`, col, argIdx))
			args = append(args, filter.Value)
			argIdx++
		case "neq":
			whereClauses = append(whereClauses, fmt.Sprintf(`%s != $%d`, col, argIdx))
			args = append(args, filter.Value)
			argIdx++
		case "contains":
			whereClauses = append(whereClauses, fmt.Sprintf(`%s ILIKE $%d`, col, argIdx))
			args = append(args, "%"+filter.Value+"%")
			argIdx++
		case "is_empty":
			whereClauses = append(whereClauses, fmt.Sprintf(`(%s IS NULL OR %s = '')`, col, col))
		case "is_not_empty":
			whereClauses = append(whereClauses, fmt.Sprintf(`(%s IS NOT NULL AND %s != '')`, col, col))
		case "gt":
			whereClauses = append(whereClauses, fmt.Sprintf(`%s > $%d`, col, argIdx))
			args = append(args, filter.Value)
			argIdx++
		case "lt":
			whereClauses = append(whereClauses, fmt.Sprintf(`%s < $%d`, col, argIdx))
			args = append(args, filter.Value)
			argIdx++
		}
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// ORDER BY
	orderClauses := []string{}
	for _, sort := range params.Sorts {
		colName, ok := fieldMap[sort.FieldID]
		if !ok {
			continue
		}
		dir := "ASC"
		if strings.ToLower(sort.Direction) == "desc" {
			dir = "DESC"
		}
		orderClauses = append(orderClauses, fmt.Sprintf(`"%s" %s`, colName, dir))
	}
	orderClauses = append(orderClauses, "_position ASC NULLS LAST", "_created_at ASC")
	orderSQL := "ORDER BY " + strings.Join(orderClauses, ", ")

	tableRef := fmt.Sprintf(`"%s"."%s"`, schema, table)

	// COUNT
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, tableRef, whereSQL)
	var total int
	if err := r.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count rows: %w", err)
	}

	// SELECT
	limitArgs := append(args, params.Limit, params.Offset)
	selectSQL := fmt.Sprintf(`
        SELECT * FROM %s
        %s
        %s
        LIMIT $%d OFFSET $%d
    `, tableRef, whereSQL, orderSQL, argIdx, argIdx+1)

	rows, err := r.pool.Query(ctx, selectSQL, limitArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query rows: %w", err)
	}
	defer rows.Close()

	return scanRows(rows)
}

func (r *Repository) Create(ctx context.Context, schema, table string, data RowData, createdBy string) (RowData, error) {
	cols := []string{"_created_by"}
	placeholders := []string{"$1"}
	args := []interface{}{createdBy}
	idx := 2

	for col, val := range data {
		cols = append(cols, fmt.Sprintf(`"%s"`, col))
		placeholders = append(placeholders, fmt.Sprintf("$%d", idx))
		args = append(args, val)
		idx++
	}

	sql := fmt.Sprintf(
		`INSERT INTO "%s"."%s" (%s) VALUES (%s) RETURNING *`,
		schema, table,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("insert row: %w", err)
	}
	defer rows.Close()

	result, _, err := scanRows(rows)
	if err != nil || len(result) == 0 {
		return nil, fmt.Errorf("scan inserted row: %w", err)
	}
	return result[0], nil
}

func (r *Repository) Update(ctx context.Context, schema, table, rowID string, data RowData) (RowData, error) {
	setClauses := []string{"_updated_at = now()"}
	args := []interface{}{}
	idx := 1

	for col, val := range data {
		setClauses = append(setClauses, fmt.Sprintf(`"%s" = $%d`, col, idx))
		args = append(args, val)
		idx++
	}

	args = append(args, rowID)
	sql := fmt.Sprintf(
		`UPDATE "%s"."%s" SET %s WHERE _id = $%d RETURNING *`,
		schema, table,
		strings.Join(setClauses, ", "),
		idx,
	)

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("update row: %w", err)
	}
	defer rows.Close()

	result, _, err := scanRows(rows)
	if err != nil || len(result) == 0 {
		return nil, fmt.Errorf("scan updated row: %w", err)
	}
	return result[0], nil
}

func (r *Repository) Delete(ctx context.Context, schema, table, rowID string) error {
	sql := fmt.Sprintf(`DELETE FROM "%s"."%s" WHERE _id = $1`, schema, table)
	_, err := r.pool.Exec(ctx, sql, rowID)
	return err
}

func scanRows(rows pgx.Rows) ([]RowData, int, error) {
	descriptions := rows.FieldDescriptions()
	var result []RowData

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, 0, err
		}
		row := make(RowData, len(descriptions))
		for i, desc := range descriptions {
			row[string(desc.Name)] = values[i]
		}
		result = append(result, row)
	}

	return result, len(result), rows.Err()
}

// Добавить метод BulkDelete в Repository:

func (r *Repository) BulkDelete(ctx context.Context, schema, table string, rowIDs []string) error {
	if len(rowIDs) == 0 {
		return nil
	}

	// Строим $1, $2, ... $N
	placeholders := make([]string, len(rowIDs))
	args := make([]interface{}, len(rowIDs))
	for i, id := range rowIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	sql := fmt.Sprintf(
		`DELETE FROM "%s"."%s" WHERE _id IN (%s)`,
		schema, table,
		strings.Join(placeholders, ", "),
	)

	_, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("bulk delete rows: %w", err)
	}

	return nil
}

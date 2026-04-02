package field

import (
	"context"
	"encoding/json"
	"fmt"

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

func (r *Repository) Create(ctx context.Context, tx pgx.Tx, f *types.Field) (*types.Field, error) {
	optionsJSON, _ := json.Marshal(f.Options)

	result := &types.Field{}
	var optRaw []byte

	err := tx.QueryRow(ctx, `
		INSERT INTO meta.fields
			(spreadsheet_id, name, column_name, field_type, position,
			 is_required, is_unique, default_value, options)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, spreadsheet_id, name, column_name, field_type,
		          position, is_required, is_unique, default_value, options
	`, f.SpreadsheetID, f.Name, f.ColumnName, f.FieldType,
		f.Position, f.IsRequired, f.IsUnique, f.DefaultValue, optionsJSON,
	).Scan(
		&result.ID, &result.SpreadsheetID, &result.Name, &result.ColumnName,
		&result.FieldType, &result.Position, &result.IsRequired, &result.IsUnique,
		&result.DefaultValue, &optRaw,
	)
	if err != nil {
		return nil, fmt.Errorf("create field: %w", err)
	}

	if optRaw != nil {
		_ = json.Unmarshal(optRaw, &result.Options)
	}
	return result, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*types.Field, error) {
	f := &types.Field{}
	var optRaw []byte

	err := r.pool.QueryRow(ctx, `
		SELECT id, spreadsheet_id, name, column_name, field_type,
		       position, is_required, is_unique, default_value, options
		FROM meta.fields WHERE id = $1
	`, id).Scan(
		&f.ID, &f.SpreadsheetID, &f.Name, &f.ColumnName, &f.FieldType,
		&f.Position, &f.IsRequired, &f.IsUnique, &f.DefaultValue, &optRaw,
	)
	if err != nil {
		return nil, fmt.Errorf("get field: %w", err)
	}

	if optRaw != nil {
		_ = json.Unmarshal(optRaw, &f.Options)
	}
	return f, nil
}

func (r *Repository) ListBySpreadsheet(ctx context.Context, spreadsheetID string) ([]*types.Field, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, spreadsheet_id, name, column_name, field_type,
		       position, is_required, is_unique, default_value, options
		FROM meta.fields
		WHERE spreadsheet_id = $1
		ORDER BY position
	`, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("list fields: %w", err)
	}
	defer rows.Close()

	var fields []*types.Field
	for rows.Next() {
		f := &types.Field{}
		var optRaw []byte
		if err := rows.Scan(
			&f.ID, &f.SpreadsheetID, &f.Name, &f.ColumnName, &f.FieldType,
			&f.Position, &f.IsRequired, &f.IsUnique, &f.DefaultValue, &optRaw,
		); err != nil {
			return nil, err
		}
		if optRaw != nil {
			_ = json.Unmarshal(optRaw, &f.Options)
		}
		fields = append(fields, f)
	}

	return fields, rows.Err()
}

func (r *Repository) Update(ctx context.Context, tx pgx.Tx, f *types.Field) (*types.Field, error) {
	optionsJSON, _ := json.Marshal(f.Options)
	result := &types.Field{}
	var optRaw []byte

	err := tx.QueryRow(ctx, `
		UPDATE meta.fields SET
			name          = $1,
			field_type    = $2,
			position      = $3,
			is_required   = $4,
			is_unique     = $5,
			default_value = $6,
			options       = $7
		WHERE id = $8
		RETURNING id, spreadsheet_id, name, column_name, field_type,
		          position, is_required, is_unique, default_value, options
	`, f.Name, f.FieldType, f.Position, f.IsRequired, f.IsUnique,
		f.DefaultValue, optionsJSON, f.ID,
	).Scan(
		&result.ID, &result.SpreadsheetID, &result.Name, &result.ColumnName,
		&result.FieldType, &result.Position, &result.IsRequired, &result.IsUnique,
		&result.DefaultValue, &optRaw,
	)
	if err != nil {
		return nil, fmt.Errorf("update field: %w", err)
	}

	if optRaw != nil {
		_ = json.Unmarshal(optRaw, &result.Options)
	}
	return result, nil
}

func (r *Repository) Delete(ctx context.Context, tx pgx.Tx, id string) error {
	_, err := tx.Exec(ctx, `DELETE FROM meta.fields WHERE id = $1`, id)
	return err
}

func (r *Repository) GetMaxPosition(ctx context.Context, spreadsheetID string) (int, error) {
	var pos int
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(position), -1)
		FROM meta.fields
		WHERE spreadsheet_id = $1
	`, spreadsheetID).Scan(&pos)
	return pos, err
}

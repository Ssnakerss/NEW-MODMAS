package spreadsheet

import (
	"context"
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

func (r *Repository) Create(ctx context.Context, tx pgx.Tx, s *types.Spreadsheet) (*types.Spreadsheet, error) {
	result := &types.Spreadsheet{}
	err := tx.QueryRow(ctx, `
		INSERT INTO meta.spreadsheets (workspace_id, name, table_name, description, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, workspace_id, name, table_name,
		          COALESCE(description,''), created_by, created_at, updated_at
	`, s.WorkspaceID, s.Name, s.TableName, s.Description, s.CreatedBy,
	).Scan(
		&result.ID, &result.WorkspaceID, &result.Name, &result.TableName,
		&result.Description, &result.CreatedBy, &result.CreatedAt, &result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create spreadsheet: %w", err)
	}
	return result, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*types.Spreadsheet, error) {
	s := &types.Spreadsheet{}
	err := r.pool.QueryRow(ctx, `
		SELECT s.id, s.workspace_id, s.name, s.table_name,
		       COALESCE(s.description,''), s.created_by,
		       s.created_at, s.updated_at,
		       w.db_schema
		FROM meta.spreadsheets s
		JOIN auth.workspaces w ON w.id = s.workspace_id
		WHERE s.id = $1
	`, id).Scan(
		&s.ID, &s.WorkspaceID, &s.Name, &s.TableName,
		&s.Description, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
		&s.DBSchema,
	)
	if err != nil {
		return nil, fmt.Errorf("get spreadsheet: %w", err)
	}
	return s, nil
}

func (r *Repository) ListByWorkspace(ctx context.Context, workspaceID string) ([]*types.Spreadsheet, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, workspace_id, name, table_name,
		       COALESCE(description,''), created_by, created_at, updated_at
		FROM meta.spreadsheets
		WHERE workspace_id = $1
		ORDER BY created_at DESC
	`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list spreadsheets: %w", err)
	}
	defer rows.Close()

	var result []*types.Spreadsheet
	for rows.Next() {
		s := &types.Spreadsheet{}
		if err := rows.Scan(
			&s.ID, &s.WorkspaceID, &s.Name, &s.TableName,
			&s.Description, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, s)
	}

	return result, rows.Err()
}

func (r *Repository) Update(ctx context.Context, id, name, description string) (*types.Spreadsheet, error) {
	s := &types.Spreadsheet{}
	err := r.pool.QueryRow(ctx, `
		UPDATE meta.spreadsheets
		SET name = $1, description = $2, updated_at = now()
		WHERE id = $3
		RETURNING id, workspace_id, name, table_name,
		          COALESCE(description,''), created_by, created_at, updated_at
	`, name, description, id).Scan(
		&s.ID, &s.WorkspaceID, &s.Name, &s.TableName,
		&s.Description, &s.CreatedBy, &s.CreatedAt, &s.UpdatedAt,
	)
	return s, err
}

func (r *Repository) Delete(ctx context.Context, tx pgx.Tx, id string) error {
	_, err := tx.Exec(ctx, `DELETE FROM meta.spreadsheets WHERE id = $1`, id)
	return err
}

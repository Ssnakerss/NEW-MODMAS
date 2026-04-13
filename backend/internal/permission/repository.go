package permission

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Ssnakerss/modmas/internal/types"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetSpreadsheetAccess(ctx context.Context, spreadsheetID string) ([]*types.SpreadsheetAccess, error) {
	rows, err := r.pool.Query(ctx, `
        SELECT sa.id, sa.spreadsheet_id, sa.principal_id, sa.principal_type,
               COALESCE(u.name, sa.principal_type) as principal_name,
               sa.can_view, sa.can_insert, sa.can_edit, sa.can_delete, sa.can_manage
        FROM meta.spreadsheet_access sa
        LEFT JOIN auth.users u ON u.id::text = sa.principal_id::text
        WHERE sa.spreadsheet_id = $1
    `, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("get spreadsheet access: %w", err)
	}
	defer rows.Close()

	var result []*types.SpreadsheetAccess
	for rows.Next() {
		a := &types.SpreadsheetAccess{}
		if err := rows.Scan(
			&a.ID, &a.SpreadsheetID, &a.PrincipalID, &a.PrincipalType,
			&a.PrincipalName, &a.CanView, &a.CanInsert, &a.CanEdit, &a.CanDelete, &a.CanManage,
		); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, nil
}

func (r *Repository) UpsertSpreadsheetAccess(ctx context.Context, a *types.SpreadsheetAccess) error {
	_, err := r.pool.Exec(ctx, `
        INSERT INTO meta.spreadsheet_access
            (spreadsheet_id, principal_id, principal_type, can_view, can_insert, can_edit, can_delete, can_manage)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
        ON CONFLICT (spreadsheet_id, principal_id, principal_type)
        DO UPDATE SET
            can_view = $4, can_insert = $5, can_edit = $6,
            can_delete = $7, can_manage = $8
    `, a.SpreadsheetID, a.PrincipalID, a.PrincipalType,
		a.CanView, a.CanInsert, a.CanEdit, a.CanDelete, a.CanManage,
	)
	return err
}

func (r *Repository) RemoveSpreadsheetAccess(ctx context.Context, spreadsheetID, principalID string) error {
	_, err := r.pool.Exec(ctx, `
        DELETE FROM meta.spreadsheet_access
        WHERE spreadsheet_id = $1 AND principal_id = $2
    `, spreadsheetID, principalID)
	return err
}

func (r *Repository) CheckAccess(ctx context.Context, spreadsheetID, userID string) (*types.SpreadsheetAccess, error) {
	a := &types.SpreadsheetAccess{}
	err := r.pool.QueryRow(ctx, `
        SELECT can_view, can_insert, can_edit, can_delete, can_manage
        FROM meta.spreadsheet_access
        WHERE spreadsheet_id = $1 AND principal_id = $2
    `, spreadsheetID, userID).Scan(
		&a.CanView, &a.CanInsert, &a.CanEdit, &a.CanDelete, &a.CanManage,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *Repository) GetFieldAccess(ctx context.Context, spreadsheetID string) ([]*types.FieldAccess, error) {
	rows, err := r.pool.Query(ctx, `
        SELECT fa.id, fa.field_id, fa.principal_id, fa.principal_type, fa.can_view, fa.can_edit
        FROM meta.field_access fa
        JOIN meta.fields f ON f.id = fa.field_id
        WHERE f.spreadsheet_id = $1
    `, spreadsheetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*types.FieldAccess
	for rows.Next() {
		fa := &types.FieldAccess{}
		if err := rows.Scan(&fa.ID, &fa.FieldID, &fa.PrincipalID, &fa.PrincipalType, &fa.CanView, &fa.CanEdit); err != nil {
			return nil, err
		}
		result = append(result, fa)
	}
	return result, nil
}

func (r *Repository) UpsertFieldAccess(ctx context.Context, fa *types.FieldAccess) error {
	_, err := r.pool.Exec(ctx, `
        INSERT INTO meta.field_access (field_id, principal_id, principal_type, can_view, can_edit)
        VALUES ($1,$2,$3,$4,$5)
        ON CONFLICT (field_id, principal_id, principal_type)
        DO UPDATE SET can_view = $4, can_edit = $5
    `, fa.FieldID, fa.PrincipalID, fa.PrincipalType, fa.CanView, fa.CanEdit)
	return err
}

// Добавить следующие методы в Repository:

func (r *Repository) GetRowRules(ctx context.Context, spreadsheetID string) ([]*RowAccessRule, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, spreadsheet_id, principal_id, principal_type,
		       condition, can_view, can_edit
		FROM meta.row_access_rules
		WHERE spreadsheet_id = $1
		ORDER BY id
	`, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("get row rules: %w", err)
	}
	defer rows.Close()

	var result []*RowAccessRule
	for rows.Next() {
		rule := &RowAccessRule{}
		var condJSON []byte
		if err := rows.Scan(
			&rule.ID, &rule.SpreadsheetID, &rule.PrincipalID,
			&rule.PrincipalType, &condJSON, &rule.CanView, &rule.CanEdit,
		); err != nil {
			return nil, err
		}
		if condJSON != nil {
			_ = json.Unmarshal(condJSON, &rule.Condition)
		}
		result = append(result, rule)
	}
	return result, rows.Err()
}

func (r *Repository) UpsertRowRule(ctx context.Context, rule *RowAccessRule) (*RowAccessRule, error) {
	condJSON, err := json.Marshal(rule.Condition)
	if err != nil {
		return nil, fmt.Errorf("marshal condition: %w", err)
	}

	result := &RowAccessRule{}
	var condRaw []byte

	err = r.pool.QueryRow(ctx, `
		INSERT INTO meta.row_access_rules
		    (spreadsheet_id, principal_id, principal_type, condition, can_view, can_edit)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (spreadsheet_id, principal_id, principal_type)
		DO UPDATE SET
		    condition = $4,
		    can_view  = $5,
		    can_edit  = $6
		RETURNING id, spreadsheet_id, principal_id, principal_type,
		          condition, can_view, can_edit
	`, rule.SpreadsheetID, rule.PrincipalID, rule.PrincipalType,
		condJSON, rule.CanView, rule.CanEdit,
	).Scan(
		&result.ID, &result.SpreadsheetID, &result.PrincipalID,
		&result.PrincipalType, &condRaw, &result.CanView, &result.CanEdit,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert row rule: %w", err)
	}

	if condRaw != nil {
		_ = json.Unmarshal(condRaw, &result.Condition)
	}

	return result, nil
}

func (r *Repository) DeleteRowRule(ctx context.Context, ruleID string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM meta.row_access_rules WHERE id = $1
	`, ruleID)
	return err
}

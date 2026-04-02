package workspace

import (
	"context"
	"fmt"

	"github.com/Ssnakerss/modmas/internal/types"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Workspace struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	OwnerID   string `json:"owner_id"`
	DBSchema  string `json:"db_schema"`
	CreatedAt string `json:"created_at"`
	Role      string `json:"role,omitempty"`
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, name, ownerID, schema string) (*Workspace, error) {
	ws := &Workspace{}
	err := r.pool.QueryRow(ctx, `
        INSERT INTO auth.workspaces (name, owner_id, db_schema)
        VALUES ($1, $2, $3)
        RETURNING id, name, owner_id, db_schema, created_at
    `, name, ownerID, schema).Scan(
		&ws.ID, &ws.Name, &ws.OwnerID, &ws.DBSchema, &ws.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}
	return ws, nil
}

func (r *Repository) AddMember(ctx context.Context, workspaceID, userID, role string) error {
	_, err := r.pool.Exec(ctx, `
        INSERT INTO auth.workspace_members (workspace_id, user_id, role)
        VALUES ($1, $2, $3)
        ON CONFLICT (workspace_id, user_id) DO UPDATE SET role = $3
    `, workspaceID, userID, role)
	return err
}

func (r *Repository) ListByUser(ctx context.Context, userID string) ([]*Workspace, error) {
	rows, err := r.pool.Query(ctx, `
        SELECT w.id, w.name, w.owner_id, w.db_schema, w.created_at, wm.role
        FROM auth.workspaces w
        JOIN auth.workspace_members wm ON wm.workspace_id = w.id
        WHERE wm.user_id = $1
        ORDER BY w.created_at DESC
    `, userID)
	if err != nil {
		return nil, fmt.Errorf("list workspaces: %w", err)
	}
	defer rows.Close()

	var result []*Workspace
	for rows.Next() {
		ws := &Workspace{}
		if err := rows.Scan(&ws.ID, &ws.Name, &ws.OwnerID, &ws.DBSchema, &ws.CreatedAt, &ws.Role); err != nil {
			return nil, err
		}
		result = append(result, ws)
	}
	return result, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*types.Workspace, error) {
	ws := &types.Workspace{}
	err := r.pool.QueryRow(ctx, `
        SELECT id, name, owner_id, db_schema, created_at
        FROM auth.workspaces WHERE id = $1
    `, id).Scan(&ws.ID, &ws.Name, &ws.OwnerID, &ws.DBSchema, &ws.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get workspace: %w", err)
	}
	return ws, nil
}

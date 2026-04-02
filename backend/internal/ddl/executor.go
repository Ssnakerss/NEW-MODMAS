package ddl

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Executor struct {
	pool *pgxpool.Pool
}

func NewExecutor(pool *pgxpool.Pool) *Executor {
	return &Executor{pool: pool}
}

// ExecInTx выполняет DDL-запрос и метафункцию в одной транзакции.
// Если ddlQuery пустой — выполняется только метафункция.
func (e *Executor) ExecInTx(ctx context.Context, ddlQuery string, metaFn func(pgx.Tx) error) error {
	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if ddlQuery != "" {
		if _, err := tx.Exec(ctx, ddlQuery); err != nil {
			return fmt.Errorf("exec ddl %q: %w", ddlQuery, err)
		}
	}

	if err := metaFn(tx); err != nil {
		return fmt.Errorf("meta fn: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

// ExecRaw выполняет DDL без транзакции (например CREATE SCHEMA).
func (e *Executor) ExecRaw(ctx context.Context, ddlQuery string) error {
	if _, err := e.pool.Exec(ctx, ddlQuery); err != nil {
		return fmt.Errorf("exec raw ddl: %w", err)
	}
	return nil
}

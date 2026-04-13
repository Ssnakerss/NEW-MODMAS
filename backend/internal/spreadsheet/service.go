package spreadsheet

import (
	"context"
	"fmt"
	"strings"

	"github.com/Ssnakerss/modmas/internal/ddl"
	"github.com/Ssnakerss/modmas/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// FieldRepository — интерфейс для работы с полями.
// Позволяет избежать циклической зависимости с field-пакетом.
type FieldRepository interface {
	Create(ctx context.Context, tx pgx.Tx, f *types.Field) (*types.Field, error)
	ListBySpreadsheet(ctx context.Context, spreadsheetID string) ([]*types.Field, error)
}

// WorkspaceGetter — интерфейс для получения workspace.
type WorkspaceGetter interface {
	GetByID(ctx context.Context, id string) (*types.Workspace, error)
}

type Service struct {
	repo      *Repository
	fieldRepo FieldRepository
	wsGetter  WorkspaceGetter
	ddlExec   *ddl.Executor
}

func NewService(
	repo *Repository,
	fieldRepo FieldRepository,
	wsGetter WorkspaceGetter,
	ddlExec *ddl.Executor,
) *Service {
	return &Service{
		repo:      repo,
		fieldRepo: fieldRepo,
		wsGetter:  wsGetter,
		ddlExec:   ddlExec,
	}
}

// ─── Create ───────────────────────────────────────────────────────────────────

func (s *Service) Create(ctx context.Context, input types.CreateSpreadsheetInput, createdBy string) (*types.SpreadsheetWithFields, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if input.WorkspaceID == "" {
		return nil, fmt.Errorf("workspace_id is required")
	}

	ws, err := s.wsGetter.GetByID(ctx, input.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("workspace not found: id:%s error: %w", input.WorkspaceID, err)
	}

	tableName := "tbl_" + strings.ReplaceAll(uuid.New().String(), "-", "")

	// Готовим DDL-колонки и метаданные полей
	colDefs := make([]ddl.ColumnDef, 0, len(input.Fields))
	fieldRecords := make([]*types.Field, 0, len(input.Fields))

	for i, fi := range input.Fields {
		colName := "col_" + strings.ReplaceAll(uuid.New().String(), "-", "")
		colDefs = append(colDefs, ddl.ColumnDef{
			Name:         colName,
			FieldType:    fi.FieldType,
			IsRequired:   fi.IsRequired,
			IsUnique:     fi.IsUnique,
			DefaultValue: fi.DefaultValue,
			Options:      fi.Options,
		})
		fieldRecords = append(fieldRecords, &types.Field{
			Name:         fi.Name,
			ColumnName:   colName,
			FieldType:    fi.FieldType,
			Position:     i,
			IsRequired:   fi.IsRequired,
			IsUnique:     fi.IsUnique,
			DefaultValue: fi.DefaultValue,
			Options:      fi.Options,
		})
	}

	ddlQuery := ddl.BuildCreateTable(ws.DBSchema, tableName, colDefs)

	spread := &types.Spreadsheet{
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		TableName:   tableName,
		Description: input.Description,
		CreatedBy:   createdBy,
	}

	var result *types.SpreadsheetWithFields

	err = s.ddlExec.ExecInTx(ctx, ddlQuery, func(tx pgx.Tx) error {
		created, err := s.repo.Create(ctx, tx, spread)
		if err != nil {
			return fmt.Errorf("create spreadsheet: %w", err)
		}

		createdFields := make([]*types.Field, 0, len(fieldRecords))
		for _, fr := range fieldRecords {
			fr.SpreadsheetID = created.ID
			cf, err := s.fieldRepo.Create(ctx, tx, fr)
			if err != nil {
				return fmt.Errorf("create field '%s': %w", fr.Name, err)
			}
			createdFields = append(createdFields, cf)
		}

		result = &types.SpreadsheetWithFields{
			Spreadsheet: created,
			Fields:      createdFields,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ─── Get ──────────────────────────────────────────────────────────────────────

func (s *Service) GetWithFields(ctx context.Context, id string) (*types.SpreadsheetWithFields, error) {
	spread, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("spreadsheet not found: %w", err)
	}

	fields, err := s.fieldRepo.ListBySpreadsheet(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list fields: %w", err)
	}

	return &types.SpreadsheetWithFields{
		Spreadsheet: spread,
		Fields:      fields,
	}, nil
}

// ─── List ─────────────────────────────────────────────────────────────────────

func (s *Service) ListByWorkspace(ctx context.Context, workspaceID string) ([]*types.SpreadsheetWithFields, error) {
	spreads, err := s.repo.ListByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list spreadsheets: %w", err)
	}

	result := make([]*types.SpreadsheetWithFields, 0, len(spreads))
	for _, sp := range spreads {
		fields, _ := s.fieldRepo.ListBySpreadsheet(ctx, sp.ID)
		result = append(result, &types.SpreadsheetWithFields{
			Spreadsheet: sp,
			Fields:      fields,
		})
	}
	return result, nil
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (s *Service) Update(ctx context.Context, id, name, description string) (*types.Spreadsheet, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	return s.repo.Update(ctx, id, name, description)
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (s *Service) Delete(ctx context.Context, id string) error {
	spread, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("spreadsheet not found: %w", err)
	}

	dropQuery := fmt.Sprintf(
		`DROP TABLE IF EXISTS "%s"."%s"`,
		spread.DBSchema, spread.TableName,
	)

	return s.ddlExec.ExecInTx(ctx, dropQuery, func(tx pgx.Tx) error {
		return s.repo.Delete(ctx, tx, id)
	})
}

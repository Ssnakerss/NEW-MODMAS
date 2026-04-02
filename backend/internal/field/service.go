package field

import (
	"context"
	"fmt"
	"strings"

	"github.com/Ssnakerss/modmas/internal/ddl"
	"github.com/Ssnakerss/modmas/internal/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// SpreadsheetGetter — интерфейс для получения таблицы.
// Позволяет избежать циклической зависимости со spreadsheet-пакетом.
type SpreadsheetGetter interface {
	GetByID(ctx context.Context, id string) (*types.Spreadsheet, error)
}

type Service struct {
	repo         *Repository
	spreadGetter SpreadsheetGetter
	ddlExec      *ddl.Executor
}

func NewService(repo *Repository, spreadGetter SpreadsheetGetter, ddlExec *ddl.Executor) *Service {
	return &Service{
		repo:         repo,
		spreadGetter: spreadGetter,
		ddlExec:      ddlExec,
	}
}

// ─── Create ───────────────────────────────────────────────────────────────────

func (s *Service) Create(ctx context.Context, spreadsheetID string, input types.CreateFieldInput) (*types.Field, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("field name is required")
	}
	if !ValidFieldType(input.FieldType) {
		return nil, fmt.Errorf("invalid field type: %s", input.FieldType)
	}
	if input.Options != nil {
		if err := ValidateOptions(input.FieldType, input.Options); err != nil {
			return nil, err
		}
	}

	spread, err := s.spreadGetter.GetByID(ctx, spreadsheetID)
	if err != nil {
		return nil, fmt.Errorf("spreadsheet not found: %w", err)
	}

	maxPos, _ := s.repo.GetMaxPosition(ctx, spreadsheetID)
	columnName := "col_" + strings.ReplaceAll(uuid.New().String(), "-", "")

	colDef := ddl.ColumnDef{
		Name:         columnName,
		FieldType:    input.FieldType,
		IsRequired:   input.IsRequired,
		IsUnique:     input.IsUnique,
		DefaultValue: input.DefaultValue,
		Options:      input.Options,
	}

	ddlQuery := ddl.BuildAddColumn(spread.DBSchema, spread.TableName, colDef)

	newField := &types.Field{
		SpreadsheetID: spreadsheetID,
		Name:          input.Name,
		ColumnName:    columnName,
		FieldType:     input.FieldType,
		Position:      maxPos + 1,
		IsRequired:    input.IsRequired,
		IsUnique:      input.IsUnique,
		DefaultValue:  input.DefaultValue,
		Options:       input.Options,
	}

	var created *types.Field

	err = s.ddlExec.ExecInTx(ctx, ddlQuery, func(tx pgx.Tx) error {
		var txErr error
		created, txErr = s.repo.Create(ctx, tx, newField)
		return txErr
	})
	if err != nil {
		return nil, fmt.Errorf("create field: %w", err)
	}

	return created, nil
}

// ─── Update ───────────────────────────────────────────────────────────────────

func (s *Service) Update(ctx context.Context, fieldID string, input types.UpdateFieldInput) (*types.Field, error) {
	existing, err := s.repo.GetByID(ctx, fieldID)
	if err != nil {
		return nil, fmt.Errorf("field not found: %w", err)
	}

	if input.Name != nil {
		existing.Name = *input.Name
	}
	if input.IsRequired != nil {
		existing.IsRequired = *input.IsRequired
	}
	if input.IsUnique != nil {
		existing.IsUnique = *input.IsUnique
	}
	if input.DefaultValue != nil {
		existing.DefaultValue = input.DefaultValue
	}
	if input.Options != nil {
		existing.Options = input.Options
	}

	// Если меняется тип — нужен ALTER COLUMN
	if input.FieldType != nil && *input.FieldType != existing.FieldType {
		if !IsConversionAllowed(existing.FieldType, *input.FieldType) {
			return nil, fmt.Errorf(
				"type conversion from '%s' to '%s' is not allowed",
				existing.FieldType, *input.FieldType,
			)
		}

		spread, err := s.spreadGetter.GetByID(ctx, existing.SpreadsheetID)
		if err != nil {
			return nil, fmt.Errorf("spreadsheet not found: %w", err)
		}

		oldType := existing.FieldType
		existing.FieldType = *input.FieldType

		colDef := ddl.ColumnDef{
			Name:      existing.ColumnName,
			FieldType: existing.FieldType,
			Options:   existing.Options,
		}

		ddlQuery := ddl.BuildAlterColumnType(spread.DBSchema, spread.TableName, colDef)

		var updated *types.Field
		err = s.ddlExec.ExecInTx(ctx, ddlQuery, func(tx pgx.Tx) error {
			var txErr error
			updated, txErr = s.repo.Update(ctx, tx, existing)
			return txErr
		})
		if err != nil {
			existing.FieldType = oldType
			return nil, fmt.Errorf("alter column type: %w", err)
		}

		return updated, nil
	}

	// Тип не меняется — только метаданные
	var updated *types.Field

	err = s.ddlExec.ExecInTx(ctx, "", func(tx pgx.Tx) error {
		var txErr error
		updated, txErr = s.repo.Update(ctx, tx, existing)
		return txErr
	})
	if err != nil {
		return nil, fmt.Errorf("update field metadata: %w", err)
	}

	return updated, nil
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func (s *Service) Delete(ctx context.Context, fieldID string) error {
	f, err := s.repo.GetByID(ctx, fieldID)
	if err != nil {
		return fmt.Errorf("field not found: %w", err)
	}

	spread, err := s.spreadGetter.GetByID(ctx, f.SpreadsheetID)
	if err != nil {
		return fmt.Errorf("spreadsheet not found: %w", err)
	}

	ddlQuery := ddl.BuildDropColumn(spread.DBSchema, spread.TableName, f.ColumnName)

	return s.ddlExec.ExecInTx(ctx, ddlQuery, func(tx pgx.Tx) error {
		return s.repo.Delete(ctx, tx, fieldID)
	})
}

// ─── List ─────────────────────────────────────────────────────────────────────

func (s *Service) ListBySpreadsheet(ctx context.Context, spreadsheetID string) ([]*types.Field, error) {
	return s.repo.ListBySpreadsheet(ctx, spreadsheetID)
}

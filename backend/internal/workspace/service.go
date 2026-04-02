package workspace

import (
	"context"
	"fmt"
	"strings"

	"github.com/Ssnakerss/modmas/internal/ddl"
	"github.com/Ssnakerss/modmas/internal/types"
	"github.com/google/uuid"
)

type Service struct {
	repo    *Repository
	ddlExec *ddl.Executor
}

func NewService(repo *Repository, ddlExec *ddl.Executor) *Service {
	return &Service{repo: repo, ddlExec: ddlExec}
}

func (s *Service) Create(ctx context.Context, name, ownerID string) (*Workspace, error) {
	schemaName := "data_" + strings.ReplaceAll(uuid.New().String(), "-", "")

	if err := s.ddlExec.ExecRaw(ctx, ddl.BuildCreateSchema(schemaName)); err != nil {
		return nil, fmt.Errorf("create schema: %w", err)
	}

	ws, err := s.repo.Create(ctx, name, ownerID, schemaName)
	if err != nil {
		return nil, err
	}

	if err := s.repo.AddMember(ctx, ws.ID, ownerID, "owner"); err != nil {
		return nil, fmt.Errorf("add owner member: %w", err)
	}

	return ws, nil
}

func (s *Service) ListByUser(ctx context.Context, userID string) ([]*Workspace, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) GetByID(ctx context.Context, id string) (*types.Workspace, error) {
	return s.repo.GetByID(ctx, id)
}

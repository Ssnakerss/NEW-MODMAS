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

// NewService создает новый экземпляр сервиса для работы с рабочими пространствами
// Принимает репозиторий и исполнитель DDL-команд для работы с базой данных
func NewService(repo *Repository, ddlExec *ddl.Executor) *Service {
	return &Service{repo: repo, ddlExec: ddlExec}
}

// Create реализует бизнес-логику создания нового рабочего пространства
// Создает новую схему в базе данных, запись рабочего пространства и добавляет владельца
// Возвращает созданное рабочее пространство или ошибку операции
func (s *Service) Create(ctx context.Context, name, ownerID string) (*types.Workspace, error) {
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

// ListByUser возвращает список всех рабочих пространств, к которым имеет доступ пользователь
// Используется для отображения доступных рабочих пространств в интерфейсе
func (s *Service) ListByUser(ctx context.Context, userID string) ([]*types.Workspace, error) {
	return s.repo.ListByUser(ctx, userID)
}

// GetByID возвращает информацию о конкретном рабочем пространстве по его идентификатору
// Используется для получения детальной информации о рабочем пространстве
func (s *Service) GetByID(ctx context.Context, id string) (*types.Workspace, error) {
	return s.repo.GetByID(ctx, id)
}

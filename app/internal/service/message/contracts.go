package message

import (
	"context"

	"github.com/google/uuid"

	"github.com/devian2011/msgchute/internal/dto"
)

type taskRepo interface {
	GetByID(context.Context, uuid.UUID) (*dto.Task, error)
	GetByMessageIDs(ctx context.Context, messageIDs []uuid.UUID) (map[uuid.UUID][]dto.Task, error)
	Unlock(ctx context.Context, IDs []uuid.UUID) error
}

type taskResultRepo interface {
	GetByTaskIDs(context.Context, []uuid.UUID) (map[uuid.UUID][]dto.TaskExecutionResult, error)
}

type messageRepo interface {
	Find(ctx context.Context, filter *dto.MessageFilter) ([]*dto.Message, uint64, error)
	GetByID(context.Context, uuid.UUID) (*dto.Message, error)
	GetByIDs(ctx context.Context, IDs []uuid.UUID) ([]dto.Message, error)
	UpdateStatus(context.Context, uuid.UUID, dto.MessageStatus) error
}

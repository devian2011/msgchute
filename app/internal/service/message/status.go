package message

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/devian2011/msgchute/internal/dto"
	"github.com/devian2011/msgchute/internal/io/storage"
)

type StatusUpdater struct {
	db       *sqlx.DB
	msgRepo  messageRepo
	taskRepo taskRepo
}

func NewStatusUpdater(
	db *sqlx.DB,
	msgRepo messageRepo,
	taskRepo taskRepo,
) *StatusUpdater {
	return &StatusUpdater{
		db:       db,
		msgRepo:  msgRepo,
		taskRepo: taskRepo,
	}
}

func (s *StatusUpdater) UpdateStatusByTaskID(taskID uuid.UUID, status dto.MessageStatus) error {
	return storage.InTransaction(context.Background(), s.db, func(ctx context.Context) error {
		task, getTaskErr := s.taskRepo.GetByID(ctx, taskID)
		if getTaskErr != nil {
			return getTaskErr
		}
		unlockErr := s.taskRepo.Unlock(ctx, []uuid.UUID{task.ID})
		if unlockErr != nil {
			return unlockErr
		}

		m, getErr := s.msgRepo.GetByID(ctx, task.MessageID)
		if getErr != nil {
			return getErr
		}
		if m.Status == status || m.Status == dto.MessageStatusSucceeded {
			return nil
		}
		return s.msgRepo.UpdateStatus(ctx, m.ID, status)
	})
}

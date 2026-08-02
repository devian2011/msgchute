package event

import (
	"context"
	"log/slog"

	"github.com/devian2011/retrier"
	"github.com/google/uuid"

	"github.com/devian2011/msgchute/internal/dto"
)

type MessageStatusUpdater interface {
	UpdateStatusByTaskID(taskID uuid.UUID, status dto.MessageStatus) error
}

type Bus struct {
	ctx              context.Context
	msgStatusUpdater MessageStatusUpdater
}

func NewBus(ctx context.Context, updater MessageStatusUpdater) *Bus {
	return &Bus{
		ctx:              ctx,
		msgStatusUpdater: updater,
	}
}

// Publish process worker result
func (b *Bus) Publish(e retrier.WorkerExecutionResult) {
	b.updateMessageState(e)
}

func (b *Bus) updateMessageState(e retrier.WorkerExecutionResult) {
	var messageStatus dto.MessageStatus
	switch e.Task.Status {
	case retrier.StatusSuccess:
		messageStatus = dto.MessageStatusSucceeded
	case retrier.StatusFailure:
		messageStatus = dto.MessageStatusFailed
	}
	if len(string(messageStatus)) > 0 {
		updateStatusErr := b.msgStatusUpdater.UpdateStatusByTaskID(e.Task.ID, messageStatus)
		if updateStatusErr != nil {
			slog.Error("failed to update status by task",
				slog.Any("taskID", e.Task.ID),
				slog.Any("status", e.Task.Status),
				slog.Any("error", updateStatusErr))
		}
	}
}

package sender

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/devian2011/retrier"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/devian2011/msgchute/internal/dto"
	"github.com/devian2011/msgchute/internal/io/storage"
	"github.com/devian2011/msgchute/pkg/generate"
)

type Queue struct {
	ctx         context.Context
	db          *sqlx.DB
	taskRepo    taskRepo
	messageRepo messageRepo
}

func NewQueue(ctx context.Context, db *sqlx.DB, taskRepo taskRepo, messageRepo messageRepo) *Queue {
	return &Queue{
		ctx:         ctx,
		db:          db,
		taskRepo:    taskRepo,
		messageRepo: messageRepo,
	}
}

func (s *Queue) Add(message *dto.Message) (*dto.Message, *dto.Task, error) {
	message.ID = generate.ID()
	now := time.Now()

	if message.Retry == nil {
		message.Retry = &dto.Retry{
			Retries:  1,
			Strategy: retrier.JitterLinearBackOff,
			Params: map[retrier.BackOffParam]interface{}{
				retrier.DurationKey: time.Second,
			},
		}
	}

	task := &dto.Task{
		ID:            generate.ID(),
		MessageID:     message.ID,
		Worker:        message.Transport,
		Status:        retrier.StatusPending,
		Retries:       0,
		MaxRetries:    message.Retry.Retries,
		BackOffCode:   message.Retry.Strategy,
		BackOffParams: message.Retry.Params,
		Deadline:      message.Deadline,
		IsProcessed:   false,

		LockUntil: time.Time{},
		CreatedAt: now,
		LastRun:   time.Time{},
		NextRun:   now,
	}

	if !message.Deadline.IsZero() {
		task.Deadline = message.Deadline
	}

	getErr := storage.InTransaction(context.TODO(), s.db, func(ctx context.Context) error {
		messageCreateErr := s.messageRepo.Create(ctx, message)
		if messageCreateErr != nil {
			slog.Error("Failed to create message", "error", messageCreateErr)
			return messageCreateErr
		}

		_, taskCreateErr := s.taskRepo.Create(ctx, task)
		if taskCreateErr != nil {
			slog.Error("Failed to create task", "error", taskCreateErr)
			return taskCreateErr
		}

		return nil
	})

	if getErr != nil {
		slog.Error("Failed to add message to queue", "error", getErr.Error())
		return nil, nil, getErr
	}

	return message, task, nil
}

// Retry send repeat action for message
func (s *Queue) Retry(mrr *dto.MessageRetryRequest) (*dto.Message, *dto.Task, error) {
	var msg *dto.Message
	var task *dto.Task
	err := storage.InTransaction(context.Background(), s.db, func(ctx context.Context) error {
		now := time.Now()
		var getErr error
		msg, getErr = s.messageRepo.GetByID(ctx, mrr.ID)
		if getErr != nil {
			return getErr
		}

		taskMap, selectErr := s.taskRepo.List(ctx, dto.TaskFilter{
			MessageIDs: []uuid.UUID{msg.ID},
		})
		if selectErr != nil {
			return selectErr
		}
		// Check that all tasks is finished
		isFinished := true
		if len(taskMap[msg.ID]) > 0 {
			for _, t := range taskMap[msg.ID] {
				if !t.IsFinished() {
					isFinished = false
					break
				}
			}
		}

		if !isFinished {
			return errors.New("all task not finished, make retry after all tasks will be closed")
		}

		// If set new retry policy
		msgRetry := msg.Retry
		if mrr.Retry != nil {
			msgRetry = mrr.Retry
		}
		// If msg.Retry is nil, set default
		if msgRetry == nil {
			msgRetry = &dto.Retry{
				Retries:  1,
				Strategy: retrier.JitterLinearBackOff,
				Params: map[retrier.BackOffParam]interface{}{
					retrier.DurationKey: time.Second,
				},
			}
		}

		// If set new schedule
		nextRun := now
		if !mrr.Schedule.IsZero() {
			nextRun = mrr.Schedule
		}

		task = &dto.Task{
			ID:            generate.ID(),
			MessageID:     msg.ID,
			Worker:        msg.Transport,
			Status:        retrier.StatusPending,
			Retries:       0,
			MaxRetries:    msgRetry.Retries,
			BackOffCode:   msgRetry.Strategy,
			BackOffParams: msgRetry.Params,
			Deadline:      mrr.Deadline,
			IsProcessed:   false,

			LockUntil: time.Time{},
			CreatedAt: now,
			LastRun:   time.Time{},
			NextRun:   nextRun,
		}

		_, taskCreateErr := s.taskRepo.Create(ctx, task)
		if taskCreateErr != nil {
			slog.Error("Failed to create task", "error", taskCreateErr)
			return taskCreateErr
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}
	return msg, task, nil
}

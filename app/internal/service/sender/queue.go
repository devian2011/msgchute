package sender

import (
	"context"
	"time"

	"github.com/devian2011/retrier"
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

	// Инициализируем Retry значениями по умолчанию, если он nil
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
		CreatedAt:     now,
		LastRun:       time.Time{},
		NextRun:       now,
	}

	if !message.Deadline.IsZero() {
		task.Deadline = message.Deadline
	}

	getErr := storage.InTransaction(context.TODO(), s.db, func(ctx context.Context) error {
		_, taskCreateErr := s.taskRepo.Create(ctx, task)
		if taskCreateErr != nil {
			return taskCreateErr
		}
		messageCreateErr := s.messageRepo.Create(ctx, message)
		if messageCreateErr != nil {
			return messageCreateErr
		}
		return nil
	})

	if getErr != nil {
		return nil, nil, getErr
	}

	return message, task, nil
}

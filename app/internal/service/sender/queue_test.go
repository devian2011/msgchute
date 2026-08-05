package sender

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/devian2011/retrier"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/devian2011/msgchute/internal/dto"
)

func TestQueue_Add(t *testing.T) {
	tests := []struct {
		name        string
		message     *dto.Message
		mockTaskErr error
		mockMsgErr  error
		expectedErr bool
		checkTaskFn func(t *testing.T, task *dto.Task)
		checkMsgFn  func(t *testing.T, msg *dto.Message)
	}{
		{
			name: "success with retry",
			message: &dto.Message{
				Transport: "email",
				Retry: &dto.Retry{
					Retries:  3,
					Strategy: retrier.ExponentialBackOff,
					Params: map[retrier.BackOffParam]interface{}{
						retrier.BaseDelayKey: 5 * time.Second,
					},
				},
				Deadline: time.Now().Add(time.Hour),
			},
			expectedErr: false,
			checkTaskFn: func(t *testing.T, task *dto.Task) {
				assert.Equal(t, retrier.StatusPending, task.Status)
				assert.Equal(t, 3, task.MaxRetries)
				assert.Equal(t, retrier.ExponentialBackOff, task.BackOffCode)
				expectedParams := dto.BackOffParams{
					retrier.BaseDelayKey: 5 * time.Second,
				}
				assert.Equal(t, expectedParams, task.BackOffParams)
				assert.NotZero(t, task.Deadline)
				assert.Equal(t, "email", task.Worker)
			},
			checkMsgFn: func(t *testing.T, msg *dto.Message) {
				assert.Equal(t, "email", msg.Transport)
				assert.NotZero(t, msg.ID)
			},
		},
		{
			name: "success with nil retry (defaults)",
			message: &dto.Message{
				Transport: "sms",
				Retry:     nil,
				Deadline:  time.Time{},
			},
			expectedErr: false,
			checkTaskFn: func(t *testing.T, task *dto.Task) {
				assert.Equal(t, 1, task.MaxRetries)
				assert.Equal(t, retrier.JitterLinearBackOff, task.BackOffCode)
				expectedParams := dto.BackOffParams{
					retrier.DurationKey: time.Second,
				}
				assert.Equal(t, expectedParams, task.BackOffParams)
				assert.True(t, task.Deadline.IsZero())
			},
			checkMsgFn: func(t *testing.T, msg *dto.Message) {
				assert.Equal(t, "sms", msg.Transport)
			},
		},
		{
			name: "task creation error",
			message: &dto.Message{
				Transport: "push",
				Retry:     &dto.Retry{Retries: 1, Strategy: "fixed"},
			},
			mockTaskErr: errors.New("task db error"),
			expectedErr: true,
		},
		{
			name: "message creation error",
			message: &dto.Message{
				Transport: "email",
				Retry:     &dto.Retry{Retries: 2, Strategy: "linear"},
			},
			mockMsgErr:  errors.New("message db error"),
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, sqlMock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			sqlxDB := sqlx.NewDb(db, "postgres")

			sqlMock.ExpectBegin()

			mockTaskRepo := new(MockTaskRepo)
			mockMsgRepo := new(MockMessageRepo)

			queue := NewQueue(context.Background(), sqlxDB, mockTaskRepo, mockMsgRepo)

			switch {
			case tt.mockMsgErr != nil:
				mockMsgRepo.On("Create", mock.Anything, mock.Anything).Return(tt.mockMsgErr).Once()
				sqlMock.ExpectRollback()
			case tt.mockTaskErr != nil:
				mockMsgRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				mockTaskRepo.On("Create", mock.Anything, mock.Anything).Return(nil, tt.mockTaskErr).Once()
				sqlMock.ExpectRollback()
			default:
				mockMsgRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
				mockTaskRepo.On("Create", mock.Anything, mock.Anything).Return(nil, nil).Once()
				sqlMock.ExpectCommit()
			}

			msg, task, err := queue.Add(tt.message)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, msg)
				assert.Nil(t, task)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, msg)
				assert.NotNil(t, task)

				assert.NotZero(t, msg.ID)
				assert.NotZero(t, task.ID)
				assert.Equal(t, msg.ID, task.MessageID)

				assert.Equal(t, tt.message.Transport, task.Worker)
				assert.Equal(t, retrier.StatusPending, task.Status)
				assert.Zero(t, task.LastRun)
				assert.NotZero(t, task.CreatedAt)
				assert.NotZero(t, task.NextRun)

				if tt.checkTaskFn != nil {
					tt.checkTaskFn(t, task)
				}
				if tt.checkMsgFn != nil {
					tt.checkMsgFn(t, msg)
				}
			}

			mockTaskRepo.AssertExpectations(t)
			mockMsgRepo.AssertExpectations(t)
			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

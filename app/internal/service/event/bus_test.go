package event

import (
	"context"
	"errors"
	"testing"

	"github.com/devian2011/retrier"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/devian2011/msgchute/internal/dto"
)

type mockMessageStatusUpdater struct {
	mock.Mock
}

func (m *mockMessageStatusUpdater) UpdateStatusByTaskID(taskID uuid.UUID, status dto.MessageStatus) error {
	args := m.Called(taskID, status)
	return args.Error(0)
}

func TestNewBus(t *testing.T) {
	ctx := context.Background()
	updater := &mockMessageStatusUpdater{}
	bus := NewBus(ctx, updater)

	assert.NotNil(t, bus)
	assert.Equal(t, ctx, bus.ctx)
	assert.Equal(t, updater, bus.msgStatusUpdater)
}

func TestBus_Publish(t *testing.T) {
	tests := []struct {
		name           string
		taskStatus     retrier.TaskStatus
		expectedStatus dto.MessageStatus
		updateError    error
		expectUpdate   bool
	}{
		{
			name:           "success status",
			taskStatus:     retrier.StatusSuccess,
			expectedStatus: dto.MessageStatusSucceeded,
			updateError:    nil,
			expectUpdate:   true,
		},
		{
			name:           "failure status",
			taskStatus:     retrier.StatusFailure,
			expectedStatus: dto.MessageStatusFailed,
			updateError:    nil,
			expectUpdate:   true,
		},
		{
			name:           "pending status (no update)",
			taskStatus:     retrier.StatusPending,
			expectedStatus: "",
			updateError:    nil,
			expectUpdate:   false,
		},
		{
			name:           "success with update error",
			taskStatus:     retrier.StatusSuccess,
			expectedStatus: dto.MessageStatusSucceeded,
			updateError:    errors.New("db error"),
			expectUpdate:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updater := &mockMessageStatusUpdater{}
			bus := NewBus(context.Background(), updater)

			taskID := uuid.New()
			e := retrier.WorkerExecutionResult{
				Task: &retrier.Task{
					ID:     taskID,
					Status: tt.taskStatus,
				},
			}

			if tt.expectUpdate {
				updater.On("UpdateStatusByTaskID", taskID, tt.expectedStatus).
					Return(tt.updateError).
					Once()
			}

			bus.Publish(e)

			if tt.expectUpdate {
				updater.AssertExpectations(t)
			} else {
				updater.AssertNotCalled(t, "UpdateStatusByTaskID")
			}
		})
	}
}

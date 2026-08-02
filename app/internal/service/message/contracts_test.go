package message

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/devian2011/msgchute/internal/dto"
)

type MockTaskRepo struct {
	mock.Mock
}

func (m *MockTaskRepo) GetByID(ctx context.Context, id uuid.UUID) (*dto.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Task), args.Error(1)
}

func (m *MockTaskRepo) GetByMessageIDs(ctx context.Context, messageIDs []uuid.UUID) (map[uuid.UUID][]dto.Task, error) {
	args := m.Called(ctx, messageIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID][]dto.Task), args.Error(1)
}

func (m *MockTaskRepo) Unlock(ctx context.Context, IDs []uuid.UUID) error {
	args := m.Called(ctx, IDs)
	return args.Error(0)
}

type MockMessageRepo struct {
	mock.Mock
}

func (m *MockMessageRepo) Find(ctx context.Context, filter *dto.MessageFilter) ([]*dto.Message, uint64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(uint64), args.Error(2)
	}
	return args.Get(0).([]*dto.Message), args.Get(1).(uint64), args.Error(2)
}

func (m *MockMessageRepo) GetByID(ctx context.Context, id uuid.UUID) (*dto.Message, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Message), args.Error(1)
}

func (m *MockMessageRepo) GetByIDs(ctx context.Context, IDs []uuid.UUID) ([]dto.Message, error) {
	args := m.Called(ctx, IDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.Message), args.Error(1)
}

func (m *MockMessageRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status dto.MessageStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

type MockTaskResultRepo struct {
	mock.Mock
}

func (m *MockTaskResultRepo) GetByTaskIDs(ctx context.Context, taskIDs []uuid.UUID) (map[uuid.UUID][]dto.TaskExecutionResult, error) {
	args := m.Called(ctx, taskIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID][]dto.TaskExecutionResult), args.Error(1)
}

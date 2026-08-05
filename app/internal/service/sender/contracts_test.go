package sender

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/devian2011/msgchute/internal/dto"
)

// MockTaskRepo is a mock implementation of taskRepo.
type MockTaskRepo struct {
	mock.Mock
}

func (m *MockTaskRepo) GetByID(_ context.Context, id uuid.UUID) (*dto.Task, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Task), args.Error(1)
}

func (m *MockTaskRepo) List(_ context.Context, filter dto.TaskFilter) (map[uuid.UUID][]dto.Task, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID][]dto.Task), args.Error(1)
}

func (m *MockTaskRepo) Create(_ context.Context, task *dto.Task) (*dto.Task, error) {
	args := m.Called(task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Task), args.Error(1)
}

func (m *MockTaskRepo) Update(_ context.Context, task *dto.Task) (*dto.Task, error) {
	args := m.Called(task)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Task), args.Error(1)
}

// Lock implements the taskRepo interface for blocking tasks.
func (m *MockTaskRepo) Lock(_ context.Context, ids []uuid.UUID) error {
	args := m.Called(ids)
	return args.Error(0)
}

// MockTaskResultRepo is a mock implementation of taskResultRepo.
type MockTaskResultRepo struct {
	mock.Mock
}

func (m *MockTaskResultRepo) Create(_ context.Context, result *dto.TaskExecutionResult) error {
	args := m.Called(result)
	return args.Error(0)
}

// MockMessageRepo is a mock implementation of messageRepo.
type MockMessageRepo struct {
	mock.Mock
}

func (m *MockMessageRepo) GetByID(_ context.Context, id uuid.UUID) (*dto.Message, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.Message), args.Error(1)
}

func (m *MockMessageRepo) GetByIDs(_ context.Context, ids []uuid.UUID) ([]dto.Message, error) {
	args := m.Called(ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.Message), args.Error(1)
}

func (m *MockMessageRepo) Create(_ context.Context, msg *dto.Message) error {
	args := m.Called(msg)
	return args.Error(0)
}

package message

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/devian2011/msgchute/internal/dto"
)

func contains(ids []uuid.UUID, id uuid.UUID) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}

func TestFinder_Find(t *testing.T) {
	msgRepo := new(MockMessageRepo)
	taskRepo := new(MockTaskRepo)
	taskResultRepo := new(MockTaskResultRepo)

	db := &sqlx.DB{}
	finder := NewFinder(db, msgRepo, taskRepo, taskResultRepo)

	ctx := context.Background()
	filter := &dto.MessageFilter{Limit: 10}

	t.Run("success", func(t *testing.T) {
		msgID1 := uuid.New()
		msgID2 := uuid.New()
		msgs := []*dto.Message{
			{ID: msgID1, Subject: "subj1"},
			{ID: msgID2, Subject: "subj2"},
		}
		total := uint64(2)

		task1 := dto.Task{ID: uuid.New(), MessageID: msgID1}
		task2 := dto.Task{ID: uuid.New(), MessageID: msgID2}
		tasksMap := map[uuid.UUID][]dto.Task{
			msgID1: {task1},
			msgID2: {task2},
		}

		taskResult1 := dto.TaskExecutionResult{ID: uuid.New(), TaskID: task1.ID}
		taskResult2 := dto.TaskExecutionResult{ID: uuid.New(), TaskID: task2.ID}
		taskResultsMap := map[uuid.UUID][]dto.TaskExecutionResult{
			task1.ID: {taskResult1},
			task2.ID: {taskResult2},
		}

		msgRepo.On("Find", ctx, filter).Return(msgs, total, nil).Once()
		taskRepo.On("GetByMessageIDs", ctx, []uuid.UUID{msgID1, msgID2}).Return(tasksMap, nil).Once()
		taskResultRepo.On("GetByTaskIDs", ctx, mock.MatchedBy(func(ids []uuid.UUID) bool {
			if len(ids) != 2 {
				return false
			}
			return contains(ids, task1.ID) && contains(ids, task2.ID)
		})).Return(taskResultsMap, nil).Once()

		result, pages, err := finder.Find(filter)
		require.NoError(t, err)
		assert.Equal(t, int(total), pages)
		assert.Len(t, result, 2)
		assert.Equal(t, msgID1, result[0].Message.ID)
		assert.Len(t, result[0].Tasks, 1)
		assert.Equal(t, task1.ID, result[0].Tasks[0].Task.ID)
		assert.Len(t, result[0].Tasks[0].Results, 1)

		msgRepo.AssertExpectations(t)
		taskRepo.AssertExpectations(t)
		taskResultRepo.AssertExpectations(t)
	})

	t.Run("no messages", func(t *testing.T) {
		msgRepo.On("Find", ctx, filter).Return([]*dto.Message{}, uint64(0), nil).Once()
		result, pages, err := finder.Find(filter)
		require.NoError(t, err)
		assert.Equal(t, 0, pages)
		assert.Empty(t, result)
		msgRepo.AssertExpectations(t)
		// taskRepo и taskResultRepo не вызываются
	})

	t.Run("messageRepo error", func(t *testing.T) {
		msgRepo.On("Find", ctx, filter).Return(nil, uint64(0), errors.New("msg error")).Once()
		result, pages, err := finder.Find(filter)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, 0, pages)
	})

	t.Run("taskRepo error", func(t *testing.T) {
		msgRepo.On("Find", ctx, filter).Return([]*dto.Message{{ID: uuid.New()}}, uint64(1), nil).Once()
		taskRepo.On("GetByMessageIDs", ctx, mock.Anything).Return(nil, errors.New("task error")).Once()
		result, pages, err := finder.Find(filter)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, 0, pages)
	})

	t.Run("taskResultRepo error", func(t *testing.T) {
		msgID := uuid.New()
		task := dto.Task{ID: uuid.New(), MessageID: msgID}
		tasksMap := map[uuid.UUID][]dto.Task{msgID: {task}}
		msgRepo.On("Find", ctx, filter).Return([]*dto.Message{{ID: msgID}}, uint64(1), nil).Once()
		taskRepo.On("GetByMessageIDs", ctx, []uuid.UUID{msgID}).Return(tasksMap, nil).Once()
		taskResultRepo.On("GetByTaskIDs", ctx, []uuid.UUID{task.ID}).Return(nil, errors.New("result error")).Once()
		result, pages, err := finder.Find(filter)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, 0, pages)
	})
}

func TestFinder_FindByID(t *testing.T) {
	msgRepo := new(MockMessageRepo)
	taskRepo := new(MockTaskRepo)
	taskResultRepo := new(MockTaskResultRepo)

	db := &sqlx.DB{}
	finder := NewFinder(db, msgRepo, taskRepo, taskResultRepo)

	ctx := context.Background()
	msgID := uuid.New()

	t.Run("success", func(t *testing.T) {
		msg := &dto.Message{ID: msgID, Subject: "test"}
		task1 := dto.Task{ID: uuid.New(), MessageID: msgID}
		task2 := dto.Task{ID: uuid.New(), MessageID: msgID}
		tasksMap := map[uuid.UUID][]dto.Task{
			msgID: {task1, task2},
		}
		taskResult1 := dto.TaskExecutionResult{ID: uuid.New(), TaskID: task1.ID}
		taskResult2 := dto.TaskExecutionResult{ID: uuid.New(), TaskID: task2.ID}
		taskResultsMap := map[uuid.UUID][]dto.TaskExecutionResult{
			task1.ID: {taskResult1},
			task2.ID: {taskResult2},
		}

		msgRepo.On("GetByID", ctx, msgID).Return(msg, nil).Once()
		taskRepo.On("GetByMessageIDs", ctx, []uuid.UUID{msgID}).Return(tasksMap, nil).Once()
		taskResultRepo.On("GetByTaskIDs", ctx, mock.MatchedBy(func(ids []uuid.UUID) bool {
			if len(ids) != 2 {
				return false
			}
			return contains(ids, task1.ID) && contains(ids, task2.ID)
		})).Return(taskResultsMap, nil).Once()

		fullMsg, err := finder.FindByID(msgID)
		require.NoError(t, err)
		assert.Equal(t, msgID, fullMsg.Message.ID)
		assert.Len(t, fullMsg.Tasks, 2)
		assert.Equal(t, task1.ID, fullMsg.Tasks[0].Task.ID)
		assert.Len(t, fullMsg.Tasks[0].Results, 1)

		msgRepo.AssertExpectations(t)
		taskRepo.AssertExpectations(t)
		taskResultRepo.AssertExpectations(t)
	})

	t.Run("message not found", func(t *testing.T) {
		msgRepo.On("GetByID", ctx, msgID).Return(nil, errors.New("not found")).Once()
		fullMsg, err := finder.FindByID(msgID)
		assert.Error(t, err)
		assert.Nil(t, fullMsg)
	})

	t.Run("taskRepo error", func(t *testing.T) {
		msg := &dto.Message{ID: msgID}
		msgRepo.On("GetByID", ctx, msgID).Return(msg, nil).Once()
		taskRepo.On("GetByMessageIDs", ctx, []uuid.UUID{msgID}).Return(nil, errors.New("task error")).Once()
		fullMsg, err := finder.FindByID(msgID)
		assert.Error(t, err)
		assert.Nil(t, fullMsg)
	})

	t.Run("taskResultRepo error", func(t *testing.T) {
		msg := &dto.Message{ID: msgID}
		task := dto.Task{ID: uuid.New(), MessageID: msgID}
		tasksMap := map[uuid.UUID][]dto.Task{msgID: {task}}
		msgRepo.On("GetByID", ctx, msgID).Return(msg, nil).Once()
		taskRepo.On("GetByMessageIDs", ctx, []uuid.UUID{msgID}).Return(tasksMap, nil).Once()
		taskResultRepo.On("GetByTaskIDs", ctx, []uuid.UUID{task.ID}).Return(nil, errors.New("result error")).Once()
		fullMsg, err := finder.FindByID(msgID)
		assert.Error(t, err)
		assert.Nil(t, fullMsg)
	})
}

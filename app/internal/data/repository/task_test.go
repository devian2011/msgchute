package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/devian2011/retrier"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/devian2011/msgchute/internal/dto"
	"github.com/devian2011/msgchute/internal/io/storage"
)

// newTestTask returns a fully populated Task for testing.
func newTestTask() *dto.Task {
	now := time.Now().Truncate(time.Second)
	return &dto.Task{
		ID:            uuid.New(),
		MessageID:     uuid.New(),
		Worker:        "email_worker",
		Status:        retrier.StatusPending,
		Retries:       1,
		MaxRetries:    3,
		BackOffCode:   "exponential",
		BackOffParams: dto.BackOffParams{retrier.BaseDelayKey: "5s"},
		Deadline:      now.Add(time.Hour),
		IsProcessed:   false,
		CreatedAt:     now,
		LastRun:       now,
		NextRun:       now.Add(5 * time.Second),
	}
}

// addTaskRow adds a task row to sqlmock.Rows.
func addTaskRow(rows *sqlmock.Rows, task *dto.Task) *sqlmock.Rows {
	paramsJSON, err := task.BackOffParams.Value()
	if err != nil {
		panic(err)
	}
	return rows.AddRow(
		task.ID, task.MessageID, task.Worker, string(task.Status), task.Retries, task.MaxRetries,
		task.BackOffCode, paramsJSON, task.Deadline, task.IsProcessed, task.CreatedAt, task.LastRun, task.NextRun,
	)
}

// TestTaskRepository_GetByID tests the GetByID method.
func TestTaskRepository_GetByID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewTaskRepository(db)
	ctx := context.Background()
	task := newTestTask()

	expectedSQL := "SELECT id, message_id, worker, status, retries, max_retries, backoff_code, backoff_params, deadline, is_processed, created_at, last_run, next_run FROM tasks WHERE id = $1 FOR UPDATE"

	t.Run("found", func(t *testing.T) {
		rows := addTaskRow(sqlmock.NewRows(taskColumns), task)
		mock.ExpectQuery(expectedSQL).
			WithArgs(task.ID).
			WillReturnRows(rows)

		result, err := repo.GetByID(ctx, task.ID)
		assert.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, task.ID, result.ID)
		assert.Equal(t, task.IsProcessed, result.IsProcessed)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(expectedSQL).
			WithArgs(task.ID).
			WillReturnError(sql.ErrNoRows)

		result, err := repo.GetByID(ctx, task.ID)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, errors.Is(err, ErrTaskNotFound))
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTaskRepository_GetByMessageID tests the GetByMessageID method.
func TestTaskRepository_GetByMessageID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewTaskRepository(db)
	ctx := context.Background()
	task := newTestTask()

	expectedSQL := "SELECT id, message_id, worker, status, retries, max_retries, backoff_code, backoff_params, deadline, is_processed, created_at, last_run, next_run FROM tasks WHERE message_id = $1 ORDER BY created_at ASC FOR UPDATE"

	t.Run("found multiple", func(t *testing.T) {
		second := newTestTask()
		second.ID = uuid.New()
		second.MessageID = task.MessageID

		rows := addTaskRow(sqlmock.NewRows(taskColumns), task)
		rows = addTaskRow(rows, second)

		mock.ExpectQuery(expectedSQL).
			WithArgs(task.MessageID).
			WillReturnRows(rows)

		result, err := repo.GetByMessageID(ctx, task.MessageID)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, task.ID, result[0].ID)
		assert.Equal(t, second.ID, result[1].ID)
	})

	t.Run("empty result", func(t *testing.T) {
		mock.ExpectQuery(expectedSQL).
			WithArgs(task.MessageID).
			WillReturnRows(sqlmock.NewRows(taskColumns))

		result, err := repo.GetByMessageID(ctx, task.MessageID)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTaskRepository_Create tests the Create method.
func TestTaskRepository_Create(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewTaskRepository(db)
	ctx := context.Background()
	task := newTestTask()

	// Поле is_processed добавлено в список столбцов и значений
	expectedSQL := "INSERT INTO tasks (id,message_id,worker,status,retries,max_retries,backoff_code,backoff_params,deadline,is_processed,last_run,next_run) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)"

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(expectedSQL).
			WithArgs(
				task.ID, task.MessageID, task.Worker, task.Status, task.Retries,
				task.MaxRetries, task.BackOffCode, task.BackOffParams,
				task.Deadline, task.IsProcessed, task.LastRun, task.NextRun,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		created, err := repo.Create(ctx, task)
		assert.NoError(t, err)
		require.NotNil(t, created)
		assert.Equal(t, task.ID, created.ID)
	})

	t.Run("zero times are converted to NULL", func(t *testing.T) {
		empty := &dto.Task{
			ID:            uuid.New(),
			MessageID:     uuid.New(),
			Worker:        "email_worker",
			Status:        retrier.StatusPending,
			MaxRetries:    3,
			BackOffCode:   "exponential",
			BackOffParams: dto.BackOffParams{},
			IsProcessed:   false,
		}

		mock.ExpectExec(expectedSQL).
			WithArgs(
				empty.ID, empty.MessageID, empty.Worker, empty.Status, empty.Retries,
				empty.MaxRetries, empty.BackOffCode, empty.BackOffParams,
				nil, empty.IsProcessed, nil, nil,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		created, err := repo.Create(ctx, empty)
		assert.NoError(t, err)
		assert.NotNil(t, created)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTaskRepository_Update tests the Update method.
func TestTaskRepository_Update(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewTaskRepository(db)
	ctx := context.Background()
	task := newTestTask()
	task.Status = retrier.StatusFailure
	task.Retries = 2

	// is_processed добавлено в SET
	expectedSQL := "UPDATE tasks SET message_id = $1, worker = $2, status = $3, retries = $4, max_retries = $5, backoff_code = $6, backoff_params = $7, deadline = $8, is_processed = $9, last_run = $10, next_run = $11 WHERE id = $12"

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(expectedSQL).
			WithArgs(
				task.MessageID, task.Worker, task.Status, task.Retries, task.MaxRetries,
				task.BackOffCode, task.BackOffParams, task.Deadline,
				task.IsProcessed, task.LastRun, task.NextRun,
				task.ID,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		updated, err := repo.Update(ctx, task)
		assert.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, task.ID, updated.ID)
		assert.Equal(t, retrier.StatusFailure, updated.Status)
		assert.Equal(t, 2, updated.Retries)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(expectedSQL).
			WithArgs(
				task.MessageID, task.Worker, task.Status, task.Retries, task.MaxRetries,
				task.BackOffCode, task.BackOffParams, task.Deadline,
				task.IsProcessed, task.LastRun, task.NextRun,
				task.ID,
			).
			WillReturnResult(sqlmock.NewResult(0, 0))

		updated, err := repo.Update(ctx, task)
		assert.Error(t, err)
		assert.Nil(t, updated)
		assert.True(t, errors.Is(err, ErrTaskNotFound))
	})

	t.Run("zero times are converted to NULL", func(t *testing.T) {
		zeroTask := newTestTask()
		zeroTask.Deadline = time.Time{}
		zeroTask.LastRun = time.Time{}
		zeroTask.NextRun = time.Time{}

		mock.ExpectExec(expectedSQL).
			WithArgs(
				zeroTask.MessageID, zeroTask.Worker, zeroTask.Status, zeroTask.Retries, zeroTask.MaxRetries,
				zeroTask.BackOffCode, zeroTask.BackOffParams, nil,
				zeroTask.IsProcessed, nil, nil,
				zeroTask.ID,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		updated, err := repo.Update(ctx, zeroTask)
		assert.NoError(t, err)
		assert.NotNil(t, updated)
		assert.True(t, updated.Deadline.IsZero())
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTaskRepository_Delete tests the Delete method.
func TestTaskRepository_Delete(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewTaskRepository(db)
	ctx := context.Background()
	taskID := uuid.New()

	expectedSQL := "DELETE FROM tasks WHERE id = $1"

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(expectedSQL).
			WithArgs(taskID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(ctx, taskID)
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectExec(expectedSQL).
			WithArgs(taskID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(ctx, taskID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrTaskNotFound))
	})

	t.Run("error on rows affected", func(t *testing.T) {
		mock.ExpectExec(expectedSQL).
			WithArgs(taskID).
			WillReturnResult(sqlmock.NewErrorResult(errors.New("db error")))

		err := repo.Delete(ctx, taskID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rows affected")
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTaskRepository_List tests the List method with filters, sorting, pagination.
func TestTaskRepository_List(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewTaskRepository(db)
	ctx := context.Background()
	task := newTestTask()

	t.Run("with all filters and pagination", func(t *testing.T) {
		now := time.Now().Truncate(time.Second)
		before := now.Add(time.Minute)
		after := now.Add(-time.Minute)
		msgID := uuid.New()
		worker := "test_worker"

		filter := dto.TaskFilter{
			MessageIDs:    []uuid.UUID{msgID},
			Statuses:      []retrier.TaskStatus{retrier.StatusPending, retrier.StatusFailure},
			Worker:        &worker,
			NextRunBefore: &before,
			NextRunAfter:  &after,
			Limit:         10,
			Offset:        5,
			SortBy:        "created_at",
			SortOrder:     "DESC",
		}

		expectedSQL := "SELECT id, message_id, worker, status, retries, max_retries, backoff_code, backoff_params, deadline, is_processed, created_at, last_run, next_run FROM tasks WHERE message_id IN ($1) AND status IN ($2,$3) AND worker = $4 AND next_run <= $5 AND next_run >= $6 ORDER BY created_at DESC LIMIT 10 OFFSET 5 FOR UPDATE"

		rows := addTaskRow(sqlmock.NewRows(taskColumns), task)
		mock.ExpectQuery(expectedSQL).
			WithArgs(msgID, retrier.StatusPending, retrier.StatusFailure, worker, before, after).
			WillReturnRows(rows)

		result, err := repo.List(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		tasksForMsg := result[task.MessageID]
		assert.NotEmpty(t, tasksForMsg)
		assert.Equal(t, task.ID, tasksForMsg[0].ID)
		assert.Equal(t, task.IsProcessed, tasksForMsg[0].IsProcessed)
	})

	t.Run("default ordering and no filters", func(t *testing.T) {
		filter := dto.TaskFilter{}
		expectedSQL := "SELECT id, message_id, worker, status, retries, max_retries, backoff_code, backoff_params, deadline, is_processed, created_at, last_run, next_run FROM tasks ORDER BY next_run ASC NULLS LAST, created_at ASC FOR UPDATE"

		mock.ExpectQuery(expectedSQL).
			WillReturnRows(sqlmock.NewRows(taskColumns))

		result, err := repo.List(ctx, filter)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("invalid sort field", func(t *testing.T) {
		filter := dto.TaskFilter{
			SortBy: "invalid_column",
		}
		_, err := repo.List(ctx, filter)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid sort field")
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTaskRepository_Lock tests the Lock method.
func TestTaskRepository_Lock(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewTaskRepository(db)
	ctx := context.Background()
	ids := []uuid.UUID{uuid.New(), uuid.New()}

	expectedSQL := "UPDATE tasks SET is_processed = $1 WHERE id IN ($2,$3)"

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(expectedSQL).
			WithArgs(true, ids[0], ids[1]).
			WillReturnResult(sqlmock.NewResult(0, 2))

		err := repo.Lock(ctx, ids)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mock.ExpectExec(expectedSQL).
			WithArgs(true, ids[0], ids[1]).
			WillReturnError(errors.New("lock error"))

		err := repo.Lock(ctx, ids)
		assert.Error(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTaskRepository_Unlock tests the Unlock method.
func TestTaskRepository_Unlock(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewTaskRepository(db)
	ctx := context.Background()
	ids := []uuid.UUID{uuid.New(), uuid.New()}

	expectedSQL := "UPDATE tasks SET is_processed = $1 WHERE id IN ($2,$3)"

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(expectedSQL).
			WithArgs(false, ids[0], ids[1]).
			WillReturnResult(sqlmock.NewResult(0, 2))

		err := repo.Unlock(ctx, ids)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mock.ExpectExec(expectedSQL).
			WithArgs(false, ids[0], ids[1]).
			WillReturnError(errors.New("unlock error"))

		err := repo.Unlock(ctx, ids)
		assert.Error(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTaskRepository_GetByMessageIDs tests the GetByMessageIDs method.
func TestTaskRepository_GetByMessageIDs(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewTaskRepository(db)
	ctx := context.Background()

	msgID1 := uuid.New()
	msgID2 := uuid.New()
	task1 := newTestTask()
	task1.MessageID = msgID1
	task2 := newTestTask()
	task2.MessageID = msgID2

	expectedSQL := "SELECT id, message_id, worker, status, retries, max_retries, backoff_code, backoff_params, deadline, is_processed, created_at, last_run, next_run FROM tasks WHERE message_id IN ($1,$2) ORDER BY next_run ASC NULLS LAST, created_at ASC FOR UPDATE"

	t.Run("success", func(t *testing.T) {
		rows := addTaskRow(sqlmock.NewRows(taskColumns), task1)
		rows = addTaskRow(rows, task2)
		mock.ExpectQuery(expectedSQL).
			WithArgs(msgID1, msgID2).
			WillReturnRows(rows)

		result, err := repo.GetByMessageIDs(ctx, []uuid.UUID{msgID1, msgID2})
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Len(t, result[msgID1], 1)
		assert.Equal(t, task1.ID, result[msgID1][0].ID)
		assert.Len(t, result[msgID2], 1)
		assert.Equal(t, task2.ID, result[msgID2][0].ID)
	})

	t.Run("empty input", func(t *testing.T) {
		result, err := repo.GetByMessageIDs(ctx, []uuid.UUID{})
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("error", func(t *testing.T) {
		mock.ExpectQuery(expectedSQL).
			WithArgs(msgID1, msgID2).
			WillReturnError(errors.New("db error"))

		result, err := repo.GetByMessageIDs(ctx, []uuid.UUID{msgID1, msgID2})
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTaskRepository_WithTx tests repository operations within a transaction.
func TestTaskRepository_WithTx(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewTaskRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err)

	ctxWithTx := storage.WithTx(ctx, tx)

	task := newTestTask()

	insertSQL := "INSERT INTO tasks (id,message_id,worker,status,retries,max_retries,backoff_code,backoff_params,deadline,is_processed,last_run,next_run) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)"
	mock.ExpectExec(insertSQL).
		WithArgs(
			task.ID, task.MessageID, task.Worker, task.Status, task.Retries,
			task.MaxRetries, task.BackOffCode, task.BackOffParams,
			task.Deadline, task.IsProcessed, task.LastRun, task.NextRun,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	created, err := repo.Create(ctxWithTx, task)
	assert.NoError(t, err)
	assert.NotNil(t, created)

	selectSQL := "SELECT id, message_id, worker, status, retries, max_retries, backoff_code, backoff_params, deadline, is_processed, created_at, last_run, next_run FROM tasks WHERE id = $1 FOR UPDATE"
	mock.ExpectQuery(selectSQL).
		WithArgs(task.ID).
		WillReturnRows(sqlmock.NewRows(taskColumns).AddRow(
			task.ID, task.MessageID, task.Worker, task.Status, task.Retries, task.MaxRetries,
			task.BackOffCode, []byte(`{"interval":"5s"}`), task.Deadline, task.IsProcessed, task.CreatedAt, task.LastRun, task.NextRun,
		))

	fetched, err := repo.GetByID(ctxWithTx, task.ID)
	assert.NoError(t, err)
	assert.Equal(t, task.ID, fetched.ID)

	mock.ExpectCommit()
	assert.NoError(t, tx.Commit())

	assert.NoError(t, mock.ExpectationsWereMet())
}

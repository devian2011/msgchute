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

// resultColumns lists all columns in the task_execution_results table.
var resultColumns = []string{
	"id", "task_id", "status", "run_at", "result", "is_critical", "execution_time",
}

// newTestTaskExecutionResult returns a fully populated TaskExecutionResult for testing.
func newTestTaskExecutionResult() *dto.TaskExecutionResult {
	now := time.Now().Truncate(time.Second)
	return &dto.TaskExecutionResult{
		ID:            uuid.New(),
		TaskID:        uuid.New(),
		Status:        retrier.StatusSuccess,
		RunAt:         now,
		Result:        []byte(`{"ok":true}`),
		IsCritical:    false,
		ExecutionTime: 100 * time.Millisecond,
	}
}

// addResultRow adds a task execution result row to sqlmock.Rows.
func addResultRow(rows *sqlmock.Rows, result *dto.TaskExecutionResult) *sqlmock.Rows {
	return rows.AddRow(
		result.ID,
		result.TaskID,
		result.Status,
		result.RunAt,
		result.Result,
		result.IsCritical,
		result.ExecutionTime,
	)
}

// TestTaskResultRepository_GetByID tests the GetByID method.
func TestTaskResultRepository_GetByID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewTaskResultRepository(db)
	ctx := context.Background()
	result := newTestTaskExecutionResult()

	expectedSQL := "SELECT id, task_id, status, run_at, result, is_critical, execution_time FROM task_execution_results WHERE id = $1"

	t.Run("found", func(t *testing.T) {
		rows := addResultRow(sqlmock.NewRows(resultColumns), result)
		mock.ExpectQuery(expectedSQL).
			WithArgs(result.ID).
			WillReturnRows(rows)

		fetched, err := repo.GetByID(ctx, result.ID)
		assert.NoError(t, err)
		require.NotNil(t, fetched)
		assert.Equal(t, result.ID, fetched.ID)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(expectedSQL).
			WithArgs(result.ID).
			WillReturnError(sql.ErrNoRows)

		fetched, err := repo.GetByID(ctx, result.ID)
		assert.NoError(t, err) // returns nil, nil
		assert.Nil(t, fetched)
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery(expectedSQL).
			WithArgs(result.ID).
			WillReturnError(errors.New("connection lost"))

		fetched, err := repo.GetByID(ctx, result.ID)
		assert.Error(t, err)
		assert.Nil(t, fetched)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTaskResultRepository_List tests the List method with filters, sorting, pagination.
func TestTaskResultRepository_List(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewTaskResultRepository(db)
	ctx := context.Background()
	result := newTestTaskExecutionResult()

	t.Run("with task_id filter and sorting", func(t *testing.T) {
		taskID := uuid.New()
		filter := dto.TaskExecutionResultFilter{
			TaskID:    &taskID,
			Statuses:  []retrier.TaskStatus{retrier.StatusSuccess, retrier.StatusFailure},
			Limit:     10,
			Offset:    5,
			SortBy:    "run_at",
			SortOrder: "DESC",
		}

		countSQL := "SELECT COUNT(*) FROM task_execution_results WHERE (task_id = $1 AND status IN ($2,$3))"
		mock.ExpectQuery(countSQL).
			WithArgs(taskID, retrier.StatusSuccess, retrier.StatusFailure).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		listSQL := "SELECT id, task_id, status, run_at, result, is_critical, execution_time FROM task_execution_results WHERE (task_id = $1 AND status IN ($2,$3)) ORDER BY run_at DESC LIMIT 10 OFFSET 5"
		rows := addResultRow(sqlmock.NewRows(resultColumns), result)
		second := newTestTaskExecutionResult()
		second.TaskID = taskID
		second.Status = retrier.StatusFailure
		rows = addResultRow(rows, second)

		mock.ExpectQuery(listSQL).
			WithArgs(taskID, retrier.StatusSuccess, retrier.StatusFailure).
			WillReturnRows(rows)

		results, total, err := repo.List(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, results, 2)
	})

	t.Run("no filters", func(t *testing.T) {
		filter := dto.TaskExecutionResultFilter{
			Limit:  20,
			Offset: 0,
		}

		mock.ExpectQuery("SELECT COUNT(*) FROM task_execution_results").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		listSQL := "SELECT id, task_id, status, run_at, result, is_critical, execution_time FROM task_execution_results LIMIT 20 OFFSET 0"
		rows := addResultRow(sqlmock.NewRows(resultColumns), result)
		mock.ExpectQuery(listSQL).
			WillReturnRows(rows)

		results, total, err := repo.List(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, results, 1)
	})

	t.Run("error on list query", func(t *testing.T) {
		filter := dto.TaskExecutionResultFilter{Limit: 10}
		// Count succeeds, no WHERE
		mock.ExpectQuery("SELECT COUNT(*) FROM task_execution_results").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		listSQL := "SELECT id, task_id, status, run_at, result, is_critical, execution_time FROM task_execution_results LIMIT 10 OFFSET 0"
		mock.ExpectQuery(listSQL).
			WillReturnError(errors.New("select error"))

		_, _, err := repo.List(ctx, filter)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "execute list")
	})

	t.Run("invalid sort field", func(t *testing.T) {
		filter := dto.TaskExecutionResultFilter{
			SortBy: "invalid_column",
		}
		_, _, err := repo.List(ctx, filter)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid sort column")
	})

	t.Run("error on count query", func(t *testing.T) {
		filter := dto.TaskExecutionResultFilter{}
		// No WHERE
		mock.ExpectQuery("SELECT COUNT(*) FROM task_execution_results").
			WillReturnError(errors.New("count error"))

		_, _, err := repo.List(ctx, filter)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "execute count")
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTaskResultRepository_Create tests the Create method.
func TestTaskResultRepository_Create(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewTaskResultRepository(db)
	ctx := context.Background()
	result := newTestTaskExecutionResult()

	expectedSQL := "INSERT INTO task_execution_results (id,task_id,status,run_at,result,is_critical,execution_time) VALUES ($1,$2,$3,$4,$5,$6,$7)"

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(expectedSQL).
			WithArgs(
				result.ID, result.TaskID, result.Status, result.RunAt,
				result.Result, result.IsCritical, result.ExecutionTime,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, result)
		assert.NoError(t, err)
	})

	t.Run("auto-generate ID when zero", func(t *testing.T) {
		zeroResult := &dto.TaskExecutionResult{
			TaskID:        uuid.New(),
			Status:        retrier.StatusPending,
			RunAt:         time.Now(),
			Result:        nil,
			IsCritical:    true,
			ExecutionTime: 0,
		}
		mock.ExpectExec(expectedSQL).
			WithArgs(
				sqlmock.AnyArg(), // id will be generated
				zeroResult.TaskID,
				zeroResult.Status,
				zeroResult.RunAt,
				zeroResult.Result,
				zeroResult.IsCritical,
				zeroResult.ExecutionTime,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(ctx, zeroResult)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, zeroResult.ID)
	})

	t.Run("db error", func(t *testing.T) {
		badResult := newTestTaskExecutionResult()
		mock.ExpectExec(expectedSQL).
			WithArgs(
				badResult.ID, badResult.TaskID, badResult.Status, badResult.RunAt,
				badResult.Result, badResult.IsCritical, badResult.ExecutionTime,
			).
			WillReturnError(errors.New("duplicate key"))

		err := repo.Create(ctx, badResult)
		assert.Error(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTaskResultRepository_Delete tests the Delete method.
func TestTaskResultRepository_Delete(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewTaskResultRepository(db)
	ctx := context.Background()
	id := uuid.New()

	expectedSQL := "DELETE FROM task_execution_results WHERE id = $1"

	t.Run("success", func(t *testing.T) {
		mock.ExpectExec(expectedSQL).
			WithArgs(id).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(ctx, id)
		assert.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectExec(expectedSQL).
			WithArgs(id).
			WillReturnError(errors.New("delete error"))

		err := repo.Delete(ctx, id)
		assert.Error(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestTaskResultRepository_WithTx tests repository operations within a transaction.
func TestTaskResultRepository_WithTx(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewTaskResultRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err)

	ctxWithTx := storage.WithTx(ctx, tx)

	result := newTestTaskExecutionResult()
	insertSQL := "INSERT INTO task_execution_results (id,task_id,status,run_at,result,is_critical,execution_time) VALUES ($1,$2,$3,$4,$5,$6,$7)"

	// Insert inside transaction
	mock.ExpectExec(insertSQL).
		WithArgs(
			result.ID, result.TaskID, result.Status, result.RunAt,
			result.Result, result.IsCritical, result.ExecutionTime,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(ctxWithTx, result)
	assert.NoError(t, err)

	// Select inside transaction
	selectSQL := "SELECT id, task_id, status, run_at, result, is_critical, execution_time FROM task_execution_results WHERE id = $1"
	rows := addResultRow(sqlmock.NewRows(resultColumns), result)
	mock.ExpectQuery(selectSQL).
		WithArgs(result.ID).
		WillReturnRows(rows)

	fetched, err := repo.GetByID(ctxWithTx, result.ID)
	assert.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, result.ID, fetched.ID)

	// Commit transaction
	mock.ExpectCommit()
	assert.NoError(t, tx.Commit())

	assert.NoError(t, mock.ExpectationsWereMet())
}

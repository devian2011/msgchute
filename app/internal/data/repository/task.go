package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	"github.com/devian2011/msgchute/internal/dto"
	"github.com/devian2011/msgchute/internal/io/storage"
)

const taskTable = "tasks"

// ErrTaskNotFound is returned when a task cannot be found by ID.
var ErrTaskNotFound = errors.New("task not found")

// TaskRepository provides access to the tasks storage.
type TaskRepository struct {
	db      DBContext
	builder squirrel.StatementBuilderType
}

// NewTaskRepository creates a new TaskRepository instance.
func NewTaskRepository(db DBContext) *TaskRepository {
	return &TaskRepository{
		db:      db,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// getDB returns the appropriate DBContext (transaction from context or main db).
func (r *TaskRepository) getDB(ctx context.Context) DBContext {
	if tx := storage.ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.db
}

// taskColumns lists all columns of the tasks table.
var taskColumns = []string{
	"id", "message_id", "worker", "status", "retries", "max_retries",
	"backoff_code", "backoff_params", "deadline", "is_processed", "created_at", "last_run", "next_run",
}

// GetByID retrieves a task by its UUID.
// Returns ErrTaskNotFound if no task exists with the given ID.
func (r *TaskRepository) GetByID(ctx context.Context, ID uuid.UUID) (*dto.Task, error) {
	query, args, err := r.builder.
		Select(taskColumns...).
		From(taskTable).
		Where(squirrel.Eq{"id": ID}).
		Suffix("FOR UPDATE").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var task dto.Task
	db := r.getDB(ctx)
	err = db.Get(&task, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("get task by id: %w", err)
	}
	return &task, nil
}

// GetByMessageID returns all tasks associated with the given message ID.
// Results are ordered by created_at in ascending order.
func (r *TaskRepository) GetByMessageID(ctx context.Context, messageID uuid.UUID) ([]dto.Task, error) {
	query, args, err := r.builder.
		Select(taskColumns...).
		From(taskTable).
		Where(squirrel.Eq{"message_id": messageID}).
		OrderBy("created_at ASC").
		Suffix("FOR UPDATE").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var tasks []dto.Task
	db := r.getDB(ctx)
	err = db.Select(&tasks, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get tasks by message_id: %w", err)
	}
	return tasks, nil
}

// Create inserts a new task into the database.
// Returns the inserted task (with generated fields filled) or an error.
func (r *TaskRepository) Create(ctx context.Context, task *dto.Task) (*dto.Task, error) {
	query, args, err := r.builder.
		Insert(taskTable).
		Columns("id", "message_id", "worker", "status", "retries", "max_retries",
			"backoff_code", "backoff_params", "deadline", "is_processed", "last_run", "next_run").
		Values(task.ID, task.MessageID, task.Worker, task.Status, task.Retries,
			task.MaxRetries, task.BackOffCode, task.BackOffParams,
			task.Deadline, task.IsProcessed, task.LastRun, task.NextRun).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build insert: %w", err)
	}

	db := r.getDB(ctx)
	_, err = db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	return task, nil
}

// Update replaces an existing task with the provided data.
// Returns the updated task or an error if the update fails.
// Returns ErrTaskNotFound if no task exists with the given ID.
func (r *TaskRepository) Update(ctx context.Context, task *dto.Task) (*dto.Task, error) {
	query, args, err := r.builder.
		Update(taskTable).
		Set("message_id", task.MessageID).
		Set("worker", task.Worker).
		Set("status", task.Status).
		Set("retries", task.Retries).
		Set("max_retries", task.MaxRetries).
		Set("backoff_code", task.BackOffCode).
		Set("backoff_params", task.BackOffParams).
		Set("deadline", task.Deadline).
		Set("is_processed", task.IsProcessed).
		Set("last_run", task.LastRun).
		Set("next_run", task.NextRun).
		Where(squirrel.Eq{"id": task.ID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build update: %w", err)
	}

	db := r.getDB(ctx)
	res, err := db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

// Delete removes a task by its ID.
// Returns ErrTaskNotFound if no task exists with the given ID.
func (r *TaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query, args, err := r.builder.
		Delete(taskTable).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build delete: %w", err)
	}

	db := r.getDB(ctx)
	res, err := db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return ErrTaskNotFound
	}
	return nil
}

// GetByMessageIDs retrieves tasks for given message IDs, grouped by message_id.
func (r *TaskRepository) GetByMessageIDs(ctx context.Context, messageIDs []uuid.UUID) (map[uuid.UUID][]dto.Task, error) {
	if len(messageIDs) == 0 {
		return map[uuid.UUID][]dto.Task{}, nil
	}
	return r.List(ctx, dto.TaskFilter{MessageIDs: messageIDs})
}

// Lock sets is_processed = true for tasks with given IDs.
func (r *TaskRepository) Lock(ctx context.Context, IDs []uuid.UUID) error {
	if len(IDs) == 0 {
		return nil
	}
	db := r.getDB(ctx)
	query, args, err := r.builder.Update(taskTable).
		Set("is_processed", true).
		Where(squirrel.Eq{"id": IDs}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build lock query: %w", err)
	}
	_, err = db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("lock tasks: %w", err)
	}
	return nil
}

// Unlock sets is_processed = false for tasks with given IDs.
func (r *TaskRepository) Unlock(ctx context.Context, IDs []uuid.UUID) error {
	if len(IDs) == 0 {
		return nil
	}
	db := r.getDB(ctx)
	query, args, err := r.builder.Update(taskTable).
		Set("is_processed", false).
		Where(squirrel.Eq{"id": IDs}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build unlock query: %w", err)
	}
	_, err = db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("unlock tasks: %w", err)
	}
	return nil
}

// List returns tasks that match the given filter, grouped by message_id.
// It applies all filters, sorting, and pagination.
// If SortBy is empty, default ordering is applied (next_run ASC NULLS LAST, created_at ASC).
func (r *TaskRepository) List(ctx context.Context, filter dto.TaskFilter) (map[uuid.UUID][]dto.Task, error) {
	b := r.builder.Select(taskColumns...).From(taskTable)

	if len(filter.MessageIDs) > 0 {
		b = b.Where(squirrel.Eq{"message_id": filter.MessageIDs})
	}
	if len(filter.Statuses) > 0 {
		b = b.Where(squirrel.Eq{"status": filter.Statuses})
	}
	if filter.Worker != nil {
		b = b.Where(squirrel.Eq{"worker": *filter.Worker})
	}
	if filter.NextRunBefore != nil {
		b = b.Where(squirrel.LtOrEq{"next_run": *filter.NextRunBefore})
	}
	if filter.NextRunAfter != nil {
		b = b.Where(squirrel.GtOrEq{"next_run": *filter.NextRunAfter})
	}
	if filter.IsProcessed != nil {
		b = b.Where(squirrel.Eq{"is_processed": *filter.IsProcessed})
	}

	if filter.SortBy != "" {
		if !isAllowedColumn(filter.SortBy) {
			return nil, fmt.Errorf("invalid sort field: %s", filter.SortBy)
		}
		order := strings.ToUpper(filter.SortOrder)
		if order != "ASC" && order != "DESC" {
			order = "ASC"
		}
		b = b.OrderBy(filter.SortBy + " " + order)
	} else {
		b = b.OrderBy("next_run ASC NULLS LAST", "created_at ASC")
	}

	if filter.Limit > 0 {
		b = b.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		b = b.Offset(filter.Offset)
	}

	query, args, err := b.Suffix("FOR UPDATE").ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list query: %w", err)
	}

	var tasks []dto.Task
	db := r.getDB(ctx)
	err = db.Select(&tasks, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}

	result := make(map[uuid.UUID][]dto.Task)
	for _, task := range tasks {
		result[task.MessageID] = append(result[task.MessageID], task)
	}
	return result, nil
}

// isAllowedColumn checks if the given column name is a valid column of the tasks table.
func isAllowedColumn(col string) bool {
	for _, c := range taskColumns {
		if c == col {
			return true
		}
	}
	return false
}

// toNullTime returns nil if the time is zero, otherwise returns the time itself.
func toNullTime(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t
}

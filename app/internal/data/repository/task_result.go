package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	"github.com/devian2011/msgchute/internal/dto"
	"github.com/devian2011/msgchute/internal/io/storage"
)

const taskResultsTable = "task_execution_results"

// TaskResultRepository handles storage operations for task execution results.
type TaskResultRepository struct {
	db      DBContext
	builder squirrel.StatementBuilderType
}

// NewTaskResultRepository creates a new repository instance.
func NewTaskResultRepository(db DBContext) *TaskResultRepository {
	return &TaskResultRepository{
		db:      db,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// getDB returns the appropriate DBContext (transaction from context or main db).
func (r *TaskResultRepository) getDB(ctx context.Context) DBContext {
	if tx := storage.ExtractTx(ctx); tx != nil {
		return tx
	}
	return r.db
}

// GetByID retrieves a single record by primary key.
// Returns nil, nil if no record is found.
func (r *TaskResultRepository) GetByID(ctx context.Context, ID uuid.UUID) (*dto.TaskExecutionResult, error) {
	query, args, err := r.builder.
		Select("id", "task_id", "status", "run_at", "result", "is_critical", "execution_time").
		From(taskResultsTable).
		Where(squirrel.Eq{"id": ID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build select query: %w", err)
	}

	var result dto.TaskExecutionResult
	db := r.getDB(ctx)
	err = db.Get(&result, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("execute select: %w", err)
	}
	return &result, nil
}

// List returns records matching the filter, along with the total count (ignoring pagination).
// Sorting and pagination are applied to the main query, while the total count is unaffected.
func (r *TaskResultRepository) List(
	ctx context.Context,
	filter dto.TaskExecutionResultFilter,
) ([]*dto.TaskExecutionResult, int, error) {
	base := r.builder.
		Select("id", "task_id", "status", "run_at", "result", "is_critical", "execution_time").
		From(taskResultsTable)

	whereConditions := squirrel.And{}
	if filter.TaskID != nil {
		whereConditions = append(whereConditions, squirrel.Eq{"task_id": *filter.TaskID})
	}
	if len(filter.Statuses) > 0 {
		whereConditions = append(whereConditions, squirrel.Eq{"status": filter.Statuses})
	}
	if len(whereConditions) > 0 {
		base = base.Where(whereConditions)
	}

	if filter.SortBy != "" {
		order := filter.SortOrder
		if order != "ASC" && order != "DESC" {
			order = "ASC"
		}
		allowedColumns := map[string]bool{
			"id": true, "task_id": true, "status": true,
			"run_at": true, "is_critical": true, "execution_time": true,
		}
		if !allowedColumns[filter.SortBy] {
			return nil, 0, fmt.Errorf("invalid sort column: %s", filter.SortBy)
		}
		base = base.OrderBy(fmt.Sprintf("%s %s", filter.SortBy, order))
	}

	countBuilder := r.builder.Select("COUNT(*)").From(taskResultsTable)
	if len(whereConditions) > 0 {
		countBuilder = countBuilder.Where(whereConditions)
	}
	countQuery, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build count query: %w", err)
	}

	var total int
	db := r.getDB(ctx)
	err = db.Get(&total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("execute count: %w", err)
	}

	base = base.Limit(filter.Limit).Offset(filter.Offset)
	query, args, err := base.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build list query: %w", err)
	}

	var results []*dto.TaskExecutionResult
	err = db.Select(&results, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("execute list: %w", err)
	}

	return results, total, nil
}

// GetByTaskIDs retrieves results for multiple task IDs, grouped by task_id.
// Returns a map where keys are task IDs and values are slices of results.
func (r *TaskResultRepository) GetByTaskIDs(
	ctx context.Context,
	taskIDs []uuid.UUID,
) (map[uuid.UUID][]dto.TaskExecutionResult, error) {
	if len(taskIDs) == 0 {
		return map[uuid.UUID][]dto.TaskExecutionResult{}, nil
	}

	query, args, err := r.builder.
		Select("id", "task_id", "status", "run_at", "result", "is_critical", "execution_time").
		From(taskResultsTable).
		Where(squirrel.Eq{"task_id": taskIDs}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var taskResults []dto.TaskExecutionResult
	db := r.getDB(ctx)
	err = db.Select(&taskResults, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return map[uuid.UUID][]dto.TaskExecutionResult{}, nil
		}
		return nil, fmt.Errorf("execute list: %w", err)
	}

	result := make(map[uuid.UUID][]dto.TaskExecutionResult, len(taskIDs))
	for i := range taskResults {
		taskID := taskResults[i].TaskID
		result[taskID] = append(result[taskID], taskResults[i])
	}
	return result, nil
}

// Create inserts a new record. If ID is zero, a new UUID is generated.
func (r *TaskResultRepository) Create(ctx context.Context, result *dto.TaskExecutionResult) error {
	if result.ID == uuid.Nil {
		result.ID = uuid.New()
	}

	query, args, err := r.builder.
		Insert(taskResultsTable).
		Columns("id", "task_id", "status", "run_at", "result", "is_critical", "execution_time").
		Values(result.ID, result.TaskID, result.Status, result.RunAt, result.Result, result.IsCritical, result.ExecutionTime).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	db := r.getDB(ctx)
	_, err = db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("execute insert: %w", err)
	}
	return nil
}

// Delete removes a record by ID.
func (r *TaskResultRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query, args, err := r.builder.
		Delete(taskResultsTable).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build delete query: %w", err)
	}

	db := r.getDB(ctx)
	_, err = db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("execute delete: %w", err)
	}
	return nil
}

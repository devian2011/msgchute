package dto

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/devian2011/retrier"
	"github.com/google/uuid"
)

// TaskFilter defines the filtering, pagination, and sorting parameters for List and Count.
type TaskFilter struct {
	// MessageIDs filters tasks by the associated message ID.
	MessageIDs []uuid.UUID `json:"message_ids"`
	// Statuses filters tasks by one or more statuses (IN clause).
	Statuses []retrier.TaskStatus `json:"statuses"`
	// Worker filters tasks by the assigned worker name.
	Worker *string `json:"worker"`
	// NextRunBefore includes tasks with next_run <= this value.
	NextRunBefore *time.Time `json:"next_run_before"`
	// NextRunAfter includes tasks with next_run >= this value.
	NextRunAfter *time.Time `json:"next_run_after"`
	// IsProcessed filter processed tasks
	IsProcessed *bool `json:"is_processed"`

	// Limit sets the maximum number of records to return (pagination).
	Limit uint64 `json:"limit"`
	// Offset sets the number of records to skip (pagination).
	Offset uint64 `json:"offset"`

	// SortBy specifies the column to sort by (must be one of the taskColumns).
	SortBy string `json:"sort_by"`
	// SortOrder specifies the sort direction: "ASC" or "DESC" (case-insensitive).
	SortOrder string `json:"sort_order"`
}

// Task retries task store struct
type Task struct {
	ID            uuid.UUID          `db:"id" json:"id"`
	MessageID     uuid.UUID          `db:"message_id" json:"message_id"`
	Worker        string             `db:"worker" json:"worker"`
	Status        retrier.TaskStatus `db:"status" json:"status"`
	Retries       int                `db:"retries" json:"retries"`
	MaxRetries    int                `db:"max_retries" json:"max_retries"`
	BackOffCode   string             `db:"backoff_code" json:"backoff_code"`
	BackOffParams BackOffParams      `db:"backoff_params" json:"backoff_params"`
	Deadline      time.Time          `db:"deadline" json:"deadline"`
	IsProcessed   bool               `db:"is_processed" json:"is_processed"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	LastRun   time.Time `db:"last_run" json:"last_run"`
	NextRun   time.Time `db:"next_run" json:"next_run"`
}

type BackOffParams map[retrier.BackOffParam]interface{}

func (b BackOffParams) Value() (driver.Value, error) {
	return sonic.Marshal(b)
}

func (b *BackOffParams) Scan(src interface{}) error {
	if src == nil {
		*b = BackOffParams{}
		return nil
	}
	data, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("unexpected type %T", src)
	}
	return sonic.Unmarshal(data, b)
}

// TaskExecutionResultFilter defines filtering, pagination, and sorting parameters.
type TaskExecutionResultFilter struct {
	// TaskID filters by the associated task ID (optional).
	TaskID *uuid.UUID `json:"task_id"`
	// Statuses filters by one or more statuses (IN clause).
	Statuses []retrier.TaskStatus `json:"statuses"`

	// Limit sets the maximum number of records to return.
	Limit uint64 `json:"limit"`
	// Offset sets the number of records to skip.
	Offset uint64 `json:"offset"`

	// SortBy specifies the column to sort by (must be one of the allowed columns).
	SortBy string `json:"sort_by"`
	// SortOrder specifies the sort direction: "ASC" or "DESC" (case-insensitive).
	SortOrder string `json:"sort_order"`
}

// TaskExecutionResult result store struct
type TaskExecutionResult struct {
	ID            uuid.UUID          `db:"id" json:"id"`
	TaskID        uuid.UUID          `db:"task_id" json:"task_id"`
	Status        retrier.TaskStatus `db:"status" json:"status"`
	RunAt         time.Time          `db:"run_at" json:"run_at"`
	Result        []byte             `db:"result" json:"result"`
	IsCritical    bool               `db:"is_critical" json:"is_critical"`
	ExecutionTime time.Duration      `db:"execution_time" json:"execution_time"`
}

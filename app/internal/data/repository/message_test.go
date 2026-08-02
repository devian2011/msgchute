package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/devian2011/msgchute/internal/dto"
	"github.com/devian2011/msgchute/internal/io/storage"
)

func TestMessageRepository_Create(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewMessageRepository(db)
	ctx := context.Background()

	msgID := uuid.New()
	msg := &dto.Message{
		ID:         msgID,
		SenderID:   "srv_auth",
		Transport:  "email",
		Code:       "otp_verify",
		Recipients: dto.Recipients{"user@example.com"},
		Status:     dto.MessageStatusRunning,
		Meta:       dto.MessageMeta{"priority": "high"},
		Params:     dto.MessageParams{"code": {Value: "1234"}},
		Retry:      &dto.Retry{Retries: 3},
		Schedule:   time.Now(),
		Deadline:   time.Now().Add(time.Hour),
		Subject:    "Verification",
		Body:       "Code: 1234",
	}

	expectedSQL := "INSERT INTO messages (id,sender_id,transport,template_code,recipients,params,retry,schedule,deadline,subject,body,status,meta) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)"

	mock.ExpectExec(expectedSQL).
		WithArgs(
			msg.ID, msg.SenderID, msg.Transport, msg.Code,
			msg.Recipients, msg.Params, msg.Retry, msg.Schedule,
			msg.Deadline, msg.Subject, msg.Body, msg.Status, msg.Meta,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(ctx, msg)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_Find(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewMessageRepository(db)
	ctx := context.Background()

	t.Run("full advanced filtering and query execution", func(t *testing.T) {
		recipients := []string{"user@example.com"}
		senderIDs := []string{"billing"}
		codes := []string{"invoice_remind"}
		transports := []string{"email"}
		statuses := []dto.MessageStatus{dto.MessageStatusSucceeded}
		sortField := "schedule"
		sortOrder := "DESC"

		filter := &dto.MessageFilter{
			Limit:     10,
			Offset:    0,
			Recipient: recipients,
			SenderIDs: senderIDs,
			Code:      codes,
			Transport: transports,
			Status:    statuses,
			SortField: &sortField,
			SortOrder: &sortOrder,
		}

		expectedCountSQL := "SELECT COUNT(*) FROM messages WHERE sender_id IN ($1) AND template_code IN ($2) AND transport IN ($3) AND status IN ($4) AND recipients @> $5"
		mock.ExpectQuery(expectedCountSQL).
			WithArgs("billing", "invoice_remind", "email", dto.MessageStatusSucceeded, `["user@example.com"]`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		expectedSelectSQL := "SELECT id, sender_id, transport, template_code AS code, recipients, params, retry, schedule, deadline, subject, body, status, meta FROM messages WHERE sender_id IN ($1) AND template_code IN ($2) AND transport IN ($3) AND status IN ($4) AND recipients @> $5 ORDER BY schedule DESC LIMIT 10 FOR UPDATE"
		rows := sqlmock.NewRows([]string{
			"id", "sender_id", "transport", "code",
			"recipients", "params", "retry", "schedule",
			"deadline", "subject", "body", "status", "meta",
		}).AddRow(
			uuid.New(), "billing", "email", "invoice_remind",
			[]byte(`["user@example.com"]`), []byte(`{"code":{"value":"1234"}}`), nil,
			time.Now(), time.Now(), "Overdue invoice", "Body text",
			dto.MessageStatusSucceeded, []byte(`{"priority":"high"}`),
		)

		mock.ExpectQuery(expectedSelectSQL).
			WithArgs("billing", "invoice_remind", "email", dto.MessageStatusSucceeded, `["user@example.com"]`).
			WillReturnRows(rows)

		messages, total, err := repo.Find(ctx, filter)
		assert.NoError(t, err)
		assert.Equal(t, uint64(1), total)
		assert.Len(t, messages, 1)
		assert.Equal(t, "invoice_remind", messages[0].Code)
	})

	t.Run("fallback sorting on invalid sort field", func(t *testing.T) {
		invalidField := "body; DROP TABLE messages;--"
		filter := &dto.MessageFilter{
			Limit:     5,
			SortField: &invalidField,
		}

		mock.ExpectQuery("SELECT COUNT(*) FROM messages").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		expectedSelectSQL := "SELECT id, sender_id, transport, template_code AS code, recipients, params, retry, schedule, deadline, subject, body, status, meta FROM messages ORDER BY schedule DESC, id DESC LIMIT 5 FOR UPDATE"
		mock.ExpectQuery(expectedSelectSQL).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "sender_id", "transport", "code",
				"recipients", "params", "retry", "schedule",
				"deadline", "subject", "body", "status", "meta",
			}))

		_, _, err := repo.Find(ctx, filter)
		assert.NoError(t, err)
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMessageRepository_CreateWithTransaction(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewMessageRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	tx, err := db.BeginTxx(ctx, nil)
	require.NoError(t, err)

	// Помещаем транзакцию в контекст
	ctxWithTx := storage.WithTx(ctx, tx)

	msg := &dto.Message{
		ID:        uuid.New(),
		SenderID:  "tx_sender",
		Transport: "sms",
		Status:    dto.MessageStatusRunning,
		Subject:   "Tx Subj",
		Body:      "Tx Body",
	}

	expectedSQL := "INSERT INTO messages (id,sender_id,transport,template_code,recipients,params,retry,schedule,deadline,subject,body,status,meta) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)"

	mock.ExpectExec(expectedSQL).
		WithArgs(msg.ID, msg.SenderID, msg.Transport, nil, msg.Recipients, msg.Params, msg.Retry, nil, msg.Deadline, msg.Subject, msg.Body, msg.Status, msg.Meta).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.Create(ctxWithTx, msg)
	assert.NoError(t, err)

	mock.ExpectCommit()
	assert.NoError(t, tx.Commit())

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewMessageRepository(db)
	ctx := context.Background()
	id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	expectedQuery := `SELECT id, sender_id, transport, template_code AS code, recipients, params, retry, schedule, deadline, subject, body, status, meta FROM messages WHERE id = $1 FOR UPDATE`

	deadline := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	schedule := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"id", "sender_id", "transport", "code",
		"recipients", "params", "retry", "schedule",
		"deadline", "subject", "body", "status", "meta",
	}).AddRow(
		id.String(),
		"sender-1",
		"email",
		"tmpl_001",
		[]byte(`["user@example.com"]`),
		[]byte(`{"p1":{"value":"v1"}}`),
		[]byte(`{"retries":3,"strategy":"fixed"}`),
		schedule,
		deadline,
		"Hello",
		"Body text",
		dto.MessageStatusRunning,
		[]byte(`{"key":"value"}`),
	)

	mock.ExpectQuery(expectedQuery).
		WithArgs(id.String()).
		WillReturnRows(rows)

	msg, err := repo.GetByID(ctx, id)
	assert.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, id.String(), msg.ID.String())
	assert.Equal(t, "sender-1", msg.SenderID)
	assert.Equal(t, "tmpl_001", msg.Code)
	assert.Equal(t, dto.MessageStatusRunning, msg.Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByID_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewMessageRepository(db)
	ctx := context.Background()
	id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")

	expectedSQL := "SELECT id, sender_id, transport, template_code AS code, recipients, params, retry, schedule, deadline, subject, body, status, meta FROM messages WHERE id = $1 FOR UPDATE"
	mock.ExpectQuery(expectedSQL).
		WithArgs(id.String()).
		WillReturnError(sql.ErrNoRows)

	msg, err := repo.GetByID(ctx, id)
	assert.Nil(t, msg)
	assert.ErrorIs(t, err, sql.ErrNoRows)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByIDs_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewMessageRepository(db)
	ctx := context.Background()

	expectedQuery := `SELECT id, sender_id, transport, template_code AS code, recipients, params, retry, schedule, deadline, subject, body, status, meta FROM messages WHERE id IN ($1,$2)`

	id1 := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	id2 := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	ids := []uuid.UUID{id1, id2}

	deadline1 := time.Date(2025, 6, 15, 18, 30, 0, 0, time.UTC)
	schedule1 := time.Date(2025, 6, 10, 9, 0, 0, 0, time.UTC)
	deadline2 := time.Date(2025, 7, 1, 12, 0, 0, 0, time.UTC)
	schedule2 := time.Date(2025, 6, 20, 14, 45, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"id", "sender_id", "transport", "code",
		"recipients", "params", "retry", "schedule",
		"deadline", "subject", "body", "status", "meta",
	}).AddRow(
		id1.String(),
		"sender-A",
		"sms",
		"code-1",
		[]byte(`["+79001234567"]`),
		[]byte(`{"p1":{"value":"hello"},"p2":{"value":123}}`),
		[]byte(`{"retries":3,"strategy":"linear"}`),
		schedule1,
		deadline1,
		"Welcome!",
		"Hello, this is a test message.",
		dto.MessageStatusSucceeded,
		[]byte(`{"source":"web"}`),
	).AddRow(
		id2.String(),
		"sender-B",
		"push",
		"code-2",
		[]byte(`["user@example.com"]`),
		[]byte(`{"x":{"value":99}}`),
		nil,
		schedule2,
		deadline2,
		"Alert",
		"Server error occurred.",
		dto.MessageStatusFailed,
		[]byte(`{"priority":"high"}`),
	)

	mock.ExpectQuery(expectedQuery).
		WithArgs(ids[0].String(), ids[1].String()).
		WillReturnRows(rows)

	messages, err := repo.GetByIDs(ctx, ids)
	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, ids[0].String(), messages[0].ID.String())
	assert.Equal(t, ids[1].String(), messages[1].ID.String())
	assert.Equal(t, dto.MessageStatusSucceeded, messages[0].Status)
	assert.Equal(t, dto.MessageStatusFailed, messages[1].Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByIDs_EmptyInput(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewMessageRepository(db)
	ctx := context.Background()

	messages, err := repo.GetByIDs(ctx, []uuid.UUID{})
	assert.NoError(t, err)
	assert.Empty(t, messages)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetByIDs_DBError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewMessageRepository(db)
	ctx := context.Background()

	testUUID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	ids := []uuid.UUID{testUUID}

	expectedSQL := "SELECT id, sender_id, transport, template_code AS code, recipients, params, retry, schedule, deadline, subject, body, status, meta FROM messages WHERE id IN ($1)"
	mock.ExpectQuery(expectedSQL).
		WithArgs(testUUID.String()).
		WillReturnError(fmt.Errorf("connection refused"))

	messages, err := repo.GetByIDs(ctx, ids)
	assert.Nil(t, messages)
	assert.ErrorContains(t, err, "get messages by IDs")
	assert.ErrorContains(t, err, "connection refused")

	assert.NoError(t, mock.ExpectationsWereMet())
}

package public

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/devian2011/msgchute/internal/dto"
)

// --- mocks ---

type mockValidator struct {
	hasErrors bool
	errs      map[string]error
	calls     int
	lastMsg   *dto.Message
}

func (m *mockValidator) Validate(msg *dto.Message) (bool, map[string]error) {
	m.calls++
	m.lastMsg = msg
	return m.hasErrors, m.errs
}

type mockQueue struct {
	retMsg  *dto.Message
	retTask *dto.Task
	retErr  error
	calls   int
	lastMsg *dto.Message
}

func (m *mockQueue) Add(msg *dto.Message) (*dto.Message, *dto.Task, error) {
	m.calls++
	m.lastMsg = msg
	return m.retMsg, m.retTask, m.retErr
}

// sequentialValidator returns a different result per Validate call, in order.
type sequentialValidator struct {
	results []struct {
		hasErrors bool
		errs      map[string]error
	}
	calls int
}

func (s *sequentialValidator) Validate(_ *dto.Message) (bool, map[string]error) {
	r := s.results[s.calls]
	s.calls++
	return r.hasErrors, r.errs
}

// --- SenderHandler ---

func TestNewSenderHandler(t *testing.T) {
	q := &mockQueue{}
	v := &mockValidator{}

	h := NewSenderHandler(q, v)

	assert.NotNil(t, h)
	assert.Equal(t, q, h.q)
	assert.Equal(t, v, h.v)
}

func TestSenderHandler_Handle_Success(t *testing.T) {
	inMsg := &dto.Message{Transport: "email"}
	outMsg := &dto.Message{Transport: "email"}
	outTask := &dto.Task{}

	q := &mockQueue{retMsg: outMsg, retTask: outTask}
	v := &mockValidator{hasErrors: false, errs: map[string]error{}}
	h := NewSenderHandler(q, v)

	msg, task, err := h.Handle(inMsg)

	assert.NoError(t, err)
	assert.Equal(t, outMsg, msg)
	assert.Equal(t, outTask, task)
	assert.Equal(t, 1, v.calls)
	assert.Equal(t, 1, q.calls)
	assert.Same(t, inMsg, v.lastMsg, "validator receives the original message")
	assert.Same(t, inMsg, q.lastMsg, "queue receives the original message")
}

func TestSenderHandler_Handle_QueueError(t *testing.T) {
	queueErr := errors.New("db is down")

	q := &mockQueue{retErr: queueErr}
	v := &mockValidator{hasErrors: false}
	h := NewSenderHandler(q, v)

	msg, task, err := h.Handle(&dto.Message{Transport: "email"})

	assert.Nil(t, msg)
	assert.Nil(t, task)
	assert.EqualError(t, err, "db is down")
	assert.Equal(t, 1, q.calls)
}

func TestSenderHandler_Handle_ValidationError(t *testing.T) {
	q := &mockQueue{}
	v := &mockValidator{
		hasErrors: true,
		errs: map[string]error{
			"transport": errors.New("unknown transport: push"),
		},
	}
	h := NewSenderHandler(q, v)

	msg, task, err := h.Handle(&dto.Message{Transport: "push"})

	assert.Nil(t, msg)
	assert.Nil(t, task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transport")
	assert.Contains(t, err.Error(), "unknown transport: push")
	assert.Equal(t, 0, q.calls, "queue.Add must not be called on validation failure")
}

// --- BatchSenderHandler ---

func TestNewBatchSenderHandler(t *testing.T) {
	q := &mockQueue{}
	v := &mockValidator{}

	h := NewBatchSenderHandler(q, v)

	assert.NotNil(t, h)
	assert.Equal(t, q, h.q)
	assert.Equal(t, v, h.v)
}

func TestBatchSenderHandler_Handle_EmptyInput(t *testing.T) {
	v := &mockValidator{}
	q := &mockQueue{}
	h := NewBatchSenderHandler(q, v)

	assert.Empty(t, h.Handle(nil))
	assert.Empty(t, h.Handle([]*dto.Message{}))
	assert.Equal(t, 0, v.calls, "validator not called for empty input")
	assert.Equal(t, 0, q.calls, "queue not called for empty input")
}

func TestBatchSenderHandler_Handle_AllValid(t *testing.T) {
	msg1 := &dto.Message{Transport: "email"}
	msg2 := &dto.Message{Transport: "sms"}
	task := &dto.Task{}

	q := &mockQueue{retMsg: msg1, retTask: task}
	v := &mockValidator{hasErrors: false, errs: map[string]error{}}
	h := NewBatchSenderHandler(q, v)

	results := h.Handle([]*dto.Message{msg1, msg2})

	assert.Len(t, results, 2)
	for _, r := range results {
		assert.NotNil(t, r)
		assert.Nil(t, r.Err)
		assert.Equal(t, task, r.Task)
		assert.Empty(t, r.VErr)
	}
	assert.Equal(t, 2, q.calls)
	assert.Equal(t, 2, v.calls)
}

func TestBatchSenderHandler_Handle_AllInvalid(t *testing.T) {
	msg1 := &dto.Message{Transport: "push"}
	msg2 := &dto.Message{Transport: "fax"}

	q := &mockQueue{}
	v := &mockValidator{
		hasErrors: true,
		errs:      map[string]error{"transport": errors.New("unknown transport")},
	}
	h := NewBatchSenderHandler(q, v)

	results := h.Handle([]*dto.Message{msg1, msg2})

	assert.Len(t, results, 2)
	for _, r := range results {
		assert.Nil(t, r.Message)
		assert.Nil(t, r.Task)
		assert.Nil(t, r.Err)
		assert.Contains(t, r.VErr, "transport")
	}
	assert.Equal(t, 0, q.calls, "queue.Add must not be called for invalid messages")
	assert.Equal(t, 2, v.calls)
}

func TestBatchSenderHandler_Handle_Mixed(t *testing.T) {
	validMsg := &dto.Message{Transport: "email"}
	invalidMsg := &dto.Message{Transport: "push"}
	task := &dto.Task{}

	v := &sequentialValidator{
		results: []struct {
			hasErrors bool
			errs      map[string]error
		}{
			{false, map[string]error{}},
			{true, map[string]error{"transport": errors.New("unknown transport: push")}},
		},
	}
	q := &mockQueue{retMsg: validMsg, retTask: task}
	h := NewBatchSenderHandler(q, v)

	results := h.Handle([]*dto.Message{validMsg, invalidMsg})

	assert.Len(t, results, 2)

	// first: valid
	assert.Equal(t, validMsg, results[0].Message)
	assert.Equal(t, task, results[0].Task)
	assert.Nil(t, results[0].Err)
	assert.Empty(t, results[0].VErr)

	// second: invalid
	assert.Nil(t, results[1].Message)
	assert.Nil(t, results[1].Task)
	assert.Nil(t, results[1].Err)
	assert.Contains(t, results[1].VErr, "transport")

	assert.Equal(t, 1, q.calls, "queue.Add called only for valid messages")
	assert.Equal(t, 2, v.calls, "validator runs for every message")
}

func TestBatchSenderHandler_Handle_PreservesOrder(t *testing.T) {
	inputs := []*dto.Message{
		{Transport: "email"},
		{Transport: "sms"},
		{Transport: "push"},
	}

	v := &sequentialValidator{
		results: []struct {
			hasErrors bool
			errs      map[string]error
		}{
			{false, map[string]error{}},
			{false, map[string]error{}},
			{true, map[string]error{"transport": errors.New("unknown transport: push")}},
		},
	}
	h := NewBatchSenderHandler(&mockQueue{}, v)

	results := h.Handle(inputs)

	assert.Len(t, results, 3)
	assert.Empty(t, results[0].VErr)
	assert.Empty(t, results[1].VErr)
	assert.Contains(t, results[2].VErr, "transport")
}

func TestBatchSenderHandler_Handle_QueueErrorPropagated(t *testing.T) {
	queueErr := errors.New("db is down")

	q := &mockQueue{retErr: queueErr}
	v := &mockValidator{hasErrors: false, errs: map[string]error{}}
	h := NewBatchSenderHandler(q, v)

	results := h.Handle([]*dto.Message{{Transport: "email"}})

	assert.Len(t, results, 1)
	assert.EqualError(t, results[0].Err, "db is down")
	assert.Empty(t, results[0].VErr)
}

// --- RetryHandler ---

type mockRetryableQueue struct {
	retMsg  *dto.Message
	retTask *dto.Task
	retErr  error
	calls   int
	lastReq *dto.MessageRetryRequest
}

func (m *mockRetryableQueue) Retry(req *dto.MessageRetryRequest) (*dto.Message, *dto.Task, error) {
	m.calls++
	m.lastReq = req
	return m.retMsg, m.retTask, m.retErr
}

func TestNewMessageRetryHandler(t *testing.T) {
	q := &mockRetryableQueue{}

	h := NewMessageRetryHandler(q)

	assert.NotNil(t, h)
	assert.Equal(t, q, h.queue)
}

func TestRetryHandler_Handle_Success(t *testing.T) {
	req := &dto.MessageRetryRequest{}
	outMsg := &dto.Message{Transport: "email"}
	outTask := &dto.Task{}

	q := &mockRetryableQueue{retMsg: outMsg, retTask: outTask}
	h := NewMessageRetryHandler(q)

	msg, task, err := h.Handle(req)

	assert.NoError(t, err)
	assert.Equal(t, outMsg, msg)
	assert.Equal(t, outTask, task)
	assert.Equal(t, 1, q.calls)
	assert.Same(t, req, q.lastReq, "the same request is passed through")
}

func TestRetryHandler_Handle_QueueError(t *testing.T) {
	queueErr := errors.New("not found")

	q := &mockRetryableQueue{retErr: queueErr}
	h := NewMessageRetryHandler(q)

	msg, task, err := h.Handle(&dto.MessageRetryRequest{})

	assert.Nil(t, msg)
	assert.Nil(t, task)
	assert.EqualError(t, err, "not found")
	assert.Equal(t, 1, q.calls)
}

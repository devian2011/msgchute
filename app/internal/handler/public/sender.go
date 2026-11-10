package public

import (
	"github.com/devian2011/msgchute/internal/dto"
	"github.com/devian2011/msgchute/pkg/helper"
)

type validator interface {
	Validate(msg *dto.Message) (bool, map[string]error)
}

type queue interface {
	Add(*dto.Message) (*dto.Message, *dto.Task, error)
}

type SenderHandler struct {
	q queue
	v validator
}

func NewSenderHandler(q queue, v validator) *SenderHandler {
	return &SenderHandler{q: q, v: v}
}

func (h *SenderHandler) Handle(msg *dto.Message) (*dto.Message, *dto.Task, error) {
	hasErrors, vErr := h.v.Validate(msg)
	if hasErrors {
		return nil, nil, helper.ErrMapToError(vErr)
	}

	return h.q.Add(msg)
}

// Batch

type BatchSenderHandler struct {
	q queue
	v validator
}

func NewBatchSenderHandler(q queue, v validator) *BatchSenderHandler {
	return &BatchSenderHandler{q: q, v: v}
}

func (h *BatchSenderHandler) Handle(msg []*dto.Message) []*dto.AddBatchMessageResponse {
	var results []*dto.AddBatchMessageResponse
	for i := range msg {
		hasErrors, vR := h.v.Validate(msg[i])
		var m *dto.Message
		var t *dto.Task
		var e error
		if !hasErrors {
			m, t, e = h.q.Add(msg[i])
		}

		results = append(results, &dto.AddBatchMessageResponse{
			Message: m,
			Task:    t,
			Err:     e,
			VErr:    vR,
		})
	}

	return results
}

// Retry

type RetryableQueue interface {
	Retry(request *dto.MessageRetryRequest) (*dto.Message, *dto.Task, error)
}

type RetryHandler struct {
	queue RetryableQueue
}

func NewMessageRetryHandler(queue RetryableQueue) *RetryHandler {
	return &RetryHandler{queue: queue}
}

func (h *RetryHandler) Handle(r *dto.MessageRetryRequest) (*dto.Message, *dto.Task, error) {
	return h.queue.Retry(r)
}

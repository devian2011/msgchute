package admin

import "github.com/devian2011/retrier"

type WorkerManager interface {
	GetWorkerStatuses() map[string]retrier.FullWorkerState
}

type WorkerStatusHandler struct {
	m WorkerManager
}

func NewWorkerHandler(m WorkerManager) *WorkerStatusHandler {
	return &WorkerStatusHandler{m: m}
}

func (h *WorkerStatusHandler) Handle() map[string]retrier.FullWorkerState {
	return h.m.GetWorkerStatuses()
}

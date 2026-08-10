package admin

import (
	"net/http"

	"github.com/devian2011/retrier"

	"github.com/devian2011/msgchute/pkg/http/response"
)

type workerStatusHandler interface {
	Handle() map[string]retrier.FullWorkerState
}

type WorkerStatusEndpoint struct {
	h workerStatusHandler
}

func NewWorkerStatusEndpoint(h workerStatusHandler) *WorkerStatusEndpoint {
	return &WorkerStatusEndpoint{h: h}
}

func (h *WorkerStatusEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	states := h.h.Handle()
	response.WriteSuccessResponse(w, r, http.StatusOK, states)
}

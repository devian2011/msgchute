package public

import (
	"github.com/devian2011/msgchute/internal/dto"
)

type queue interface {
	Add(*dto.Message) (*dto.Message, *dto.Task, error)
}

type SenderHandler struct {
	q queue
}

func (h *SenderHandler) Handle(msg *dto.Message) (*dto.Message, *dto.Task, error) {
	return h.q.Add(msg)
}

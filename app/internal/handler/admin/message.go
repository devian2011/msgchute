package admin

import (
	"github.com/google/uuid"

	"github.com/devian2011/msgchute/internal/dto"
)

type messageFinder interface {
	Find(filter *dto.MessageFilter) ([]dto.FullMessageInfo, int, error)
	FindByID(ID uuid.UUID) (*dto.FullMessageInfo, error)
}

type MessageFindHandler struct {
	finder messageFinder
}

func NewMessageFindHandler(finder messageFinder) *MessageFindHandler {
	return &MessageFindHandler{
		finder: finder,
	}
}

func (h *MessageFindHandler) Handle(filter *dto.MessageFilter) ([]dto.FullMessageInfo, int, error) {
	return h.finder.Find(filter)
}

type MessageFindByIDHandler struct {
	finder messageFinder
}

func NewMessageFindByIDHandler(finder messageFinder) *MessageFindByIDHandler {
	return &MessageFindByIDHandler{
		finder: finder,
	}
}

func (h *MessageFindByIDHandler) Handle(ID uuid.UUID) (*dto.FullMessageInfo, error) {
	return h.finder.FindByID(ID)
}

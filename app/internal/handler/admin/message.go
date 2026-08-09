package admin

import (
	"context"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"

	"github.com/devian2011/msgchute/internal/dto"
)

type messageRecipientFinder interface {
	GetRecipients(ctx context.Context, search string) ([]string, error)
}

type MessageRecipientFindHandler struct {
	getter messageRecipientFinder
}

func NewMessageRecipientFindHandler(getter messageRecipientFinder) *MessageRecipientFindHandler {
	return &MessageRecipientFindHandler{getter: getter}
}

func (h *MessageRecipientFindHandler) Handle(ctx context.Context, search string) ([]string, error) {
	return h.getter.GetRecipients(ctx, search)
}

type messageDictionaryGetter interface {
	GetSenders(ctx context.Context) ([]string, error)
	GetTransports(ctx context.Context) ([]string, error)
	GetTemplates(ctx context.Context) ([]string, error)
}

type MessageDictionaryHandler struct {
	getter messageDictionaryGetter
}

func NewMessageDictionaryHandler(getter messageDictionaryGetter) *MessageDictionaryHandler {
	return &MessageDictionaryHandler{getter: getter}
}

func (h *MessageDictionaryHandler) Handle(ctx context.Context) (*dto.MessageDictionaries, error) {
	g, _ := errgroup.WithContext(ctx)
	result := &dto.MessageDictionaries{}
	g.Go(func() error {
		var getterErr error
		result.SenderIDs, getterErr = h.getter.GetSenders(ctx)
		return getterErr
	})
	g.Go(func() error {
		var getterErr error
		result.Transports, getterErr = h.getter.GetTransports(ctx)
		return getterErr
	})
	g.Go(func() error {
		var getterErr error
		result.Templates, getterErr = h.getter.GetTemplates(ctx)
		return getterErr
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}

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

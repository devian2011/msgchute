package public

import (
	"github.com/devian2011/msgchute/internal/dto"
)

type previewGenerator interface {
	GenerateMessage(t *dto.Message) (subject string, body string, err error)
}

type PreviewHandler struct {
	generator previewGenerator
}

func NewPreviewHandler(generator previewGenerator) *PreviewHandler {
	return &PreviewHandler{generator: generator}
}

func (h *PreviewHandler) Handle(msg *dto.Message) (dto.MessagePreview, error) {
	s, b, err := h.generator.GenerateMessage(msg)
	if err != nil {
		return dto.MessagePreview{}, err
	}
	return dto.MessagePreview{
		Subject: s,
		Body:    b,
	}, nil
}

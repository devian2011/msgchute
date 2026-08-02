package public

import (
	"net/http"

	"github.com/devian2011/msgchute/internal/dto"
)

type PreviewHandler interface {
	Handle(msg *dto.Message) (dto.MessagePreview, error)
}

type MessagePreviewEndpoint struct {
	h PreviewHandler
}

func NewMessagePreviewEndpoint(h PreviewHandler) *MessagePreviewEndpoint {
	return &MessagePreviewEndpoint{h: h}
}

func (e *MessagePreviewEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) () {

}

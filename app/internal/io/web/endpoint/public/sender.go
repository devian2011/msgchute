package public

import (
	"net/http"

	"github.com/devian2011/msgchute/internal/dto"
)

type SenderHandler interface {
	Handle(msg *dto.Message) (*dto.Message, *dto.Task, error)
}

type SenderEndpoint struct {
	h SenderHandler
}

func NewSenderEndpoint(h SenderHandler) *SenderEndpoint {
	return &SenderEndpoint{h}
}

func (e *SenderEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

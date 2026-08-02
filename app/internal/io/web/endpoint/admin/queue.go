package admin

import "net/http"

type MessageFinderHandler interface {
}

type MessageFinderEndpoint struct {
	h MessageFinderHandler
}

func NewMessageFinderEndpoint(h MessageFinderHandler) *MessageFinderEndpoint {
	return &MessageFinderEndpoint{h: h}
}

func (e *MessageFinderEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

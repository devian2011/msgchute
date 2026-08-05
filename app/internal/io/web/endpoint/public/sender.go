package public

import (
	"net/http"

	"github.com/bytedance/sonic"

	"github.com/devian2011/msgchute/internal/dto"
	"github.com/devian2011/msgchute/pkg/http/response"
)

type SenderResponse struct {
	Message *dto.Message `json:"message"`
	Task    *dto.Task    `json:"task"`
}

type SenderHandler interface {
	Handle(msg *dto.Message) (*dto.Message, *dto.Task, error)
}

type SenderEndpoint struct {
	h SenderHandler
}

func NewSenderEndpoint(h SenderHandler) *SenderEndpoint {
	return &SenderEndpoint{h}
}

// ServeHTTP handles the message delivery process
//	@Summary		Send message
//	@Description	Accepts a message payload, processes its delivery, and registers the tracking task.
//	@Tags			messages
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.Message			true	"Message details to be processed and sent"
//	@Success		200		{object}	SenderResponse		"Message processed and delivery task created successfully"
//	@Failure		400		{object}	response.Response	"Invalid JSON payload or structural error"
//	@Failure		500		{object}	response.Response	"Internal server error during processing or delivery dispatch"
//	@Router			/api/v1/send [post]
func (e *SenderEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var msg *dto.Message
	if decodeErr := sonic.ConfigDefault.NewDecoder(r.Body).Decode(&msg); decodeErr != nil {
		response.WriteErrorResponse(w, http.StatusBadRequest, decodeErr)
		return
	}

	msgResult, taskResult, sendErr := e.h.Handle(msg)
	if sendErr != nil {
		response.WriteErrorResponse(w, http.StatusInternalServerError, sendErr)
		return
	}

	response.WriteSuccessResponse(w, http.StatusOK, &SenderResponse{
		Message: msgResult,
		Task:    taskResult,
	})
}

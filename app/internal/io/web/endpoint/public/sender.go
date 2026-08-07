package public

import (
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"

	"github.com/devian2011/msgchute/internal/dto"
	"github.com/devian2011/msgchute/pkg/http/response"
)

type SenderMessageRequest struct {
	SenderID   string            `json:"sender_id" db:"sender_id" validate:"required"`
	Recipients dto.Recipients    `json:"recipients" db:"recipients" validate:"required"`
	Meta       dto.MessageMeta   `json:"meta" db:"meta"`
	Code       *string           `json:"code,omitempty" db:"code"`
	Params     dto.MessageParams `json:"params,omitempty" db:"params"`
	Transport  string            `json:"transport" db:"transport" validate:"required"` // Transport message provider
	Subject    string            `json:"subject" db:"subject"`
	Body       string            `json:"body" db:"body"`
	Deadline   time.Time         `json:"deadline" db:"deadline"`
	Retry      *dto.Retry        `json:"retry,omitempty" db:"retry"`
	Schedule   time.Time         `json:"schedule,omitempty" db:"schedule"`
}

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
//
//	@Summary		Send message
//	@Description	Accepts a message payload, processes its delivery, and registers the tracking task.
//	@Tags			messages
//	@Accept			json
//	@Produce		json
//	@Param			request	body		SenderMessageRequest	true	"Message details to be processed and sent"
//	@Success		200		{object}	SenderResponse		"Message processed and delivery task created successfully"
//	@Failure		400		{object}	response.Response 		"Invalid JSON payload or structural error"
//	@Failure		500		{object}	response.Response		"Internal server error during processing or delivery dispatch"
//	@Router			/api/v1/send [post]
func (e *SenderEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var msgRequest SenderMessageRequest
	if decodeErr := sonic.ConfigDefault.NewDecoder(r.Body).Decode(&msgRequest); decodeErr != nil {
		response.WriteErrorResponse(w, http.StatusBadRequest, decodeErr)
		return
	}

	msgResult, taskResult, sendErr := e.h.Handle(&dto.Message{
		ID:         uuid.UUID{},
		SenderID:   msgRequest.SenderID,
		Recipients: msgRequest.Recipients,
		Status:     "",
		Meta:       msgRequest.Meta,
		Code:       msgRequest.Code,
		Params:     msgRequest.Params,
		Transport:  msgRequest.Transport,
		Subject:    msgRequest.Subject,
		Body:       msgRequest.Body,
		Deadline:   msgRequest.Deadline,
		Retry:      msgRequest.Retry,
		Schedule:   msgRequest.Schedule,
	})
	if sendErr != nil {
		response.WriteErrorResponse(w, http.StatusInternalServerError, sendErr)
		return
	}

	response.WriteSuccessResponse(w, http.StatusOK, &SenderResponse{
		Message: msgResult,
		Task:    taskResult,
	})
}

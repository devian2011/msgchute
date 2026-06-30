package public

import (
	"errors"
	"fmt"

	"github.com/devian2011/msgchute/internal/dto"

	dto2 "github.com/devian2011/msgchute/app/internal/dto"
)

type sender struct {
	storage    TemplateStorage
	transports TransportCollection
}

func (s *sender) execute(request dto.MessageSendRequest) (interface{}, error) {
	if len(request.Messages) == 0 {
		return nil, errors.New("no messages")
	}

	type msgResponse struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}

	response := make([]msgResponse, 0, len(request.Messages))

	for _, mReq := range request.Messages {
		if len(mReq.To) == 0 {
			response = append(response, msgResponse{
				Success: false,
				Error:   "empty destinations request",
			})
			continue
		}

		msg := dto2.Message{}

		if mReq.Code != "" {
			tmpl, tmplLoadErr := s.storage.Load(mReq.Code)
			if tmplLoadErr != nil {
				response = append(response, msgResponse{
					Success: false,
					Error:   fmt.Sprintf("error on load template: %s", tmplLoadErr.Error()),
				})
				continue
			}

			var msgErr error
			msg, msgErr = buildMessage(tmpl, mReq.Params)
			if msgErr != nil {
				response = append(response, msgResponse{
					Success: false,
					Error:   fmt.Sprintf("error on build message template: %s", msgErr.Error()),
				})
				continue
			}
		} else {
			msg = *mReq.Message
		}

		sendErr := s.transports.Send(mReq.Transport, mReq.To, &msg, mReq.Meta)
		if sendErr != nil {
			response = append(response, msgResponse{
				Success: false,
				Error:   fmt.Sprintf("error on send message: %s", sendErr.Error()),
			})
		}
	}

	return response, nil
}

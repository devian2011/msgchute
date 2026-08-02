package response

import (
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/sirupsen/logrus"
)

type HandleFn func(w http.ResponseWriter, r *http.Request)

func WriteSuccessResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := NewSuccess(data)

	if encodeErr := sonic.ConfigDefault.NewEncoder(w).Encode(response); encodeErr != nil {
		logrus.WithError(encodeErr).Errorln("Failed to encode response")
	}
}

func WriteErrorResponse(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := NewError(err.Error())

	if encodeErr := sonic.ConfigDefault.NewEncoder(w).Encode(errorResponse); encodeErr != nil {
		logrus.WithError(encodeErr).Errorln("Failed to encode error response")
	}
}

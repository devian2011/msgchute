package response

import (
	"log/slog"
	"net/http"

	"github.com/bytedance/sonic"
)

func WriteErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, err error) {
	// 1. Автоматически логируем бизнес-ошибку
	slog.ErrorContext(r.Context(), "business logic error occurred",
		"error", err,
		"status", statusCode,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := NewError(err.Error())

	if encodeErr := sonic.ConfigDefault.NewEncoder(w).Encode(errorResponse); encodeErr != nil {
		slog.ErrorContext(r.Context(), "Failed to encode error response",
			"error", encodeErr,
			"status", statusCode,
		)
	}
}

func WriteSuccessResponse(w http.ResponseWriter, r *http.Request, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := NewSuccess(data)

	if encodeErr := sonic.ConfigDefault.NewEncoder(w).Encode(response); encodeErr != nil {
		slog.ErrorContext(r.Context(), "Failed to encode response",
			"error", encodeErr,
			"status", statusCode,
		)
	}
}

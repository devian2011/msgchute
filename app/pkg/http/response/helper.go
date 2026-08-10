package response

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func GetUUIDParam(key string, w http.ResponseWriter, r *http.Request) (uuid.UUID, error) {
	param := chi.URLParam(r, key)
	if len(param) == 0 {
		WriteErrorResponse(w, r, http.StatusBadRequest, fmt.Errorf("not set %s", key))
		return uuid.Nil, errors.New("not set param: " + key)
	}
	cId, parseErr := uuid.Parse(param)
	if parseErr != nil {
		slog.Error("wrong id", key, param)
		WriteErrorResponse(w, r, http.StatusBadRequest, errors.New("wrong id"))
		return uuid.Nil, errors.New("cannot parse param: " + key)
	}

	return cId, nil
}

func GetParams(keys []string, r *http.Request) map[string]string {
	result := make(map[string]string, len(keys))

	for _, k := range keys {
		param := chi.URLParam(r, k)
		if len(param) == 0 {
			continue
		}

		result[k] = param
	}

	return result
}

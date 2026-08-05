package response

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func GetUUIDParam(key string, w http.ResponseWriter, r *http.Request) (uuid.UUID, error) {
	params := mux.Vars(r)
	if _, exists := params["id"]; !exists {
		WriteErrorResponse(w, http.StatusBadRequest, errors.New("not set id"))
		return uuid.Nil, errors.New("not set param: " + key)
	}
	cId, parseErr := uuid.Parse(params["id"])
	if parseErr != nil {
		slog.Error("wrong id: %s", params["id"])
		WriteErrorResponse(w, http.StatusBadRequest, errors.New("wrong id"))
		return uuid.Nil, errors.New("cannot parse param: " + key)
	}

	return cId, nil
}

func GetParams(keys []string, r *http.Request) map[string]string {
	params := mux.Vars(r)
	result := make(map[string]string, len(keys))

	for _, k := range keys {
		if val, exists := params[k]; exists {
			result[k] = val
		}
	}

	return result
}

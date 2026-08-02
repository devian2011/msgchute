package response

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func GetUUIDParam(key string, w http.ResponseWriter, r *http.Request) (uuid.UUID, error) {
	params := mux.Vars(r)
	if _, exists := params["id"]; !exists {
		WriteErrorResponse(w, http.StatusBadRequest, errors.New("not set id"))
		return uuid.Nil, errors.New("not set param: " + key)
	}
	cId, parseErr := uuid.Parse(params["id"])
	if parseErr != nil {
		logrus.WithError(parseErr).Errorf("wrong id: %s", params["id"])
		WriteErrorResponse(w, http.StatusBadRequest, errors.New("wrong id"))
		return uuid.Nil, errors.New("cannot parse param: " + key)
	}

	return cId, nil
}

package auth

import (
	"errors"
	"net/http"

	"github.com/devian2011/msgchute/pkg/http/response"
	"github.com/devian2011/msgchute/pkg/shared/auth"
)

type HttpMiddleware struct {
	p auth.Provider
}

func NewMiddleware(p auth.Provider) *HttpMiddleware {
	return &HttpMiddleware{p: p}
}

func (m *HttpMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allow, authErr := m.p.Allow(r.Context(), r)
		if authErr != nil {
			response.WriteErrorResponse(w, r, http.StatusInternalServerError, authErr)
			return
		}
		if !allow {
			response.WriteErrorResponse(w, r, http.StatusForbidden, errors.New("forbidden"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

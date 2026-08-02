package auth

import (
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
			response.WriteErrorResponse(w, 500, authErr)
			return
		}
		if !allow {
			response.WriteSuccessResponse(w, http.StatusForbidden, "forbidden")
			return
		}

		next.ServeHTTP(w, r)
	})
}

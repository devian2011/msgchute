package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockProvider struct {
	allow bool
	err   error
}

func (m *mockProvider) Allow(ctx context.Context, r *http.Request) (bool, error) {
	return m.allow, m.err
}

func (m *mockProvider) Configure(payload []byte) error {
	return nil
}

func TestHttpMiddleware_Middleware(t *testing.T) {
	tests := []struct {
		name           string
		allow          bool
		providerErr    error
		expectedStatus int
		expectedBody   string
		shouldCallNext bool
	}{
		{
			name:           "success - allow",
			allow:          true,
			providerErr:    nil,
			expectedStatus: http.StatusOK,
			expectedBody:   "next called",
			shouldCallNext: true,
		},
		{
			name:           "forbidden",
			allow:          false,
			providerErr:    nil,
			expectedStatus: http.StatusForbidden,
			expectedBody:   "forbidden",
			shouldCallNext: false,
		},
		{
			name:           "provider error",
			allow:          false,
			providerErr:    errors.New("something went wrong"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "something went wrong",
			shouldCallNext: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockP := &mockProvider{
				allow: tt.allow,
				err:   tt.providerErr,
			}

			middleware := NewMiddleware(mockP)

			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("next called"))
			})

			handler := middleware.Middleware(nextHandler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			body := rec.Body.String()
			if tt.expectedStatus == http.StatusOK && tt.shouldCallNext {
				assert.Contains(t, body, tt.expectedBody)
			} else {
				assert.Contains(t, body, tt.expectedBody)
			}

			assert.Equal(t, tt.shouldCallNext, nextCalled)
		})
	}
}

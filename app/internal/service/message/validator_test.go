package message

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/devian2011/msgchute/internal/dto"
)

// mockProviderMapper is a stub for the ProviderMapper interface.
type mockProviderMapper struct {
	providers []string
	calls     int
}

func (m *mockProviderMapper) GetProviders() []string {
	m.calls++
	return m.providers
}

func TestNewValidator(t *testing.T) {
	mapper := &mockProviderMapper{providers: []string{"email"}}
	v := NewValidator(mapper)

	assert.NotNil(t, v)
	assert.Equal(t, mapper, v.providerMapper)
}

func TestValidator_Validate(t *testing.T) {
	tests := []struct {
		name              string
		providers         []string
		transport         string
		expectedHasErrors bool
		expectedErrorKeys []string
	}{
		{
			name:              "valid transport",
			providers:         []string{"email", "sms", "telegram"},
			transport:         "email",
			expectedHasErrors: false,
		},
		{
			name:              "valid transport — last in list",
			providers:         []string{"email", "sms", "telegram"},
			transport:         "telegram",
			expectedHasErrors: false,
		},
		{
			name:              "unknown transport",
			providers:         []string{"email", "sms"},
			transport:         "push",
			expectedHasErrors: true,
			expectedErrorKeys: []string{"transport"},
		},
		{
			name:              "empty transport",
			providers:         []string{"email"},
			transport:         "",
			expectedHasErrors: true,
			expectedErrorKeys: []string{"transport"},
		},
		{
			name:              "empty providers list",
			providers:         []string{},
			transport:         "email",
			expectedHasErrors: true,
			expectedErrorKeys: []string{"transport"},
		},
		{
			name:              "nil providers list",
			providers:         nil,
			transport:         "email",
			expectedHasErrors: true,
			expectedErrorKeys: []string{"transport"},
		},
		{
			name:              "case sensitivity — Email != email",
			providers:         []string{"email"},
			transport:         "Email",
			expectedHasErrors: true,
			expectedErrorKeys: []string{"transport"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := &mockProviderMapper{providers: tt.providers}
			v := NewValidator(mapper)

			hasErrors, errs := v.Validate(&dto.Message{Transport: tt.transport})

			assert.Equal(t, tt.expectedHasErrors, hasErrors, "hasErrors flag")
			if !tt.expectedHasErrors {
				assert.Empty(t, errs)
				return
			}
			assert.Len(t, errs, len(tt.expectedErrorKeys))
			for _, key := range tt.expectedErrorKeys {
				assert.Contains(t, errs, key)
			}
		})
	}
}

func TestValidator_Validate_ErrorMessage(t *testing.T) {
	mapper := &mockProviderMapper{providers: []string{"email"}}
	v := NewValidator(mapper)

	hasErrors, errs := v.Validate(&dto.Message{Transport: "unknown"})

	assert.True(t, hasErrors)
	assert.Contains(t, errs, "transport")
	assert.EqualError(t, errs["transport"], "unknown transport: unknown")
}

func TestValidator_Validate_CallsGetProvidersOnce(t *testing.T) {
	mapper := &mockProviderMapper{providers: []string{"email"}}
	v := NewValidator(mapper)

	v.Validate(&dto.Message{Transport: "email"})

	assert.Equal(t, 1, mapper.calls, "GetProviders must be called exactly once")
}

func TestValidator_Validate_DoesNotMutateMessage(t *testing.T) {
	mapper := &mockProviderMapper{providers: []string{"email"}}
	v := NewValidator(mapper)

	msg := &dto.Message{Transport: "unknown"}
	_, _ = v.Validate(msg)

	assert.Equal(t, "unknown", msg.Transport, "Validate must not modify the message")
}

// --- Nil message handling ---

func TestValidator_Validate_NilMessage(t *testing.T) {
	mapper := &mockProviderMapper{providers: []string{"email"}}
	v := NewValidator(mapper)

	// nil message is reported as "has errors" — consistent with len(result) > 0.
	hasErrors, errs := v.Validate(nil)

	assert.True(t, hasErrors, "true is returned for nil — has errors")
	assert.Contains(t, errs, "message")
	assert.EqualError(t, errs["message"], "empty message")
}

func TestValidator_Validate_NilMessage_DoesNotCallGetProviders(t *testing.T) {
	mapper := &mockProviderMapper{providers: []string{"email"}}
	v := NewValidator(mapper)

	v.Validate(nil)

	assert.Equal(t, 0, mapper.calls, "GetProviders must not be called for a nil message")
}

// --- Concurrency (run with -race) ---

func TestValidator_Validate_Concurrent(t *testing.T) {
	// Validator holds no mutable state and is safe for concurrent use.
	// This test guards against future regressions (e.g., a cache without a mutex).
	mapper := &mockProviderMapper{providers: []string{"email", "sms"}}
	v := NewValidator(mapper)

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()

			transport := "email"
			if i%2 == 0 {
				transport = "unknown"
			}

			hasErrors, errs := v.Validate(&dto.Message{Transport: transport})
			if transport == "email" {
				assert.False(t, hasErrors)
				assert.Empty(t, errs)
			} else {
				assert.True(t, hasErrors)
				assert.Contains(t, errs, "transport")
			}
		}(i)
	}

	wg.Wait()
}

package auth

import (
	"context"
	"net/http"
)

type EmptyAuthProvider struct{}

func (p *EmptyAuthProvider) Configure(payload []byte) error {
	return nil
}

func (p *EmptyAuthProvider) Allow(context.Context, *http.Request) (bool, error) {
	return true, nil
}

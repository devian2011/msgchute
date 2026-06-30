package internal

import (
	"context"

	"github.com/devian2011/msgchute/internal/bootstrap"
	"github.com/devian2011/msgchute/internal/registry"
)

type App struct {
	ctx context.Context
	r   *registry.AppRegistry
}

func NewApp(ctx context.Context, cfgFilePath string) (*App, error) {
	r, err := bootstrap.Boostrap(ctx, cfgFilePath)
	if err != nil {
		return nil, err
	}

	return &App{
		ctx: ctx,
		r:   r,
	}, nil
}

func (a *App) Run() error {
	errCh := make(chan error)

	select {
	case <-a.ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

package internal

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/devian2011/msgchute/internal/bootstrap"
	"github.com/devian2011/msgchute/internal/io/web/route"
	"github.com/devian2011/msgchute/internal/registry"
)

type App struct {
	ctx context.Context
	r   *registry.AppRegistry
}

func NewApp(ctx context.Context, configPath string) (*App, error) {
	r, err := bootstrap.Bootstrap(ctx, configPath)
	if err != nil {
		return nil, err
	}

	return &App{
		ctx: ctx,
		r:   r,
	}, nil
}

func (a *App) Run() error {
	errChan := make(chan error)

	initErr := a.r.Services.Sender.Init()
	if initErr != nil {
		return initErr
	}
	a.r.Services.Sender.Run()

	route.RegisterRoutes(a.r.Http, a.r.Handlers, a.r.Middlewares)
	go a.r.Http.Run(errChan)

	select {
	case <-a.ctx.Done():
		slog.Info("shutdown signal received, initiating graceful shutdown...")
		return a.shutdown()

	case err := <-errChan:
		return fmt.Errorf("immediate lifecycle shutdown due to error: %w", err)
	}
}

func (a *App) shutdown() error {
	_, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	srvShutdown, _ := context.WithTimeout(a.ctx, 2*time.Second)

	shutdownErr := a.r.Http.Shutdown(srvShutdown)
	if shutdownErr != nil {
		slog.Error("http shutdown failed due to error", slog.Any("error", shutdownErr))
	}

	a.r.Services.Sender.Shutdown()
	if a.r.AuthProvider != nil {
		a.r.AuthProvider.Close()
	}
	a.r.AuthProvider.Close()

	slog.Info("plugin sub-processes terminated successfully")
	slog.Info("all application resources released. goodbye!")
	return nil
}

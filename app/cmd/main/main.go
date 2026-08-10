package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/devian2011/msgchute/internal"
)

// @title						Notification service API
// @version					1.0
// @description				HTTP API for send any messages
// @BasePath					/
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description				Bearer token authentication
func main() {
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(jsonHandler))

	cfgFilePath := flag.String("config", "./config/config.yml", "config file path")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGABRT)
	defer stop()

	var exitCode int
	defer func() {
		if err := recover(); err != nil {
			slog.Error("critical application panic",
				"error", err,
				"stack", string(debug.Stack()),
			)
			os.Exit(1)
		}
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	}()

	slog.Info("application init")
	app, initAppErr := internal.NewApp(ctx, *cfgFilePath)
	if initAppErr != nil {
		slog.Error("error on application init", "error", initAppErr)
		exitCode = 1
		return
	}
	slog.Info("application init complete")

	slog.Info("application running...")
	if execErr := app.Run(); execErr != nil {
		slog.Error("error on application run", "error", execErr)
		exitCode = 1
		return
	}

}

package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/devian2011/msgchute/internal"
)

func main() {
	cfgFilePath := flag.String("config", "./config/config.yaml", "config file path")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGABRT)
	defer stop()

	defer func() {
		if err := recover(); err != nil {
			logrus.WithField("error", err).Errorf("critical application panic")
			os.Exit(1)
		}
	}()

	logrus.Infoln("application init")
	app, initAppErr := internal.NewApp(ctx, *cfgFilePath)
	if initAppErr != nil {
		logrus.WithField("error", initAppErr).Errorf("error on application init")
		os.Exit(1)
	}
	logrus.Infoln("application init complete")

	logrus.Infoln("application running...")
	if execErr := app.Run(); execErr != nil {
		logrus.WithField("error", execErr).Errorf("error on application run")
		os.Exit(1)
	}
}

package sender

import "log/slog"

type Logger struct{}

func (l *Logger) Infof(format string, args ...interface{}) {
	slog.Info(format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	slog.Error(format, args...)
}

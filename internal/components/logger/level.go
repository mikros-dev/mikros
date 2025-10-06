package logger

import (
	"log/slog"
)

type logLeveler struct {
	level slog.Level
}

func newLogLeveler(level slog.Level) *logLeveler {
	return &logLeveler{
		level: level,
	}
}

func (l *logLeveler) Level() slog.Level {
	return l.level
}

func (l *logLeveler) setLevel(level slog.Level) {
	l.level = level
}

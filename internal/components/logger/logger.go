package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
	logger_fields "github.com/mikros-dev/mikros/components/logger"
)

const (
	levelFatal    = slog.Level(12)
	levelInternal = slog.Level(-2)
	fatalExitCode = 1
	loggerPkgHint = "/internal/components/logger"
)

var levelNames = map[slog.Leveler]string{
	levelFatal:    "FATAL",
	levelInternal: "INTERNAL",
}

type ErrorStackTraceMode string

const (
	ErrorStackTraceModeDisabled   ErrorStackTraceMode = "disabled"
	ErrorStackTraceModeDefault    ErrorStackTraceMode = "default"
	ErrorStackTraceModeStructured ErrorStackTraceMode = "structured"
)

type (
	// ContextFieldExtractor is a function that receives a context and should
	// return a slice of logger_api.Attribute to be added into every log call.
	ContextFieldExtractor func(ctx context.Context) []logger_api.Attribute
)

type Logger struct {
	errorStacktrace ErrorStackTraceMode
	logger          *slog.Logger
	errorLogger     *slog.Logger
	level           *logLeveler
	fieldExtractor  ContextFieldExtractor
}

type Options struct {
	TextOutput      bool
	DiscardMessages bool
	ErrorStacktrace string
	FixedAttributes map[string]string
}

// New creates a new Logger interface for applications.
func New(options Options) *Logger {
	var (
		attrs []slog.Attr
		level = newLogLeveler(slog.LevelInfo)
		opts  = &slog.HandlerOptions{
			Level: level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				// Prints our custom log level label.
				if a.Key == slog.LevelKey {
					if level, ok := a.Value.Any().(slog.Level); ok {
						levelLabel, exists := levelNames[level]
						if !exists {
							levelLabel = level.String()
						}

						a.Value = slog.StringValue(levelLabel)
					}
				}

				// Change the source path to only 'dir/file.go'
				if a.Key == slog.SourceKey {
					if source, ok := a.Value.Any().(*slog.Source); ok {
						filename := filepath.Base(source.File)
						source.File = filepath.Join(filepath.Base(filepath.Dir(source.File)), filename)
					}
				}

				return a
			},
		}
	)

	// Adds custom fixed attributes into every log message.
	for k, v := range options.FixedAttributes {
		attrs = append(attrs, slog.String(k, v))
	}

	logHandler := slog.NewJSONHandler(os.Stdout, opts).WithAttrs(attrs)
	if options.TextOutput {
		logHandler = slog.NewTextHandler(os.Stdout, opts).WithAttrs(attrs)
	}

	// Creates a specific log handler so every error message can have its source
	// in the output.
	opts.AddSource = false
	errHandler := slog.NewJSONHandler(os.Stderr, opts).WithAttrs(attrs)
	if options.TextOutput {
		errHandler = slog.NewTextHandler(os.Stderr, opts).WithAttrs(attrs)
	}

	// Create our handlers
	l := slog.New(logHandler)
	e := slog.New(errHandler)

	if options.DiscardMessages {
		l = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
		e = l
	}

	return &Logger{
		errorStacktrace: ErrorStackTraceMode(options.ErrorStacktrace),
		logger:          l,
		errorLogger:     e,
		level:           level,
	}
}

// Debug outputs messages using debug level.
func (l *Logger) Debug(ctx context.Context, msg string, attrs ...logger_api.Attribute) {
	mFields := l.mergeFieldsWithCtx(ctx, attrs)
	l.logger.Debug(msg, mFields...)
}

// Info outputs messages using the info level.
func (l *Logger) Info(ctx context.Context, msg string, attrs ...logger_api.Attribute) {
	mFields := l.mergeFieldsWithCtx(ctx, attrs)
	l.logger.Info(msg, mFields...)
}

// Warn outputs messages using warning level.
func (l *Logger) Warn(ctx context.Context, msg string, attrs ...logger_api.Attribute) {
	mFields := l.mergeFieldsWithCtx(ctx, attrs)
	l.logger.Warn(msg, mFields...)
}

// Error outputs messages using error level.
func (l *Logger) Error(ctx context.Context, msg string, attrs ...logger_api.Attribute) {
	l.error(ctx, msg, attrs...)
}

func (l *Logger) error(ctx context.Context, msg string, attrs ...logger_api.Attribute) {
	var (
		mFields = l.mergeFieldsWithCtx(ctx, attrs)
		pc      uintptr
	)

	if l.level.Level() > slog.LevelError {
		return
	}

	fr, idx, ok := pickCallerFrame(2)
	if ok {
		pc = fr.PC
	}

	r := slog.NewRecord(time.Now(), slog.LevelError, msg, pc)

	if len(mFields) > 0 {
		r.Add(mFields...)
	}

	if ok {
		var (
			funcName = fr.Function
			file     = fr.File
			fileBase = filepath.Base(file)
		)

		file = filepath.Join(filepath.Base(filepath.Dir(file)), fileBase)

		r.AddAttrs(slog.Any(slog.SourceKey, &slog.Source{
			Function: funcName,
			File:     file,
			Line:     fr.Line,
		}))
	}

	l.printErrorStackTrace(&r, 2+idx)

	if err := l.errorLogger.Handler().Handle(ctx, r); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error logging error: %v\n", err)
	}
}

func (l *Logger) printErrorStackTrace(record *slog.Record, skip int) {
	if l.errorStacktrace == ErrorStackTraceModeDisabled || l.errorStacktrace == "" {
		return
	}

	if l.errorStacktrace == ErrorStackTraceModeDefault {
		_, _ = fmt.Print(takeStacktrace(skip))
	}

	record.AddAttrs(slog.String("stack", takeStacktrace(skip)))
}

func pickCallerFrame(startSkip int) (runtime.Frame, int, bool) {
	var (
		pcs [32]uintptr
		n   = runtime.Callers(startSkip, pcs[:])
	)

	if n == 0 {
		return runtime.Frame{}, 0, false
	}

	var (
		idx         = 0
		frames      = runtime.CallersFrames(pcs[:n])
		isLogMethod = func(name string) bool {
			switch name {
			case "Debug", "Debugf", "Debugw",
				"Info", "Infof", "Infow",
				"Warn", "Warnf", "Warnw",
				"Error", "Errorf", "Errorw",
				"Fatal", "Fatalf", "Fatalw":
				return true
			}

			return false
		}
	)

	for {
		fr, more := frames.Next()
		full := fr.Function

		// skip frames from the logger package itself
		if strings.Contains(full, loggerPkgHint) {
			idx++

			if !more {
				break
			}

			continue
		}

		// skip wrapper frames that implement logger-like methods, if any
		name := full
		if i := strings.LastIndex(full, "."); i >= 0 && i < len(full)-1 {
			name = full[i+1:]
		}

		if isLogMethod(name) {
			idx++

			if !more {
				break
			}

			continue
		}

		return fr, idx, true
	}

	return runtime.Frame{}, idx, false
}

// Fatal outputs message using fatal level.
func (l *Logger) Fatal(ctx context.Context, msg string, attrs ...logger_api.Attribute) {
	mFields := l.mergeFieldsWithCtx(ctx, attrs)
	l.logger.Log(ctx, levelFatal, msg, mFields...)
	os.Exit(fatalExitCode)
}

func (l *Logger) mergeFieldsWithCtx(ctx context.Context, attrs []logger_api.Attribute) []any {
	var (
		appendedFields = l.appendServiceContext(ctx, attrs)
		mergedFields   = make([]any, len(appendedFields))
	)

	for i, field := range appendedFields {
		mergedFields[i] = slog.Any(field.Key(), field.Value())
	}

	return mergedFields
}

// DisableDebugMessages is a helper method to disable Debug level messages.
func (l *Logger) DisableDebugMessages() {
	l.level.setLevel(slog.LevelInfo)
}

// appendServiceContext executes a custom field extractor from the current
// context to add more fields into the message.
func (l *Logger) appendServiceContext(ctx context.Context, attrs []logger_api.Attribute) []logger_api.Attribute {
	if l.fieldExtractor != nil {
		attrs = append(attrs, l.fieldExtractor(ctx)...)
	}

	return attrs
}

// SetLogLevel changes the current messages log level.
func (l *Logger) SetLogLevel(level string) (string, error) {
	var newLevel slog.Level

	switch strings.ToLower(level) {
	case "debug":
		newLevel = slog.LevelDebug
	case "info":
		newLevel = slog.LevelInfo
	case "warn":
		newLevel = slog.LevelWarn
	case "error":
		newLevel = slog.LevelError
	case "fatal":
		newLevel = levelFatal
	case "internal":
		newLevel = levelInternal
	default:
		return "", fmt.Errorf("unknown log level '%v'", level)
	}

	l.level.setLevel(newLevel)
	return level, nil
}

// Level gets the current log level.
func (l *Logger) Level() string {
	switch l.level.Level() {
	case slog.LevelDebug:
		return "debug"
	case slog.LevelInfo:
		return "info"
	case slog.LevelWarn:
		return "warn"
	case slog.LevelError:
		return "error"
	case levelFatal:
		return "fatal"
	case levelInternal:
		return "internal"
	}

	return "unknown"
}

// SetErrorStacktrace lets one choose the runtime stacktrace format that
// error messages can show.
func (l *Logger) SetErrorStacktrace(mode ErrorStackTraceMode) {
	l.errorStacktrace = mode
}

// SetContextFieldExtractor adds a custom function to extract values from the
// context and add them into the log messages.
func (l *Logger) SetContextFieldExtractor(extractor ContextFieldExtractor) {
	l.fieldExtractor = extractor
}

func (l *Logger) Debugf(ctx context.Context, msg string, attrs ...map[string]interface{}) {
	var loggerFields []logger_api.Attribute
	if len(attrs) > 0 {
		for k, v := range attrs[0] {
			loggerFields = append(loggerFields, logger_fields.Any(k, v))
		}
	}

	l.Debug(ctx, msg, loggerFields...)
}

func (l *Logger) Infof(ctx context.Context, msg string, attrs ...map[string]interface{}) {
	var loggerFields []logger_api.Attribute
	if len(attrs) > 0 {
		for k, v := range attrs[0] {
			loggerFields = append(loggerFields, logger_fields.Any(k, v))
		}
	}

	l.Info(ctx, msg, loggerFields...)
}

func (l *Logger) Warnf(ctx context.Context, msg string, attrs ...map[string]interface{}) {
	var loggerFields []logger_api.Attribute
	if len(attrs) > 0 {
		for k, v := range attrs[0] {
			loggerFields = append(loggerFields, logger_fields.Any(k, v))
		}
	}

	l.Warn(ctx, msg, loggerFields...)
}

func (l *Logger) Errorf(ctx context.Context, msg string, attrs ...map[string]interface{}) {
	var loggerFields []logger_api.Attribute
	if len(attrs) > 0 {
		for k, v := range attrs[0] {
			loggerFields = append(loggerFields, logger_fields.Any(k, v))
		}
	}

	l.Error(ctx, msg, loggerFields...)
}

func (l *Logger) Fatalf(ctx context.Context, msg string, attrs ...map[string]interface{}) {
	var loggerFields []logger_api.Attribute
	if len(attrs) > 0 {
		for k, v := range attrs[0] {
			loggerFields = append(loggerFields, logger_fields.Any(k, v))
		}
	}

	l.Fatal(ctx, msg, loggerFields...)
}

func (l *Logger) InnerLogger() *slog.Logger {
	return l.logger
}

// Internal outputs messages using the internal level.
func (l *Logger) Internal(ctx context.Context, msg string, attrs ...logger_api.Attribute) {
	mFields := l.mergeFieldsWithCtx(ctx, attrs)
	l.logger.Log(ctx, levelInternal, msg, mFields...)
}

func (l *Logger) Internalf(ctx context.Context, msg string, attrs ...map[string]interface{}) {
	var loggerFields []logger_api.Attribute
	if len(attrs) > 0 {
		for k, v := range attrs[0] {
			loggerFields = append(loggerFields, logger_fields.Any(k, v))
		}
	}

	l.Internal(ctx, msg, loggerFields...)
}

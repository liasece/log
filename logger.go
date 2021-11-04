package log

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/liasece/log/encoder"
	logsentry "github.com/liasece/log/sentry"
	"go.elastic.co/apm"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	_globalL *zap.Logger
)

const (
	_traceIDKey = "trace.traceid"
	_spanIDKey  = "trace.spanid"
)

// L return global logger
func L(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return _globalL
	}

	span := apm.SpanFromContext(ctx)

	if span != nil {
		return withTraceContext(span.TraceContext())
	}

	tx := apm.TransactionFromContext(ctx)
	if tx != nil {
		return withTraceContext(tx.TraceContext())
	}

	return _globalL
}

func withTraceContext(tc apm.TraceContext) *zap.Logger {
	return _globalL.With(
		zap.String(_traceIDKey, tc.Trace.String()),
		zap.String(_spanIDKey, tc.Span.String()),
	)
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Debug(msg string, fields ...zap.Field) {
	_globalL.Debug(msg, fields...)
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Info(msg string, fields ...zap.Field) {
	_globalL.Info(msg, fields...)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Warn(msg string, fields ...zap.Field) {
	_globalL.Warn(msg, fields...)
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Error(msg string, fields ...zap.Field) {
	_globalL.Error(msg, fields...)
}

// DPanic logs a message at DPanicLevel. The message includes any fields
// passed at the log site, as well as any fields accumulated on the logger.
//
// If the logger is in development mode, it then panics (DPanic means
// "development panic"). This is useful for catching errors that are
// recoverable, but shouldn't ever happen.
func DPanic(msg string, fields ...zap.Field) {
	_globalL.DPanic(msg, fields...)
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func Panic(msg string, fields ...zap.Field) {
	_globalL.Panic(msg, fields...)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func Fatal(msg string, fields ...zap.Field) {
	_globalL.Fatal(msg, fields...)
}

// With creates a child logger and adds structured context to it. Fields added
// to the child don't affect the parent, and vice versa.
func With(fields ...zap.Field) *zap.Logger {
	return _globalL.With(fields...)
}

// Sync flushes buffered logs (if any).
func Sync() error {
	return _globalL.Core().Sync()
}

func getZapLevelEnablerFunc(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	case "panic":
		return zapcore.PanicLevel
	case "info":
		return zapcore.InfoLevel
	default:
		return zapcore.DebugLevel
	}
}

func getConsoleCore(colorLevel bool, level string) (zapcore.Core, error) {
	consoleEncoder := zap.NewProductionEncoderConfig()
	consoleEncoder.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[")
		enc.AppendString(t.Format("2006-01-02T15:04:05.000"))
		enc.AppendString("]")
	}
	consoleEncoder.EncodeLevel = encoder.MyLevelEncoder
	consoleEncoder.EncodeCaller = encoder.MyCallerEncode

	if colorLevel {
		consoleEncoder.EncodeLevel = encoder.MyColorLevelEncoder
	}

	return zapcore.NewCore(encoder.NewConsoleEncoder(consoleEncoder), zapcore.Lock(os.Stdout), getZapLevelEnablerFunc(level)), nil
}

func getSentryCore(options sentry.ClientOptions) (zapcore.Core, error) {
	cfg := logsentry.Configuration{
		Level: zapcore.ErrorLevel, //when to send message to sentry
		Tags: map[string]string{
			"component": "system",
		},
		FlushTimeout: time.Second * 5,
	}
	core, err := logsentry.NewCore(cfg, logsentry.NewSentryClientFromOptions(options))
	//in case of err it will return noop core. so we can safely attach it
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to init zap [%v]", err)
	}
	return core, err
}

func initZapLogger(level string) (err error) {
	consoleCore, err := getConsoleCore(true, level)
	if err != nil {
		return
	}

	_globalL = zap.New(zapcore.NewTee(consoleCore), zap.AddCaller(), zap.AddCallerSkip(1))
	return
}

func init() {
	consoleCore, err := getConsoleCore(true, "debug")
	encoder.CheckIfTerminal(os.Stdout)
	if err == nil {
		_globalL = zap.New(zapcore.NewTee(consoleCore), zap.AddCaller(), zap.AddCallerSkip(1))
	} else {
		panic(err)
	}
}

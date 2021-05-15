package log

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	logsentry "github.com/liasece/log/sentry"
	"github.com/spf13/viper"
)

// InitLog Init logging
func InitLog(fileName string, cfg *viper.Viper) error {
	return initZapLogger(cfg.GetString("logging.level"))
}

// InitSentry initialize sentry client and log sentry hook
func InitSentry(options sentry.ClientOptions) {
	sentrycore, err := getSentryCore(options)
	if err != nil || _globalL == nil {
		Panic("initialize sentry failed", NamedError("init sentry", err))
	} else {
		_globalL = logsentry.AttachCoreToLogger(sentrycore, _globalL)
		Info("initialize sentry ok")
	}
}

// isPanicFromLogger check the goroutine's "skip+2" number of stack frames is zap@v1.10.0/zapcore/entry.go:229
// Where "+2" is derived from isPanicFromLogger which can determine at least the following callers need to be popped:
// 1.  isPanicFromLogger()
// 2.  zapcore/entry.go:229 panic(msg)
// The "skip" originates from the number of stack layers of the caller of isPanicFromLogger itself after defer:
// 1.  RecoverWithSentry()
// so, In this code use isPanicFromLogger(1)
func isPanicFromLogger(skip int) bool {
	_, file, _, _ := runtime.Caller(skip + 2)
	return strings.Contains(file, "/zapcore/")
}

// RecoverWithSentry capture panic and send to sentry
func RecoverWithSentry(rethrow bool) {
	err := recover()
	if err == nil {
		return
	}
	{
		Error("RecoverWithSentry panic", Reflect("err", err), Reflect("stack", string(debug.Stack())))
	}
	hub := sentry.CurrentHub()
	if isPanicFromLogger(1) {
		// send to sentry by log hook, skip here
		// event := CreateSentryEvent(entry)
		// hub.CaptureEvent(event)
		return
	}
	event := hub.Recover(err)
	if event == nil {
		// unknown panic type, convert to string
		hub.Recover(fmt.Sprintf("%+v", err))
	}
	// rethrow raw panic
	// maybe we could do better here: keep the origin panic stacktrace
	if rethrow {
		panic(err)
	}
}

// FlushSentry flush sentry, call before process exit
func FlushSentry() {
	sentry.Flush(time.Second * 5)
}

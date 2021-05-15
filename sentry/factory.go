package logsentry

import (
	"fmt"

	"github.com/getsentry/sentry-go"
)

// NewSentryClientFromOptions create a sentry factory with sentry.ClientOptions
func NewSentryClientFromOptions(options sentry.ClientOptions) SentryClientFactory {
	err := sentry.Init(options)
	if err != nil {
		fmt.Printf("entry.Init error %v\n", err)
	}
	return func() (*sentry.Client, error) {
		return sentry.CurrentHub().Client(), nil
	}
}

// NewSentryClientFromClient new a sentry client function with sentry.Client
func NewSentryClientFromClient(client *sentry.Client) SentryClientFactory {
	return func() (*sentry.Client, error) {
		return client, nil
	}
}

// SentryClientFactory is a type of sentry factory returned function
type SentryClientFactory func() (*sentry.Client, error)

package tools

import (
	"context"
	"sync"
)

// Logger implementations should send internal errors to stderr

type LoggerProvider interface {
	Start(stop context.Context, await *sync.WaitGroup) error
	New(source string) LoggerInstance
}

type LoggerInstance interface {
	Warn(message string, data any)
	Info(message string, data any)
	Debug(message string, data any)
	Error(message string, data any)
	Fatal(message string, data any)
}

var Logger LoggerProvider

func SetupLogger(stop context.Context, await *sync.WaitGroup) {
	switch LOGGER_PROVIDER {
	case "console":
		Logger = &loggerProviderConsole{}
	default:
		panic("unknown logger provider: " + LOGGER_PROVIDER)
	}
	if err := Logger.Start(stop, await); err != nil {
		panic("logger startup failed: " + err.Error())
	}
}

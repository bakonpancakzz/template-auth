package tools

import (
	"context"
	"os"
	"sync"
	"time"
)

var (
	LoggerMain        = NewLoggerInstance("main")
	LoggerHttp        = NewLoggerInstance("http")
	LoggerRatelimit   = NewLoggerInstance("ratelimit")
	LoggerStorage     = NewLoggerInstance("storage")
	LoggerGeolocation = NewLoggerInstance("geolocation")
	LoggerDatabase    = NewLoggerInstance("database")
	LoggerEmail       = NewLoggerInstance("email")
	LoggerLogger      = NewLoggerInstance("logger")
)

type LoggerProvider interface {
	Start(stop context.Context, await *sync.WaitGroup) error
	Entry(level, source, message string, data any)
}

var Logger LoggerProvider = &loggerProviderConsole{}

func SetupLogger(stop context.Context, await *sync.WaitGroup) {
	t := time.Now()

	switch LOGGER_PROVIDER {
	case "console":
		// enabled by fallback
	default:
		panic("unknown logger provider: " + LOGGER_PROVIDER)
	}

	if err := Logger.Start(stop, await); err != nil {
		LoggerLogger.Fatal("Startup Failed", err.Error())
	}
	LoggerLogger.Info("Ready", map[string]any{
		"time": time.Since(t).String(),
	})
}

// Helper Function to Create a new Logger Instance which includes Standardized Functions
func NewLoggerInstance(source string) *LoggerInstance {
	return &LoggerInstance{source: source}
}

type LoggerInstance struct {
	source string
}

func (p *LoggerInstance) log(level, message string, data any) {
	Logger.Entry(level, p.source, message, data)
}

func (p *LoggerInstance) Info(message string, data any) {
	p.log("INFO", message, data)
}

func (p *LoggerInstance) Warn(message string, data any) {
	p.log("WARN", message, data)
}

func (p *LoggerInstance) Debug(message string, data any) {
	p.log("DEBUG", message, data)
}

func (p *LoggerInstance) Error(message string, data any) {
	p.log("ERROR", message, data)
}

func (p *LoggerInstance) Fatal(message string, data any) {
	p.log("FATAL", message, data)
	os.Exit(1)
}

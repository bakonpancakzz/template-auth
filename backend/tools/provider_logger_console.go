package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type loggerProviderConsole struct{}

type loggerProviderConsoleInstance struct {
	source string
}

func (p *loggerProviderConsole) Start(stop context.Context, await *sync.WaitGroup) error {
	return nil // no background tasks needed
}

func (p *loggerProviderConsole) New(source string) LoggerInstance {
	return &loggerProviderConsoleInstance{source: source}
}

func (p *loggerProviderConsoleInstance) log(level, message string, data any) {

	target := os.Stdout
	if level == "ERROR" || level == "FATAL" {
		target = os.Stderr
	}

	entryData := "null"
	if data != nil {
		if b, err := json.Marshal(data); err != nil {
			entryData = fmt.Sprintf("marshal_error:%q", err)
		} else {
			entryData = string(b)
		}
	}

	entryTime := time.Now().Format(time.RFC3339)
	fmt.Fprintf(
		target,
		"%s level=%s source=%s message=%q data=%s\n",
		entryTime, level, p.source, message, entryData,
	)
}

func (p *loggerProviderConsoleInstance) Info(message string, data any) {
	p.log("INFO", message, data)
}

func (p *loggerProviderConsoleInstance) Warn(message string, data any) {
	p.log("WARN", message, data)
}

func (p *loggerProviderConsoleInstance) Debug(message string, data any) {
	p.log("DEBUG", message, data)
}

func (p *loggerProviderConsoleInstance) Error(message string, data any) {
	p.log("ERROR", message, data)
}

func (p *loggerProviderConsoleInstance) Fatal(message string, data any) {
	p.log("FATAL", message, data)
	os.Exit(1)
}

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

func (p *loggerProviderConsole) Start(stop context.Context, await *sync.WaitGroup) error {
	// this should always be empty
	return nil
}

func (p *loggerProviderConsole) Entry(level, source, message string, data any) {
	entryData := "null"
	if data != nil {
		if b, err := json.Marshal(data); err != nil {
			entryData = fmt.Sprintf("marshal_error:%q", err)
		} else {
			entryData = string(b)
		}
	}
	target := os.Stdout
	if level == "ERROR" || level == "FATAL" {
		target = os.Stderr
	}
	fmt.Fprintf(target,
		"%s level=%s source=%s message=%q data=%s\n",
		time.Now().Format(time.RFC3339), level, source, message, entryData,
	)
}

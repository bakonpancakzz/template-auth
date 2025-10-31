package tools

import (
	"context"
	"sync"
	"time"
)

type RatelimitProvider interface {
	// Interface should mimic the following redis behavior:
	// 	- TTL Non-existing key will equal zero duration
	// 	- Decrement Non-existing key will equal -1
	// 	- Increment Non-existing key will equal 1
	// 	- Increment Existing Key will +1

	Start(stop context.Context, await *sync.WaitGroup) error
	Increment(key string, period time.Duration) (int64, error)
	Decrement(key string) (int64, error)
	TTL(key string) (time.Duration, error)
}

var Ratelimit RatelimitProvider

func SetupRatelimitProvider(stop context.Context, await *sync.WaitGroup) {
	t := time.Now()

	switch RATELIMIT_PROVIDER {
	case "local":
		Ratelimit = &ratelimitProviderLocal{}
	case "redis":
		Ratelimit = &ratelimitProviderRedis{}
	default:
		LoggerRatelimit.Fatal("Unknown Provider", RATELIMIT_PROVIDER)
	}

	if err := Ratelimit.Start(stop, await); err != nil {
		LoggerRatelimit.Fatal("Startup Failed", err.Error())
	}
	LoggerRatelimit.Info("Ready", map[string]any{
		"time": time.Since(t).String(),
	})
}

package tools

import (
	"context"
	"sync"
	"time"
)

type rateLimitProviderNone struct {
}

func (p *rateLimitProviderNone) Start(stop context.Context, await *sync.WaitGroup) error {
	return nil
}

func (p *rateLimitProviderNone) Increment(key string, period time.Duration) (int64, error) {
	return 0, nil
}

func (p *rateLimitProviderNone) Decrement(key string) (int64, error) {
	return 0, nil
}

func (p *rateLimitProviderNone) TTL(key string) (time.Duration, error) {
	return 0, nil
}

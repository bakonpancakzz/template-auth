package tools

import (
	"context"
	"sync"
)

// Debug or Dummy Email Provider

type emailProviderNone struct{}

func (e *emailProviderNone) Start(stop context.Context, await *sync.WaitGroup) error {
	return nil
}

func (o *emailProviderNone) Send(toAddress, subject, html string) error {
	return nil
}

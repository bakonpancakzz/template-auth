package tools

import (
	"context"
	"sync"
	"time"
)

type StorageProvider interface {
	Start(stop context.Context, await *sync.WaitGroup) error
	Put(key, contentType string, data []byte) error
	Delete(keys ...string) error
}

var Storage StorageProvider

func SetupStorageProvider(stop context.Context, await *sync.WaitGroup) {
	t := time.Now()

	switch STORAGE_PROVIDER {
	case "none":
		Storage = &storageProviderNone{}
	case "disk":
		Storage = &storageProviderDisk{}
	case "s3":
		Storage = &storageProviderS3{}
	default:
		LoggerStorage.Fatal("Unknown Provider", STORAGE_PROVIDER)
	}

	if err := Storage.Start(stop, await); err != nil {
		LoggerStorage.Fatal("Startup Failed", err)
	}
	LoggerStorage.Info("Ready", map[string]any{
		"time": time.Since(t).String(),
	})
}

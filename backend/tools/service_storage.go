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
var StorageLogger LoggerInstance

func SetupStorageProvider(stop context.Context, await *sync.WaitGroup) {
	StorageLogger = Logger.New("storage")
	t := time.Now()

	switch STORAGE_PROVIDER {
	case "none":
		Storage = &storageProviderNone{}
	case "disk":
		Storage = &storageProviderDisk{}
	case "s3":
		Storage = &storageProviderS3{}
	default:
		StorageLogger.Fatal("Unknown Provider", STORAGE_PROVIDER)
	}

	if err := Storage.Start(stop, await); err != nil {
		StorageLogger.Fatal("Startup Failed", err)
	}
	StorageLogger.Info("Ready", map[string]any{
		"time": time.Since(t).String(),
	})
}

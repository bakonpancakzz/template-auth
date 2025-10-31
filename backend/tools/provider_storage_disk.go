package tools

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
)

// NOTE: Mainly for testing or small scales

type storageProviderDisk struct {
	Base string
	Mode os.FileMode
}

func (o *storageProviderDisk) Start(stop context.Context, await *sync.WaitGroup) error {
	o.Base = STORAGE_DISK_DIRECTORY
	o.Mode = os.FileMode(STORAGE_DISK_PERMISSIONS)
	return os.MkdirAll(o.Base, o.Mode)
}

func (o *storageProviderDisk) Put(key, contentType string, data []byte) error {
	full := path.Join(o.Base, path.Clean(key))
	return os.WriteFile(full, data, o.Mode)
}

func (o *storageProviderDisk) Delete(keys ...string) error {
	var errs []string
	for _, k := range keys {
		full := path.Join(o.Base, path.Clean(k))
		if err := os.Remove(full); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("fs errors:\n%s", strings.Join(errs, "\n"))
	}
	return nil
}

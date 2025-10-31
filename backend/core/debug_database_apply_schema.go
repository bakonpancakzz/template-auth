package core

import (
	"context"
	"os"
	"sync"

	"github.com/bakonpancakzz/template-auth/include"
	"github.com/bakonpancakzz/template-auth/tools"
)

func DebugDatabaseApplySchema() {
	var stopCtx, stop = context.WithCancel(context.Background())
	var stopWg sync.WaitGroup

	tools.SetupLogger(stopCtx, &stopWg)
	tools.SetupDatabase(stopCtx, &stopWg)
	_, err := tools.Database.Exec(context.Background(), include.DatabaseSchema)
	if err != nil {
		panic("failed to apply schema: " + err.Error())
	}
	tools.DatabaseLogger.Info("Schema Applied", nil)

	stop()
	stopWg.Wait()
	os.Exit(0)
}

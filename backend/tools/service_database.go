package tools

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Database *pgxpool.Pool
var DatabaseLogger LoggerInstance

func SetupDatabase(stop context.Context, await *sync.WaitGroup) {
	DatabaseLogger = Logger.New("database")
	t := time.Now()

	// Setup Connection Pool
	ctx, cancel := NewContext()
	defer cancel()
	var err error

	if Database, err = pgxpool.New(ctx, DATABASE_URL); err != nil {
		DatabaseLogger.Fatal("Failed to create pool", err)
	}
	if err = Database.Ping(ctx); err != nil {
		DatabaseLogger.Fatal("Failed to ping database", err)
	}

	// Shutdown Logic
	await.Add(1)
	go func() {
		defer await.Done()
		<-stop.Done()
		Database.Close()
		DatabaseLogger.Info("Closed", nil)
	}()
	DatabaseLogger.Info("Ready", map[string]any{
		"time": time.Since(t).String(),
	})
}

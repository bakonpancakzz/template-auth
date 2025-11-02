package tools

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Database *pgxpool.Pool

func SetupDatabase(stop context.Context, await *sync.WaitGroup) {
	t := time.Now()

	// Setup Connection Pool
	ctx, cancel := NewContext()
	defer cancel()
	var err error

	// Parse URL and TLS Configuration
	cfg, err := pgxpool.ParseConfig(DATABASE_URL)
	if err != nil {
		LoggerDatabase.Fatal("Invalid Database URI", err.Error())
	}
	if DATABASE_TLS_ENABLED {
		tls, err := NewTLSConfig(
			DATABASE_TLS_CERT,
			DATABASE_TLS_KEY,
			DATABASE_TLS_CA,
		)
		if err != nil {
			LoggerDatabase.Fatal("Failed to parse tls:", err.Error())
		}
		cfg.ConnConfig.TLSConfig = tls
	}

	// Create and Test Client
	if Database, err = pgxpool.NewWithConfig(ctx, cfg); err != nil {
		LoggerDatabase.Fatal("Failed to create pool", err.Error())
	}
	if err = Database.Ping(ctx); err != nil {
		LoggerDatabase.Fatal("Failed to ping database", err.Error())
	}

	// Shutdown Logic
	await.Add(1)
	go func() {
		defer await.Done()
		<-stop.Done()
		Database.Close()
		LoggerDatabase.Info("Closed", nil)
	}()
	LoggerDatabase.Info("Ready", map[string]any{
		"time": time.Since(t).String(),
	})
}

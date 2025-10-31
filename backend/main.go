package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bakonpancakzz/template-auth/core"
	"github.com/bakonpancakzz/template-auth/tools"
)

func main() {
	time.Local = time.UTC

	// Startup Services
	// 	Logger are unique and must be started specifically,
	// 	everything else can be started at the same time
	var stopCtx, stop = context.WithCancel(context.Background())
	var stopWg sync.WaitGroup
	var syncWg sync.WaitGroup

	tools.LoggerMain.Info("Starting Services", nil)
	for _, fn := range []func(stop context.Context, await *sync.WaitGroup){
		tools.SetupLogger,
		tools.SetupDatabase,
		tools.SetupGeolocation,
		tools.SetupEmailProvider,
		tools.SetupRatelimitProvider,
		tools.SetupStorageProvider,
	} {
		syncWg.Add(1)
		go func() {
			defer syncWg.Done()
			fn(stopCtx, &stopWg)
		}()
	}
	syncWg.Wait()
	go StartupHTTP(stopCtx, &stopWg)

	// Await Shutdown Signal
	cancel := make(chan os.Signal, 1)
	signal.Notify(cancel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-cancel
	stop()

	// Begin Shutdown Process
	timeout, finish := context.WithTimeout(context.Background(), time.Minute)
	defer finish()
	go func() {
		<-timeout.Done()
		if timeout.Err() == context.DeadlineExceeded {
			tools.LoggerMain.Fatal("Shutdown Deadline Exceeded", nil)
		}
	}()
	stopWg.Wait()
	os.Exit(0)
}

func StartupHTTP(stop context.Context, await *sync.WaitGroup) {

	// Optimized to prevent malicious attacks but shouldn't
	// really bother devices on slower networks :)

	svr := http.Server{
		Handler:           core.SetupMux(),
		Addr:              tools.HTTP_ADDRESS,
		MaxHeaderBytes:    4096,
		IdleTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadTimeout:       10 * time.Second,
	}
	if tools.HTTP_TLS_ENABLED {
		tls, err := tools.NewTLSConfig(
			tools.HTTP_TLS_CERT,
			tools.HTTP_TLS_KEY,
			tools.HTTP_TLS_CA,
		)
		if err != nil {
			tools.LoggerHttp.Fatal("TLS Configuration Error", err)
			return
		}
		svr.TLSConfig = tls
	}

	// Shutdown Logic
	await.Add(1)
	go func() {
		defer await.Done()
		<-stop.Done()
		svr.Shutdown(context.Background())
		tools.LoggerHttp.Info("Server Closed", nil)
	}()

	// Server Startup
	var err error
	tools.LoggerHttp.Info("Listening", svr.Addr)
	if tools.HTTP_TLS_ENABLED {
		err = svr.ListenAndServeTLS("", "")
	} else {
		err = svr.ListenAndServe()
	}
	if err != http.ErrServerClosed {
		tools.LoggerHttp.Fatal("Startup Failed", err)
	}
}

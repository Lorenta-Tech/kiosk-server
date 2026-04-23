package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Lorenta-Tech/kiosk-server/internal/app"
	"github.com/Lorenta-Tech/kiosk-server/internal/env"
	"github.com/Lorenta-Tech/kiosk-server/internal/routes"
)

func main() {
	port := env.GetInt("PORT", 17069)

	application, err := app.NewApplication()

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize application:%v\n", err)
		os.Exit(1)
	}
	defer application.DB.Close()

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           routes.SetupRoutes(application),
		IdleTimeout:       time.Minute,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	serverErr := make(chan error, 1)

	go func() {
		application.Logger.Info("server listening", "port", port)
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err = <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			application.Logger.Error("server error", "error", err)
			os.Exit(1)
		}
	case sig := <-quit:
		application.Logger.Info("shutting down", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
		defer cancel()

		if err = server.Shutdown(ctx); err != nil {
			application.Logger.Error("gracceful shutdown failed", "error", err)
		} else {
			application.Logger.Info("server stopped cleanly")
		}
	}
}

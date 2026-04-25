package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/tobiasheinloth/CoffeeOracle/backend/internal/config"
	"github.com/tobiasheinloth/CoffeeOracle/backend/internal/logger"
	"github.com/tobiasheinloth/CoffeeOracle/backend/internal/oracle"
	"github.com/tobiasheinloth/CoffeeOracle/backend/internal/server"
	"github.com/tobiasheinloth/CoffeeOracle/backend/internal/server/handlers"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	logger.Configure(cfg.LogLevel, cfg.LogEnabled)
	logger.Info("log level set to %s (enabled=%t)", cfg.LogLevel, cfg.LogEnabled)

	oracleSvc := oracle.NewService(cfg.OpenAIAPIKey)
	oracleHandler := handlers.NewOracleHandler(oracleSvc)

	middleware := []func(http.Handler) http.Handler{
		server.LoggingMiddleware,
		server.CORSMiddleware("*"),
		server.TimeoutMiddleware(2 * time.Minute),
	}

	router := server.NewRouter(server.RouterOptions{
		OracleHandler: oracleHandler,
		Middleware:    middleware,
	})

	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      0,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("CoffeeOracle backend listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}

	logger.Info("server stopped")
}

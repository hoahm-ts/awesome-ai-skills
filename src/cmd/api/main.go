package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"

	"github.com/hoahm-ts/awesome-ai-skills/pkg/config"
	"github.com/hoahm-ts/awesome-ai-skills/pkg/logger"
	appMiddleware "github.com/hoahm-ts/awesome-ai-skills/pkg/middleware"
)

const (
	_readTimeout     = 15 * time.Second
	_writeTimeout    = 15 * time.Second
	_idleTimeout     = 60 * time.Second
	_shutdownTimeout = 30 * time.Second
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := logger.New(cfg.App.Name, cfg.App.Env, zerolog.InfoLevel)

	r := chi.NewRouter()
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RequestID)
	r.Use(appMiddleware.RequestLogger(log))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      r,
		ReadTimeout:  _readTimeout,
		WriteTimeout: _writeTimeout,
		IdleTimeout:  _idleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("api server starting")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("api server failed")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("shutting down api server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), _shutdownTimeout)
	defer cancel()

	return srv.Shutdown(shutdownCtx)
}

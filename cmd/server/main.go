package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Isvane/gomen/internal/api"
	"github.com/Isvane/gomen/internal/repository"
)

func main() {
	logger := slog.Default()

	pprofServer := &http.Server{
		Addr: "localhost:6060",
	}

	go func() {
		slog.Info("pprof server listening on localhost:6060")
		if err := pprofServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("pprof server error", "error", err)
		}
	}()

	repo := repository.NewUserRepo()
	db := &api.Database{
		Repo: repo,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /echo/{arg}", api.EchoHandler)
	mux.HandleFunc("POST /user", db.RegisterUserHandler)
	mux.HandleFunc("GET /user/{name}", db.GetUserHandler)
	mux.HandleFunc("DELETE /user/{name}", db.DeleteUserHandler)
	mux.HandleFunc("PUT /user/{name}", db.UpdateUserHandler)

	s := &http.Server{
		Addr:           ":8080",
		Handler:        api.LogMiddleware(logger)(mux),
		ErrorLog:       slog.NewLogLogger(logger.Handler(), slog.LevelError),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		slog.Info("Server listening on http://localhost:8080")
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.Shutdown(shutdownCtx); err != nil {
		slog.Error("Shutdown error", "error", err)
	}

	if err := pprofServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("pprof server shutdown error", "error", err)
	}

	slog.Info("Gracefully shutting down.")
}

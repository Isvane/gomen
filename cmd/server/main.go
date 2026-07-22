package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Isvane/gohan/internal/api"
	"github.com/Isvane/gohan/internal/repository"
)

func main() {
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
		Handler:        api.LogMiddleware(mux),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		log.Printf("Server listening on http://localhost:8080")
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}

	log.Println("Gracefully shutting down.")
}

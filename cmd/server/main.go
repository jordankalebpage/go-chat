package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	gochat "go-chat"
	"go-chat/internal/chat"
	"go-chat/internal/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	staticFS, err := gochat.StaticFS()
	if err != nil {
		log.Fatalf("load embedded assets: %v", err)
	}

	hub := chat.NewHub([]string{"general", "golang", "streaming"})
	go hub.Run(ctx)

	config := server.Config{
		DemoAccessPassword: os.Getenv("DEMO_ACCESS_PASSWORD"),
		Port:               port,
		ShutdownTimeout:    10 * time.Second,
		StaticFS:           staticFS,
	}

	appServer := server.New(config, hub)

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
		defer cancel()

		if err := appServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	log.Printf("server listening on :%s", port)

	err = appServer.Start()
	if err == nil {
		return
	}

	if errors.Is(err, http.ErrServerClosed) {
		return
	}

	log.Fatalf("server error: %v", err)
}

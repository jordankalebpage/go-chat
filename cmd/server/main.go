package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	gochat "go-chat"
	"go-chat/internal/chat"
	"go-chat/internal/server"
)

func main() {
	if err := loadEnvFile(".env"); err != nil {
		log.Fatalf("load .env: %v", err)
	}

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

func loadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			return fmt.Errorf("line %d: expected KEY=VALUE", lineNumber)
		}

		key = strings.TrimSpace(key)
		if key == "" {
			return fmt.Errorf("line %d: missing key", lineNumber)
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)

		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

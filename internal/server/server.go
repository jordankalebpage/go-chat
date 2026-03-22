package server

import (
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"

	"go-chat/internal/chat"
	appmiddleware "go-chat/internal/middleware"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type Config struct {
	Port            string
	ShutdownTimeout time.Duration
	StaticFS        fs.FS
}

type Server struct {
	config     Config
	hub        *chat.Hub
	httpServer *http.Server
}

func New(config Config, hub *chat.Hub) *Server {
	server := &Server{
		config: config,
		hub:    hub,
	}

	router := chi.NewRouter()
	router.Use(chimiddleware.RealIP)
	router.Use(appmiddleware.Logging)
	router.Use(allowCORS)

	wsLimiter := appmiddleware.NewIPRateLimiter(1, 5, 15*time.Minute)

	router.Route("/api", func(r chi.Router) {
		r.Get("/health", server.handleHealth)
		r.Get("/rooms", server.handleRooms)
	})

	router.With(wsLimiter.Middleware).Get("/ws", server.handleWebSocket)

	if config.StaticFS != nil {
		router.Handle("/", spaHandler(config.StaticFS))
		router.Handle("/*", spaHandler(config.StaticFS))
	}

	server.httpServer = &http.Server{
		Addr:              ":" + config.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return server
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (s *Server) handleRooms(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.hub.ListRooms())
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	room, err := chat.NormalizeRoom(r.URL.Query().Get("room"))
	if err != nil {
		http.Error(w, chat.ErrInvalidRoom.Error(), http.StatusBadRequest)
		return
	}

	username, err := chat.NormalizeUsername(r.URL.Query().Get("username"))
	if err != nil {
		http.Error(w, chat.ErrInvalidUsername.Error(), http.StatusBadRequest)
		return
	}

	conn, err := chat.Upgrade(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	client := chat.NewClient(s.hub, conn, room, username)

	s.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}

func spaHandler(staticFS fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(staticFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cleanedPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if cleanedPath == "." || cleanedPath == "" {
			fileServer.ServeHTTP(w, r)
			return
		}

		_, err := fs.Stat(staticFS, cleanedPath)
		if err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(value)
}

func allowCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}

		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

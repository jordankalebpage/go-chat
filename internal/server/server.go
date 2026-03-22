package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"log"
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
	DemoAccessPassword string
	Port               string
	ShutdownTimeout    time.Duration
	StaticFS           fs.FS
}

type Server struct {
	config     Config
	hub        *chat.Hub
	httpServer *http.Server
}

const demoAccessCookieName = "go_chat_demo_access"

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
		r.Get("/session", server.handleSession)
		r.Post("/session", server.handleSessionLogin)
		r.Get("/health", server.handleHealth)
		r.With(server.requireDemoAccess).Get("/rooms", server.handleRooms)
	})

	router.With(server.requireDemoAccess, wsLimiter.Middleware).Get("/ws", server.handleWebSocket)

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

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"requiresPassword": s.demoAccessEnabled(),
		"unlocked":         s.hasDemoAccess(r),
	})
}

func (s *Server) handleSessionLogin(w http.ResponseWriter, r *http.Request) {
	if !s.demoAccessEnabled() {
		writeJSON(w, http.StatusOK, map[string]bool{
			"unlocked": true,
		})
		return
	}

	type sessionLoginRequest struct {
		Password string `json:"password"`
	}

	var request sessionLoginRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if request.Password != s.config.DemoAccessPassword {
		http.Error(w, "invalid password", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     demoAccessCookieName,
		Value:    demoAccessToken(s.config.DemoAccessPassword),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   requestScheme(r) == "https",
	})

	writeJSON(w, http.StatusOK, map[string]bool{
		"unlocked": true,
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
		log.Printf("websocket upgrade failed: %v", err)
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

func (s *Server) requireDemoAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.demoAccessEnabled() || s.hasDemoAccess(r) {
			next.ServeHTTP(w, r)
			return
		}

		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	})
}

func (s *Server) hasDemoAccess(r *http.Request) bool {
	if !s.demoAccessEnabled() {
		return true
	}

	cookie, err := r.Cookie(demoAccessCookieName)
	if err != nil {
		return false
	}

	return cookie.Value == demoAccessToken(s.config.DemoAccessPassword)
}

func (s *Server) demoAccessEnabled() bool {
	return strings.TrimSpace(s.config.DemoAccessPassword) != ""
}

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(value)
}

func allowCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && sameOrigin(origin, r.Host, requestScheme(r)) {
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

func sameOrigin(origin string, host string, scheme string) bool {
	return origin == scheme+"://"+host
}

func requestScheme(r *http.Request) string {
	forwardedProto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))
	if forwardedProto != "" {
		return forwardedProto
	}

	if r.TLS != nil {
		return "https"
	}

	return "http"
}

func demoAccessToken(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

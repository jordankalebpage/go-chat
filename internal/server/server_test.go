package server

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"go-chat/internal/chat"

	"github.com/gorilla/websocket"
)

func TestWebSocketUpgradeSucceedsWithLoggingMiddleware(t *testing.T) {
	hub := chat.NewHub([]string{"general"})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	server := New(Config{Port: "0"}, hub)
	testServer := httptest.NewServer(server.httpServer.Handler)
	defer testServer.Close()

	wsURL := "ws" + testServer.URL[len("http"):] + "/ws?room=general&username=gopher"

	conn, response, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		if response != nil {
			t.Fatalf("websocket dial failed with status %d: %v", response.StatusCode, err)
		}

		t.Fatalf("websocket dial failed: %v", err)
	}

	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}

	var message chat.ServerMessage
	if err := conn.ReadJSON(&message); err != nil {
		t.Fatalf("read join message: %v", err)
	}

	if message.Type != chat.MessageTypeJoin {
		t.Fatalf("expected join message, got %q", message.Type)
	}

	if message.Room != "general" {
		t.Fatalf("expected room %q, got %q", "general", message.Room)
	}

	if message.Username != "gopher" {
		t.Fatalf("expected username %q, got %q", "gopher", message.Username)
	}
}

package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestSessionReportsLockedStateWhenPasswordConfigured(t *testing.T) {
	hub := chat.NewHub([]string{"general"})
	server := New(Config{Port: "0", DemoAccessPassword: "secret"}, hub)
	testServer := httptest.NewServer(server.httpServer.Handler)
	defer testServer.Close()

	response, err := http.Get(testServer.URL + "/api/session")
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	defer response.Body.Close()

	var state struct {
		RequiresPassword bool `json:"requiresPassword"`
		Unlocked         bool `json:"unlocked"`
	}

	if err := json.NewDecoder(response.Body).Decode(&state); err != nil {
		t.Fatalf("decode session: %v", err)
	}

	if !state.RequiresPassword {
		t.Fatalf("expected requiresPassword to be true")
	}

	if state.Unlocked {
		t.Fatalf("expected unlocked to be false")
	}
}

func TestDemoAccessRequiredForRoomsAndWebSocket(t *testing.T) {
	hub := chat.NewHub([]string{"general"})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	server := New(Config{Port: "0", DemoAccessPassword: "secret"}, hub)
	testServer := httptest.NewServer(server.httpServer.Handler)
	defer testServer.Close()

	response, err := http.Get(testServer.URL + "/api/rooms")
	if err != nil {
		t.Fatalf("get rooms: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, response.StatusCode)
	}

	wsURL := "ws" + testServer.URL[len("http"):] + "/ws?room=general&username=gopher"

	_, wsResponse, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatalf("expected websocket dial to fail without demo access")
	}

	if wsResponse == nil {
		t.Fatalf("expected websocket response for unauthorized dial")
	}

	if wsResponse.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected websocket status %d, got %d", http.StatusUnauthorized, wsResponse.StatusCode)
	}
}

func TestSessionLoginUnlocksRoomsAndWebSocket(t *testing.T) {
	hub := chat.NewHub([]string{"general"})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	server := New(Config{Port: "0", DemoAccessPassword: "secret"}, hub)
	testServer := httptest.NewServer(server.httpServer.Handler)
	defer testServer.Close()

	request, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/session", strings.NewReader(`{"password":"secret"}`))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("login request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, response.StatusCode)
	}

	cookies := response.Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected login cookie")
	}

	roomsRequest, err := http.NewRequest(http.MethodGet, testServer.URL+"/api/rooms", nil)
	if err != nil {
		t.Fatalf("new rooms request: %v", err)
	}

	roomsRequest.AddCookie(cookies[0])

	roomsResponse, err := http.DefaultClient.Do(roomsRequest)
	if err != nil {
		t.Fatalf("rooms request: %v", err)
	}
	defer roomsResponse.Body.Close()

	if roomsResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, roomsResponse.StatusCode)
	}

	wsURL := "ws" + testServer.URL[len("http"):] + "/ws?room=general&username=gopher"

	header := http.Header{}
	header.Add("Cookie", cookies[0].String())

	conn, wsResponse, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		if wsResponse != nil {
			t.Fatalf("websocket dial failed with status %d: %v", wsResponse.StatusCode, err)
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
}

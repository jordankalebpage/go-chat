package chat

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestHubListsRooms(t *testing.T) {
	t.Parallel()

	hub, cancel := startTestHub(t)
	defer cancel()

	client := newTestClient(hub, "general", "alice")
	hub.Register(client)

	event := readEvent(t, client.send)
	if event.Type != MessageTypeJoin {
		t.Fatalf("expected join event, got %q", event.Type)
	}

	rooms := hub.ListRooms()
	if len(rooms) != 2 {
		t.Fatalf("expected 2 rooms, got %d", len(rooms))
	}

	if rooms[0].Name != "general" || rooms[0].MemberCount != 1 {
		t.Fatalf("unexpected general room summary: %+v", rooms[0])
	}

	if rooms[1].Name != "golang" || rooms[1].MemberCount != 0 {
		t.Fatalf("unexpected golang room summary: %+v", rooms[1])
	}
}

func TestHubBroadcastsWithinRoom(t *testing.T) {
	t.Parallel()

	hub, cancel := startTestHub(t)
	defer cancel()

	alice := newTestClient(hub, "general", "alice")
	bob := newTestClient(hub, "general", "bob")

	hub.Register(alice)
	hub.Register(bob)

	flushEvents(alice.send)
	flushEvents(bob.send)

	hub.Broadcast(Broadcast{
		Room:     "general",
		Username: "alice",
		Type:     MessageTypeMessage,
		Content:  "hello gophers",
	})

	aliceEvent := readEvent(t, alice.send)
	bobEvent := readEvent(t, bob.send)

	if aliceEvent.Content != "hello gophers" || bobEvent.Content != "hello gophers" {
		t.Fatalf("expected broadcast content for both clients, got %#v and %#v", aliceEvent, bobEvent)
	}
}

func TestHubKeepsRoomsIsolated(t *testing.T) {
	t.Parallel()

	hub, cancel := startTestHub(t)
	defer cancel()

	alice := newTestClient(hub, "general", "alice")
	bob := newTestClient(hub, "golang", "bob")

	hub.Register(alice)
	hub.Register(bob)

	flushEvents(alice.send)
	flushEvents(bob.send)

	hub.Broadcast(Broadcast{
		Room:     "general",
		Username: "alice",
		Type:     MessageTypeMessage,
		Content:  "general only",
	})

	aliceEvent := readEvent(t, alice.send)
	if aliceEvent.Content != "general only" {
		t.Fatalf("expected alice to receive room message, got %#v", aliceEvent)
	}

	select {
	case message := <-bob.send:
		var event ServerMessage
		if err := json.Unmarshal(message, &event); err != nil {
			t.Fatalf("unmarshal event: %v", err)
		}

		t.Fatalf("expected bob to receive no message, got %#v", event)
	case <-time.After(150 * time.Millisecond):
	}
}

func startTestHub(t *testing.T) (*Hub, context.CancelFunc) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	hub := NewHub([]string{"general", "golang"})

	go hub.Run(ctx)

	return hub, cancel
}

func newTestClient(hub *Hub, room string, username string) *Client {
	return &Client{
		hub:      hub,
		send:     make(chan []byte, 8),
		room:     room,
		username: username,
	}
}

func readEvent(t *testing.T, messages <-chan []byte) ServerMessage {
	t.Helper()

	select {
	case payload := <-messages:
		var event ServerMessage

		if err := json.Unmarshal(payload, &event); err != nil {
			t.Fatalf("unmarshal event: %v", err)
		}

		return event
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}

	return ServerMessage{}
}

func flushEvents(messages <-chan []byte) {
	for {
		select {
		case <-messages:
		default:
			return
		}
	}
}

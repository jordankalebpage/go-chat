package chat

import (
	"context"
	"encoding/json"
	"sort"
)

type roomListRequest struct {
	response chan []RoomSummary
}

type Hub struct {
	defaultRooms []string
	rooms        map[string]*Room

	register  chan *Client
	unregister chan *Client
	broadcast chan Broadcast
	listRooms chan roomListRequest
}

func NewHub(defaultRooms []string) *Hub {
	return &Hub{
		defaultRooms: append([]string(nil), defaultRooms...),
		rooms:        make(map[string]*Room),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		broadcast:    make(chan Broadcast, 64),
		listRooms:    make(chan roomListRequest),
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) Broadcast(message Broadcast) {
	h.broadcast <- message
}

func (h *Hub) ListRooms() []RoomSummary {
	response := make(chan []RoomSummary, 1)

	h.listRooms <- roomListRequest{
		response: response,
	}

	return <-response
}

func (h *Hub) Run(ctx context.Context) {
	for _, roomName := range h.defaultRooms {
		h.getOrCreateRoom(roomName)
	}

	for {
		select {
		case <-ctx.Done():
			h.shutdown()
			return
		case client := <-h.register:
			h.handleRegister(client)
		case client := <-h.unregister:
			h.handleUnregister(client)
		case message := <-h.broadcast:
			h.handleBroadcast(message)
		case request := <-h.listRooms:
			request.response <- h.roomSummaries()
		}
	}
}

func (h *Hub) handleRegister(client *Client) {
	room := h.getOrCreateRoom(client.room)
	room.Join(client)

	message := NewServerMessage(
		MessageTypeJoin,
		client.room,
		client.username,
		client.username+" joined the room",
		room.Usernames(),
	)

	h.broadcastToRoom(room, message)
}

func (h *Hub) handleUnregister(client *Client) {
	room, ok := h.rooms[client.room]
	if !ok {
		client.CloseSend()
		_ = client.CloseConn()
		return
	}

	if room.MemberCount() == 0 {
		client.CloseSend()
		_ = client.CloseConn()
		return
	}

	room.Leave(client)

	message := NewServerMessage(
		MessageTypeLeave,
		client.room,
		client.username,
		client.username+" left the room",
		room.Usernames(),
	)

	h.broadcastToRoom(room, message)

	if room.MemberCount() == 0 && !h.isDefaultRoom(room.Name) {
		delete(h.rooms, room.Name)
	}

	client.CloseSend()
	_ = client.CloseConn()
}

func (h *Hub) handleBroadcast(message Broadcast) {
	room, ok := h.rooms[message.Room]
	if !ok {
		return
	}

	event := NewServerMessage(
		message.Type,
		message.Room,
		message.Username,
		message.Content,
		room.Usernames(),
	)

	h.broadcastToRoom(room, event)
}

func (h *Hub) broadcastToRoom(room *Room, message ServerMessage) {
	payload, err := json.Marshal(message)
	if err != nil {
		return
	}

	droppedClients := room.Broadcast(payload)

	for _, client := range droppedClients {
		room.Leave(client)
		client.CloseSend()
		_ = client.CloseConn()
	}
}

func (h *Hub) roomSummaries() []RoomSummary {
	names := make([]string, 0, len(h.rooms))

	for name := range h.rooms {
		names = append(names, name)
	}

	sort.Strings(names)

	summaries := make([]RoomSummary, 0, len(names))

	for _, name := range names {
		summaries = append(summaries, RoomSummary{
			Name:        name,
			MemberCount: h.rooms[name].MemberCount(),
		})
	}

	return summaries
}

func (h *Hub) getOrCreateRoom(name string) *Room {
	if room, ok := h.rooms[name]; ok {
		return room
	}

	room := NewRoom(name)
	h.rooms[name] = room

	return room
}

func (h *Hub) isDefaultRoom(name string) bool {
	for _, roomName := range h.defaultRooms {
		if roomName == name {
			return true
		}
	}

	return false
}

func (h *Hub) shutdown() {
	for _, room := range h.rooms {
		for _, client := range room.Members() {
			room.Leave(client)
			client.CloseSend()
			_ = client.CloseConn()
		}
	}
}

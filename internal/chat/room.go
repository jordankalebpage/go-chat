package chat

import (
	"sort"
	"sync"
)

type Room struct {
	Name string

	mu      sync.RWMutex
	members map[*Client]struct{}
}

func NewRoom(name string) *Room {
	return &Room{
		Name:    name,
		members: make(map[*Client]struct{}),
	}
}

func (r *Room) Join(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.members[client] = struct{}{}
}

func (r *Room) Leave(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.members, client)
}

func (r *Room) Members() []*Client {
	r.mu.RLock()
	defer r.mu.RUnlock()

	members := make([]*Client, 0, len(r.members))

	for client := range r.members {
		members = append(members, client)
	}

	return members
}

func (r *Room) Usernames() []string {
	members := r.Members()
	usernames := make([]string, 0, len(members))

	for _, client := range members {
		usernames = append(usernames, client.username)
	}

	sort.Strings(usernames)

	return usernames
}

func (r *Room) MemberCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.members)
}

func (r *Room) Broadcast(payload []byte) []*Client {
	members := r.Members()
	droppedClients := make([]*Client, 0)

	for _, client := range members {
		select {
		case client.send <- payload:
		default:
			droppedClients = append(droppedClients, client)
		}
	}

	return droppedClients
}

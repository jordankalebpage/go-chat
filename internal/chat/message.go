package chat

import (
	"errors"
	"strings"
	"time"
	"unicode"
)

const (
	MessageTypeJoin    = "join"
	MessageTypeLeave   = "leave"
	MessageTypeMessage = "message"
)

const (
	maxUsernameLength = 24
	maxRoomLength     = 32
	maxContentLength  = 400
)

var (
	ErrInvalidUsername = errors.New("invalid username")
	ErrInvalidRoom     = errors.New("invalid room")
	ErrInvalidContent  = errors.New("invalid content")
)

type ClientMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type ServerMessage struct {
	Type      string   `json:"type"`
	Room      string   `json:"room,omitempty"`
	Username  string   `json:"username,omitempty"`
	Content   string   `json:"content,omitempty"`
	Users     []string `json:"users,omitempty"`
	Timestamp string   `json:"timestamp,omitempty"`
}

type RoomSummary struct {
	Name        string `json:"name"`
	MemberCount int    `json:"memberCount"`
}

type Broadcast struct {
	Room     string
	Username string
	Type     string
	Content  string
}

func NewServerMessage(messageType string, room string, username string, content string, users []string) ServerMessage {
	return ServerMessage{
		Type:      messageType,
		Room:      room,
		Username:  username,
		Content:   content,
		Users:     users,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

func NormalizeUsername(value string) (string, error) {
	trimmed := strings.TrimSpace(value)

	if trimmed == "" {
		return "", ErrInvalidUsername
	}

	if len([]rune(trimmed)) > maxUsernameLength {
		return "", ErrInvalidUsername
	}

	for _, char := range trimmed {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			continue
		}

		if char == '-' || char == '_' {
			continue
		}

		return "", ErrInvalidUsername
	}

	return trimmed, nil
}

func NormalizeRoom(value string) (string, error) {
	trimmed := strings.TrimSpace(strings.ToLower(value))

	if trimmed == "" {
		return "", ErrInvalidRoom
	}

	if len([]rune(trimmed)) > maxRoomLength {
		return "", ErrInvalidRoom
	}

	for _, char := range trimmed {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			continue
		}

		if char == '-' || char == '_' {
			continue
		}

		return "", ErrInvalidRoom
	}

	return trimmed, nil
}

func NormalizeContent(value string) (string, error) {
	trimmed := strings.TrimSpace(value)

	if trimmed == "" {
		return "", ErrInvalidContent
	}

	if len([]rune(trimmed)) > maxContentLength {
		return "", ErrInvalidContent
	}

	return trimmed, nil
}

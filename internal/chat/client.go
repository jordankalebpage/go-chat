package chat

import (
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024
	sendBufferSize = 32
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

var ErrUpgradeFailed = errors.New("websocket upgrade failed")

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	room     string
	username string
	limiter  *rate.Limiter
	sendOnce  sync.Once
	connOnce  sync.Once
}

func NewClient(hub *Hub, conn *websocket.Conn, room string, username string) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, sendBufferSize),
		room:     room,
		username: username,
		limiter:  rate.NewLimiter(rate.Every(250*time.Millisecond), 4),
	}
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, errors.Join(ErrUpgradeFailed, err)
	}

	return conn, nil
}

func (c *Client) Close() error {
	c.CloseSend()
	return c.CloseConn()
}

func (c *Client) CloseSend() {
	c.sendOnce.Do(func() {
		if c.send != nil {
			close(c.send)
		}
	})
}

func (c *Client) CloseConn() error {
	var closeErr error

	c.connOnce.Do(func() {
		if c.conn == nil {
			return
		}

		closeErr = c.conn.Close()
	})

	return closeErr
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)

		if err := c.CloseConn(); err != nil && !errors.Is(err, websocket.ErrCloseSent) {
			log.Printf("readPump close error: %v", err)
		}
	}()

	c.conn.SetReadLimit(maxMessageSize)

	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Printf("read deadline error: %v", err)
		return
	}

	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		var incoming ClientMessage

		if err := c.conn.ReadJSON(&incoming); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket read error: %v", err)
			}

			return
		}

		if incoming.Type != MessageTypeMessage {
			c.closeWithMessage(websocket.CloseUnsupportedData, "unsupported message type")
			return
		}

		if !c.limiter.Allow() {
			c.closeWithMessage(websocket.ClosePolicyViolation, "message rate limit exceeded")
			return
		}

		content, err := NormalizeContent(incoming.Content)
		if err != nil {
			c.closeWithMessage(websocket.CloseInvalidFramePayloadData, err.Error())
			return
		}

		c.hub.Broadcast(Broadcast{
			Room:     c.room,
			Username: c.username,
			Type:     MessageTypeMessage,
			Content:  content,
		})
	}
}

func (c *Client) closeWithMessage(code int, message string) {
	if c.conn == nil {
		return
	}

	deadline := time.Now().Add(writeWait)
	if err := c.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(code, message), deadline); err != nil {
		log.Printf("write close control error: %v", err)
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()

		if err := c.CloseConn(); err != nil && !errors.Is(err, websocket.ErrCloseSent) {
			log.Printf("writePump close error: %v", err)
		}
	}()

	for {
		select {
		case message, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Printf("write deadline error: %v", err)
				return
			}

			if !ok {
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil && !errors.Is(err, websocket.ErrCloseSent) {
					log.Printf("close frame error: %v", err)
				}

				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("websocket write error: %v", err)
				return
			}
		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Printf("ping deadline error: %v", err)
				return
			}

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

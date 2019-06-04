package gosocket

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Socket is a middleman between the websocket connection and the server.
type Socket struct {
	server *Server
	conn   *websocket.Conn
	send   chan []byte

	ID string
}

type subscription struct {
	socket   *Socket
	socketID string
	room     string
	message  Message
}

// Message ...
type Message struct {
	Type    string                 `json:"type"`
	Content map[string]interface{} `json:"content"`
}

// Router ..
func Router(server *Server, w http.ResponseWriter, r *http.Request) {
	// Origin domain
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	conn, err := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity
	if err != nil {
		log.Println("conn err", err)
	}
	socket := &Socket{
		server: server,
		conn:   conn,
		send:   make(chan []byte, 256),

		ID: createID(32),
	}

	socket.server.register <- socket

	message := Message{
		Type:    "connection",
		Content: make(map[string]interface{}),
	}
	go socket.server.onPacket(socket, message)
	go socket.listenReadPump()
	go socket.listenWritePump()
}

func (s *Socket) listenReadPump() {
	defer func() {
		s.server.unregister <- s
		s.conn.Close()
	}()
	s.conn.SetReadLimit(maxMessageSize)
	s.conn.SetReadDeadline(time.Now().Add(pongWait))
	s.conn.SetPongHandler(func(string) error { s.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, readDataByte, err := s.conn.ReadMessage()
		if err != nil {
			s.Disconnect()
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		message := parseByteToMessage(readDataByte)

		go s.server.onPacket(s, message)
		// s.Broadcast(readDataByte)
	}
}

func (s *Socket) listenWritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		s.conn.Close()
	}()
	for {
		select {
		case message, ok := <-s.send:
			s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The server closed the channel.
				s.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := s.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(s.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-s.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Broadcast ...
func (s *Socket) Broadcast(message Message) {
	subscription := subscription{}
	subscription.message = message
	s.server.broadcast <- subscription
}

// BroadcastTo ...BroadcastTo
func (s *Socket) BroadcastTo(socketID string, message Message) {
	subscription := subscription{}
	subscription.socketID = socketID
	subscription.message = message
	s.server.broadcastTo <- subscription
}

// BroadcastEmit ...
func (s *Socket) BroadcastEmit(message Message) {
	subscription := subscription{}
	subscription.message = message
	subscription.socket = s
	s.server.broadcastEmit <- subscription
}

// Emit ...
func (s *Socket) Emit(message Message) {
	subscription := subscription{}
	subscription.message = message
	subscription.socket = s
	s.server.emit <- subscription
}

// Join ...
func (s *Socket) Join(name string) {
	subscription := subscription{}
	subscription.room = name
	subscription.socket = s
	s.server.Join(subscription)
}

// Leave ...
func (s *Socket) Leave(name string) {
	subscription := subscription{}
	subscription.room = name
	subscription.socket = s
	s.server.Leave(subscription)
}

// BroadcastRoom ...
func (s *Socket) BroadcastRoom(name string, message Message) {
	subscription := subscription{}
	subscription.socket = s
	subscription.room = name
	subscription.message = message
	s.server.broadcastRoom <- subscription
}

// Disconnect ...
func (s *Socket) Disconnect() {
	s.server.unregister <- s
}

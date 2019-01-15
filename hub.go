package socket

import (
	"encoding/json"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages to the
type Hub struct {
	events map[string]*caller
	evMu   sync.Mutex

	sockets       map[*Socket]bool
	broadcast     chan []byte
	broadcastTo   chan []byte
	broadcastEmit chan map[*Socket][]byte
	emit          chan map[*Socket][]byte
	register      chan *Socket
	unregister    chan *Socket

	rooms         map[*Room]map[*Socket]bool
	join          chan map[*Room]*Socket
	leave         chan map[*Room]*Socket
	broadcastRoom chan map[*Room][]byte
}

// New ...
func New() *Hub {
	hub := &Hub{
		events: make(map[string]*caller),
		evMu:   sync.Mutex{},

		sockets: make(map[*Socket]bool),
		rooms:   make(map[*Room]map[*Socket]bool),

		emit:          make(chan map[*Socket][]byte),
		broadcast:     make(chan []byte),
		broadcastTo:   make(chan []byte),
		broadcastRoom: make(chan map[*Room][]byte),
		broadcastEmit: make(chan map[*Socket][]byte),

		register:   make(chan *Socket),
		unregister: make(chan *Socket),

		join:  make(chan map[*Room]*Socket),
		leave: make(chan map[*Room]*Socket),
	}
	go hub.run()
	return hub
}

// Run ...
func (h *Hub) run() {
	for {
		select {

		/* Case about socket event
		* Register
		* Unregister
		* Broadcast
		* BroadcastTo
		* BroadcastEmit
		* Emit
		 */
		case socket := <-h.register:
			h.sockets[socket] = true
		case socket := <-h.unregister:
			if _, ok := h.sockets[socket]; ok {
				message := Message{
					Type:    "disconnect",
					Content: make(map[string]interface{}),
				}
				go h.onPacket(socket, message)
				delete(h.sockets, socket)
				close(socket.send)
			}

		//Send message to all socket
		case message := <-h.broadcast:
			for socket := range h.sockets {
				select {
				case socket.send <- message:
				default:
					close(socket.send)
					delete(h.sockets, socket)
				}
			}

		//Send message to specific socket by socket_id
		case message := <-h.broadcastTo:
			var msg Message
			json.Unmarshal(message, &msg)
			for socket := range h.sockets {
				if socket.ID == msg.SocketID {
					socket.send <- message
				}
			}

		// Send message to other user except myself
		case message := <-h.broadcastEmit:
			for so, msg := range message {
				for socket := range h.sockets {
					if socket.ID != so.ID {
						socket.send <- msg
					}
				}
			}

		// Send message to myself
		case message := <-h.emit:
			for so, msg := range message {
				for socket := range h.sockets {
					if socket.ID == so.ID {
						socket.send <- msg
					}
				}
			}

			// End event about socket

		/* Event about room
		* Join
		* Leave
		* BroadcastRoom
		 */

		// A socket join a room
		case data := <-h.join:
			for room, so := range data {
				h.rooms[room][so] = true
			}

		// A socket leave room
		case data := <-h.leave:
			for room, so := range data {
				if _, ok := h.rooms[room][so]; ok {
					delete(h.rooms[room], so)
					// close(so.send) Rời phòng
				}
			}

		// Send message to every socket in specific room by room_id
		case message := <-h.broadcastRoom:
			for ro, msg := range message {
				for room, sockets := range h.rooms {
					if room.ID == ro.ID {
						for socket := range sockets {
							socket.send <- msg
						}
					}
				}
			}
			// End event about room
		}
	}
}

// On registers the function f to handle an event.
func (h *Hub) On(event string, f interface{}) error {
	c, err := newCaller(f)
	if err != nil {
		return err
	}
	h.evMu.Lock()
	h.events[event] = c
	h.evMu.Unlock()
	return nil
}

func (h *Hub) onPacket(so *Socket, message Message) ([]interface{}, error) {
	h.evMu.Lock()
	c, ok := h.events[message.Type]
	h.evMu.Unlock()
	if !ok {
		return nil, nil
	}

	retV := c.Call(so, message)
	if len(retV) == 0 {
		return nil, nil
	}

	var err error
	if last, ok := retV[len(retV)-1].Interface().(error); ok {
		err = last
		retV = retV[0 : len(retV)-1]
	}
	ret := make([]interface{}, len(retV))
	for i, v := range retV {
		ret[i] = v.Interface()
	}
	return ret, err
}

// Broadcast ..
func (h *Hub) Broadcast(message Message) {
	empData, _ := json.Marshal(message)
	h.broadcast <- empData
}

// BroadcastTo ...
func (h *Hub) BroadcastTo(message Message) {
	empData, _ := json.Marshal(message)
	h.broadcastTo <- empData
}

// BroadcastRoom ...
func (h *Hub) BroadcastRoom(name string, message Message) {
	room := &Room{
		ID: name,
	}
	data := make(map[*Room][]byte)
	empData, _ := json.Marshal(message)
	data[room] = empData
	h.broadcastRoom <- data
}

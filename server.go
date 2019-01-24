package gosocket

import (
	"encoding/json"
	"sync"
)

// Server maintains the set of active clients and broadcasts messages to the
type Server struct {
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
func New() *Server {
	Server := &Server{
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
	go Server.run()
	return Server
}

// Run ...
func (s *Server) run() {
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
		case socket := <-s.register:
			s.sockets[socket] = true
		case socket := <-s.unregister:
			if _, ok := s.sockets[socket]; ok {
				message := Message{
					Type:    "disconnect",
					Content: make(map[string]interface{}),
				}
				go s.onPacket(socket, message)
				delete(s.sockets, socket)
				close(socket.send)
			}

		//Send message to all socket
		case message := <-s.broadcast:
			for socket := range s.sockets {
				select {
				case socket.send <- message:
				default:
					close(socket.send)
					delete(s.sockets, socket)
				}
			}

		//Send message to specific socket by socket_id
		case message := <-s.broadcastTo:
			var msg Message
			json.Unmarshal(message, &msg)
			for socket := range s.sockets {
				if socket.ID == msg.SocketID {
					socket.send <- message
				}
			}

		// Send message to other user except myself
		case message := <-s.broadcastEmit:
			for so, msg := range message {
				for socket := range s.sockets {
					if socket.ID != so.ID {
						socket.send <- msg
					}
				}
			}

		// Send message to myself
		case message := <-s.emit:
			for so, msg := range message {
				for socket := range s.sockets {
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
		case data := <-s.join:
			for room, so := range data {
				s.rooms[room][so] = true
			}

		// A socket leave room
		case data := <-s.leave:
			for room, so := range data {
				if _, ok := s.rooms[room][so]; ok {
					delete(s.rooms[room], so)
					// close(so.send) Rời phòng
				}
			}

		// Send message to every socket in specific room by room_id
		case message := <-s.broadcastRoom:
			for ro, msg := range message {
				for room, sockets := range s.rooms {
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
func (s *Server) On(event string, f interface{}) error {
	c, err := newCaller(f)
	if err != nil {
		return err
	}
	s.evMu.Lock()
	s.events[event] = c
	s.evMu.Unlock()
	return nil
}

func (s *Server) onPacket(so *Socket, message Message) ([]interface{}, error) {
	s.evMu.Lock()
	c, ok := s.events[message.Type]
	s.evMu.Unlock()
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
func (s *Server) Broadcast(message Message) error {
	empData, err := json.Marshal(message)
	if err != nil {
		return err
	}
	s.broadcast <- empData
	return nil
}

// BroadcastTo ...
func (s *Server) BroadcastTo(socketID interface{}, message Message) error {
	message.SocketID = socketID.(string)
	empData, err := json.Marshal(message)
	if err != nil {
		return err
	}
	s.broadcastTo <- empData
	return nil
}

// BroadcastRoom ...
func (s *Server) BroadcastRoom(name string, message Message) error {
	room := &Room{
		ID: name,
	}
	data := make(map[*Room][]byte)
	empData, err := json.Marshal(message)
	if err != nil {
		return err
	}
	data[room] = empData
	s.broadcastRoom <- data
	return nil
}

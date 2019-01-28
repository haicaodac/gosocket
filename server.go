package gosocket

import (
	"sync"
)

// Server maintains the set of active clients and broadcasts messages to the
type Server struct {
	events map[string]*caller
	evMu   sync.Mutex

	sockets       map[*Socket]bool
	broadcast     chan subscription
	broadcastTo   chan subscription
	broadcastEmit chan subscription
	emit          chan subscription
	register      chan *Socket
	unregister    chan *Socket

	rooms         map[string]map[*Socket]bool
	join          chan subscription
	leave         chan subscription
	broadcastRoom chan subscription
}

// New ...
func New() *Server {
	Server := &Server{
		events: make(map[string]*caller),
		evMu:   sync.Mutex{},

		sockets: make(map[*Socket]bool),
		rooms:   make(map[string]map[*Socket]bool),

		emit:          make(chan subscription),
		broadcast:     make(chan subscription),
		broadcastTo:   make(chan subscription),
		broadcastRoom: make(chan subscription),
		broadcastEmit: make(chan subscription),

		register:   make(chan *Socket),
		unregister: make(chan *Socket),

		join:  make(chan subscription),
		leave: make(chan subscription),
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

				// Client out rooms
				for room, sockets := range s.rooms {
					for socketRom := range sockets {
						if socket.ID == socketRom.ID {
							delete(s.rooms[room], socket)
						}
					}
				}
			}

		//Send message to all socket
		case subscription := <-s.broadcast:
			message := parseMessage(subscription.message)
			for socket := range s.sockets {
				select {
				case socket.send <- message:
				default:
					s.unregister <- socket
				}
			}

		//Send message to specific socket by socket_id
		case subscription := <-s.broadcastTo:
			message := parseMessage(subscription.message)
			for socket := range s.sockets {
				if socket.ID == subscription.socketID {
					socket.send <- message
				}
			}

		// Send message to other user except myself
		case subscription := <-s.broadcastEmit:
			message := parseMessage(subscription.message)
			for socket := range s.sockets {
				if socket.ID != subscription.socket.ID {
					socket.send <- message
				}
			}

		// Send message to myself
		case subscription := <-s.emit:
			message := parseMessage(subscription.message)
			subscription.socket.send <- message

			// End event about socket

		/* Event about room
		* Join
		* Leave
		* BroadcastRoom
		 */

		// A socket join a room
		case subscription := <-s.join:
			if len(s.rooms[subscription.room]) == 0 {
				s.rooms[subscription.room] = make(map[*Socket]bool)
			}
			s.rooms[subscription.room][subscription.socket] = true

		// A socket leave room
		case subscription := <-s.leave:
			if _, ok := s.rooms[subscription.room][subscription.socket]; ok {
				delete(s.rooms[subscription.room], subscription.socket)
				// close(so.send) Rời phòng
			}

		// Send message to every socket in specific room by room_id
		case subscription := <-s.broadcastRoom:
			message := parseMessage(subscription.message)
			flag := true // Validator socket in room
			if subscription.socket != nil {
				flag = false
				for socket := range s.rooms[subscription.room] {
					if socket.ID == subscription.socket.ID {
						flag = true
					}
				}
			}
			if flag {
				for socket := range s.rooms[subscription.room] {
					select {
					case socket.send <- message:
					default:
						s.unregister <- socket
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
func (s *Server) Broadcast(message Message) {
	subscription := subscription{}
	subscription.message = message
	s.broadcast <- subscription
}

// BroadcastTo ...
func (s *Server) BroadcastTo(socketID string, message Message) {
	subscription := subscription{}
	subscription.socketID = socketID
	subscription.message = message
	s.broadcastTo <- subscription
}

// BroadcastRoom ...
func (s *Server) BroadcastRoom(name string, message Message) {
	subscription := subscription{}
	subscription.room = name
	subscription.message = message
	s.broadcastRoom <- subscription
}

// CountRoom ...
func (s *Server) CountRoom() int {
	return len(s.rooms)
}

// CountSocketInRoom ...
func (s *Server) CountSocketInRoom(name string) int {
	count := len(s.rooms[name])
	return count
}

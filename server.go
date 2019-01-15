// main ...
package main

import (
	"net/http"
	socket "socket/socket"
)

func main() {
	// Websocket
	server := socket.New()
	server.On("connection", func(so *socket.Socket, message socket.Message) {

		message.Content["id"] = so.ID
		so.Broadcast(message)

		server.On("msg", func(so *socket.Socket, message socket.Message) {
			if message.SocketID != "" {
				so.BroadcastTo(message)
			} else {
				so.Emit(message)
			}
		})
	})

	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		socket.Router(server, w, r)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.ListenAndServe(":8080", nil)
}

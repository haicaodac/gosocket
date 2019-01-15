// main ...
package main

import (
	"fmt"
	"log"
	"net/http"

	gosocket "github.com/haicaodac/gosocket"
)

func main() {
	// Websocket
	server := gosocket.New()
	server.On("connection", func(so *gosocket.Socket, message gosocket.Message) {

		message.Content["id"] = so.ID
		so.Broadcast(message)

		server.On("msg", func(so *gosocket.Socket, message gosocket.Message) {
			if message.SocketID != "" {
				so.BroadcastTo(message)
			} else {
				so.Emit(message)
			}
		})
	})

	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		gosocket.Router(server, w, r)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	fmt.Println("Server run ...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

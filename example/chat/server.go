// main ...
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/haicaodac/gosocket"
)

func main() {
	// Websocket
	server := gosocket.New()
	server.On("connection", func(so *gosocket.Socket, message gosocket.Message) {

		message.Content["id"] = so.ID
		so.Broadcast(message)

		server.On("msg", func(so *gosocket.Socket, message gosocket.Message) {
			if !gosocket.IsBlank(message.Content["socket_id"]) {
				socketID := message.Content["socket_id"].(string)
				so.BroadcastTo(socketID, message)
			} else {
				so.Broadcast(message)
			}
		})

		server.On("disconnect", func(so *gosocket.Socket, message gosocket.Message) {
			message.Content["message"] = "User: " + so.ID + " out room"
			server.Broadcast(message)
		})
	})

	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		gosocket.Router(server, w, r)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	fmt.Println("Server run port: 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

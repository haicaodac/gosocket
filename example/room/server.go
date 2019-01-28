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
			if !gosocket.IsBlank(message.Content["room"]) {
				room := message.Content["room"].(string)
				so.BroadcastRoom(room, message)
			}
		})

		server.On("join", func(so *gosocket.Socket, message gosocket.Message) {
			if !gosocket.IsBlank(message.Content["room"]) {
				room := message.Content["room"].(string)
				so.Join(room)
				// fmt.Println(server.CountSocketInRoom(room))
				message.Content["count_user"] = server.CountSocketInRoom(room)
				so.Emit(message)
				// fmt.Println(server.CountSocketInRoom(room))
			}
		})

		server.On("leave", func(so *gosocket.Socket, message gosocket.Message) {
			if !gosocket.IsBlank(message.Content["room"]) {
				room := message.Content["room"].(string)
				so.Leave(room)
				message.Content["message"] = "User: " + so.ID + " out room"
				server.BroadcastRoom(room, message)
			}
		})
	})

	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		gosocket.Router(server, w, r)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "room.html")
	})

	fmt.Println("Server run ...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

# Go Socket

# Struct 
#### Message
```sh
type Message struct {
	Type     string                 `json:"type"`
	Room     string                 `json:"room"`
	SocketID string                 `json:"socket_id"`
	Content  map[string]interface{} `json:"content"`
}
```
#### Socket
```sh
type Socket struct {
    ID string
    ...
}
```
#### Server
```sh
type Server struct {
    ...
}
```
# API

#### Struct Message


#### Broadcast
```sh
so.Broadcast(message)

or

server.Broadcast(message)
```
#### BroadcastTo
```sh
if message.Content["socket_id"] != nil {
    so.BroadcastTo(message.Content["socket_id"], message)
}

or 

server.BroadcastTo(message.Content["socket_id"], message)
```
#### BroadcastEmit
```sh
so.BroadcastEmit(message)
```
#### Emit
```sh
so.Emit(message)
```
#### Join
```sh
so.Join("abc")
```
#### Leave
```sh
so.Leave("abc")
```
#### BroadcastRoom
```sh
so.BroadcastRoom("abc", message)

or 

server.BroadcastRoom("abc", message)
```

# Example

Create file **index.html**
```sh
<!DOCTYPE html>
<html>

<head></head>

<body>
    <!-- websockets.html -->
    <input id="input" type="text" value="test" />
    <input id="idSocket" type="text" placeholder="ID socket" />
    <button onclick="send()">Send</button>
    <pre id="output"></pre>
    <script src="https://hanyny.com/gosocket/gosocket.js"></script>
    <script>
        var _url = "ws://localhost:8080/echo"
        gosocket = gosocket(_url);

        var input = document.getElementById("input");
        var id = document.getElementById("idSocket");

        var output = document.getElementById("output");

        gosocket.on("connection", function (data) {
            output.innerHTML += "Server: ID - " + data.id + "\n";
        })

        gosocket.on("msg", function (data) {
            output.innerHTML += "Server: " + data.value + "\n";
        })

        function send() {
            if (id.value) {
                var data = {
                    "socket_id": id.value,
                    "value": input.value
                }
            } else {
                var data = {
                    "value": input.value
                }
            }
            gosocket.emit("msg", data)
            input.value = ""
        }
    </script>
</body>

</html>
```
Create file **server.go**
```sh
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
			if message.Content["socket_id"] != nil {
				so.BroadcastTo(message.Content["socket_id"], message)
			} else {
				so.Broadcast(message)
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
```
# Run server
```sh
go run server.go
```
```sh
http://127.0.0.1:8080
```

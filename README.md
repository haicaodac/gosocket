# Websocket in Golang

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
    <script>
        var _url = "ws://localhost:8080/echo"
        var socket = null
        socket = new WebSocket(_url);

        var input = document.getElementById("input");
        var id = document.getElementById("idSocket");

        var output = document.getElementById("output");

        socket.onmessage = function (e) {
            var data = JSON.parse(e.data);
            console.log(data);
            switch (data.type) {
                case "connection":
                    connection(data);
                    break;
                case "msg":
                    receiveMessage(data);
                    break;
            }
        };

        function connection(data) {
            output.innerHTML += "Server: ID - " + data.content.id + "\n";
        }

        function receiveMessage(data) {
            console.log(data);
            output.innerHTML += "Server: " + data.content.value + "\n";
        }

        function send() {
            var data = {
                "type": "msg",
                "socket_id": id.value,
                "content": {
                    "value": input.value
                }
            }
            data = JSON.stringify(data);
            socket.send(data);
            input.value = "";
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

```
# Run server
```sh
go run server.go
```
```sh
http://127.0.0.1:8080
```

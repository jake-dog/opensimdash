package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func init() {
	http.HandleFunc("/sock", wsEndpoint)
	http.Handle("/", http.FileServer(http.Dir("static")))
}

type dataPoint struct {
	Speed int
}

// define a reader which will listen for
// new messages being sent to our WebSocket
// endpoint
/*func reader(conn *websocket.Conn) {
	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			logger.Println(err)
			return
		}
		// print out that message for clarity
		fmt.Println(string(p))

		if err := conn.WriteMessage(messageType, p); err != nil {
			logger.Println(err)
			return
		}
	}
}*/

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// upgrade this connection to a WebSocket
	// connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Println(err)
	}

	logger.Println("Client Connected")

	var i int
	for {
		b, _ := json.Marshal(&dataPoint{Speed: (50 + i) % 700})
		i = i + 1

		//err = ws.WriteMessage(1, []byte("Hi Client!"))
		err = ws.WriteMessage(1, b)
		if err != nil {
			logger.Println(err)
		}

		time.Sleep(2 * time.Second)
	}

	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	//reader(ws)
}

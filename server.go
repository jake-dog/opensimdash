package main

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/jake-dog/opensimdash/hid"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	defaultWriter = &WebSockWriter{}

	// ws is a wrapper around defaultWriter for faster JSON marshalling
	ws = &webSockPackSender{
		WebSockWriter: defaultWriter,
		Buf:           make([]byte, 0, 100), // Size large enough for marshalling!
	}
)

func init() {
	http.HandleFunc("/sock", wsEndpoint)
	http.Handle("/", http.FileServer(http.Dir("static")))
}

type webSockPackSender struct {
	*WebSockWriter
	Buf []byte
}

// SendPack is about 6 times faster than json.Marshaler
func (ws *webSockPackSender) SendPack(d hid.TelemetryPack) {
	ws.Buf = ws.Buf[:0]
	ws.Buf = append(ws.Buf, `{"Gear":`...)
	ws.Buf = strconv.AppendInt(ws.Buf, int64(d.GetGear()), 10) // Avoid allocs!
	ws.Buf = append(ws.Buf, `,"Speed":`...)
	ws.Buf = strconv.AppendInt(ws.Buf, int64(d.GetSpeed()), 10) // Avoid allocs!
	ws.Buf = append(ws.Buf, `}`...)
	ws.Write(ws.Buf)
}

// WebSockWriter provides a thread safe mechanism for performing synchronous
// writes to multiple websockets.  Clients which disconnect are removed from the
// pool automatically.
type WebSockWriter struct {
	writersmu sync.Mutex
	writers   []*websocket.Conn
}

// Write binary data to all connected websockets synchronously
func (w *WebSockWriter) Write(b []byte) (n int, err error) {
	w.writersmu.Lock()
	defer w.writersmu.Unlock()

	// Send data to each WS client, and remove clients who throw errors
	var i int
	for _, ws := range w.writers {
		if erro := ws.WriteMessage(1, b); erro != nil {
			err = erro
			// ws.Close() // shoudl we force close or let connection timeout?
		} else {
			w.writers[i] = ws
			i++
		}
	}
	w.writers = w.writers[:i]

	if err != nil {
		n = len(b)
	}
	return n, err
}

// Add a websocket to the pool of websockets being written to.  Websockets that
// disconnect are automatically removed.
func (w *WebSockWriter) Add(ws *websocket.Conn) {
	w.writersmu.Lock()
	defer w.writersmu.Unlock()

	// Add a WS, and create array if its not already there
	if w.writers == nil {
		w.writers = make([]*websocket.Conn, 0, 1)
	}
	w.writers = append(w.writers, ws)
}

// WriteMessage to all websockets
func WriteMessage(b []byte) (n int, err error) {
	return defaultWriter.Write(b)
}

// AddWebSock to the pool of websockets
func AddWebSock(ws *websocket.Conn) {
	defaultWriter.Add(ws)
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// Upgrade connection to a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Println(err)
	}

	logger.Println("Client Connected")

	// Add our new WS connection to the global WebSocketWriter
	AddWebSock(ws)
}

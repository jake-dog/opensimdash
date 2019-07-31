package main

import (
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/jake-dog/opendash/codemasters"
	"github.com/jake-dog/opendash/hid"
)

var logger = log.New(os.Stdout, "", log.LstdFlags|log.LUTC|log.Lshortfile)

func main() {
	// We're doing a lot of system calls here so minimum three system threads
	runtime.GOMAXPROCS(3)

	// TODO need to make this configurable, and catch errors it throws
	go http.ListenAndServe(":8080", nil)

	// Handle USB device add/remove
	r := hid.Registrar()
	AddSubscriber(r)

	// Create a new UDP connection to read telemetry
	// TODO this all needs to be configurable/optional/etc.
	s, err := NewTelemetry("") // Defaults to ":20777"
	if err != nil {
		logger.Println(err)
		os.Exit(-1)
	}

	// Loop receive packet from UDP client, and ship them off to HID and WebSocket
	// TODO this all needs to be configurable/optional/etc.
	rcv := make([]byte, codemasters.DirtPacketSize) // Maybe 1500 (standard frame)
	p := &codemasters.DirtPacket{}                  // TODO oh well obviously...
	for {
		// Retrieve a packet
		if err := s.DecodePacket(p, rcv); err != nil {
			logger.Println(err)
			break // TODO probably need better than this for error handling...
		}

		// Send data to websocket clients
		ws.SendPack(p)

		// Send data to any connected USB HID devices
		r.SendPack(p)
	}
}

package main

import (
	"log"
	"os"

	"github.com/jake-dog/opendash/codemasters"
)

var logger = log.New(os.Stdout, "", log.LstdFlags|log.LUTC|log.Lshortfile)

func main() {
	// Start up the HID debugger for our device
	// TODO this all needs to be configurable/optional/etc.
	hdebug := &HIDDebugger{
		VendorID:  0x16c0,
		ProductID: 0x0480,
		Usage:     0xFF31,
		UsagePage: 0x0074,
		Log:       logger,
	}
	go hdebug.ReadLoop()

	// Get our real HID Device
	// TODO this all needs to be configurable/optional/etc.
	c, err := GetDevice(0x16c0, 0x0480, 0xFFAB, 0x200)
	if err != nil {
		logger.Println(err)
		os.Exit(-1)
	}

	// Create a new UDP connection to read telemetry
	// TODO this all needs to be configurable/optional/etc.
	s, err := NewTelemetry("") // Defaults to ":20777"
	if err != nil {
		logger.Println(err)
		os.Exit(-1)
	}

	// Loop receive packet from UDP client, and ship them off to our HID Device
	// TODO this all needs to be configurable/optional/etc.
	rcv := make([]byte, codemasters.DirtPacketSize) // Maybe 1500 (standard frame)
	snd := make([]byte, 64)                         // HID can use other packet size maybe...
	packet := &codemasters.DirtPacket{}             // TODO oh well obviously...
	for {
		// Retrieve a packet
		if err := s.DecodePacket(packet, rcv); err != nil {
			logger.Println(err)
			break // TODO probably need better than this for error handling...
		}

		// Do something to convert it into an HID payload

		// Send the HID payload
		if _, err := c.Write(snd); err != nil {
			logger.Println(err)
			break // TODO probably need better than this for error handling...
		}
	}
}

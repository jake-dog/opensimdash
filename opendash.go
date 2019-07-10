package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/jake-dog/opendash/codemasters"
)

// Speed appears to be in meters per second (m/s), so convert to MPH
const mslashs float32 = 2.23694

var logger = log.New(os.Stdout, "", log.LstdFlags|log.LUTC|log.Lshortfile)

func setLeds(leds []byte, levels []int, signal int) {
	for i, level := range levels {
		if signal >= level {
			leds[i] = 1
		} else {
			leds[i] = 0
		}
	}
}

func main() {
	// TODO need to make this configurable, and catch errors it throws
	go http.ListenAndServe(":8080", nil)

	// Start up the HID debugger for our device
	// TODO this all needs to be configurable/optional/etc.
	hdebug := &HIDDebugger{
		VendorID:  0x16c0,
		ProductID: 0x0480,
		UsagePage: 0xFF31,
		Usage:     0x0074,
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
	p := &codemasters.DirtPacket{}                  // TODO oh well obviously...
	leds := make([]byte, 8)                         // Just for testing
	levels := []int{80, 83, 85, 87, 89, 91, 93, 95} // LED thresholds
	var ledByte byte
	for {
		// Retrieve a packet
		if err := s.DecodePacket(p, rcv); err != nil {
			logger.Println(err)
			break // TODO probably need better than this for error handling...
		}

		// Send data to websockets
		b, _ := json.Marshal(&dataPoint{
			Speed: int(p.Speed * mslashs),
			Gear:  int(p.Gear),
		})
		WriteMessage(b)

		// Do something to convert it into an HID payload
		// TODO this is just for testing
		rpmPercent := (100 * p.EngineRate) / p.Max_rpm
		setLeds(leds, levels, int(rpmPercent))
		ledByte = 0
		for i, b := range leds {
			ledByte |= b << uint(i)
		}
		snd[0] = ledByte

		// Send the HID payload
		if _, err := c.Write(snd); err != nil {
			logger.Println(err)
			break // TODO probably need better than this for error handling...
		}
	}
}

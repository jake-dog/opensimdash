package hid

import (
	"fmt"

	"github.com/karalabe/hid"
)

var (
	// Teensy is the regular teensy 2.0++.  Just sending rev light indicator using
	// 64-byte payloads, but it only sets the first byte to reflect the 8 LEDs
	Teensy = &teensy{
		snd:    make([]byte, 64),
		leds:   make([]byte, 8),
		levels: []int{80, 83, 85, 87, 89, 91, 93, 95},
		SimDashDevice: &SimDashDevice{
			DeviceInfo: &hid.DeviceInfo{
				VendorID:  0x16c0,
				ProductID: 0x0480,
				UsagePage: 0xFFAB,
				Usage:     0x200,
			},
		},
	}

	// TeensyDebug is a separate Usage/UsagePage possibly used for sending debug
	// messages.  It doesn't receive data, only sends it.
	TeensyDebug = &teensy{
		SimDashDevice: &SimDashDevice{
			DeviceInfo: &hid.DeviceInfo{
				VendorID:  0x16c0,
				ProductID: 0x0480,
				UsagePage: 0xFF31,
				Usage:     0x0074,
			},
		},
	}
)

func init() {
	Register(Teensy)
	//Register(TeensyDebug)
}

type teensy struct {
	snd         []byte
	leds        []byte
	levels      []int
	ledByte     byte
	prevLedByte byte
	*SimDashDevice
}

func setLeds(leds []byte, levels []int, signal int) {
	for i, level := range levels {
		if signal >= level {
			leds[i] = 1
		} else {
			leds[i] = 0
		}
	}
}

func (t *teensy) Write(p TelemetryPack) {
	// Compute which of the eight LEDs to turn on based on the revLightPercent
	setLeds(t.leds, t.levels, p.GetRevLightPercent())
	t.ledByte = 0
	for i, b := range t.leds {
		t.ledByte |= b << uint(i)
	}
	t.snd[0] = t.ledByte

	// Skip sending the HID payload if current and last byte are zero
	if t.prevLedByte|t.ledByte == 0 {
		return
	}
	t.prevLedByte = t.ledByte
	if _, err := t.Device.Write(t.snd); err != nil {
		// TODO figure out the logging . . .
		fmt.Println(err)
	}
}

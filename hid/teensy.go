package hid

import (
	"fmt"

	"github.com/karalabe/hid"
)

func init() {
	// Teensy is the regular teensy 2.0++.  Just sending rev light indicator using
	// 64-byte payloads, but it only sets the first byte to reflect the 8 LEDs
	Register(&teensy{
		snd:    make([]byte, 64),
		levels: []int{80, 83, 85, 87, 89, 91, 93, 95},
		SimDashDevice: &SimDashDevice{
			DeviceInfo: &hid.DeviceInfo{
				VendorID:  0x16c0,
				ProductID: 0x0480,
				UsagePage: 0xFFAB,
				Usage:     0x200,
			},
		},
	})

	// TeensyDebug is a separate Usage/UsagePage possibly used for sending debug
	// messages.  It doesn't receive data, only sends it.
	/*Register(&teensy{
		SimDashDevice: &SimDashDevice{
			DeviceInfo: &hid.DeviceInfo{
				VendorID:  0x16c0,
				ProductID: 0x0480,
				UsagePage: 0xFF31,
				Usage:     0x0074,
			},
		},
	})*/
}

type teensy struct {
	snd     []byte
	levels  []int
	ledByte byte
	*SimDashDevice
}

func (t *teensy) Write(p TelemetryPack) {
	// Compute which of the eight LEDs to turn on based on the revLightPercent
	revLights := p.GetRevLightPercent()
	t.ledByte = 0
	for i, level := range t.levels {
		if revLights >= level {
			t.ledByte |= 1 << uint(i)
		}
	}

	// Skip sending the HID payload if current and last byte are zero
	if t.snd[0]|t.ledByte == 0 {
		return
	}
	t.snd[0] = t.ledByte

	if _, err := t.Device.Write(t.snd); err != nil {
		// TODO figure out the logging . . .
		fmt.Println(err)
	}
}

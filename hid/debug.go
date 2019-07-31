package hid

import (
	"log"

	"github.com/karalabe/hid"
)

// HIDDebugger simplifies creating a HID device connection which simply reads
// bytes and prints them to a logger.
type HIDDebugger struct {
	Device *hid.Device
	Log    *log.Logger
}

func (d *HIDDebugger) logf(format string, args ...interface{}) {
	if d.Log != nil {
		d.Log.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

// ReadLoop simply loops read from the HID device and writes to a logger.  This
// call is blocking.
func (d *HIDDebugger) ReadLoop() {
	data := make([]byte, 64) // TODO packet doesn't need to be 64 bytes
	for {
		if i, err := d.Device.Read(data); err != nil {
			d.logf("ERROR reading from debug USB device %s", err)
			return
		} else {
			d.logf(string(data[:i]))
		}
	}
}

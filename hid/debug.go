package hid

import (
	"bufio"
	"log"

	"github.com/karalabe/hid"
)

// Debugger simplifies creating a HID device connection which simply reads
// bytes and prints them to a logger.
type Debugger struct {
	Device *hid.Device
	Log    *log.Logger
}

func (d *Debugger) logf(format string, args ...interface{}) {
	if d.Log != nil {
		d.Log.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

// ReadLoop simply loops read from the HID device and writes to a logger.  This
// call is blocking.
func (d *Debugger) ReadLoop() {
	scanner := bufio.NewScanner(d.Device)
	for scanner.Scan() {
		d.logf(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		d.logf("Aborting debugging of USB device : %v", err)
	}
}

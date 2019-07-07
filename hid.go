package main

import (
	"errors"
	"log"

	"github.com/karalabe/hid"
)

// HIDDebugger simplifies creating a HID device connection which simply reads
// bytes and prints them to a logger.
type HIDDebugger struct {
	VendorID  uint16
	ProductID uint16
	UsagePage uint16
	Usage     uint16
	Log       *log.Logger
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
	dev, err := GetDevice(d.VendorID, d.ProductID, d.UsagePage, d.Usage)
	if err != nil {
		d.logf("ERROR unable to locate debug USB device VID=%d PID=%d", d.VendorID, d.ProductID)
		return
	}
	data := make([]byte, 64) // TODO packet doesn't need to be 64 bytes
	for {
		if i, err := dev.Read(data); err != nil {
			d.logf("ERROR reading from debug USB device %s", err)
			return
		} else {
			d.logf(string(data[:i]))
		}
	}
}

// GetDevice retrieves an HID device as identified by the usual HID device info
func GetDevice(vid, pid, usagepage, usage uint16) (*hid.Device, error) {
	for _, dev := range hid.Enumerate(vid, pid) {
		if dev.UsagePage == usagepage && dev.Usage == usage {
			return dev.Open()
		}
	}
	return nil, errors.New("Unable to find device")
}

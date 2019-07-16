package hid

import (
	"fmt"
	"io"
	"sync"

	"github.com/karalabe/hid"
)

// HIDWriter is a generic interface for a USB HID allowing writing
type HIDWriter interface {
	Write(TelemetryPack)
	GetDeviceInfo() *hid.DeviceInfo
	GetDevice() io.WriteCloser
	SetDevice(io.WriteCloser)
	Equals(*hid.DeviceInfo) bool
}

type SimDashDevice struct {
	Device     io.WriteCloser
	DeviceInfo *hid.DeviceInfo
}

func (d *SimDashDevice) SetDevice(dev io.WriteCloser) {
	d.Device = dev
}

func (d *SimDashDevice) GetDevice() io.WriteCloser {
	return d.Device
}

func (d *SimDashDevice) GetDeviceInfo() *hid.DeviceInfo {
	return d.DeviceInfo
}

func (d *SimDashDevice) Equals(h *hid.DeviceInfo) bool {
	if d.DeviceInfo.VendorID == h.VendorID &&
		d.DeviceInfo.ProductID == h.ProductID &&
		d.DeviceInfo.UsagePage == h.UsagePage &&
		d.DeviceInfo.Usage == h.Usage {
		return true
	}
	return false
}

// HIDRegistrar fulfills the UsbDeviceNotifier interface, but adds Write method
type HIDRegistrar interface {
	Write(TelemetryPack)
	Add(uintptr)
	Remove(uintptr)
}

type registrar struct {
	mu      sync.Mutex
	devices []HIDWriter
	writers []HIDWriter
}

var r = &registrar{}

func Register(d HIDWriter) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.devices = append(r.devices, d)
}

func Registrar() HIDRegistrar {
	return r
}

func (r *registrar) Write(p TelemetryPack) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, d := range r.writers {
		d.Write(p)
	}
}

func (r *registrar) Remove(_ uintptr) {
	devices := hid.Enumerate(0, 0)

	r.mu.Lock()
	defer r.mu.Unlock()

	// Remove any writers that aren't connected
	var i int
	for _, conn := range r.writers {
		var found bool
		for _, d := range devices {
			// If device is still connected, make sure its in the writers array
			if conn.Equals(&d) {
				r.writers[i] = conn
				i++
				found = true
			}
		}
		// Attempt to close device and remove all record of it
		if !found && conn.GetDevice() != nil {
			conn.GetDevice().Close()
			conn.SetDevice(nil)
		}
	}
	r.writers = r.writers[:i]
}

func (r *registrar) Add(_ uintptr) {
	devices := hid.Enumerate(0, 0)

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check which writers are supported
	for _, dev := range r.devices {
		for _, d := range devices {
			if dev.Equals(&d) && dev.GetDevice() == nil {
				if device, err := d.Open(); err != nil {
					fmt.Printf("Unable to open device %v\n", d)
				} else {
					dev.SetDevice(device)
					r.writers = append(r.writers, dev)
				}
			}
		}
	}
}

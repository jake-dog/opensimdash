package hid

import (
	"fmt"
	"io"
	"sync"

	"github.com/karalabe/hid"
)

// PackSender is a generic interface for a USB HID allowing writing
type PackSender interface {
	// SendPack to the user supplied code so that it can be converted to device
	// specific []byte, then sent to the device via provided Write method.
	SendPack(TelemetryPack)

	// Sealed methods only implemented by SimDashDevice
	getDevice() io.WriteCloser
	setDevice(io.WriteCloser)
	equals(*hid.DeviceInfo) bool
}

type SimDashDevice struct {
	VendorID  uint16
	ProductID uint16
	UsagePage uint16
	Usage     uint16

	// TODO allow multiple instances of the same device
	device io.WriteCloser
}

func (d *SimDashDevice) Write(p []byte) (int, error) {
	return d.device.Write(p)
}

func (d *SimDashDevice) setDevice(dev io.WriteCloser) {
	d.device = dev
}

func (d *SimDashDevice) getDevice() io.WriteCloser {
	return d.device
}

func (d *SimDashDevice) equals(h *hid.DeviceInfo) bool {
	if d.VendorID == h.VendorID &&
		d.ProductID == h.ProductID &&
		d.UsagePage == h.UsagePage &&
		d.Usage == h.Usage {
		return true
	}
	return false
}

// HIDRegistrar fulfills the UsbDeviceNotifier interface, but adds Write method
type HIDRegistrar interface {
	SendPack(TelemetryPack)
	Add(uintptr)
	Remove(uintptr)
}

type registrar struct {
	once    sync.Once
	mu      sync.Mutex
	devices []PackSender
	writers []PackSender
}

var r = &registrar{}

func Register(d PackSender) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.devices = append(r.devices, d)
}

// Registrar returns the HIDRegistrar which fullfills the UsbDeviceNotifier
// interface and is intended to receive notifications on device changes via
// WM_DEVICECHANGE messages.  Any HID devices which are detected can be written
// to using the Write method.
func Registrar() HIDRegistrar {
	// Add all recognized devices the first time the registrar is invoked
	r.once.Do(func() { r.Add(uintptr(0)) })
	return r
}

func (r *registrar) SendPack(p TelemetryPack) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, d := range r.writers {
		d.SendPack(p)
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
			if conn.equals(&d) {
				r.writers[i] = conn
				i++
				found = true
			}
		}
		// Attempt to close device and remove all record of it
		if !found && conn.getDevice() != nil {
			conn.getDevice().Close()
			conn.setDevice(nil)
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
			if dev.equals(&d) && dev.getDevice() == nil {
				if device, err := d.Open(); err != nil {
					fmt.Printf("Unable to open device %v\n", d)
				} else {
					dev.setDevice(device)
					r.writers = append(r.writers, dev)
				}
			}
		}
	}
}

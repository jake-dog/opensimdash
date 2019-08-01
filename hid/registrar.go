package hid

import (
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/karalabe/hid"
)

// PackSender is a generic interface for a writing telemetry packs
type PackSender interface {
	// SendPack to the user supplied code so that it can be converted to device
	// specific []byte, then sent to the device via provided Write method.
	SendPack(TelemetryPack)
}

// HIDPackSender has some sealed methods added to PackSender for device tracking
type HIDPackSender interface {
	PackSender

	// Sealed methods only implemented by SimDashDevice
	getDevice() io.WriteCloser
	setDevice(io.WriteCloser)
	equals(*hid.DeviceInfo) bool
	debug() bool // TODO change to an enumerated type to allow more device types
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

func (d *SimDashDevice) String() string {
	return fmt.Sprintf(
		"VID=%d PID=%d UsagePage=%d Usage=%d",
		d.VendorID,
		d.ProductID,
		d.UsagePage,
		d.Usage)
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

func (d *SimDashDevice) debug() bool {
	return false
}

// DebugDevice is the same as SimDashDevice but telmetry is not sent to it,
// instead an infinite loop reads messages from the device and logs the output.
type DebugDevice struct {
	PackSender
	*SimDashDevice
}

func (d *DebugDevice) debug() bool {
	return true
}

// HIDRegistrar fulfills UsbDeviceNotifier interface but adds SendPack method
type HIDRegistrar interface {
	PackSender

	Add(uintptr)
	Remove(uintptr)
}

type registrar struct {
	logger  *log.Logger // TODO probably better to use an interface
	once    sync.Once
	mu      sync.Mutex
	devices []HIDPackSender
	writers []HIDPackSender
}

var r = &registrar{}

func Register(d HIDPackSender) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.devices = append(r.devices, d)
}

// Registrar returns the HIDRegistrar which fullfills the UsbDeviceNotifier
// interface and is intended to receive notifications on device changes via
// WM_DEVICECHANGE messages.  Any HID devices which are detected can be written
// to using the Write method.
func Registrar(logger *log.Logger) HIDRegistrar {
	// Add all recognized devices the first time the registrar is invoked
	r.once.Do(func() {
		r.logger = logger
		r.Add(uintptr(0))
	})
	return r
}

func (r *registrar) logf(format string, args ...interface{}) {
	if r.logger != nil {
		r.logger.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
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
		for _, d := range devices {
			// If device is still connected, make sure its in the writers array
			if conn.equals(&d) {
				r.writers[i] = conn
				i++
			}
		}
	}
	r.writers = r.writers[:i]

	// Close devices that are disconnected
	for _, dev := range r.devices {
		var found bool
		for _, d := range devices {
			if dev.equals(&d) {
				found = true
			}
		}

		if !found && dev.getDevice() != nil {
			r.logf("HID device disconnected : %v", dev)
			dev.getDevice().Close()
			dev.setDevice(nil)
		}
	}
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
					r.logf("%v : %v", err, dev)
				} else if dev.debug() {
					r.logf("HID debug device connected : %v", dev)
					dev.setDevice(device)
					debugger := &Debugger{
						Device: device,
						Log:    r.logger,
					}
					go debugger.ReadLoop()
				} else {
					r.logf("HID telemetry device connected : %v", dev)
					dev.setDevice(device)
					r.writers = append(r.writers, dev)
				}
			}
		}
	}
}

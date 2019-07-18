package hid

// TelemetryPack is the interface for a structure sent to the SendPack method of
// an HID device which fullfills the PackSender interface.  It contains methods
// which can be used to generate a binary payload to be sent to an HID device.
type TelemetryPack interface {
	GetGear() int
	GetRevLightPercent() int
	GetSpeed() int
}

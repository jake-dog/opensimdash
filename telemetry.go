package main

import (
	"encoding/binary"
	"net"
)

// Decodable interface is a high performance interface for parsing binary
// structs.  Structs which implement Decodable (with custom parsing) can be
// several times faster than using binary.Read()
type Decodable interface {
	Decode(b []byte)
	Size() int
}

// Telemetry wraps net.UDPConn providing extra methods for parsing UDP telemetry
type Telemetry struct {
	*net.UDPConn
}

// NewTelemetry returns a new Telemetry UDP connection, wrapping *net.UDPConn,
// which can rapidly process various telemetry packets.
// If provided address is empty, ":20777", will be used.
func NewTelemetry(address string) (*Telemetry, error) {
	addr := address
	if addr == "" {
		// Default port for codemasters
		addr = ":20777"
	}
	conn, err := net.ListenPacket("udp", addr) // Convenience function
	if err != nil {
		return nil, err
	}
	// The ListenPacket interface sucks, so just convert back to net.UDPConn
	return &Telemetry{conn.(*net.UDPConn)}, nil
}

// DecodePacket from client's UDP stream using optional buffer.  If provided
// buffer is nil, then a byte array will be allocated to read data which will
// result in unnecessary memory allocations.  In all cases it is preferred that
// a buffer is provided to avoid memory allocations.
func (c *Telemetry) DecodePacket(d Decodable, buf []byte) error {
	b := buf
	if b == nil {
		b = make([]byte, d.Size())
	}
	if _, err := c.Read(b); err != nil {
		return err
	}
	d.Decode(b)
	return nil
}

// ReadPacket inefficiently from the UDP stream and attempt to decode it into
// the provided abstract interface.  DecodePacket is preferred over this
// function as it avoids unnecessary allocations and reflection.
func (c *Telemetry) ReadPacket(data interface{}) error {
	if err := binary.Read(c, binary.LittleEndian, &data); err != nil {
		return err
	}
	return nil
}

// TODO Detect the packet type by retrieving the first packet from client
// and checking length, syncword, and decode.  Not sure what the return type
// should be on this.  Maybe an enumeration so that client can create relevant
// struct?
//func (c *Client) Detect() interface{} {}

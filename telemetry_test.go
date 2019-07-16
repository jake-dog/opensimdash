package main

import "testing"

func TestNewTelemetry(t *testing.T) {
	c, err := NewTelemetry("127.0.0.1:20777")
	if err != nil {
		t.Error(err)
	}
	if c == nil {
		t.Error("Client is nil")
	}
}

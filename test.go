package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/karalabe/hid"
)

func getDevice() (*hid.Device, error) {
	for _, dev := range hid.Enumerate(0x16c0, 0x0480) {
		if dev.UsagePage == 0xFFAB && dev.Usage == 0x200 {
			return dev.Open()
		}
	}
	return nil, errors.New("Unable to find device")
}

func main() {
	c, err := getDevice()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	data := make([]byte, 64)
	sndto := make([]byte, 64)
	if i, err := c.Read(data); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(i)
		fmt.Println(data[:i])
	}

	sndto[0] = 103 // "g" in decimal
	//sndto[0] = 99 // "c" in decimal
	fmt.Println(sndto)
	if i, err := c.Write(sndto); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(i)
	}
}

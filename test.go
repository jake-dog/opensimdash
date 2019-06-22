package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/karalabe/hid"
	"github.com/zetamatta/go-getch"
)

func getDevice() (*hid.Device, error) {
	for _, dev := range hid.Enumerate(0x16c0, 0x0480) {
		if dev.UsagePage == 0xFFAB && dev.Usage == 0x200 {
			return dev.Open()
		}
	}
	return nil, errors.New("Unable to find device")
}

func debugger() {
	for _, dev := range hid.Enumerate(0x16c0, 0x0480) {
		if dev.UsagePage == 0xFF31 && dev.Usage == 0x0074 {
			c, err := dev.Open()
			if err != nil {
				fmt.Println(err)
				return
			}
			data := make([]byte, 64)
			for {
				if i, err := c.Read(data); err != nil {
					fmt.Println(err)
					break
				} else {
					fmt.Println("Debugger read: ", i)
					fmt.Println(string(data[:i]))
				}
			}
		}
	}
	fmt.Println("Unable to find debugger")
}

func main() {
	go debugger()

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

	for v := 0; v < 10; v++ {
		e := getch.Rune()
		setting := byte(e & 0xFF)
		if e >= 32 {
			fmt.Println("received char: ", e, setting)
			sndto[0] = setting
			if i, err := c.Write(sndto); err != nil {
				fmt.Println(err)
				break
			} else {
				fmt.Println(i)
			}
		}

		//sndto[0] = e
		//sndto[0] = 103 // "g" in decimal
		//sndto[0] = 99 // "c" in decimal
		//fmt.Println(sndto)
	}
}

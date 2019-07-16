package main

import (
	"fmt"
	"sync"
	"syscall"
	"unsafe"
)

const (
	// https://docs.microsoft.com/en-us/windows/win32/devio/wm-devicechange
	// https://godoc.org/github.com/AllenDang/w32#WM_DEVICECHANGE
	WM_DEVICECHANGE = 537

	// https://godoc.org/github.com/AllenDang/w32#HWND_MESSAGE
	HWND_MESSAGE = ^uintptr(2)

	// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-registerdevicenotificationa#device_notify_all_interface_classes
	DEVICE_NOTIFY_ALL_INTERFACE_CLASSES = 4

	// https://docs.microsoft.com/en-us/windows/win32/api/dbt/ns-dbt-_dev_broadcast_hdr#DBT_DEVTYP_DEVICEINTERFACE
	DBT_DEVTYP_DEVICEINTERFACE = 5

	// https://docs.microsoft.com/en-us/windows/win32/devio/wm-devicechange#DBT_DEVICEARRIVAL
	DBT_DEVICEARRIVAL = 0x8000

	// https://docs.microsoft.com/en-us/windows/win32/devio/wm-devicechange#DBT_DEVICEREMOVECOMPLETE
	DBT_DEVICEREMOVECOMPLETE = 0x8004
)

const (
	// Used only for clients listening for USB device events
	addUSBDevice = iota
	removeUSBDevice
)

var (
	user32                      = syscall.NewLazyDLL("user32.dll")
	kernel32                    = syscall.NewLazyDLL("kernel32.dll")
	pDefWindowProc              = user32.NewProc("DefWindowProcW")
	pCreateWindowEx             = user32.NewProc("CreateWindowExW")
	pGetModuleHandle            = kernel32.NewProc("GetModuleHandleW")
	pRegisterClassEx            = user32.NewProc("RegisterClassExW")
	pGetMessage                 = user32.NewProc("GetMessageW")
	pDispatchMessage            = user32.NewProc("DispatchMessageW")
	pRegisterDeviceNotification = user32.NewProc("RegisterDeviceNotificationW")
)

// Keep track of who is publishing
type publisher struct {
	mu          sync.Mutex
	subscribers []UsbDeviceNotifier
}

var pub = &publisher{}

func (p *publisher) addSubscriber(sub UsbDeviceNotifier) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.subscribers = append(p.subscribers, sub)
}

func (p *publisher) notify(method int, lParam uintptr) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, sub := range p.subscribers {
		switch method {
		case addUSBDevice:
			sub.Add(lParam)
		case removeUSBDevice:
			sub.Remove(lParam)
		}
	}
}

// AddSubscriber to the list of USB device notification subscribers.
func AddSubscriber(sub UsbDeviceNotifier) {
	pub.addSubscriber(sub)
}

// UsbDeviceNotifier is an interface for a resource to be notified of USB adds
// and removals
type UsbDeviceNotifier interface {
	Add(uintptr)    // Called on DBT_DEVICEARRIVAL
	Remove(uintptr) // Called on DBT_DEVICEREMOVECOMPLETE
}

// https://www.lifewire.com/device-class-guids-for-most-common-types-of-hardware-2619208
// 745A17A0-74D3-11D0-B6FE-00A0C90F57DA
var HID_DEVICE_CLASS = GUID{
	0x745a17a0,
	0x74d3,
	0x11d0,
	[8]byte{0xb6, 0xfe, 0x00, 0xa0, 0xc9, 0x0f, 0x57, 0xda},
}

// https://docs.microsoft.com/en-us/windows-hardware/drivers/install/guid-devinterface-usb-device
// A5DCBF10-6530-11D2-901F-00C04FB951ED
var GUID_DEVINTERFACE_USB_DEVICE = GUID{
	0xa5dcbf10,
	0x6530,
	0x11d2,
	[8]byte{0x90, 0x1f, 0x00, 0xc0, 0x4f, 0xb9, 0x51, 0xed},
}

// https://docs.microsoft.com/en-us/previous-versions//dd162805(v=vs.85)
type POINT struct {
	x uintptr
	y uintptr
}

// https://docs.microsoft.com/en-us/windows/win32/api/winuser/ns-winuser-tagmsg
type MSG struct {
	hWnd    syscall.Handle
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      POINT
}

// https://docs.microsoft.com/en-us/previous-versions/aa373931(v=vs.80)
type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

// https://docs.microsoft.com/en-us/windows/win32/api/dbt/ns-dbt-_dev_broadcast_deviceinterface_a
type DevBroadcastDevinterface struct {
	dwSize       uint32
	dwDeviceType uint32
	dwReserved   uint32
	classGuid    GUID
	szName       uint16
}

// https://docs.microsoft.com/en-us/windows/win32/api/winuser/ns-winuser-tagwndclassexa
// https://golang.org/src/runtime/syscall_windows_test.go
type Wndclassex struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   syscall.Handle
	Icon       syscall.Handle
	Cursor     syscall.Handle
	Background syscall.Handle
	MenuName   *uint16
	ClassName  *uint16
	IconSm     syscall.Handle
}

// https://docs.microsoft.com/en-us/previous-versions/windows/desktop/legacy/ms633573(v=vs.85)
func WndProc(hWnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	// TODO handle WM_DESTROY and deregister the hwnd class
	case WM_DEVICECHANGE:
		switch wParam {
		case uintptr(DBT_DEVICEARRIVAL):
			pub.notify(addUSBDevice, lParam)
		case uintptr(DBT_DEVICEREMOVECOMPLETE):
			pub.notify(removeUSBDevice, lParam)
		}
		return 0
	default:
		fmt.Println("doing default", msg)
		ret, _, _ := pDefWindowProc.Call(uintptr(hWnd), uintptr(msg), uintptr(wParam), uintptr(lParam))
		return ret
	}
}

func init() {
	// TODO clean this up a bit
	// The whole thing needs to be run in a single scope/closure otherwise golang
	// will GC all the structs and the message window will not work.
	go func() {
		// Create callback
		cb := syscall.NewCallback(WndProc)
		mh, _, _ := pGetModuleHandle.Call(0)

		// Create a class and window name
		lpClassName := syscall.StringToUTF16Ptr("opensimdash")
		lpWindowName := syscall.StringToUTF16Ptr("opensimdash")

		// Register our invisible window class
		// Code from: https://golang.org/src/runtime/syscall_windows_test.go
		wc := Wndclassex{
			WndProc:   cb,
			Instance:  syscall.Handle(mh),
			ClassName: lpClassName,
		}
		wc.Size = uint32(unsafe.Sizeof(wc))
		a, _, err := pRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))
		if a == 0 {
			logger.Printf("RegisterClassEx failed: %v", err)
			return
		}

		// Create a message only window
		// https://docs.microsoft.com/en-us/windows/win32/winmsg/window-features#message-only-windows
		// https://stackoverflow.com/a/4081383
		ret, _, err := pCreateWindowEx.Call(
			uintptr(0),                            //dwExStyle
			uintptr(unsafe.Pointer(lpClassName)),  //lpClassName
			uintptr(unsafe.Pointer(lpWindowName)), //lpWindowName
			uintptr(0),                            //dwStyle
			uintptr(0),                            //X
			uintptr(0),                            //Y
			uintptr(0),                            //nWidth
			uintptr(0),                            //nHeight
			HWND_MESSAGE,                          //hWndParent
			uintptr(0),                            //hMenu
			uintptr(0),                            //hInstance
			uintptr(0))                            //lpParam

		if ret == 0 {
			logger.Printf("CreateWindowEx failed: %v", err)
			return
		}
		hWnd := syscall.Handle(ret)

		// Register for device notifications
		// https://github.com/google/cloud-print-connector/blob/master/winspool/win32.go
		// https://www.lifewire.com/device-class-guids-for-most-common-types-of-hardware-2619208
		var notificationFilter DevBroadcastDevinterface
		notificationFilter.dwSize = uint32(unsafe.Sizeof(notificationFilter))
		notificationFilter.dwDeviceType = DBT_DEVTYP_DEVICEINTERFACE
		notificationFilter.dwReserved = 0
		notificationFilter.classGuid = HID_DEVICE_CLASS
		notificationFilter.szName = 0
		ret, _, err = pRegisterDeviceNotification.Call(uintptr(hWnd), uintptr(unsafe.Pointer(&notificationFilter)), DEVICE_NOTIFY_ALL_INTERFACE_CLASSES)
		if ret == 0 {
			logger.Printf("RegisterDeviceNotification failed: %v", err)
			return
		}

		// If we made it here, start the main message loop
		var msg MSG
		for {
			if ret, _, _ := pGetMessage.Call(uintptr(unsafe.Pointer(&msg)), uintptr(0), uintptr(0), uintptr(0)); ret == 0 {
				break
			}
			pDispatchMessage.Call((uintptr(unsafe.Pointer(&msg))))
		}
	}()
}

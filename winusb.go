package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	WM_DESTROY      = 2
	WM_PAINT        = 15
	WM_QUIT         = 18
	WM_COMMAND      = 273
	WM_DEVICECHANGE = 537

	HWND_MESSAGE = ^uintptr(2)

	DEVICE_NOTIFY_SERVICE_HANDLE        = 1
	DEVICE_NOTIFY_ALL_INTERFACE_CLASSES = 4

	DBT_DEVTYP_DEVICEINTERFACE = 5
)

var (
	user32                      = syscall.NewLazyDLL("user32.dll")
	kernel32                    = syscall.NewLazyDLL("kernel32.dll")
	pDefWindowProc              = user32.NewProc("DefWindowProcW")
	pCreateWindowEx             = user32.NewProc("CreateWindowExW")
	pGetModuleHandle            = kernel32.NewProc("GetModuleHandleW")
	pRegisterClassEx            = user32.NewProc("RegisterClassExW")
	pPostQuitMessage            = user32.NewProc("PostQuitMessage")
	pGetMessage                 = user32.NewProc("GetMessageW")
	pDispatchMessage            = user32.NewProc("DispatchMessageW")
	pRegisterDeviceNotification = user32.NewProc("RegisterDeviceNotificationW")
	pUpdateWindow               = user32.NewProc("UpdateWindow")
)

// https://www.lifewire.com/device-class-guids-for-most-common-types-of-hardware-2619208
// 745A17A0-74D3-11D0-B6FE-00A0C90F57DA
var HID_DEVICE_CLASS = GUID{
	0x745a17a0,
	0x74d3,
	0x11d0,
	[8]byte{0xb6, 0xfe, 0x00, 0xa0, 0xc9, 0x0f, 0x57, 0xda},
}

type POINT struct {
	x uintptr
	y uintptr
}

type MSG struct {
	hWnd    syscall.Handle
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      POINT
}

type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type DevBroadcastDevinterface struct {
	dwSize       uint32
	dwDeviceType uint32
	dwReserved   uint32
	classGuid    GUID
	szName       uint16
}

func WndProc(hWnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_PAINT:
		fmt.Println("painting")
		return 0
	case WM_COMMAND:
		fmt.Println("a command?")
		return 0
	case WM_DESTROY:
		fmt.Println("quitting")
		pPostQuitMessage.Call(uintptr(0))
		return 0
	case WM_DEVICECHANGE:
		fmt.Println("device changed!")
		return 0
	default:
		fmt.Println("doing default", msg)
		ret, _, _ := pDefWindowProc.Call(uintptr(hWnd), uintptr(msg), uintptr(wParam), uintptr(lParam))
		return ret
	}
	return 0
}

func main() {
	// Create callback
	// https://stackoverflow.com/questions/2122506/how-to-create-a-hidden-window-in-c
	cb := syscall.NewCallback(WndProc)
	mh, _, _ := pGetModuleHandle.Call(0)

	// Create a class and window name
	lpClassName := syscall.StringToUTF16Ptr("opensimdash")
	lpWindowName := syscall.StringToUTF16Ptr("opensimdash")

	// Register our invisible window class
	// From TestRegisterClass:
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
	wc := Wndclassex{
		WndProc:   cb,
		Instance:  syscall.Handle(mh),
		ClassName: lpClassName,
	}
	wc.Size = uint32(unsafe.Sizeof(wc))
	a, _, err := pRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))
	if a == 0 {
		fmt.Println("RegisterClassEx failed: %v", err)
		os.Exit(1)
	}

	// Create a message only window
	// https://docs.microsoft.com/en-us/windows/win32/winmsg/window-features#message-only-windows
	// CreateWindowEx( 0, class_name, "dummy_name", 0, 0, 0, 0, 0, HWND_MESSAGE, NULL, NULL, NULL );
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
		fmt.Println("Unable to create a window: ", err)
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
		fmt.Println("Unable to register for USB notifications: ", err)
	}

	//pUpdateWindow.Call(uintptr(hWnd))

	// Main message loop
	var msg MSG
	for {
		if ret, _, _ := pGetMessage.Call(uintptr(unsafe.Pointer(&msg)), uintptr(0), uintptr(0), uintptr(0)); ret == 0 {
			break
		}
		pDispatchMessage.Call((uintptr(unsafe.Pointer(&msg))))
	}
}

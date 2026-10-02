//go:build windows

package main

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	swRestore     = 9
	hwndTop       = 0
	swpNoSize     = 0x0001
	swpNoMove     = 0x0002
	swpShowWindow = 0x0040
)

var (
	user32              = windows.NewLazySystemDLL("user32.dll")
	findWindow          = user32.NewProc("FindWindowW")
	showWindow          = user32.NewProc("ShowWindow")
	setWindowPos        = user32.NewProc("SetWindowPos")
	bringWindowToTop    = user32.NewProc("BringWindowToTop")
	setForegroundWindow = user32.NewProc("SetForegroundWindow")
)

func activateMainWindow() {
	title, err := windows.UTF16PtrFromString(appTitle)
	if err != nil {
		return
	}

	hwnd, _, _ := findWindow.Call(0, uintptr(unsafe.Pointer(title)))
	if hwnd == 0 {
		return
	}

	showWindow.Call(hwnd, swRestore)
	setWindowPos.Call(hwnd, hwndTop, 0, 0, 0, 0, swpNoMove|swpNoSize|swpShowWindow)
	bringWindowToTop.Call(hwnd)
	setForegroundWindow.Call(hwnd)
}

//go:build windows

package logger

import (
	"syscall"
	"unsafe"
)

// isatty is the Windows implementation. We call GetConsoleMode via
// syscall.Syscall6 against kernel32.dll. A non-zero return means the
// handle is a console.
func isatty(fd uintptr) bool {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GetConsoleMode")

	var mode uint32
	// BOOL GetConsoleMode(HANDLE hConsoleHandle, LPDWORD lpMode);
	r1, _, _ := proc.Call(
		fd,
		uintptr(unsafe.Pointer(&mode)),
	)
	return r1 != 0
}

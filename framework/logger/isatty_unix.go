//go:build unix || linux || darwin

package logger

import (
	"syscall"
	"unsafe"
)

// isatty is the unix implementation of the TTY check. It issues the
// TIOCGWINSZ ioctl against the file descriptor; a zero errno means the
// fd is attached to a terminal. (We deliberately avoid
// term.IsTerminal from x/term to keep the logger package
// dependency-free.)
func isatty(fd uintptr) bool {
	// TIOCGWINSZ — see termios(3) / ioctl_console(2).
	const tiocgwinsz = 0x5413

	var ws [4]uint16 // struct winsize { ws_row, ws_col, ws_xpixel, ws_ypixel }
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		uintptr(tiocgwinsz),
		uintptr(unsafe.Pointer(&ws[0])),
	)
	return errno == 0
}

//go:build unix

package terminal

import (
	"syscall"
	"unsafe"
)

func terminalQuerySize(target any, fallbackWidth int, fallbackHeight int) (int, int) {
	file, ok := target.(terminalFdReader)
	if !ok {
		return fallbackWidth, fallbackHeight
	}
	var size struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}
	if err := terminalIoctl(file.Fd(), syscall.TIOCGWINSZ, unsafe.Pointer(&size)); err != nil {
		return fallbackWidth, fallbackHeight
	}
	if size.Col == 0 || size.Row == 0 {
		return fallbackWidth, fallbackHeight
	}
	return int(size.Col), int(size.Row)
}

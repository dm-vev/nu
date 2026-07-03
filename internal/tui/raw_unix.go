//go:build unix

package tui

import (
	"fmt"
	"io"
	"syscall"
	"unsafe"
)

type fdReader interface {
	Fd() uintptr
}

func enableRawInput(reader io.Reader) (func() error, bool, error) {
	file, ok := reader.(fdReader)
	if !ok {
		return nil, false, nil
	}
	fd := file.Fd()
	var oldState syscall.Termios
	if err := ioctlTermios(fd, syscall.TCGETS, &oldState); err != nil {
		if err == syscall.ENOTTY || err == syscall.EINVAL {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("read terminal state: %w", err)
	}

	rawState := oldState
	rawState.Iflag &^= syscall.ICRNL | syscall.IXON
	rawState.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG
	rawState.Cflag |= syscall.CS8
	rawState.Cc[syscall.VMIN] = 1
	rawState.Cc[syscall.VTIME] = 0
	if err := ioctlTermios(fd, syscall.TCSETS, &rawState); err != nil {
		return nil, false, fmt.Errorf("enable raw terminal: %w", err)
	}

	restore := func() error {
		if err := ioctlTermios(fd, syscall.TCSETS, &oldState); err != nil {
			return fmt.Errorf("restore terminal state: %w", err)
		}
		return nil
	}
	return restore, true, nil
}

func terminalSize(target any, fallbackWidth int, fallbackHeight int) (int, int) {
	file, ok := target.(fdReader)
	if !ok {
		return fallbackWidth, fallbackHeight
	}
	var size struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}
	if err := ioctl(file.Fd(), syscall.TIOCGWINSZ, unsafe.Pointer(&size)); err != nil {
		return fallbackWidth, fallbackHeight
	}
	if size.Col == 0 || size.Row == 0 {
		return fallbackWidth, fallbackHeight
	}
	return int(size.Col), int(size.Row)
}

func ioctlTermios(fd uintptr, request uintptr, state *syscall.Termios) error {
	return ioctl(fd, request, unsafe.Pointer(state))
}

func ioctl(fd uintptr, request uintptr, pointer unsafe.Pointer) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, request, uintptr(pointer))
	if errno != 0 {
		return errno
	}
	return nil
}

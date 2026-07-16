//go:build unix

package terminal

import (
	"fmt"
	"syscall"
	"unsafe"
)

type terminalFdReader interface {
	Fd() uintptr
}

// EnableRaw enables raw mode when stdin is a TTY.
func (t *Terminal) EnableRaw() (func() error, bool, error) {
	file, ok := t.stdin.(terminalFdReader)
	if !ok {
		return nil, false, nil
	}
	fd := file.Fd()
	var oldState syscall.Termios
	if err := terminalIoctlTermios(fd, syscall.TCGETS, &oldState); err != nil {
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
	if err := terminalIoctlTermios(fd, syscall.TCSETS, &rawState); err != nil {
		return nil, false, fmt.Errorf("enable raw terminal: %w", err)
	}
	restore := func() error {
		if err := terminalIoctlTermios(fd, syscall.TCSETS, &oldState); err != nil {
			return fmt.Errorf("restore terminal state: %w", err)
		}
		return nil
	}
	return restore, true, nil
}

func terminalIoctlTermios(fd uintptr, request uintptr, state *syscall.Termios) error {
	return terminalIoctl(fd, request, unsafe.Pointer(state))
}

func terminalIoctl(fd uintptr, request uintptr, pointer unsafe.Pointer) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, request, uintptr(pointer))
	if errno != 0 {
		return errno
	}
	return nil
}

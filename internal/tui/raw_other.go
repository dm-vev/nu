//go:build !unix

package tui

import "io"

func enableRawInput(reader io.Reader) (func() error, bool, error) {
	return nil, false, nil
}

func terminalSize(target any, fallbackWidth int, fallbackHeight int) (int, int) {
	return fallbackWidth, fallbackHeight
}

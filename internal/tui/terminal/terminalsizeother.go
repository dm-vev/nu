//go:build !unix

package terminal

func terminalQuerySize(target any, fallbackWidth int, fallbackHeight int) (int, int) {
	return fallbackWidth, fallbackHeight
}

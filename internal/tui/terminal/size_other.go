//go:build !unix

package terminal

func querySize(target any, fallbackWidth int, fallbackHeight int) (int, int) {
	return fallbackWidth, fallbackHeight
}

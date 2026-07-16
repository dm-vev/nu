//go:build !unix

package terminal

// WatchResize is a no-op on unsupported platforms.
func WatchResize(render func()) func() {
	return func() {}
}

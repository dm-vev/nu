//go:build !unix

package tui

func watchResize(render func()) func() {
	return func() {}
}

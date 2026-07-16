//go:build !unix

package terminal

// EnableRaw is a no-op on unsupported platforms.
func (t *Terminal) EnableRaw() (func() error, bool, error) {
	return nil, false, nil
}

//go:build !windows
// +build !windows

package deej

import "fmt"

// sendKeystrokes is a stub implementation for non‑Windows platforms.  It
// always returns an error indicating that key simulation is unsupported.
func sendKeystrokes(keys []uint16) error {
        return fmt.Errorf("sendKeystrokes is not implemented on this platform")
}
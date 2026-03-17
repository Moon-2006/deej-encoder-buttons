//go:build windows
// +build windows

package deej

import (
        "fmt"
        "unsafe"

        "golang.org/x/sys/windows"
)

// sendKeystrokes dispatches one or more virtual key codes as a single key
// sequence using the Windows SendInput API.  It presses each key in order
// and then releases them in reverse order.  If any error occurs the
// returned error will be non‑nil.
func sendKeystrokes(keys []uint16) error {
        if len(keys) == 0 {
                return nil
        }
        // Build input slice: first keydown events, then keyup events in reverse
        inputs := make([]windows.INPUT, 0, len(keys)*2)
        for _, vk := range keys {
                ki := windows.KEYBDINPUT{
                        Vk:         vk,
                        Scan:       0,
                        Flags:      0,
                        Time:       0,
                        ExtraInfo:  0,
                }
                inputs = append(inputs, windows.INPUT{
                        Type: windows.INPUT_KEYBOARD,
                        Ki:   ki,
                })
        }
        for i := len(keys) - 1; i >= 0; i-- {
                vk := keys[i]
                ki := windows.KEYBDINPUT{
                        Vk:         vk,
                        Scan:       0,
                        Flags:      windows.KEYEVENTF_KEYUP,
                        Time:       0,
                        ExtraInfo:  0,
                }
                inputs = append(inputs, windows.INPUT{
                        Type: windows.INPUT_KEYBOARD,
                        Ki:   ki,
                })
        }
        sent, err := windows.SendInput(uint32(len(inputs)), &inputs[0], int32(unsafe.Sizeof(inputs[0])))
        if err != nil {
                return err
        }
        if sent != uint32(len(inputs)) {
                return fmt.Errorf("SendInput sent %d of %d events", sent, len(inputs))
        }
        return nil
}
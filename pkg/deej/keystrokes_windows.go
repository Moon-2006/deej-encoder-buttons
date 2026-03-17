//go:build windows
// +build windows

package deej

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	inputKeyboard   = 1
	keyeventfKeyUp  = 0x0002
)

type keyboardInput struct {
	Vk         uint16
	Scan       uint16
	Flags      uint32
	Time       uint32
	ExtraInfo  uintptr
}

type input struct {
	Type uint32
	Ki   keyboardInput
	// The real Win32 INPUT union is larger, but for keyboard input this is
	// enough for the old Go/Windows build used by this repo.
}

var (
	user32         = syscall.NewLazyDLL("user32.dll")
	procSendInput  = user32.NewProc("SendInput")
)

func sendKeystrokes(keys []uint16) error {
	if len(keys) == 0 {
		return nil
	}

	inputs := make([]input, 0, len(keys)*2)

	// key down
	for _, vk := range keys {
		inputs = append(inputs, input{
			Type: inputKeyboard,
			Ki: keyboardInput{
				Vk:        vk,
				Scan:      0,
				Flags:     0,
				Time:      0,
				ExtraInfo: 0,
			},
		})
	}

	// key up in reverse order
	for i := len(keys) - 1; i >= 0; i-- {
		vk := keys[i]
		inputs = append(inputs, input{
			Type: inputKeyboard,
			Ki: keyboardInput{
				Vk:        vk,
				Scan:      0,
				Flags:     keyeventfKeyUp,
				Time:      0,
				ExtraInfo: 0,
			},
		})
	}

	ret, _, err := procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		unsafe.Sizeof(inputs[0]),
	)

	if ret == 0 {
		if err != syscall.Errno(0) {
			return err
		}
		return fmt.Errorf("SendInput failed")
	}

	return nil
}

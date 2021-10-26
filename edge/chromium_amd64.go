//go:build windows
// +build windows

package edge

import (
	"unsafe"

	"github.com/Gipcomp/win32/gdi32"
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/user32"
)

func (e *Chromium) Resize() {
	if e.controller == nil {
		return
	}
	var bounds gdi32.RECT
	user32.GetClientRect(handle.HWND(e.hwnd), &bounds)
	//	w32.User32GetClientRect.Call(e.hwnd, uintptr(unsafe.Pointer(&bounds)))
	e.controller.vtbl.PutBounds.Call(
		uintptr(unsafe.Pointer(e.controller)),
		uintptr(unsafe.Pointer(&bounds)),
	)
}

// Copyright 2013 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"syscall"
	"unsafe"

	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/kernel32"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/win"
)

const clipboardWindowClass = `\o/ Walk_Clipboard_Class \o/`

func init() {
	AppendToWalkInit(func() {
		MustRegisterWindowClassWithWndProcPtr(clipboardWindowClass, syscall.NewCallback(clipboardWndProc))
		lpClassName, _ := syscall.UTF16PtrFromString(clipboardWindowClass)
		hwnd := user32.CreateWindowEx(
			0,
			lpClassName,
			nil,
			0,
			0,
			0,
			0,
			0,
			user32.HWND_MESSAGE,
			0,
			0,
			nil)

		if hwnd == 0 {
			panic("failed to create clipboard window")
		}

		if !user32.AddClipboardFormatListener(hwnd) {
			lastError("AddClipboardFormatListener")
		}

		clipboard.hwnd = hwnd
	})
}

func clipboardWndProc(hwnd handle.HWND, msg uint32, wp, lp uintptr) uintptr {
	switch msg {
	case user32.WM_CLIPBOARDUPDATE:
		clipboard.contentsChangedPublisher.Publish()
		return 0
	}

	return user32.DefWindowProc(hwnd, msg, wp, lp)
}

var clipboard ClipboardService

// Clipboard returns an object that provides access to the system clipboard.
func Clipboard() *ClipboardService {
	return &clipboard
}

// ClipboardService provides access to the system clipboard.
type ClipboardService struct {
	hwnd                     handle.HWND
	contentsChangedPublisher EventPublisher
}

// ContentsChanged returns an Event that you can attach to for handling
// clipboard content changes.
func (c *ClipboardService) ContentsChanged() *Event {
	return c.contentsChangedPublisher.Event()
}

// Clear clears the contents of the clipboard.
func (c *ClipboardService) Clear() error {
	return c.withOpenClipboard(func() error {
		if !user32.EmptyClipboard() {
			return lastError("EmptyClipboard")
		}

		return nil
	})
}

// ContainsText returns whether the clipboard currently contains text data.
func (c *ClipboardService) ContainsText() (available bool, err error) {
	err = c.withOpenClipboard(func() error {
		available = user32.IsClipboardFormatAvailable(user32.CF_UNICODETEXT)

		return nil
	})

	return
}

// Text returns the current text data of the clipboard.
func (c *ClipboardService) Text() (text string, err error) {
	err = c.withOpenClipboard(func() error {
		hMem := kernel32.HGLOBAL(user32.GetClipboardData(user32.CF_UNICODETEXT))
		if hMem == 0 {
			return lastError("GetClipboardData")
		}

		p := kernel32.GlobalLock(hMem)
		if p == nil {
			return lastError("GlobalLock()")
		}
		defer kernel32.GlobalUnlock(hMem)

		text = win.UTF16PtrToString((*uint16)(p))

		return nil
	})

	return
}

// SetText sets the current text data of the clipboard.
func (c *ClipboardService) SetText(s string) error {
	return c.withOpenClipboard(func() error {
		utf16, err := syscall.UTF16FromString(s)
		if err != nil {
			return err
		}

		hMem := kernel32.GlobalAlloc(kernel32.GMEM_MOVEABLE, uintptr(len(utf16)*2))
		if hMem == 0 {
			return lastError("GlobalAlloc")
		}

		p := kernel32.GlobalLock(hMem)
		if p == nil {
			return lastError("GlobalLock()")
		}

		kernel32.MoveMemory(p, unsafe.Pointer(&utf16[0]), uintptr(len(utf16)*2))

		kernel32.GlobalUnlock(hMem)

		if user32.SetClipboardData(user32.CF_UNICODETEXT, handle.HANDLE(hMem)) == 0 {
			// We need to free hMem.
			defer kernel32.GlobalFree(hMem)

			return lastError("SetClipboardData")
		}

		// The system now owns the memory referred to by hMem.

		return nil
	})
}

func (c *ClipboardService) withOpenClipboard(f func() error) error {
	if !user32.OpenClipboard(c.hwnd) {
		return lastError("OpenClipboard")
	}
	defer user32.CloseClipboard()

	return f()
}

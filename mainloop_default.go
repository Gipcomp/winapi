// Copyright 2019 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows,!walk_use_cgo

package winapi

import (
	"unsafe"

	"github.com/Gipcomp/win32/kernel32"
	"github.com/Gipcomp/win32/user32"
)

func (fb *FormBase) mainLoop() int {
	msg := (*user32.MSG)(unsafe.Pointer(kernel32.GlobalAlloc(0, unsafe.Sizeof(user32.MSG{}))))
	defer kernel32.GlobalFree(kernel32.HGLOBAL(unsafe.Pointer(msg)))

	for fb.hWnd != 0 {
		switch user32.GetMessage(msg, 0, 0, 0) {
		case 0:
			return int(msg.WParam)

		case -1:
			return -1
		}

		switch msg.Message {
		case user32.WM_KEYDOWN:
			if fb.handleKeyDown(msg) {
				continue
			}
		}

		if !user32.IsDialogMessage(fb.hWnd, msg) {
			user32.TranslateMessage(msg)
			user32.DispatchMessage(msg)
		}

		fb.group.RunSynchronized()
	}

	return 0
}

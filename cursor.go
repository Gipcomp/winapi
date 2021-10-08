// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"image"

	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/win"
)

type Cursor interface {
	Dispose()
	handle() user32.HCURSOR
}

type stockCursor struct {
	hCursor user32.HCURSOR
}

func (sc stockCursor) Dispose() {
	// nop
}

func (sc stockCursor) handle() user32.HCURSOR {
	return sc.hCursor
}

func CursorArrow() Cursor {
	return stockCursor{user32.LoadCursor(0, win.MAKEINTRESOURCE(user32.IDC_ARROW))}
}

func CursorIBeam() Cursor {
	return stockCursor{user32.LoadCursor(0, win.MAKEINTRESOURCE(user32.IDC_IBEAM))}
}

func CursorWait() Cursor {
	return stockCursor{user32.LoadCursor(0, win.MAKEINTRESOURCE(user32.IDC_WAIT))}
}

func CursorCross() Cursor {
	return stockCursor{user32.LoadCursor(0, win.MAKEINTRESOURCE(user32.IDC_CROSS))}
}

func CursorUpArrow() Cursor {
	return stockCursor{user32.LoadCursor(0, win.MAKEINTRESOURCE(user32.IDC_UPARROW))}
}

func CursorSizeNWSE() Cursor {
	return stockCursor{user32.LoadCursor(0, win.MAKEINTRESOURCE(user32.IDC_SIZENWSE))}
}

func CursorSizeNESW() Cursor {
	return stockCursor{user32.LoadCursor(0, win.MAKEINTRESOURCE(user32.IDC_SIZENESW))}
}

func CursorSizeWE() Cursor {
	return stockCursor{user32.LoadCursor(0, win.MAKEINTRESOURCE(user32.IDC_SIZEWE))}
}

func CursorSizeNS() Cursor {
	return stockCursor{user32.LoadCursor(0, win.MAKEINTRESOURCE(user32.IDC_SIZENS))}
}

func CursorSizeAll() Cursor {
	return stockCursor{user32.LoadCursor(0, win.MAKEINTRESOURCE(user32.IDC_SIZEALL))}
}

func CursorNo() Cursor {
	return stockCursor{user32.LoadCursor(0, win.MAKEINTRESOURCE(user32.IDC_NO))}
}

func CursorHand() Cursor {
	return stockCursor{user32.LoadCursor(0, win.MAKEINTRESOURCE(user32.IDC_HAND))}
}

func CursorAppStarting() Cursor {
	return stockCursor{user32.LoadCursor(0, win.MAKEINTRESOURCE(user32.IDC_APPSTARTING))}
}

func CursorHelp() Cursor {
	return stockCursor{user32.LoadCursor(0, win.MAKEINTRESOURCE(user32.IDC_HELP))}
}

func CursorIcon() Cursor {
	return stockCursor{user32.LoadCursor(0, win.MAKEINTRESOURCE(user32.IDC_ICON))}
}

func CursorSize() Cursor {
	return stockCursor{user32.LoadCursor(0, win.MAKEINTRESOURCE(user32.IDC_SIZE))}
}

type customCursor struct {
	hCursor user32.HCURSOR
}

func NewCursorFromImage(im image.Image, hotspot image.Point) (Cursor, error) {
	i, err := createAlphaCursorOrIconFromImage(im, hotspot, false)
	if err != nil {
		return nil, err
	}
	return customCursor{user32.HCURSOR(i)}, nil
}

func (cc customCursor) Dispose() {
	user32.DestroyIcon(user32.HICON(cc.hCursor))
}

func (cc customCursor) handle() user32.HCURSOR {
	return cc.hCursor
}

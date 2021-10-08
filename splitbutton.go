// Copyright 2016 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"unsafe"

	"github.com/Gipcomp/win32/gdi32"
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/user32"
)

type SplitButton struct {
	Button
	menu *Menu
}

func NewSplitButton(parent Container) (*SplitButton, error) {
	sb := new(SplitButton)

	var disposables Disposables
	defer disposables.Treat()

	if err := InitWidget(
		sb,
		parent,
		"BUTTON",
		user32.WS_TABSTOP|user32.WS_VISIBLE|user32.BS_SPLITBUTTON,
		0); err != nil {
		return nil, err
	}
	disposables.Add(sb)

	sb.Button.init()

	menu, err := NewMenu()
	if err != nil {
		return nil, err
	}
	disposables.Add(menu)
	menu.window = sb
	sb.menu = menu

	sb.GraphicsEffects().Add(InteractionEffect)
	sb.GraphicsEffects().Add(FocusEffect)

	disposables.Spare()

	return sb, nil
}

func (sb *SplitButton) Dispose() {
	sb.Button.Dispose()

	sb.menu.Dispose()
}

func (sb *SplitButton) ImageAboveText() bool {
	return sb.hasStyleBits(user32.BS_TOP)
}

func (sb *SplitButton) SetImageAboveText(value bool) error {
	if err := sb.ensureStyleBits(user32.BS_TOP, value); err != nil {
		return err
	}

	// We need to set the image again, or Windows will fail to calculate the
	// button control size correctly.
	return sb.SetImage(sb.image)
}

func (sb *SplitButton) Menu() *Menu {
	return sb.menu
}

func (sb *SplitButton) WndProc(hwnd handle.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case user32.WM_NOTIFY:
		switch ((*user32.NMHDR)(unsafe.Pointer(lParam))).Code {
		case user32.BCN_DROPDOWN:
			dd := (*user32.NMBCDROPDOWN)(unsafe.Pointer(lParam))

			p := gdi32.POINT{dd.RcButton.Left, dd.RcButton.Bottom}

			user32.ClientToScreen(sb.hWnd, &p)

			user32.TrackPopupMenuEx(
				sb.menu.hMenu,
				user32.TPM_NOANIMATION,
				p.X,
				p.Y,
				sb.hWnd,
				nil)
			return 0
		}
	}

	return sb.Button.WndProc(hwnd, msg, wParam, lParam)
}

// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/user32"
)

type PushButton struct {
	Button
}

func NewPushButton(parent Container) (*PushButton, error) {
	pb := new(PushButton)

	if err := InitWidget(
		pb,
		parent,
		"BUTTON",
		user32.WS_TABSTOP|user32.WS_VISIBLE|user32.BS_PUSHBUTTON,
		0); err != nil {
		return nil, err
	}

	pb.Button.init()

	pb.GraphicsEffects().Add(InteractionEffect)
	pb.GraphicsEffects().Add(FocusEffect)

	return pb, nil
}

func (pb *PushButton) ImageAboveText() bool {
	return pb.hasStyleBits(user32.BS_TOP)
}

func (pb *PushButton) SetImageAboveText(value bool) error {
	if err := pb.ensureStyleBits(user32.BS_TOP, value); err != nil {
		return err
	}

	// We need to set the image again, or Windows will fail to calculate the
	// button control size correctly.
	return pb.SetImage(pb.image)
}

func (pb *PushButton) ensureProperDialogDefaultButton(hwndFocus handle.HWND) {
	widget := windowFromHandle(hwndFocus)
	if widget == nil {
		return
	}

	if _, ok := widget.(*PushButton); ok {
		return
	}

	form := ancestor(pb)
	if form == nil {
		return
	}

	dlg, ok := form.(dialogish)
	if !ok {
		return
	}

	defBtn := dlg.DefaultButton()
	if defBtn == nil {
		return
	}

	if err := defBtn.setAndClearStyleBits(user32.BS_DEFPUSHBUTTON, user32.BS_PUSHBUTTON); err != nil {
		return
	}

	if err := defBtn.Invalidate(); err != nil {
		return
	}
}

func (pb *PushButton) WndProc(hwnd handle.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case user32.WM_GETDLGCODE:
		hwndFocus := user32.GetFocus()
		if hwndFocus == pb.hWnd {
			form := ancestor(pb)
			if form == nil {
				break
			}

			dlg, ok := form.(dialogish)
			if !ok {
				break
			}

			defBtn := dlg.DefaultButton()
			if defBtn == pb {
				pb.setAndClearStyleBits(user32.BS_DEFPUSHBUTTON, user32.BS_PUSHBUTTON)
				return user32.DLGC_BUTTON | user32.DLGC_DEFPUSHBUTTON
			}

			break
		}

		pb.ensureProperDialogDefaultButton(hwndFocus)

	case user32.WM_KILLFOCUS:
		pb.ensureProperDialogDefaultButton(handle.HWND(wParam))
	}

	return pb.Button.WndProc(hwnd, msg, wParam, lParam)
}

func (pb *PushButton) CreateLayoutItem(ctx *LayoutContext) LayoutItem {
	return &pushButtonLayoutItem{
		buttonLayoutItem: buttonLayoutItem{
			idealSize: pb.idealSize(),
		},
	}
}

type pushButtonLayoutItem struct {
	buttonLayoutItem
}

func (*pushButtonLayoutItem) LayoutFlags() LayoutFlags {
	return GrowableHorz
}

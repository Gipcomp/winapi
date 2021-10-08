// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"strconv"

	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/win"
)

type CheckState int

const (
	CheckUnchecked     CheckState = user32.BST_UNCHECKED
	CheckChecked       CheckState = user32.BST_CHECKED
	CheckIndeterminate CheckState = user32.BST_INDETERMINATE
)

var checkBoxCheckSize Size // in native pixels

type CheckBox struct {
	Button
	checkStateChangedPublisher EventPublisher
}

func NewCheckBox(parent Container) (*CheckBox, error) {
	cb := new(CheckBox)

	if err := InitWidget(
		cb,
		parent,
		"BUTTON",
		user32.WS_TABSTOP|user32.WS_VISIBLE|user32.BS_AUTOCHECKBOX,
		0); err != nil {
		return nil, err
	}

	cb.Button.init()

	cb.SetBackground(nullBrushSingleton)

	cb.GraphicsEffects().Add(InteractionEffect)
	cb.GraphicsEffects().Add(FocusEffect)

	cb.MustRegisterProperty("CheckState", NewProperty(
		func() interface{} {
			return cb.CheckState()
		},
		func(v interface{}) error {
			cb.SetCheckState(CheckState(assertIntOr(v, 0)))

			return nil
		},
		cb.CheckStateChanged()))

	return cb, nil
}

func (cb *CheckBox) TextOnLeftSide() bool {
	return cb.hasStyleBits(user32.BS_LEFTTEXT)
}

func (cb *CheckBox) SetTextOnLeftSide(textLeft bool) error {
	return cb.ensureStyleBits(user32.BS_LEFTTEXT, textLeft)
}

func (cb *CheckBox) setChecked(checked bool) {
	cb.Button.setChecked(checked)

	cb.checkStateChangedPublisher.Publish()
}

func (cb *CheckBox) Tristate() bool {
	return cb.hasStyleBits(user32.BS_AUTO3STATE)
}

func (cb *CheckBox) SetTristate(tristate bool) error {
	var set, clear uint32
	if tristate {
		set, clear = user32.BS_AUTO3STATE, user32.BS_AUTOCHECKBOX
	} else {
		set, clear = user32.BS_AUTOCHECKBOX, user32.BS_AUTO3STATE
	}

	return cb.setAndClearStyleBits(set, clear)
}

func (cb *CheckBox) CheckState() CheckState {
	return CheckState(cb.SendMessage(user32.BM_GETCHECK, 0, 0))
}

func (cb *CheckBox) SetCheckState(state CheckState) {
	if state == cb.CheckState() {
		return
	}

	cb.SendMessage(user32.BM_SETCHECK, uintptr(state), 0)

	cb.checkedChangedPublisher.Publish()
	cb.checkStateChangedPublisher.Publish()
}

func (cb *CheckBox) CheckStateChanged() *Event {
	return cb.checkStateChangedPublisher.Event()
}

func (cb *CheckBox) SaveState() error {
	return cb.WriteState(strconv.Itoa(int(cb.CheckState())))
}

func (cb *CheckBox) RestoreState() error {
	s, err := cb.ReadState()
	if err != nil {
		return err
	}

	cs, err := strconv.Atoi(s)
	if err != nil {
		return err
	}

	cb.SetCheckState(CheckState(cs))

	return nil
}

func (cb *CheckBox) WndProc(hwnd handle.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case user32.WM_COMMAND:
		switch win.HIWORD(uint32(wParam)) {
		case user32.BN_CLICKED:
			cb.checkedChangedPublisher.Publish()
			cb.checkStateChangedPublisher.Publish()
		}
	}

	return cb.Button.WndProc(hwnd, msg, wParam, lParam)
}

// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"fmt"
	"unsafe"

	"github.com/Gipcomp/win32/comctl32"
	"github.com/Gipcomp/win32/gdi32"
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/win"
)

type clickable interface {
	raiseClicked()
}

type setCheckeder interface {
	setChecked(checked bool)
}

type Button struct {
	WidgetBase
	checkedChangedPublisher EventPublisher
	clickedPublisher        EventPublisher
	textChangedPublisher    EventPublisher
	imageChangedPublisher   EventPublisher
	image                   Image
	persistent              bool
}

func (b *Button) init() {
	b.MustRegisterProperty("Checked", NewBoolProperty(
		func() bool {
			return b.Checked()
		},
		func(v bool) error {
			b.SetChecked(v)
			return nil
		},
		b.CheckedChanged()))

	b.MustRegisterProperty("Image", NewProperty(
		func() interface{} {
			return b.Image()
		},
		func(v interface{}) error {
			img, err := ImageFrom(v)
			if err != nil {
				return err
			}

			b.SetImage(img)

			return nil
		},
		b.imageChangedPublisher.Event()))

	b.MustRegisterProperty("Text", NewProperty(
		func() interface{} {
			return b.Text()
		},
		func(v interface{}) error {
			return b.SetText(assertStringOr(v, ""))
		},
		b.textChangedPublisher.Event()))
}

func (b *Button) ApplyDPI(dpi int) {
	b.WidgetBase.ApplyDPI(dpi)

	b.SetImage(b.image)
}

func (b *Button) Image() Image {
	return b.image
}

func (b *Button) SetImage(image Image) error {
	var typ, handle uintptr
	switch img := image.(type) {
	case nil:

	case *Bitmap:
		typ = user32.IMAGE_BITMAP
		handle = uintptr(img.hBmp)

	case *Icon:
		typ = user32.IMAGE_ICON
		handle = uintptr(img.handleForDPI(b.DPI()))

	default:
		bmp, err := iconCache.Bitmap(image, b.DPI())
		if err != nil {
			return err
		}

		typ = user32.IMAGE_BITMAP
		handle = uintptr(bmp.hBmp)
	}

	b.SendMessage(user32.BM_SETIMAGE, typ, handle)

	b.image = image

	b.RequestLayout()

	b.imageChangedPublisher.Publish()

	return nil
}

func (b *Button) ImageChanged() *Event {
	return b.imageChangedPublisher.Event()
}

func (b *Button) Text() string {
	return b.text()
}

func (b *Button) SetText(value string) error {
	if value == b.Text() {
		return nil
	}

	if err := b.setText(value); err != nil {
		return err
	}

	b.RequestLayout()

	return nil
}

func (b *Button) Checked() bool {
	return b.SendMessage(user32.BM_GETCHECK, 0, 0) == user32.BST_CHECKED
}

func (b *Button) SetChecked(checked bool) {
	if checked == b.Checked() {
		return
	}

	b.window.(setCheckeder).setChecked(checked)
}

func (b *Button) setChecked(checked bool) {
	var chk uintptr

	if checked {
		chk = user32.BST_CHECKED
	} else {
		chk = user32.BST_UNCHECKED
	}

	b.SendMessage(user32.BM_SETCHECK, chk, 0)

	b.checkedChangedPublisher.Publish()
}

func (b *Button) CheckedChanged() *Event {
	return b.checkedChangedPublisher.Event()
}

func (b *Button) Persistent() bool {
	return b.persistent
}

func (b *Button) SetPersistent(value bool) {
	b.persistent = value
}

func (b *Button) SaveState() error {
	return b.WriteState(fmt.Sprintf("%t", b.Checked()))
}

func (b *Button) RestoreState() error {
	s, err := b.ReadState()
	if err != nil {
		return err
	}

	b.SetChecked(s == "true")

	return nil
}

func (b *Button) Clicked() *Event {
	return b.clickedPublisher.Event()
}

func (b *Button) raiseClicked() {
	b.clickedPublisher.Publish()
}

func (b *Button) WndProc(hwnd handle.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case user32.WM_COMMAND:
		hiWP := win.HIWORD(uint32(wParam))

		if hiWP == 0 && lParam == 0 {
			if a, ok := actionsById[win.LOWORD(uint32(wParam))]; ok {
				a.raiseTriggered()
			}
		} else {
			switch hiWP {
			case user32.BN_CLICKED:
				b.raiseClicked()
			}
		}

	case user32.WM_SETTEXT:
		b.textChangedPublisher.Publish()
	}

	return b.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}

// idealSize returns ideal button size in native pixels.
func (b *Button) idealSize() Size {
	min := b.dialogBaseUnitsToPixels(Size{50, 14})

	if b.Text() == "" {
		return min
	}

	var s gdi32.SIZE
	b.SendMessage(comctl32.BCM_GETIDEALSIZE, 0, uintptr(unsafe.Pointer(&s)))

	return maxSize(sizeFromSIZE(s), min)
}

func (b *Button) CreateLayoutItem(ctx *LayoutContext) LayoutItem {
	return &buttonLayoutItem{
		idealSize: b.idealSize(),
	}
}

type buttonLayoutItem struct {
	LayoutItemBase
	idealSize Size // in native pixels
}

func (li *buttonLayoutItem) LayoutFlags() LayoutFlags {
	return 0
}

func (li *buttonLayoutItem) IdealSize() Size {
	return li.MinSize()
}

func (li *buttonLayoutItem) MinSize() Size {
	return li.idealSize
}

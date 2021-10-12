// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/winapi/errs"
)

const splitterHandleWindowClass = `\o/ Walk_SplitterHandle_Class \o/`

func init() {
	AppendToWalkInit(func() {
		MustRegisterWindowClass(splitterHandleWindowClass)
	})
}

type splitterHandle struct {
	WidgetBase
}

func newSplitterHandle(splitter *Splitter) (*splitterHandle, error) {
	if splitter == nil {
		return nil, errs.NewError("splitter cannot be nil")
	}

	sh := new(splitterHandle)
	sh.parent = splitter

	if err := InitWindow(
		sh,
		splitter,
		splitterHandleWindowClass,
		user32.WS_CHILD|user32.WS_VISIBLE,
		0); err != nil {
		return nil, err
	}

	sh.SetBackground(NullBrush())

	if err := sh.setAndClearStyleBits(0, user32.WS_CLIPSIBLINGS); err != nil {
		return nil, err
	}

	return sh, nil
}

func (sh *splitterHandle) WndProc(hwnd handle.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case user32.WM_ERASEBKGND:
		if sh.Background() == nullBrushSingleton {
			return 1
		}

	case user32.WM_PAINT:
		if sh.Background() == nullBrushSingleton {
			var ps user32.PAINTSTRUCT

			user32.BeginPaint(hwnd, &ps)
			defer user32.EndPaint(hwnd, &ps)

			return 0
		}
	}

	return sh.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}

func (sh *splitterHandle) CreateLayoutItem(ctx *LayoutContext) LayoutItem {
	var orientation Orientation
	var handleWidth int

	if splitter, ok := sh.Parent().(*Splitter); ok {
		orientation = splitter.Orientation()
		handleWidth = splitter.HandleWidth()
	}

	return &splitterHandleLayoutItem{
		orientation: orientation,
		handleWidth: handleWidth,
	}
}

type splitterHandleLayoutItem struct {
	LayoutItemBase
	orientation Orientation
	handleWidth int
}

func (li *splitterHandleLayoutItem) LayoutFlags() LayoutFlags {
	if li.orientation == Horizontal {
		return ShrinkableVert | GrowableVert | GreedyVert
	}

	return ShrinkableHorz | GrowableHorz | GreedyHorz
}

func (li *splitterHandleLayoutItem) IdealSize() Size {
	var size Size
	dpi := int(user32.GetDpiForWindow(li.handle))

	if li.orientation == Horizontal {
		size.Width = IntFrom96DPI(li.handleWidth, dpi)
	} else {
		size.Height = IntFrom96DPI(li.handleWidth, dpi)
	}

	return size
}

func (li *splitterHandleLayoutItem) MinSize() Size {
	return li.IdealSize()
}

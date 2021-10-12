// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"unsafe"

	"github.com/Gipcomp/win32/gdi32"
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/winapi/errs"
)

const customWidgetWindowClass = `\o/ Walk_CustomWidget_Class \o/`

func init() {
	AppendToWalkInit(func() {
		MustRegisterWindowClass(customWidgetWindowClass)
	})
}

// PaintFunc paints custom widget content. updateBounds is specified in 1/96" or native pixels.
type PaintFunc func(canvas *Canvas, updateBounds Rectangle) error

type PaintMode int

const (
	PaintNormal   PaintMode = iota // erase background before PaintFunc
	PaintNoErase                   // PaintFunc clears background, single buffered
	PaintBuffered                  // PaintFunc clears background, double buffered
)

type CustomWidget struct {
	WidgetBase
	paint               PaintFunc // in 1/96" units
	paintPixels         PaintFunc // in native pixels
	invalidatesOnResize bool
	paintMode           PaintMode
}

// NewCustomWidget creates and initializes a new custom draw widget.
//
// Deprecated: PaintFunc is taking updateBounds parameter at 96dpi for backward compatibility with
// clients. On high-DPI displays this is too sparse and may incur a thin unpainted edge around
// control due to rounding errors. Newer applications should use NewCustomWidgetPixels.
func NewCustomWidget(parent Container, style uint, paint PaintFunc) (*CustomWidget, error) {
	cw := &CustomWidget{paint: paint}
	err := cw.init(parent, style)
	if err != nil {
		return nil, err
	}

	return cw, nil
}

// NewCustomWidgetPixels creates and initializes a new custom draw widget.
func NewCustomWidgetPixels(parent Container, style uint, paintPixels PaintFunc) (*CustomWidget, error) {
	cw := &CustomWidget{paintPixels: paintPixels}
	err := cw.init(parent, style)
	if err != nil {
		return nil, err
	}

	return cw, nil
}

func (cw *CustomWidget) init(parent Container, style uint) error {
	if err := InitWidget(
		cw,
		parent,
		customWidgetWindowClass,
		user32.WS_VISIBLE|uint32(style),
		0); err != nil {
		return err
	}

	return nil
}

// deprecated, use PaintMode
func (cw *CustomWidget) ClearsBackground() bool {
	return cw.paintMode != PaintNormal
}

// deprecated, use SetPaintMode
func (cw *CustomWidget) SetClearsBackground(value bool) {
	if value != cw.ClearsBackground() {
		if value {
			cw.paintMode = PaintNormal
		} else {
			cw.paintMode = PaintNoErase
		}
	}
}

func (cw *CustomWidget) InvalidatesOnResize() bool {
	return cw.invalidatesOnResize
}

func (cw *CustomWidget) SetInvalidatesOnResize(value bool) {
	cw.invalidatesOnResize = value
}

func (cw *CustomWidget) PaintMode() PaintMode {
	return cw.paintMode
}

func (cw *CustomWidget) SetPaintMode(value PaintMode) {
	cw.paintMode = value
}

func (cw *CustomWidget) WndProc(hwnd handle.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case user32.WM_PAINT:
		if cw.paint == nil && cw.paintPixels == nil {
			errs.NewError("paint(Pixels) func is nil")
			break
		}

		var ps user32.PAINTSTRUCT

		var hdc gdi32.HDC
		if wParam == 0 {
			hdc = user32.BeginPaint(cw.hWnd, &ps)
		} else {
			hdc = gdi32.HDC(wParam)
		}
		if hdc == 0 {
			errs.NewError("BeginPaint failed")
			break
		}
		defer func() {
			if wParam == 0 {
				user32.EndPaint(cw.hWnd, &ps)
			}
		}()

		canvas, err := newCanvasFromHDC(hdc)
		if err != nil {
			errs.NewError("newCanvasFromHDC failed")
			break
		}
		defer canvas.Dispose()

		bounds := rectangleFromRECT(ps.RcPaint)
		if cw.paintMode == PaintBuffered {
			err = cw.bufferedPaint(canvas, bounds)
		} else if cw.paintPixels != nil {
			err = cw.paintPixels(canvas, bounds)
		} else {
			err = cw.paint(canvas, RectangleTo96DPI(bounds, cw.DPI()))
		}

		if err != nil {
			errs.NewError("paint failed")
			break
		}

		return 0

	case user32.WM_ERASEBKGND:
		if cw.paintMode != PaintNormal {
			return 1
		}

	case user32.WM_PRINTCLIENT:
		user32.SendMessage(hwnd, user32.WM_PAINT, wParam, lParam)

	case user32.WM_WINDOWPOSCHANGED:
		wp := (*user32.WINDOWPOS)(unsafe.Pointer(lParam))

		if wp.Flags&user32.SWP_NOSIZE != 0 {
			break
		}

		if cw.invalidatesOnResize {
			cw.Invalidate()
		}
	}

	return cw.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}

// bufferedPaint draws widget on a memory buffer. updateBounds are in native pixels.
func (cw *CustomWidget) bufferedPaint(canvas *Canvas, updateBounds Rectangle) error {
	hdc := gdi32.CreateCompatibleDC(canvas.hdc)
	if hdc == 0 {
		return errs.NewError("CreateCompatibleDC failed")
	}
	defer gdi32.DeleteDC(hdc)

	buffered := Canvas{hdc: hdc, doNotDispose: true}
	if _, err := buffered.init(); err != nil {
		return err
	}

	w, h := int32(updateBounds.Width), int32(updateBounds.Height)
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	hbmp := gdi32.CreateCompatibleBitmap(canvas.hdc, w, h)
	if hbmp == 0 {
		return errs.LastError("CreateCompatibleBitmap failed")
	}
	defer gdi32.DeleteObject(gdi32.HGDIOBJ(hbmp))

	oldbmp := gdi32.SelectObject(buffered.hdc, gdi32.HGDIOBJ(hbmp))
	if oldbmp == 0 {
		return errs.NewError("SelectObject failed")
	}
	defer gdi32.SelectObject(buffered.hdc, oldbmp)

	gdi32.SetViewportOrgEx(buffered.hdc, -int32(updateBounds.X), -int32(updateBounds.Y), nil)
	gdi32.SetBrushOrgEx(buffered.hdc, -int32(updateBounds.X), -int32(updateBounds.Y), nil)

	var err error
	if cw.paintPixels != nil {
		err = cw.paintPixels(&buffered, updateBounds)
	} else {
		err = cw.paint(&buffered, RectangleTo96DPI(updateBounds, cw.DPI()))
	}

	if !gdi32.BitBlt(canvas.hdc,
		int32(updateBounds.X), int32(updateBounds.Y), w, h,
		buffered.hdc,
		int32(updateBounds.X), int32(updateBounds.Y), gdi32.SRCCOPY) {
		return errs.LastError("buffered BitBlt failed")
	}

	return err
}

func (*CustomWidget) CreateLayoutItem(ctx *LayoutContext) LayoutItem {
	return NewGreedyLayoutItem()
}

// Copyright 2018 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"syscall"
	"unsafe"

	"github.com/Gipcomp/win32/commctrl"
	"github.com/Gipcomp/win32/gdi32"
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/user32"
)

const staticWindowClass = `\o/ Walk_Static_Class \o/`

var staticWndProcPtr uintptr

func init() {
	AppendToWalkInit(func() {
		MustRegisterWindowClass(staticWindowClass)
		staticWndProcPtr = syscall.NewCallback(staticWndProc)
	})
}

type static struct {
	WidgetBase
	hwndStatic           handle.HWND
	origStaticWndProcPtr uintptr
	textAlignment        Alignment2D
	textColor            Color
}

func (s *static) init(widget Widget, parent Container, style uint32) error {
	if err := InitWidget(
		widget,
		parent,
		staticWindowClass,
		user32.WS_VISIBLE|(style&user32.WS_BORDER),
		user32.WS_EX_CONTROLPARENT); err != nil {
		return err
	}
	strPtr, err := syscall.UTF16PtrFromString("static")
	if err != nil {
		return err
	}
	if s.hwndStatic = user32.CreateWindowEx(
		0,
		strPtr,
		nil,
		user32.WS_CHILD|user32.WS_CLIPSIBLINGS|user32.WS_VISIBLE|user32.SS_LEFT|user32.SS_NOTIFY|(style&^user32.WS_BORDER),
		user32.CW_USEDEFAULT,
		user32.CW_USEDEFAULT,
		user32.CW_USEDEFAULT,
		user32.CW_USEDEFAULT,
		s.hWnd,
		0,
		0,
		nil,
	); s.hwndStatic == 0 {
		return newError("creating static failed")
	}

	if err := s.group.toolTip.AddTool(s); err != nil {
		return err
	}

	s.origStaticWndProcPtr = user32.SetWindowLongPtr(s.hwndStatic, user32.GWLP_WNDPROC, staticWndProcPtr)
	if s.origStaticWndProcPtr == 0 {
		return lastError("SetWindowLongPtr")
	}

	s.applyFont(s.Font())

	s.SetBackground(nullBrushSingleton)

	s.SetAlignment(AlignHNearVCenter)

	return nil
}

func (s *static) Dispose() {
	if s.hwndStatic != 0 {
		user32.DestroyWindow(s.hwndStatic)
		s.hwndStatic = 0
	}

	s.WidgetBase.Dispose()
}

func (s *static) handleForToolTip() handle.HWND {
	return s.hwndStatic
}

func (s *static) applyEnabled(enabled bool) {
	s.WidgetBase.applyEnabled(enabled)

	setWindowEnabled(s.hwndStatic, enabled)
}

func (s *static) applyFont(font *Font) {
	s.WidgetBase.applyFont(font)

	SetWindowFont(s.hwndStatic, font)
}

func (s *static) textAlignment1D() Alignment1D {
	switch s.textAlignment {
	case AlignHCenterVNear, AlignHCenterVCenter, AlignHCenterVFar:
		return AlignCenter

	case AlignHFarVNear, AlignHFarVCenter, AlignHFarVFar:
		return AlignFar

	default:
		return AlignNear
	}
}

func (s *static) setTextAlignment1D(alignment Alignment1D) error {
	var align Alignment2D

	switch alignment {
	case AlignCenter:
		align = AlignHCenterVCenter

	case AlignFar:
		align = AlignHFarVCenter

	default:
		align = AlignHNearVCenter
	}

	return s.setTextAlignment(align)
}

func (s *static) setTextAlignment(alignment Alignment2D) error {
	if alignment == s.textAlignment {
		return nil
	}

	var styleBit uint32

	switch alignment {
	case AlignHNearVNear, AlignHNearVCenter, AlignHNearVFar:
		styleBit |= user32.SS_LEFT

	case AlignHCenterVNear, AlignHCenterVCenter, AlignHCenterVFar:
		styleBit |= user32.SS_CENTER

	case AlignHFarVNear, AlignHFarVCenter, AlignHFarVFar:
		styleBit |= user32.SS_RIGHT
	}

	if err := setAndClearWindowLongBits(s.hwndStatic, user32.GWL_STYLE, styleBit, user32.SS_LEFT|user32.SS_CENTER|user32.SS_RIGHT); err != nil {
		return err
	}

	s.textAlignment = alignment

	s.Invalidate()

	return nil
}

func (s *static) setText(text string) (changed bool, err error) {
	if text == s.text() {
		return false, nil
	}

	if err := s.WidgetBase.setText(text); err != nil {
		return false, err
	}

	if err := setWindowText(s.hwndStatic, text); err != nil {
		return false, err
	}

	s.RequestLayout()

	return true, nil
}

func (s *static) TextColor() Color {
	return s.textColor
}

func (s *static) SetTextColor(c Color) {
	s.textColor = c

	s.Invalidate()
}

func (s *static) shrinkable() bool {
	if em, ok := s.window.(interface{ EllipsisMode() EllipsisMode }); ok {
		return em.EllipsisMode() != EllipsisNone
	}

	return false
}

func (s *static) updateStaticBounds() {
	var format DrawTextFormat

	switch s.textAlignment {
	case AlignHNearVNear, AlignHNearVCenter, AlignHNearVFar:
		format |= TextLeft

	case AlignHCenterVNear, AlignHCenterVCenter, AlignHCenterVFar:
		format |= TextCenter

	case AlignHFarVNear, AlignHFarVCenter, AlignHFarVFar:
		format |= TextRight
	}

	switch s.textAlignment {
	case AlignHNearVNear, AlignHCenterVNear, AlignHFarVNear:
		format |= TextTop

	case AlignHNearVCenter, AlignHCenterVCenter, AlignHFarVCenter:
		format |= TextVCenter

	case AlignHNearVFar, AlignHCenterVFar, AlignHFarVFar:
		format |= TextBottom
	}

	cb := s.ClientBoundsPixels()

	if shrinkable := s.shrinkable(); shrinkable || format&TextVCenter != 0 || format&TextBottom != 0 {
		var size Size
		if _, ok := s.window.(HeightForWidther); ok {
			size = s.calculateTextSizeForWidth(cb.Width)
		} else {
			size = s.calculateTextSize()
		}

		if shrinkable {
			var text string
			if size.Width > cb.Width {
				text = s.text()
			}
			s.SetToolTipText(text)
		}

		if format&TextVCenter != 0 || format&TextBottom != 0 {
			if format&TextVCenter != 0 {
				cb.Y += (cb.Height - size.Height) / 2
			} else {
				cb.Y += cb.Height - size.Height
			}

			cb.Height = size.Height
		}
	}

	user32.MoveWindow(s.hwndStatic, int32(cb.X), int32(cb.Y), int32(cb.Width), int32(cb.Height), true)

	s.Invalidate()
}

func (s *static) WndProc(hwnd handle.HWND, msg uint32, wp, lp uintptr) uintptr {
	switch msg {
	case user32.WM_CTLCOLORSTATIC:
		if hBrush := s.handleWMCTLCOLOR(wp, uintptr(s.hWnd)); hBrush != 0 {
			return hBrush
		}

	case user32.WM_WINDOWPOSCHANGED:
		wp := (*user32.WINDOWPOS)(unsafe.Pointer(lp))

		if wp.Flags&user32.SWP_NOSIZE != 0 {
			break
		}

		s.updateStaticBounds()
	}

	return s.WidgetBase.WndProc(hwnd, msg, wp, lp)
}

func staticWndProc(hwnd handle.HWND, msg uint32, wp, lp uintptr) uintptr {
	as, ok := windowFromHandle(user32.GetParent(hwnd)).(interface{ asStatic() *static })
	if !ok {
		return 0
	}

	s := as.asStatic()

	switch msg {
	case user32.WM_NCHITTEST:
		return user32.HTCLIENT

	case user32.WM_MOUSEMOVE, user32.WM_LBUTTONDOWN, user32.WM_LBUTTONUP, user32.WM_MBUTTONDOWN, user32.WM_MBUTTONUP, user32.WM_RBUTTONDOWN, user32.WM_RBUTTONUP:
		m := user32.MSG{
			HWnd:    hwnd,
			Message: msg,
			WParam:  wp,
			LParam:  lp,
			Pt:      gdi32.POINT{X: int32(user32.GET_X_LPARAM(lp)), Y: int32(user32.GET_Y_LPARAM(lp))},
		}

		return s.group.toolTip.SendMessage(commctrl.TTM_RELAYEVENT, 0, uintptr(unsafe.Pointer(&m)))
	}

	return user32.CallWindowProc(s.origStaticWndProcPtr, hwnd, msg, wp, lp)
}

func (s *static) CreateLayoutItem(ctx *LayoutContext) LayoutItem {
	var layoutFlags LayoutFlags
	if s.textAlignment1D() != AlignNear {
		layoutFlags = GrowableHorz
	} else if s.shrinkable() {
		layoutFlags = ShrinkableHorz
	}

	idealSize := s.calculateTextSize()
	if s.hasStyleBits(user32.WS_BORDER) {
		border := s.IntFrom96DPI(1) * 2
		idealSize.Width += border
		idealSize.Height += border * 2
	}

	return &staticLayoutItem{
		layoutFlags: layoutFlags,
		idealSize:   idealSize,
	}
}

type staticLayoutItem struct {
	LayoutItemBase
	layoutFlags LayoutFlags
	idealSize   Size // in native pixels
}

func (li *staticLayoutItem) LayoutFlags() LayoutFlags {
	return li.layoutFlags
}

func (li *staticLayoutItem) IdealSize() Size {
	return li.idealSize
}

func (li *staticLayoutItem) MinSize() Size {
	if li.layoutFlags&ShrinkableHorz != 0 {
		return Size{Height: li.idealSize.Height}
	}

	return li.idealSize
}

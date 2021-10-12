// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"syscall"
	"unsafe"

	"github.com/Gipcomp/win32/gdi32"
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/win"
	"github.com/Gipcomp/win32/winuser"
	"github.com/Gipcomp/winapi/errs"
)

type CaseMode uint32

const (
	CaseModeMixed CaseMode = iota
	CaseModeUpper
	CaseModeLower
)

const (
	lineEditMinChars    = 1  // 10 // number of characters needed to make a LineEdit usable
	lineEditGreedyLimit = 29 // 80 // fields with MaxLength larger than this will be greedy (default length is 32767)
)

type LineEdit struct {
	WidgetBase
	editingFinishedPublisher EventPublisher
	readOnlyChangedPublisher EventPublisher
	textChangedPublisher     EventPublisher
	charWidthFont            *Font
	charWidth                int // in native pixels
	textColor                Color
}

func newLineEdit(parent Window) (*LineEdit, error) {
	le := new(LineEdit)

	if err := InitWindow(
		le,
		parent,
		"EDIT",
		user32.WS_CHILD|user32.WS_TABSTOP|user32.WS_VISIBLE|winuser.ES_AUTOHSCROLL,
		user32.WS_EX_CLIENTEDGE); err != nil {
		return nil, err
	}

	le.GraphicsEffects().Add(InteractionEffect)
	le.GraphicsEffects().Add(FocusEffect)

	le.MustRegisterProperty("ReadOnly", NewProperty(
		func() interface{} {
			return le.ReadOnly()
		},
		func(v interface{}) error {
			return le.SetReadOnly(v.(bool))
		},
		le.readOnlyChangedPublisher.Event()))

	le.MustRegisterProperty("Text", NewProperty(
		func() interface{} {
			return le.Text()
		},
		func(v interface{}) error {
			return le.SetText(assertStringOr(v, ""))
		},
		le.textChangedPublisher.Event()))

	return le, nil
}

func NewLineEdit(parent Container) (*LineEdit, error) {
	if parent == nil {
		return nil, errs.NewError("parent cannot be nil")
	}

	le, err := newLineEdit(parent)
	if err != nil {
		return nil, err
	}

	var succeeded bool
	defer func() {
		if !succeeded {
			le.Dispose()
		}
	}()

	le.parent = parent
	if err = parent.Children().Add(le); err != nil {
		return nil, err
	}

	succeeded = true

	return le, nil
}

func (le *LineEdit) CueBanner() string {
	buf := make([]uint16, 128)
	if win.FALSE == le.SendMessage(winuser.EM_GETCUEBANNER, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf))) {
		errs.NewError("EM_GETCUEBANNER failed")
		return ""
	}

	return syscall.UTF16ToString(buf)
}

func (le *LineEdit) SetCueBanner(value string) error {
	strPtr, err := syscall.UTF16PtrFromString(value)
	if err != nil {
		errs.NewError(err.Error())
	}
	if win.FALSE == le.SendMessage(winuser.EM_SETCUEBANNER, win.FALSE, uintptr(unsafe.Pointer(strPtr))) {
		return errs.NewError("EM_SETCUEBANNER failed")
	}

	return nil
}

func (le *LineEdit) MaxLength() int {
	return int(le.SendMessage(winuser.EM_GETLIMITTEXT, 0, 0))
}

func (le *LineEdit) SetMaxLength(value int) {
	le.SendMessage(winuser.EM_LIMITTEXT, uintptr(value), 0)
}

func (le *LineEdit) Text() string {
	return le.text()
}

func (le *LineEdit) SetText(value string) error {
	return le.setText(value)
}

func (le *LineEdit) TextSelection() (start, end int) {
	le.SendMessage(winuser.EM_GETSEL, uintptr(unsafe.Pointer(&start)), uintptr(unsafe.Pointer(&end)))
	return
}

func (le *LineEdit) SetTextSelection(start, end int) {
	le.SendMessage(winuser.EM_SETSEL, uintptr(start), uintptr(end))
}

func (le *LineEdit) TextAlignment() Alignment1D {
	switch user32.GetWindowLong(le.hWnd, user32.GWL_STYLE) & (winuser.ES_LEFT | winuser.ES_CENTER | winuser.ES_RIGHT) {
	case winuser.ES_CENTER:
		return AlignCenter

	case winuser.ES_RIGHT:
		return AlignFar
	}

	return AlignNear
}

func (le *LineEdit) SetTextAlignment(alignment Alignment1D) error {
	if alignment == AlignDefault {
		alignment = AlignNear
	}

	var bit uint32

	switch alignment {
	case AlignCenter:
		bit = winuser.ES_CENTER

	case AlignFar:
		bit = winuser.ES_RIGHT

	default:
		bit = winuser.ES_LEFT
	}

	return le.setAndClearStyleBits(bit, winuser.ES_LEFT|winuser.ES_CENTER|winuser.ES_RIGHT)
}

func (le *LineEdit) CaseMode() CaseMode {
	style := uint32(user32.GetWindowLong(le.hWnd, user32.GWL_STYLE))

	if style&winuser.ES_UPPERCASE != 0 {
		return CaseModeUpper
	} else if style&winuser.ES_LOWERCASE != 0 {
		return CaseModeLower
	} else {
		return CaseModeMixed
	}
}

func (le *LineEdit) SetCaseMode(mode CaseMode) error {
	var set, clear uint32

	switch mode {
	case CaseModeMixed:
		clear = winuser.ES_UPPERCASE | winuser.ES_LOWERCASE

	case CaseModeUpper:
		set = winuser.ES_UPPERCASE
		clear = winuser.ES_LOWERCASE

	case CaseModeLower:
		set = winuser.ES_LOWERCASE
		clear = winuser.ES_UPPERCASE

	default:
		panic("invalid CaseMode")
	}

	return le.setAndClearStyleBits(set, clear)
}

func (le *LineEdit) PasswordMode() bool {
	return le.SendMessage(winuser.EM_GETPASSWORDCHAR, 0, 0) != 0
}

func (le *LineEdit) SetPasswordMode(value bool) {
	var c uintptr
	if value {
		c = uintptr('*')
	}

	le.SendMessage(winuser.EM_SETPASSWORDCHAR, c, 0)
}

func (le *LineEdit) ReadOnly() bool {
	return le.hasStyleBits(winuser.ES_READONLY)
}

func (le *LineEdit) SetReadOnly(readOnly bool) error {
	if le.SendMessage(winuser.EM_SETREADONLY, uintptr(win.BoolToBOOL(readOnly)), 0) == 0 {
		return errs.NewError("SendMessage(EM_SETREADONLY)")
	}

	if readOnly != le.ReadOnly() {
		le.invalidateBorderInParent()
	}

	le.readOnlyChangedPublisher.Publish()

	return nil
}

// sizeHintForLimit returns size hint for given limit in native pixels
func (le *LineEdit) sizeHintForLimit(limit int) (size Size) {
	size = le.dialogBaseUnitsToPixels(Size{50, 12})
	le.initCharWidth()
	n := le.MaxLength()
	if n > limit {
		n = limit
	}
	size.Width = le.charWidth * (n + 1)
	return
}

func (le *LineEdit) initCharWidth() {
	font := le.Font()
	if font == le.charWidthFont {
		return
	}
	le.charWidthFont = font
	le.charWidth = 8

	hdc := user32.GetDC(le.hWnd)
	if hdc == 0 {
		errs.NewError("GetDC failed")
		return
	}
	defer user32.ReleaseDC(le.hWnd, hdc)

	defer gdi32.SelectObject(hdc, gdi32.SelectObject(hdc, gdi32.HGDIOBJ(font.handleForDPI(le.DPI()))))

	buf := []uint16{'M'}

	var s gdi32.SIZE
	if !gdi32.GetTextExtentPoint32(hdc, &buf[0], int32(len(buf)), &s) {
		errs.NewError("GetTextExtentPoint32 failed")
		return
	}
	le.charWidth = int(s.CX)
}

func (le *LineEdit) EditingFinished() *Event {
	return le.editingFinishedPublisher.Event()
}

func (le *LineEdit) TextChanged() *Event {
	return le.textChangedPublisher.Event()
}

func (le *LineEdit) TextColor() Color {
	return le.textColor
}

func (le *LineEdit) SetTextColor(c Color) {
	le.textColor = c

	le.Invalidate()
}

func (*LineEdit) NeedsWmSize() bool {
	return true
}

func (le *LineEdit) WndProc(hwnd handle.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case user32.WM_COMMAND:
		switch win.HIWORD(uint32(wParam)) {
		case winuser.EN_CHANGE:
			le.textChangedPublisher.Publish()
		}

	case user32.WM_GETDLGCODE:
		if form := ancestor(le); form != nil {
			if dlg, ok := form.(dialogish); ok {
				if dlg.DefaultButton() != nil {
					// If the LineEdit lives in a Dialog that has a DefaultButton,
					// we won't swallow the return key.
					break
				}
			}
		}

		if wParam == user32.VK_RETURN {
			return user32.DLGC_WANTALLKEYS
		}

	case user32.WM_KEYDOWN:
		switch Key(wParam) {
		case KeyA:
			if ControlDown() {
				le.SetTextSelection(0, -1)
			}

		case KeyReturn:
			le.editingFinishedPublisher.Publish()
		}

	case user32.WM_KILLFOCUS:
		// FIXME: This may be dangerous, see remarks section:
		// http://msdn.microsoft.com/en-us/library/ms646282(v=vs.85).aspx
		le.editingFinishedPublisher.Publish()
	}

	return le.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}

func (le *LineEdit) CreateLayoutItem(ctx *LayoutContext) LayoutItem {
	lf := ShrinkableHorz | GrowableHorz
	if le.MaxLength() > lineEditGreedyLimit {
		lf |= GreedyHorz
	}

	return &lineEditLayoutItem{
		layoutFlags: lf,
		idealSize:   le.sizeHintForLimit(lineEditGreedyLimit),
		minSize:     le.sizeHintForLimit(lineEditMinChars),
	}
}

type lineEditLayoutItem struct {
	LayoutItemBase
	layoutFlags LayoutFlags
	idealSize   Size // in native pixels
	minSize     Size // in native pixels
}

func (li *lineEditLayoutItem) LayoutFlags() LayoutFlags {
	return li.layoutFlags
}

func (li *lineEditLayoutItem) IdealSize() Size {
	return li.idealSize
}

func (li *lineEditLayoutItem) MinSize() Size {
	return li.minSize
}

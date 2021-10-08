// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"syscall"
	"unsafe"

	"github.com/Gipcomp/win32/commctrl"
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/win"
)

// https://msdn.microsoft.com/en-us/library/windows/desktop/bb760416(v=vs.85).aspx says 80,
// but in reality, that hasn't been enforced for many many Windows versions. So we give it
// 1024 instead.
const maxToolTipTextLen = 1024 // including NUL terminator

type ToolTip struct {
	WindowBase
}

func NewToolTip() (*ToolTip, error) {
	tt, err := newToolTip(0)
	if err != nil {
		return nil, err
	}

	user32.SetWindowPos(tt.hWnd, user32.HWND_TOPMOST, 0, 0, 0, 0, user32.SWP_NOMOVE|user32.SWP_NOSIZE|user32.SWP_NOACTIVATE)

	return tt, nil
}

func newToolTip(style uint32) (*ToolTip, error) {
	tt := new(ToolTip)

	if err := InitWindow(
		tt,
		nil,
		"tooltips_class32",
		user32.WS_DISABLED|user32.WS_POPUP|commctrl.TTS_ALWAYSTIP|commctrl.TTS_NOPREFIX|style,
		0); err != nil {
		return nil, err
	}

	succeeded := false
	defer func() {
		if !succeeded {
			tt.Dispose()
		}
	}()

	tt.SendMessage(commctrl.TTM_SETMAXTIPWIDTH, 0, 300)

	succeeded = true

	return tt, nil
}

func (tt *ToolTip) Title() string {
	var gt commctrl.TTGETTITLE

	buf := make([]uint16, 100)

	gt.DwSize = uint32(unsafe.Sizeof(gt))
	gt.Cch = uint32(len(buf))
	gt.PszTitle = &buf[0]

	tt.SendMessage(commctrl.TTM_GETTITLE, 0, uintptr(unsafe.Pointer(&gt)))

	return syscall.UTF16ToString(buf)
}

func (tt *ToolTip) SetTitle(title string) error {
	return tt.setTitle(title, commctrl.TTI_NONE)
}

func (tt *ToolTip) SetInfoTitle(title string) error {
	return tt.setTitle(title, commctrl.TTI_INFO)
}

func (tt *ToolTip) SetWarningTitle(title string) error {
	return tt.setTitle(title, commctrl.TTI_WARNING)
}

func (tt *ToolTip) SetErrorTitle(title string) error {
	return tt.setTitle(title, commctrl.TTI_ERROR)
}

func (tt *ToolTip) setTitle(title string, icon uintptr) error {
	if len(title) > 99 {
		title = title[:99]
	}
	strPtr, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return err
	}
	if win.FALSE == tt.SendMessage(commctrl.TTM_SETTITLE, icon, uintptr(unsafe.Pointer(strPtr))) {
		return newError("TTM_SETTITLE failed")
	}

	return nil
}

func (tt *ToolTip) track(tool Widget) error {
	form := tool.Form()
	if form == nil {
		return nil
	}
	// HACK: We may have to delay this until the form is fully up to avoid glitches.
	if !form.AsFormBase().started {
		form.Starting().Once(func() {
			tt.track(tool)
		})
		return nil
	}

	ti := tt.toolInfo(tool.Handle())
	if ti == nil {
		return newError("unknown tool")
	}

	tt.SendMessage(commctrl.TTM_TRACKACTIVATE, 1, uintptr(unsafe.Pointer(ti)))

	b := tool.BoundsPixels()

	p := Point{0, b.Y + b.Height}.toPOINT()
	if form.RightToLeftLayout() {
		p.X = int32(b.X - b.Width/2)
	} else {
		p.X = int32(b.X + b.Width/2)
	}

	user32.ClientToScreen(tool.Parent().Handle(), &p)

	tt.SendMessage(commctrl.TTM_TRACKPOSITION, 0, uintptr(win.MAKELONG(uint16(p.X), uint16(p.Y))))

	var insertAfterHWND handle.HWND
	if form := tool.Form(); form != nil && user32.GetForegroundWindow() == form.Handle() {
		insertAfterHWND = user32.HWND_TOP
	} else {
		insertAfterHWND = tool.Handle()
	}
	user32.SetWindowPos(tt.hWnd, insertAfterHWND, 0, 0, 0, 0, user32.SWP_NOMOVE|user32.SWP_NOSIZE|user32.SWP_NOACTIVATE)

	return nil
}

func (tt *ToolTip) untrack(tool Widget) error {
	ti := tt.toolInfo(tool.Handle())
	if ti == nil {
		return newError("unknown tool")
	}

	tt.SendMessage(commctrl.TTM_TRACKACTIVATE, 0, uintptr(unsafe.Pointer(ti)))

	return nil
}

func (tt *ToolTip) AddTool(tool Widget) error {
	return tt.addTool(tt.hwndForTool(tool), false)
}

func (tt *ToolTip) addTrackedTool(tool Widget) error {
	return tt.addTool(tt.hwndForTool(tool), true)
}

func (tt *ToolTip) addTool(hwnd handle.HWND, track bool) error {
	if hwnd == 0 {
		return nil
	}

	var ti commctrl.TOOLINFO
	ti.CbSize = uint32(unsafe.Sizeof(ti))
	ti.Hwnd = hwnd
	ti.UFlags = commctrl.TTF_IDISHWND
	if track {
		ti.UFlags |= commctrl.TTF_TRACK
	} else {
		ti.UFlags |= commctrl.TTF_SUBCLASS
	}
	ti.UId = uintptr(hwnd)

	if win.FALSE == tt.SendMessage(commctrl.TTM_ADDTOOL, 0, uintptr(unsafe.Pointer(&ti))) {
		return newError("TTM_ADDTOOL failed")
	}

	return nil
}

func (tt *ToolTip) RemoveTool(tool Widget) error {
	return tt.removeTool(tt.hwndForTool(tool))
}

func (tt *ToolTip) removeTool(hwnd handle.HWND) error {
	var ti commctrl.TOOLINFO
	ti.CbSize = uint32(unsafe.Sizeof(ti))
	ti.Hwnd = hwnd
	ti.UId = uintptr(hwnd)

	tt.SendMessage(commctrl.TTM_DELTOOL, 0, uintptr(unsafe.Pointer(&ti)))

	return nil
}

func (tt *ToolTip) Text(tool Widget) string {
	return tt.text(tt.hwndForTool(tool))
}

func (tt *ToolTip) text(hwnd handle.HWND) string {
	ti := tt.toolInfo(hwnd)
	if ti == nil {
		return ""
	}

	return win.UTF16PtrToString(ti.LpszText)
}

func (tt *ToolTip) SetText(tool Widget, text string) error {
	return tt.setText(tt.hwndForTool(tool), text)
}

func (tt *ToolTip) setText(hwnd handle.HWND, text string) error {
	ti := tt.toolInfo(hwnd)
	if ti == nil {
		return newError("unknown tool")
	}

	n := 0
	for i, r := range text {
		if r < 0x10000 {
			n++
		} else {
			n += 2 // surrogate pair
		}
		if n >= maxToolTipTextLen {
			text = text[:i]
			break
		}
	}
	var err error
	ti.LpszText, err = syscall.UTF16PtrFromString(text)
	if err != nil {
		return err
	}

	tt.SendMessage(commctrl.TTM_SETTOOLINFO, 0, uintptr(unsafe.Pointer(ti)))

	return nil
}

func (tt *ToolTip) toolInfo(hwnd handle.HWND) *commctrl.TOOLINFO {
	var ti commctrl.TOOLINFO
	var buf [maxToolTipTextLen]uint16

	ti.CbSize = uint32(unsafe.Sizeof(ti))
	ti.Hwnd = hwnd
	ti.UId = uintptr(hwnd)
	ti.LpszText = &buf[0]

	if win.FALSE == tt.SendMessage(commctrl.TTM_GETTOOLINFO, 0, uintptr(unsafe.Pointer(&ti))) {
		return nil
	}

	return &ti
}

func (*ToolTip) hwndForTool(tool Widget) handle.HWND {
	if hftt, ok := tool.(interface{ handleForToolTip() handle.HWND }); ok {
		return hftt.handleForToolTip()
	}

	return tool.Handle()
}

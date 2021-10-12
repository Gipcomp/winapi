// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"unsafe"

	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/winapi/errs"
)

const mainWindowWindowClass = `\o/ Walk_MainWindow_Class \o/`

func init() {
	AppendToWalkInit(func() {
		MustRegisterWindowClass(mainWindowWindowClass)
	})
}

type MainWindowCfg struct {
	Name   string
	Bounds Rectangle
}

type MainWindow struct {
	FormBase
	windowPlacement *user32.WINDOWPLACEMENT
	menu            *Menu
	toolBar         *ToolBar
	statusBar       *StatusBar
}

func NewMainWindow() (*MainWindow, error) {
	return NewMainWindowWithName("")
}

func NewMainWindowWithName(name string) (*MainWindow, error) {
	return NewMainWindowWithCfg(&MainWindowCfg{Name: name})
}

func NewMainWindowWithCfg(cfg *MainWindowCfg) (*MainWindow, error) {
	mw := new(MainWindow)
	mw.SetName(cfg.Name)

	if err := initWindowWithCfg(&windowCfg{
		Window:    mw,
		ClassName: mainWindowWindowClass,
		Style:     user32.WS_OVERLAPPEDWINDOW,
		ExStyle:   user32.WS_EX_CONTROLPARENT,
		Bounds:    cfg.Bounds,
	}); err != nil {
		return nil, err
	}

	succeeded := false
	defer func() {
		if !succeeded {
			mw.Dispose()
		}
	}()

	mw.SetPersistent(true)

	var err error

	if mw.menu, err = newMenuBar(mw); err != nil {
		return nil, err
	}
	if !user32.SetMenu(mw.hWnd, mw.menu.hMenu) {
		return nil, errs.LastError("SetMenu")
	}

	tb, err := NewToolBar(mw)
	if err != nil {
		return nil, err
	}
	mw.SetToolBar(tb)

	if mw.statusBar, err = NewStatusBar(mw); err != nil {
		return nil, err
	}
	mw.statusBar.parent = nil
	mw.Children().Remove(mw.statusBar)
	mw.statusBar.parent = mw
	user32.SetParent(mw.statusBar.hWnd, mw.hWnd)
	mw.statusBar.visibleChangedPublisher.event.Attach(func() {
		mw.SetBoundsPixels(mw.BoundsPixels())
	})

	succeeded = true

	return mw, nil
}

func (mw *MainWindow) Menu() *Menu {
	return mw.menu
}

func (mw *MainWindow) ToolBar() *ToolBar {
	return mw.toolBar
}

func (mw *MainWindow) SetToolBar(tb *ToolBar) {
	if mw.toolBar != nil {
		user32.SetParent(mw.toolBar.hWnd, 0)
	}

	if tb != nil {
		parent := tb.parent
		tb.parent = nil
		parent.Children().Remove(tb)
		tb.parent = mw
		user32.SetParent(tb.hWnd, mw.hWnd)
	}

	mw.toolBar = tb
}

func (mw *MainWindow) StatusBar() *StatusBar {
	return mw.statusBar
}

func (mw *MainWindow) ClientBoundsPixels() Rectangle {
	bounds := mw.FormBase.ClientBoundsPixels()

	if mw.toolBar != nil && mw.toolBar.Actions().Len() > 0 {
		tlbBounds := mw.toolBar.BoundsPixels()

		bounds.Y += tlbBounds.Height
		bounds.Height -= tlbBounds.Height
	}

	if mw.statusBar.Visible() {
		bounds.Height -= mw.statusBar.HeightPixels()
	}

	return bounds
}

func (mw *MainWindow) SetVisible(visible bool) {
	if visible {
		user32.DrawMenuBar(mw.hWnd)

		mw.clientComposite.RequestLayout()
	}

	mw.FormBase.SetVisible(visible)
}

func (mw *MainWindow) applyFont(font *Font) {
	mw.FormBase.applyFont(font)

	if mw.toolBar != nil {
		mw.toolBar.applyFont(font)
	}

	if mw.statusBar != nil {
		mw.statusBar.applyFont(font)
	}
}

func (mw *MainWindow) Fullscreen() bool {
	return user32.GetWindowLong(mw.hWnd, user32.GWL_STYLE)&user32.WS_OVERLAPPEDWINDOW == 0
}

func (mw *MainWindow) SetFullscreen(fullscreen bool) error {
	if fullscreen == mw.Fullscreen() {
		return nil
	}

	if fullscreen {
		var mi user32.MONITORINFO
		mi.CbSize = uint32(unsafe.Sizeof(mi))

		if mw.windowPlacement == nil {
			mw.windowPlacement = new(user32.WINDOWPLACEMENT)
		}

		if !user32.GetWindowPlacement(mw.hWnd, mw.windowPlacement) {
			return errs.LastError("GetWindowPlacement")
		}
		if !user32.GetMonitorInfo(user32.MonitorFromWindow(
			mw.hWnd, user32.MONITOR_DEFAULTTOPRIMARY), &mi) {

			return errs.NewError("GetMonitorInfo")
		}

		if err := mw.ensureStyleBits(user32.WS_OVERLAPPEDWINDOW, false); err != nil {
			return err
		}

		if r := mi.RcMonitor; !user32.SetWindowPos(
			mw.hWnd, user32.HWND_TOP,
			r.Left, r.Top, r.Right-r.Left, r.Bottom-r.Top,
			user32.SWP_FRAMECHANGED|user32.SWP_NOOWNERZORDER) {

			return errs.LastError("SetWindowPos")
		}
	} else {
		if err := mw.ensureStyleBits(user32.WS_OVERLAPPEDWINDOW, true); err != nil {
			return err
		}

		if !user32.SetWindowPlacement(mw.hWnd, mw.windowPlacement) {
			return errs.LastError("SetWindowPlacement")
		}

		if !user32.SetWindowPos(mw.hWnd, 0, 0, 0, 0, 0, user32.SWP_FRAMECHANGED|user32.SWP_NOMOVE|
			user32.SWP_NOOWNERZORDER|user32.SWP_NOSIZE|user32.SWP_NOZORDER) {

			return errs.LastError("SetWindowPos")
		}
	}

	return nil
}

func (mw *MainWindow) WndProc(hwnd handle.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case user32.WM_WINDOWPOSCHANGED, user32.WM_SIZE:
		if user32.WM_WINDOWPOSCHANGED == msg {
			wp := (*user32.WINDOWPOS)(unsafe.Pointer(lParam))
			if wp.Flags&user32.SWP_NOSIZE != 0 {
				break
			}
		}

		cb := mw.ClientBoundsPixels()

		if mw.toolBar != nil {
			bounds := Rectangle{0, 0, cb.Width, mw.toolBar.HeightPixels()}
			if mw.toolBar.BoundsPixels() != bounds {
				mw.toolBar.SetBoundsPixels(bounds)
			}
		}

		bounds := Rectangle{0, cb.Y + cb.Height, cb.Width, mw.statusBar.HeightPixels()}
		if mw.statusBar.BoundsPixels() != bounds {
			mw.statusBar.SetBoundsPixels(bounds)
		}

	case user32.WM_INITMENUPOPUP:
		mw.menu.updateItemsWithImageForWindow(mw)
	}

	return mw.FormBase.WndProc(hwnd, msg, wParam, lParam)
}

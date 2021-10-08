// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"syscall"

	"github.com/Gipcomp/win32/gdi32"
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/ole32"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/win"
	"github.com/Gipcomp/win32/winuser"
)

var webViewIOleInPlaceFrameVtbl *ole32.IOleInPlaceFrameVtbl

func init() {
	AppendToWalkInit(func() {
		webViewIOleInPlaceFrameVtbl = &ole32.IOleInPlaceFrameVtbl{
			QueryInterface:       syscall.NewCallback(webView_IOleInPlaceFrame_QueryInterface),
			AddRef:               syscall.NewCallback(webView_IOleInPlaceFrame_AddRef),
			Release:              syscall.NewCallback(webView_IOleInPlaceFrame_Release),
			GetWindow:            syscall.NewCallback(webView_IOleInPlaceFrame_GetWindow),
			ContextSensitiveHelp: syscall.NewCallback(webView_IOleInPlaceFrame_ContextSensitiveHelp),
			GetBorder:            syscall.NewCallback(webView_IOleInPlaceFrame_GetBorder),
			RequestBorderSpace:   syscall.NewCallback(webView_IOleInPlaceFrame_RequestBorderSpace),
			SetBorderSpace:       syscall.NewCallback(webView_IOleInPlaceFrame_SetBorderSpace),
			SetActiveObject:      syscall.NewCallback(webView_IOleInPlaceFrame_SetActiveObject),
			InsertMenus:          syscall.NewCallback(webView_IOleInPlaceFrame_InsertMenus),
			SetMenu:              syscall.NewCallback(webView_IOleInPlaceFrame_SetMenu),
			RemoveMenus:          syscall.NewCallback(webView_IOleInPlaceFrame_RemoveMenus),
			SetStatusText:        syscall.NewCallback(webView_IOleInPlaceFrame_SetStatusText),
			EnableModeless:       syscall.NewCallback(webView_IOleInPlaceFrame_EnableModeless),
			TranslateAccelerator: syscall.NewCallback(webView_IOleInPlaceFrame_TranslateAccelerator),
		}
	})
}

type webViewIOleInPlaceFrame struct {
	ole32.IOleInPlaceFrame
	webView *WebView
}

func webView_IOleInPlaceFrame_QueryInterface(inPlaceFrame *webViewIOleInPlaceFrame, riid ole32.REFIID, ppvObj *uintptr) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleInPlaceFrame_AddRef(inPlaceFrame *webViewIOleInPlaceFrame) uintptr {
	return 1
}

func webView_IOleInPlaceFrame_Release(inPlaceFrame *webViewIOleInPlaceFrame) uintptr {
	return 1
}

func webView_IOleInPlaceFrame_GetWindow(inPlaceFrame *webViewIOleInPlaceFrame, lphwnd *handle.HWND) uintptr {
	*lphwnd = inPlaceFrame.webView.hWnd

	return win.S_OK
}

func webView_IOleInPlaceFrame_ContextSensitiveHelp(inPlaceFrame *webViewIOleInPlaceFrame, fEnterMode win.BOOL) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleInPlaceFrame_GetBorder(inPlaceFrame *webViewIOleInPlaceFrame, lprectBorder *gdi32.RECT) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleInPlaceFrame_RequestBorderSpace(inPlaceFrame *webViewIOleInPlaceFrame, pborderwidths uintptr) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleInPlaceFrame_SetBorderSpace(inPlaceFrame *webViewIOleInPlaceFrame, pborderwidths uintptr) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleInPlaceFrame_SetActiveObject(inPlaceFrame *webViewIOleInPlaceFrame, pActiveObject uintptr, pszObjName *uint16) uintptr {
	return win.S_OK
}

func webView_IOleInPlaceFrame_InsertMenus(inPlaceFrame *webViewIOleInPlaceFrame, hmenuShared winuser.HMENU, lpMenuWidths uintptr) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleInPlaceFrame_SetMenu(inPlaceFrame *webViewIOleInPlaceFrame, hmenuShared winuser.HMENU, holemenu winuser.HMENU, hwndActiveObject handle.HWND) uintptr {
	return win.S_OK
}

func webView_IOleInPlaceFrame_RemoveMenus(inPlaceFrame *webViewIOleInPlaceFrame, hmenuShared winuser.HMENU) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleInPlaceFrame_SetStatusText(inPlaceFrame *webViewIOleInPlaceFrame, pszStatusText *uint16) uintptr {
	return win.S_OK
}

func webView_IOleInPlaceFrame_EnableModeless(inPlaceFrame *webViewIOleInPlaceFrame, fEnable win.BOOL) uintptr {
	return win.S_OK
}

func webView_IOleInPlaceFrame_TranslateAccelerator(inPlaceFrame *webViewIOleInPlaceFrame, lpmsg *user32.MSG, wID uint32) uintptr {
	return win.E_NOTIMPL
}

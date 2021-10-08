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
	"github.com/Gipcomp/win32/ole32"
	"github.com/Gipcomp/win32/win"
)

var webViewIOleInPlaceSiteVtbl *ole32.IOleInPlaceSiteVtbl

func init() {
	AppendToWalkInit(func() {
		webViewIOleInPlaceSiteVtbl = &ole32.IOleInPlaceSiteVtbl{
			QueryInterface:       syscall.NewCallback(webView_IOleInPlaceSite_QueryInterface),
			AddRef:               syscall.NewCallback(webView_IOleInPlaceSite_AddRef),
			Release:              syscall.NewCallback(webView_IOleInPlaceSite_Release),
			GetWindow:            syscall.NewCallback(webView_IOleInPlaceSite_GetWindow),
			ContextSensitiveHelp: syscall.NewCallback(webView_IOleInPlaceSite_ContextSensitiveHelp),
			CanInPlaceActivate:   syscall.NewCallback(webView_IOleInPlaceSite_CanInPlaceActivate),
			OnInPlaceActivate:    syscall.NewCallback(webView_IOleInPlaceSite_OnInPlaceActivate),
			OnUIActivate:         syscall.NewCallback(webView_IOleInPlaceSite_OnUIActivate),
			GetWindowContext:     syscall.NewCallback(webView_IOleInPlaceSite_GetWindowContext),
			Scroll:               syscall.NewCallback(webView_IOleInPlaceSite_Scroll),
			OnUIDeactivate:       syscall.NewCallback(webView_IOleInPlaceSite_OnUIDeactivate),
			OnInPlaceDeactivate:  syscall.NewCallback(webView_IOleInPlaceSite_OnInPlaceDeactivate),
			DiscardUndoState:     syscall.NewCallback(webView_IOleInPlaceSite_DiscardUndoState),
			DeactivateAndUndo:    syscall.NewCallback(webView_IOleInPlaceSite_DeactivateAndUndo),
			OnPosRectChange:      syscall.NewCallback(webView_IOleInPlaceSite_OnPosRectChange),
		}
	})
}

type webViewIOleInPlaceSite struct {
	ole32.IOleInPlaceSite
	inPlaceFrame webViewIOleInPlaceFrame
}

func webView_IOleInPlaceSite_QueryInterface(inPlaceSite *webViewIOleInPlaceSite, riid ole32.REFIID, ppvObject *unsafe.Pointer) uintptr {
	// Just reuse the QueryInterface implementation we have for IOleClientSite.
	// We need to adjust object from the webViewIDocHostUIHandler to the
	// containing webViewIOleInPlaceSite.
	var clientSite ole32.IOleClientSite

	ptr := uintptr(unsafe.Pointer(inPlaceSite)) - uintptr(unsafe.Sizeof(clientSite))

	return webView_IOleClientSite_QueryInterface((*webViewIOleClientSite)(unsafe.Pointer(ptr)), riid, ppvObject)
}

func webView_IOleInPlaceSite_AddRef(inPlaceSite *webViewIOleInPlaceSite) uintptr {
	return 1
}

func webView_IOleInPlaceSite_Release(inPlaceSite *webViewIOleInPlaceSite) uintptr {
	return 1
}

func webView_IOleInPlaceSite_GetWindow(inPlaceSite *webViewIOleInPlaceSite, lphwnd *handle.HWND) uintptr {
	*lphwnd = inPlaceSite.inPlaceFrame.webView.hWnd

	return win.S_OK
}

func webView_IOleInPlaceSite_ContextSensitiveHelp(inPlaceSite *webViewIOleInPlaceSite, fEnterMode win.BOOL) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleInPlaceSite_CanInPlaceActivate(inPlaceSite *webViewIOleInPlaceSite) uintptr {
	return win.S_OK
}

func webView_IOleInPlaceSite_OnInPlaceActivate(inPlaceSite *webViewIOleInPlaceSite) uintptr {
	return win.S_OK
}

func webView_IOleInPlaceSite_OnUIActivate(inPlaceSite *webViewIOleInPlaceSite) uintptr {
	return win.S_OK
}

func webView_IOleInPlaceSite_GetWindowContext(inPlaceSite *webViewIOleInPlaceSite, lplpFrame **webViewIOleInPlaceFrame, lplpDoc *uintptr, lprcPosRect, lprcClipRect *gdi32.RECT, lpFrameInfo *ole32.OLEINPLACEFRAMEINFO) uintptr {
	*lplpFrame = &inPlaceSite.inPlaceFrame
	*lplpDoc = 0

	lpFrameInfo.FMDIApp = win.FALSE
	lpFrameInfo.HwndFrame = inPlaceSite.inPlaceFrame.webView.hWnd
	lpFrameInfo.Haccel = 0
	lpFrameInfo.CAccelEntries = 0

	return win.S_OK
}

func webView_IOleInPlaceSite_Scroll(inPlaceSite *webViewIOleInPlaceSite, scrollExtentX, scrollExtentY int32) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleInPlaceSite_OnUIDeactivate(inPlaceSite *webViewIOleInPlaceSite, fUndoable win.BOOL) uintptr {
	return win.S_OK
}

func webView_IOleInPlaceSite_OnInPlaceDeactivate(inPlaceSite *webViewIOleInPlaceSite) uintptr {
	return win.S_OK
}

func webView_IOleInPlaceSite_DiscardUndoState(inPlaceSite *webViewIOleInPlaceSite) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleInPlaceSite_DeactivateAndUndo(inPlaceSite *webViewIOleInPlaceSite) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleInPlaceSite_OnPosRectChange(inPlaceSite *webViewIOleInPlaceSite, lprcPosRect *gdi32.RECT) uintptr {
	browserObject := inPlaceSite.inPlaceFrame.webView.browserObject
	var inPlaceObjectPtr unsafe.Pointer
	if hr := browserObject.QueryInterface(&ole32.IID_IOleInPlaceObject, &inPlaceObjectPtr); win.FAILED(hr) {
		return uintptr(hr)
	}
	inPlaceObject := (*ole32.IOleInPlaceObject)(inPlaceObjectPtr)
	defer inPlaceObject.Release()

	return uintptr(inPlaceObject.SetObjectRects(lprcPosRect, lprcPosRect))
}

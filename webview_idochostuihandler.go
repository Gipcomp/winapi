// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"syscall"
	"unsafe"

	"github.com/Gipcomp/win32/gdi32"
	"github.com/Gipcomp/win32/ole32"
	"github.com/Gipcomp/win32/shdocvw"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/win"
)

var webViewIDocHostUIHandlerVtbl *shdocvw.IDocHostUIHandlerVtbl

func init() {
	AppendToWalkInit(func() {
		webViewIDocHostUIHandlerVtbl = &shdocvw.IDocHostUIHandlerVtbl{
			QueryInterface:        syscall.NewCallback(webView_IDocHostUIHandler_QueryInterface),
			AddRef:                syscall.NewCallback(webView_IDocHostUIHandler_AddRef),
			Release:               syscall.NewCallback(webView_IDocHostUIHandler_Release),
			ShowContextMenu:       syscall.NewCallback(webView_IDocHostUIHandler_ShowContextMenu),
			GetHostInfo:           syscall.NewCallback(webView_IDocHostUIHandler_GetHostInfo),
			ShowUI:                syscall.NewCallback(webView_IDocHostUIHandler_ShowUI),
			HideUI:                syscall.NewCallback(webView_IDocHostUIHandler_HideUI),
			UpdateUI:              syscall.NewCallback(webView_IDocHostUIHandler_UpdateUI),
			EnableModeless:        syscall.NewCallback(webView_IDocHostUIHandler_EnableModeless),
			OnDocWindowActivate:   syscall.NewCallback(webView_IDocHostUIHandler_OnDocWindowActivate),
			OnFrameWindowActivate: syscall.NewCallback(webView_IDocHostUIHandler_OnFrameWindowActivate),
			ResizeBorder:          syscall.NewCallback(webView_IDocHostUIHandler_ResizeBorder),
			TranslateAccelerator:  syscall.NewCallback(webView_IDocHostUIHandler_TranslateAccelerator),
			GetOptionKeyPath:      syscall.NewCallback(webView_IDocHostUIHandler_GetOptionKeyPath),
			GetDropTarget:         syscall.NewCallback(webView_IDocHostUIHandler_GetDropTarget),
			GetExternal:           syscall.NewCallback(webView_IDocHostUIHandler_GetExternal),
			TranslateUrl:          syscall.NewCallback(webView_IDocHostUIHandler_TranslateUrl),
			FilterDataObject:      syscall.NewCallback(webView_IDocHostUIHandler_FilterDataObject),
		}
	})
}

type webViewIDocHostUIHandler struct {
	shdocvw.IDocHostUIHandler
}

func webView_IDocHostUIHandler_QueryInterface(docHostUIHandler *webViewIDocHostUIHandler, riid ole32.REFIID, ppvObject *unsafe.Pointer) uintptr {
	// Just reuse the QueryInterface implementation we have for IOleClientSite.
	// We need to adjust object, which initially points at our
	// webViewIDocHostUIHandler, so it refers to the containing
	// webViewIOleClientSite for the call.
	var clientSite ole32.IOleClientSite
	var webViewInPlaceSite webViewIOleInPlaceSite

	ptr := uintptr(unsafe.Pointer(docHostUIHandler)) - uintptr(unsafe.Sizeof(clientSite)) -
		uintptr(unsafe.Sizeof(webViewInPlaceSite))

	return webView_IOleClientSite_QueryInterface((*webViewIOleClientSite)(unsafe.Pointer(ptr)), riid, ppvObject)
}

func webView_IDocHostUIHandler_AddRef(docHostUIHandler *webViewIDocHostUIHandler) uintptr {
	return 1
}

func webView_IDocHostUIHandler_Release(docHostUIHandler *webViewIDocHostUIHandler) uintptr {
	return 1
}

func webView_IDocHostUIHandler_ShowContextMenu(docHostUIHandler *webViewIDocHostUIHandler, dwID uint32, ppt *gdi32.POINT, pcmdtReserved *ole32.IUnknown, pdispReserved uintptr) uintptr {
	var webViewInPlaceSite webViewIOleInPlaceSite
	var iOleClientSite ole32.IOleClientSite
	var wb WidgetBase
	ptr := uintptr(unsafe.Pointer(docHostUIHandler)) -
		uintptr(unsafe.Sizeof(webViewInPlaceSite)) -
		uintptr(unsafe.Sizeof(iOleClientSite)) -
		uintptr(unsafe.Sizeof(wb))
	webView := (*WebView)(unsafe.Pointer(ptr))

	// show context menu
	if webView.NativeContextMenuEnabled() {
		return win.S_FALSE
	}

	return win.S_OK
}

func webView_IDocHostUIHandler_GetHostInfo(docHostUIHandler *webViewIDocHostUIHandler, pInfo *shdocvw.DOCHOSTUIINFO) uintptr {
	pInfo.CbSize = uint32(unsafe.Sizeof(*pInfo))
	pInfo.DwFlags = shdocvw.DOCHOSTUIFLAG_NO3DBORDER
	pInfo.DwDoubleClick = shdocvw.DOCHOSTUIDBLCLK_DEFAULT

	return win.S_OK
}

func webView_IDocHostUIHandler_ShowUI(docHostUIHandler *webViewIDocHostUIHandler, dwID uint32, pActiveObject uintptr, pCommandTarget uintptr, pFrame *ole32.IOleInPlaceFrame, pDoc uintptr) uintptr {
	return win.S_OK
}

func webView_IDocHostUIHandler_HideUI(docHostUIHandler *webViewIDocHostUIHandler) uintptr {
	return win.S_OK
}

func webView_IDocHostUIHandler_UpdateUI(docHostUIHandler *webViewIDocHostUIHandler) uintptr {
	return win.S_OK
}

func webView_IDocHostUIHandler_EnableModeless(docHostUIHandler *webViewIDocHostUIHandler, fEnable win.BOOL) uintptr {
	return win.S_OK
}

func webView_IDocHostUIHandler_OnDocWindowActivate(docHostUIHandler *webViewIDocHostUIHandler, fActivate win.BOOL) uintptr {
	return win.S_OK
}

func webView_IDocHostUIHandler_OnFrameWindowActivate(docHostUIHandler *webViewIDocHostUIHandler, fActivate win.BOOL) uintptr {
	return win.S_OK
}

func webView_IDocHostUIHandler_ResizeBorder(docHostUIHandler *webViewIDocHostUIHandler, prcBorder *gdi32.RECT, pUIWindow uintptr, fRameWindow win.BOOL) uintptr {
	return win.S_OK
}

func webView_IDocHostUIHandler_TranslateAccelerator(docHostUIHandler *webViewIDocHostUIHandler, lpMsg *user32.MSG, pguidCmdGroup *syscall.GUID, nCmdID uint) uintptr {
	return win.S_FALSE
}

func webView_IDocHostUIHandler_GetOptionKeyPath(docHostUIHandler *webViewIDocHostUIHandler, pchKey *uint16, dw uint) uintptr {
	return win.S_FALSE
}

func webView_IDocHostUIHandler_GetDropTarget(docHostUIHandler *webViewIDocHostUIHandler, pDropTarget uintptr, ppDropTarget *uintptr) uintptr {
	return win.S_FALSE
}

func webView_IDocHostUIHandler_GetExternal(docHostUIHandler *webViewIDocHostUIHandler, ppDispatch *uintptr) uintptr {
	*ppDispatch = 0

	return win.S_FALSE
}

func webView_IDocHostUIHandler_TranslateUrl(docHostUIHandler *webViewIDocHostUIHandler, dwTranslate uint32, pchURLIn *uint16, ppchURLOut **uint16) uintptr {
	*ppchURLOut = nil

	return win.S_FALSE
}

func webView_IDocHostUIHandler_FilterDataObject(docHostUIHandler *webViewIDocHostUIHandler, pDO uintptr, ppDORet *uintptr) uintptr {
	*ppDORet = 0

	return win.S_FALSE
}

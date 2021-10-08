// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"syscall"
	"unsafe"

	"github.com/Gipcomp/win32/ole32"
	"github.com/Gipcomp/win32/shdocvw"
	"github.com/Gipcomp/win32/win"
)

var webViewIOleClientSiteVtbl *ole32.IOleClientSiteVtbl

func init() {
	AppendToWalkInit(func() {
		webViewIOleClientSiteVtbl = &ole32.IOleClientSiteVtbl{
			QueryInterface:         syscall.NewCallback(webView_IOleClientSite_QueryInterface),
			AddRef:                 syscall.NewCallback(webView_IOleClientSite_AddRef),
			Release:                syscall.NewCallback(webView_IOleClientSite_Release),
			SaveObject:             syscall.NewCallback(webView_IOleClientSite_SaveObject),
			GetMoniker:             syscall.NewCallback(webView_IOleClientSite_GetMoniker),
			GetContainer:           syscall.NewCallback(webView_IOleClientSite_GetContainer),
			ShowObject:             syscall.NewCallback(webView_IOleClientSite_ShowObject),
			OnShowWindow:           syscall.NewCallback(webView_IOleClientSite_OnShowWindow),
			RequestNewObjectLayout: syscall.NewCallback(webView_IOleClientSite_RequestNewObjectLayout),
		}
	})
}

type webViewIOleClientSite struct {
	ole32.IOleClientSite
	inPlaceSite       webViewIOleInPlaceSite
	docHostUIHandler  webViewIDocHostUIHandler
	webBrowserEvents2 webViewDWebBrowserEvents2
}

func webView_IOleClientSite_QueryInterface(clientSite *webViewIOleClientSite, riid ole32.REFIID, ppvObject *unsafe.Pointer) uintptr {
	if ole32.EqualREFIID(riid, &ole32.IID_IUnknown) {
		*ppvObject = unsafe.Pointer(clientSite)
	} else if ole32.EqualREFIID(riid, &ole32.IID_IOleClientSite) {
		*ppvObject = unsafe.Pointer(clientSite)
	} else if ole32.EqualREFIID(riid, &ole32.IID_IOleInPlaceSite) {
		*ppvObject = unsafe.Pointer(&clientSite.inPlaceSite)
	} else if ole32.EqualREFIID(riid, &shdocvw.IID_IDocHostUIHandler) {
		*ppvObject = unsafe.Pointer(&clientSite.docHostUIHandler)
	} else if ole32.EqualREFIID(riid, &shdocvw.DIID_DWebBrowserEvents2) {
		*ppvObject = unsafe.Pointer(&clientSite.webBrowserEvents2)
	} else {
		*ppvObject = nil
		return win.E_NOINTERFACE
	}

	return win.S_OK
}

func webView_IOleClientSite_AddRef(clientSite *webViewIOleClientSite) uintptr {
	return 1
}

func webView_IOleClientSite_Release(clientSite *webViewIOleClientSite) uintptr {
	return 1
}

func webView_IOleClientSite_SaveObject(clientSite *webViewIOleClientSite) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleClientSite_GetMoniker(clientSite *webViewIOleClientSite, dwAssign, dwWhichMoniker uint32, ppmk *unsafe.Pointer) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleClientSite_GetContainer(clientSite *webViewIOleClientSite, ppContainer *unsafe.Pointer) uintptr {
	*ppContainer = nil

	return win.E_NOINTERFACE
}

func webView_IOleClientSite_ShowObject(clientSite *webViewIOleClientSite) uintptr {
	return win.S_OK
}

func webView_IOleClientSite_OnShowWindow(clientSite *webViewIOleClientSite, fShow win.BOOL) uintptr {
	return win.E_NOTIMPL
}

func webView_IOleClientSite_RequestNewObjectLayout(clientSite *webViewIOleClientSite) uintptr {
	return win.E_NOTIMPL
}

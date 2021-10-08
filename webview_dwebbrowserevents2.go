// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"syscall"
	"time"
	"unsafe"

	"github.com/Gipcomp/win32/ole32"
	"github.com/Gipcomp/win32/oleaut32"
	"github.com/Gipcomp/win32/shdocvw"
	"github.com/Gipcomp/win32/win"
)

var webViewDWebBrowserEvents2Vtbl *shdocvw.DWebBrowserEvents2Vtbl

func init() {
	AppendToWalkInit(func() {
		webViewDWebBrowserEvents2Vtbl = &shdocvw.DWebBrowserEvents2Vtbl{
			QueryInterface:   syscall.NewCallback(webView_DWebBrowserEvents2_QueryInterface),
			AddRef:           syscall.NewCallback(webView_DWebBrowserEvents2_AddRef),
			Release:          syscall.NewCallback(webView_DWebBrowserEvents2_Release),
			GetTypeInfoCount: syscall.NewCallback(webView_DWebBrowserEvents2_GetTypeInfoCount),
			GetTypeInfo:      syscall.NewCallback(webView_DWebBrowserEvents2_GetTypeInfo),
			GetIDsOfNames:    syscall.NewCallback(webView_DWebBrowserEvents2_GetIDsOfNames),
			Invoke:           syscall.NewCallback(webView_DWebBrowserEvents2_Invoke),
		}
	})
}

type webViewDWebBrowserEvents2 struct {
	shdocvw.DWebBrowserEvents2
}

func webView_DWebBrowserEvents2_QueryInterface(wbe2 *webViewDWebBrowserEvents2, riid ole32.REFIID, ppvObject *unsafe.Pointer) uintptr {
	// Just reuse the QueryInterface implementation we have for IOleClientSite.
	// We need to adjust object, which initially points at our
	// webViewDWebBrowserEvents2, so it refers to the containing
	// webViewIOleClientSite for the call.
	var clientSite ole32.IOleClientSite
	var webViewInPlaceSite webViewIOleInPlaceSite
	var docHostUIHandler webViewIDocHostUIHandler

	ptr := uintptr(unsafe.Pointer(wbe2)) -
		uintptr(unsafe.Sizeof(clientSite)) -
		uintptr(unsafe.Sizeof(webViewInPlaceSite)) -
		uintptr(unsafe.Sizeof(docHostUIHandler))

	return webView_IOleClientSite_QueryInterface((*webViewIOleClientSite)(unsafe.Pointer(ptr)), riid, ppvObject)
}

func webView_DWebBrowserEvents2_AddRef(args *uintptr) uintptr {
	return 1
}

func webView_DWebBrowserEvents2_Release(args *uintptr) uintptr {
	return 1
}

func webView_DWebBrowserEvents2_GetTypeInfoCount(args *uintptr) uintptr {
	/*	p := (*struct {
			wbe2    *webViewDWebBrowserEvents2
			pctinfo *uint
		})(unsafe.Pointer(args))

		*p.pctinfo = 0

		return S_OK*/

	return win.E_NOTIMPL
}

func webView_DWebBrowserEvents2_GetTypeInfo(args *uintptr) uintptr {
	/*	p := (*struct {
				wbe2         *webViewDWebBrowserEvents2
			})(unsafe.Pointer(args))

		    unsigned int  iTInfo,
		    LCID  lcid,
		    ITypeInfo FAR* FAR*  ppTInfo*/

	return win.E_NOTIMPL
}

func webView_DWebBrowserEvents2_GetIDsOfNames(args *uintptr) uintptr {
	/*	p := (*struct {
		wbe2      *webViewDWebBrowserEvents2
		riid      REFIID
		rgszNames **uint16
		cNames    uint32
		lcid      LCID
		rgDispId  *DISPID
	})(unsafe.Pointer(args))*/

	return win.E_NOTIMPL
}

/*
func webView_DWebBrowserEvents2_Invoke(
	wbe2 *webViewDWebBrowserEvents2,
	dispIdMember oleaut32.DISPID,
	riid ole32.REFIID,
	lcid uint32, // LCID
	wFlags uint16,
	pDispParams *oleaut32.DISPPARAMS,
	pVarResult *oleaut32.VARIANT,
	pExcepInfo unsafe.Pointer, // *EXCEPINFO
	puArgErr *uint32) uintptr {
*/
func webView_DWebBrowserEvents2_Invoke(
	arg0 uintptr,
	arg1 uintptr,
	arg2 uintptr,
	arg3 uintptr,
	arg4 uintptr,
	arg5 uintptr,
	arg6 uintptr,
	arg7 uintptr,
	arg8 uintptr) uintptr {

	wbe2 := (*webViewDWebBrowserEvents2)(unsafe.Pointer(arg0))
	dispIdMember := *(*oleaut32.DISPID)(unsafe.Pointer(&arg1))
	//riid := *(*ole32.REFIID)(unsafe.Pointer(&arg2))
	//lcid := *(*uint32)(unsafe.Pointer(&arg3))
	//wFlags := *(*uint16)(unsafe.Pointer(&arg4))
	pDispParams := (*oleaut32.DISPPARAMS)(unsafe.Pointer(arg5))
	//pVarResult := (*oleaut32.VARIANT)(unsafe.Pointer(arg6))
	//pExcepInfo := unsafe.Pointer(arg7)
	//puArgErr := (*uint32)(unsafe.Pointer(arg8))

	var wb WidgetBase
	var wvcs webViewIOleClientSite

	wv := (*WebView)(unsafe.Pointer(uintptr(unsafe.Pointer(wbe2)) +
		uintptr(unsafe.Sizeof(*wbe2)) -
		uintptr(unsafe.Sizeof(wvcs)) -
		uintptr(unsafe.Sizeof(wb))))

	switch dispIdMember {
	case oleaut32.DISPID_BEFORENAVIGATE2:
		rgvargPtr := (*[7]oleaut32.VARIANTARG)(unsafe.Pointer(pDispParams.Rgvarg))
		eventData := &WebViewNavigatingEventData{
			pDisp:           (*rgvargPtr)[6].MustPDispatch(),
			url:             (*rgvargPtr)[5].MustPVariant(),
			flags:           (*rgvargPtr)[4].MustPVariant(),
			targetFrameName: (*rgvargPtr)[3].MustPVariant(),
			postData:        (*rgvargPtr)[2].MustPVariant(),
			headers:         (*rgvargPtr)[1].MustPVariant(),
			cancel:          (*rgvargPtr)[0].MustPBool(),
		}
		wv.navigatingPublisher.Publish(eventData)

	case oleaut32.DISPID_NAVIGATECOMPLETE2:
		rgvargPtr := (*[2]oleaut32.VARIANTARG)(unsafe.Pointer(pDispParams.Rgvarg))
		url := (*rgvargPtr)[0].MustPVariant()
		urlStr := ""
		if url != nil && url.MustBSTR() != nil {
			urlStr = oleaut32.BSTRToString(url.MustBSTR())
		}
		wv.navigatedPublisher.Publish(urlStr)

		wv.urlChangedPublisher.Publish()

	case oleaut32.DISPID_DOWNLOADBEGIN:
		wv.downloadingPublisher.Publish()

	case oleaut32.DISPID_DOWNLOADCOMPLETE:
		wv.downloadedPublisher.Publish()

	case oleaut32.DISPID_DOCUMENTCOMPLETE:
		rgvargPtr := (*[2]oleaut32.VARIANTARG)(unsafe.Pointer(pDispParams.Rgvarg))
		url := (*rgvargPtr)[0].MustPVariant()
		urlStr := ""
		if url != nil && url.MustBSTR() != nil {
			urlStr = oleaut32.BSTRToString(url.MustBSTR())
		}

		// FIXME: Horrible hack to avoid glitch where the document is not displayed.
		time.AfterFunc(time.Millisecond*100, func() {
			wv.Synchronize(func() {
				b := wv.BoundsPixels()
				b.Width++
				wv.SetBoundsPixels(b)
				b.Width--
				wv.SetBoundsPixels(b)
			})
		})

		wv.documentCompletedPublisher.Publish(urlStr)

	case oleaut32.DISPID_NAVIGATEERROR:
		rgvargPtr := (*[5]oleaut32.VARIANTARG)(unsafe.Pointer(pDispParams.Rgvarg))
		eventData := &WebViewNavigatedErrorEventData{
			pDisp:           (*rgvargPtr)[4].MustPDispatch(),
			url:             (*rgvargPtr)[3].MustPVariant(),
			targetFrameName: (*rgvargPtr)[2].MustPVariant(),
			statusCode:      (*rgvargPtr)[1].MustPVariant(),
			cancel:          (*rgvargPtr)[0].MustPBool(),
		}
		wv.navigatedErrorPublisher.Publish(eventData)

	case oleaut32.DISPID_NEWWINDOW3:
		rgvargPtr := (*[5]oleaut32.VARIANTARG)(unsafe.Pointer(pDispParams.Rgvarg))
		eventData := &WebViewNewWindowEventData{
			ppDisp:         (*rgvargPtr)[4].MustPPDispatch(),
			cancel:         (*rgvargPtr)[3].MustPBool(),
			dwFlags:        (*rgvargPtr)[2].MustULong(),
			bstrUrlContext: (*rgvargPtr)[1].MustBSTR(),
			bstrUrl:        (*rgvargPtr)[0].MustBSTR(),
		}
		wv.newWindowPublisher.Publish(eventData)

	case oleaut32.DISPID_ONQUIT:
		wv.quittingPublisher.Publish()

	case oleaut32.DISPID_WINDOWCLOSING:
		rgvargPtr := (*[2]oleaut32.VARIANTARG)(unsafe.Pointer(pDispParams.Rgvarg))
		eventData := &WebViewWindowClosingEventData{
			bIsChildWindow: (*rgvargPtr)[1].MustBool(),
			cancel:         (*rgvargPtr)[0].MustPBool(),
		}
		wv.windowClosingPublisher.Publish(eventData)

	case oleaut32.DISPID_ONSTATUSBAR:
		rgvargPtr := (*[1]oleaut32.VARIANTARG)(unsafe.Pointer(pDispParams.Rgvarg))
		statusBar := (*rgvargPtr)[0].MustBool()
		if statusBar != oleaut32.VARIANT_FALSE {
			wv.statusBarVisible = true
		} else {
			wv.statusBarVisible = false
		}
		wv.statusBarVisibleChangedPublisher.Publish()

	case oleaut32.DISPID_ONTHEATERMODE:
		rgvargPtr := (*[1]oleaut32.VARIANTARG)(unsafe.Pointer(pDispParams.Rgvarg))
		theaterMode := (*rgvargPtr)[0].MustBool()
		if theaterMode != oleaut32.VARIANT_FALSE {
			wv.isTheaterMode = true
		} else {
			wv.isTheaterMode = false
		}
		wv.theaterModeChangedPublisher.Publish()

	case oleaut32.DISPID_ONTOOLBAR:
		rgvargPtr := (*[1]oleaut32.VARIANTARG)(unsafe.Pointer(pDispParams.Rgvarg))
		toolBar := (*rgvargPtr)[0].MustBool()
		if toolBar != oleaut32.VARIANT_FALSE {
			wv.toolBarVisible = true
		} else {
			wv.toolBarVisible = false
		}
		wv.toolBarVisibleChangedPublisher.Publish()

	case oleaut32.DISPID_ONVISIBLE:
		rgvargPtr := (*[1]oleaut32.VARIANTARG)(unsafe.Pointer(pDispParams.Rgvarg))
		vVisible := (*rgvargPtr)[0].MustBool()
		if vVisible != oleaut32.VARIANT_FALSE {
			wv.browserVisible = true
		} else {
			wv.browserVisible = false
		}
		wv.browserVisibleChangedPublisher.Publish()

	case oleaut32.DISPID_COMMANDSTATECHANGE:
		rgvargPtr := (*[2]oleaut32.VARIANTARG)(unsafe.Pointer(pDispParams.Rgvarg))
		command := (*rgvargPtr)[1].MustLong()
		enable := (*rgvargPtr)[0].MustBool()
		enableBool := (enable != oleaut32.VARIANT_FALSE)
		switch command {
		case oleaut32.CSC_UPDATECOMMANDS:
			wv.toolBarEnabled = enableBool
			wv.toolBarEnabledChangedPublisher.Publish()

		case oleaut32.CSC_NAVIGATEFORWARD:
			wv.canGoForward = enableBool
			wv.canGoForwardChangedPublisher.Publish()

		case oleaut32.CSC_NAVIGATEBACK:
			wv.canGoBack = enableBool
			wv.canGoBackChangedPublisher.Publish()
		}

	case oleaut32.DISPID_PROGRESSCHANGE:
		rgvargPtr := (*[2]oleaut32.VARIANTARG)(unsafe.Pointer(pDispParams.Rgvarg))
		wv.progressValue = (*rgvargPtr)[1].MustLong()
		wv.progressMax = (*rgvargPtr)[0].MustLong()
		wv.progressChangedPublisher.Publish()

	case oleaut32.DISPID_STATUSTEXTCHANGE:
		rgvargPtr := (*[1]oleaut32.VARIANTARG)(unsafe.Pointer(pDispParams.Rgvarg))
		sText := (*rgvargPtr)[0].MustBSTR()
		if sText != nil {
			wv.statusText = oleaut32.BSTRToString(sText)
		} else {
			wv.statusText = ""
		}
		wv.statusTextChangedPublisher.Publish()

	case oleaut32.DISPID_TITLECHANGE:
		rgvargPtr := (*[1]oleaut32.VARIANTARG)(unsafe.Pointer(pDispParams.Rgvarg))
		sText := (*rgvargPtr)[0].MustBSTR()
		if sText != nil {
			wv.documentTitle = oleaut32.BSTRToString(sText)
		} else {
			wv.documentTitle = ""
		}
		wv.documentTitleChangedPublisher.Publish()
	}

	return oleaut32.DISP_E_MEMBERNOTFOUND
}

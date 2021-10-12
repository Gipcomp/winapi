// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/Gipcomp/win32/gdi32"
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/ole32"
	"github.com/Gipcomp/win32/oleaut32"
	"github.com/Gipcomp/win32/shdocvw"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/win"
	"github.com/Gipcomp/winapi/errs"
)

const webViewWindowClass = `\o/ Walk_WebView_Class \o/`

func init() {
	AppendToWalkInit(func() {
		MustRegisterWindowClass(webViewWindowClass)
	})
}

type WebView struct {
	WidgetBase
	clientSite                               webViewIOleClientSite // IMPORTANT: Must remain first member after WidgetBase
	browserObject                            *ole32.IOleObject
	urlChangedPublisher                      EventPublisher
	shortcutsEnabled                         bool
	shortcutsEnabledChangedPublisher         EventPublisher
	nativeContextMenuEnabled                 bool
	nativeContextMenuEnabledChangedPublisher EventPublisher
	navigatingPublisher                      WebViewNavigatingEventPublisher
	navigatedPublisher                       StringEventPublisher
	downloadingPublisher                     EventPublisher
	downloadedPublisher                      EventPublisher
	documentCompletedPublisher               StringEventPublisher
	navigatedErrorPublisher                  WebViewNavigatedErrorEventPublisher
	newWindowPublisher                       WebViewNewWindowEventPublisher
	quittingPublisher                        EventPublisher
	windowClosingPublisher                   WebViewWindowClosingEventPublisher
	statusBarVisible                         bool
	statusBarVisibleChangedPublisher         EventPublisher
	isTheaterMode                            bool
	theaterModeChangedPublisher              EventPublisher
	toolBarVisible                           bool
	toolBarVisibleChangedPublisher           EventPublisher
	browserVisible                           bool
	browserVisibleChangedPublisher           EventPublisher
	toolBarEnabled                           bool
	toolBarEnabledChangedPublisher           EventPublisher
	canGoBack                                bool
	canGoBackChangedPublisher                EventPublisher
	canGoForward                             bool
	canGoForwardChangedPublisher             EventPublisher
	progressValue                            int32
	progressMax                              int32
	progressChangedPublisher                 EventPublisher
	statusText                               string
	statusTextChangedPublisher               EventPublisher
	documentTitle                            string
	documentTitleChangedPublisher            EventPublisher
}

func NewWebView(parent Container) (*WebView, error) {
	if hr := ole32.OleInitialize(); hr != win.S_OK && hr != win.S_FALSE {
		return nil, errs.NewError(fmt.Sprint("OleInitialize Error: ", hr))
	}

	wv := &WebView{
		clientSite: webViewIOleClientSite{
			IOleClientSite: ole32.IOleClientSite{
				LpVtbl: webViewIOleClientSiteVtbl,
			},
			inPlaceSite: webViewIOleInPlaceSite{
				IOleInPlaceSite: ole32.IOleInPlaceSite{
					LpVtbl: webViewIOleInPlaceSiteVtbl,
				},
				inPlaceFrame: webViewIOleInPlaceFrame{
					IOleInPlaceFrame: ole32.IOleInPlaceFrame{
						LpVtbl: webViewIOleInPlaceFrameVtbl,
					},
				},
			},
			docHostUIHandler: webViewIDocHostUIHandler{
				IDocHostUIHandler: shdocvw.IDocHostUIHandler{
					LpVtbl: webViewIDocHostUIHandlerVtbl,
				},
			},
			webBrowserEvents2: webViewDWebBrowserEvents2{
				DWebBrowserEvents2: shdocvw.DWebBrowserEvents2{
					LpVtbl: webViewDWebBrowserEvents2Vtbl,
				},
			},
		},
		shortcutsEnabled:         false,
		nativeContextMenuEnabled: false,
	}

	if err := InitWidget(
		wv,
		parent,
		webViewWindowClass,
		user32.WS_CLIPCHILDREN|user32.WS_VISIBLE,
		0); err != nil {
		return nil, err
	}

	wv.clientSite.inPlaceSite.inPlaceFrame.webView = wv

	succeeded := false

	defer func() {
		if !succeeded {
			wv.Dispose()
		}
	}()

	var classFactoryPtr unsafe.Pointer
	if hr := ole32.CoGetClassObject(&shdocvw.CLSID_WebBrowser, ole32.CLSCTX_INPROC_HANDLER|ole32.CLSCTX_INPROC_SERVER, nil, &ole32.IID_IClassFactory, &classFactoryPtr); win.FAILED(hr) {
		return nil, errs.ErrorFromHRESULT("CoGetClassObject", hr)
	}
	classFactory := (*ole32.IClassFactory)(classFactoryPtr)
	defer classFactory.Release()

	var browserObjectPtr unsafe.Pointer
	if hr := classFactory.CreateInstance(nil, &ole32.IID_IOleObject, &browserObjectPtr); win.FAILED(hr) {
		return nil, errs.ErrorFromHRESULT("IClassFactory.CreateInstance", hr)
	}
	browserObject := (*ole32.IOleObject)(browserObjectPtr)

	wv.browserObject = browserObject

	if hr := browserObject.SetClientSite((*ole32.IOleClientSite)(unsafe.Pointer(&wv.clientSite))); win.FAILED(hr) {
		return nil, errs.ErrorFromHRESULT("IOleObject.SetClientSite", hr)
	}
	strPtr, err := syscall.UTF16PtrFromString("Walk.WebView")
	if err != nil {
		return nil, err
	}
	if hr := browserObject.SetHostNames(strPtr, nil); win.FAILED(hr) {
		return nil, errs.ErrorFromHRESULT("IOleObject.SetHostNames", hr)
	}

	if hr := ole32.OleSetContainedObject((*ole32.IUnknown)(unsafe.Pointer(browserObject)), true); win.FAILED(hr) {
		return nil, errs.ErrorFromHRESULT("OleSetContainedObject", hr)
	}

	var rect gdi32.RECT
	user32.GetClientRect(wv.hWnd, &rect)

	if hr := browserObject.DoVerb(ole32.OLEIVERB_SHOW, nil, (*ole32.IOleClientSite)(unsafe.Pointer(&wv.clientSite)), 0, wv.hWnd, &rect); win.FAILED(hr) {
		return nil, errs.ErrorFromHRESULT("IOleObject.DoVerb", hr)
	}

	var cpcPtr unsafe.Pointer
	if hr := browserObject.QueryInterface(&ole32.IID_IConnectionPointContainer, &cpcPtr); win.FAILED(hr) {
		return nil, errs.ErrorFromHRESULT("IOleObject.QueryInterface(IID_IConnectionPointContainer)", hr)
	}
	cpc := (*ole32.IConnectionPointContainer)(cpcPtr)
	defer cpc.Release()

	var cp *ole32.IConnectionPoint
	if hr := cpc.FindConnectionPoint(&shdocvw.DIID_DWebBrowserEvents2, &cp); win.FAILED(hr) {
		return nil, errs.ErrorFromHRESULT("IConnectionPointContainer.FindConnectionPoint(DIID_DWebBrowserEvents2)", hr)
	}
	defer cp.Release()

	var cookie uint32
	if hr := cp.Advise(unsafe.Pointer(&wv.clientSite.webBrowserEvents2), &cookie); win.FAILED(hr) {
		return nil, errs.ErrorFromHRESULT("IConnectionPoint.Advise", hr)
	}

	wv.onResize()

	wv.MustRegisterProperty("URL", NewProperty(
		func() interface{} {
			url, _ := wv.URL()
			return url
		},
		func(v interface{}) error {
			return wv.SetURL(assertStringOr(v, ""))
		},
		wv.urlChangedPublisher.Event()))

	wv.MustRegisterProperty("ShortcutsEnabled", NewProperty(
		func() interface{} {
			return wv.ShortcutsEnabled()
		},
		func(v interface{}) error {
			wv.SetShortcutsEnabled(v.(bool))
			return nil
		},
		wv.shortcutsEnabledChangedPublisher.Event()))

	wv.MustRegisterProperty("NativeContextMenuEnabled", NewProperty(
		func() interface{} {
			return wv.NativeContextMenuEnabled()
		},
		func(v interface{}) error {
			wv.SetNativeContextMenuEnabled(v.(bool))
			return nil
		},
		wv.nativeContextMenuEnabledChangedPublisher.Event()))

	succeeded = true

	return wv, nil
}

func (wv *WebView) Dispose() {
	if wv.browserObject != nil {
		wv.browserObject.Close(ole32.OLECLOSE_NOSAVE)
		wv.browserObject.Release()

		wv.browserObject = nil

		ole32.OleUninitialize()
	}

	wv.WidgetBase.Dispose()
}

func (wv *WebView) URL() (url string, err error) {
	err = wv.withWebBrowser2(func(webBrowser2 *shdocvw.IWebBrowser2) error {
		var urlBstr *uint16 /*BSTR*/
		if hr := webBrowser2.Get_LocationURL(&urlBstr); win.FAILED(hr) {
			return errs.ErrorFromHRESULT("IWebBrowser2.Get_LocationURL", hr)
		}
		defer oleaut32.SysFreeString(urlBstr)

		url = oleaut32.BSTRToString(urlBstr)

		return nil
	})

	return
}

func (wv *WebView) SetURL(url string) error {
	return wv.withWebBrowser2(func(webBrowser2 *shdocvw.IWebBrowser2) error {
		urlBstr := oleaut32.StringToVariantBSTR(url)
		flags := oleaut32.IntToVariantI4(0)
		targetFrameName := oleaut32.StringToVariantBSTR("_self")

		if hr := webBrowser2.Navigate2(urlBstr, flags, targetFrameName, nil, nil); win.FAILED(hr) {
			return errs.ErrorFromHRESULT("IWebBrowser2.Navigate2", hr)
		}

		return nil
	})
}

func (wv *WebView) URLChanged() *Event {
	return wv.urlChangedPublisher.Event()
}

func (wv *WebView) ShortcutsEnabled() bool {
	return wv.shortcutsEnabled
}

func (wv *WebView) SetShortcutsEnabled(value bool) {
	wv.shortcutsEnabled = value
	wv.shortcutsEnabledChangedPublisher.Publish()
}

func (wv *WebView) ShortcutsEnabledChanged() *Event {
	return wv.shortcutsEnabledChangedPublisher.Event()
}

func (wv *WebView) NativeContextMenuEnabled() bool {
	return wv.nativeContextMenuEnabled
}

func (wv *WebView) SetNativeContextMenuEnabled(value bool) {
	wv.nativeContextMenuEnabled = value
	wv.nativeContextMenuEnabledChangedPublisher.Publish()
}

func (wv *WebView) NativeContextMenuEnabledChanged() *Event {
	return wv.nativeContextMenuEnabledChangedPublisher.Event()
}

func (wv *WebView) Navigating() *WebViewNavigatingEvent {
	return wv.navigatingPublisher.Event()
}

func (wv *WebView) Navigated() *StringEvent {
	return wv.navigatedPublisher.Event()
}

func (wv *WebView) Downloading() *Event {
	return wv.downloadingPublisher.Event()
}

func (wv *WebView) Downloaded() *Event {
	return wv.downloadedPublisher.Event()
}

func (wv *WebView) DocumentCompleted() *StringEvent {
	return wv.documentCompletedPublisher.Event()
}

func (wv *WebView) NavigatedError() *WebViewNavigatedErrorEvent {
	return wv.navigatedErrorPublisher.Event()
}

func (wv *WebView) NewWindow() *WebViewNewWindowEvent {
	return wv.newWindowPublisher.Event()
}

func (wv *WebView) Quitting() *Event {
	return wv.quittingPublisher.Event()
}

func (wv *WebView) WindowClosing() *WebViewWindowClosingEvent {
	return wv.windowClosingPublisher.Event()
}

func (wv *WebView) StatusBarVisible() bool {
	return wv.statusBarVisible
}

func (wv *WebView) StatusBarVisibleChanged() *Event {
	return wv.statusBarVisibleChangedPublisher.Event()
}

func (wv *WebView) IsTheaterMode() bool {
	return wv.isTheaterMode
}

func (wv *WebView) TheaterModeChanged() *Event {
	return wv.theaterModeChangedPublisher.Event()
}

func (wv *WebView) ToolBarVisible() bool {
	return wv.toolBarVisible
}

func (wv *WebView) ToolBarVisibleChanged() *Event {
	return wv.toolBarVisibleChangedPublisher.Event()
}

func (wv *WebView) BrowserVisible() bool {
	return wv.browserVisible
}

func (wv *WebView) BrowserVisibleChanged() *Event {
	return wv.browserVisibleChangedPublisher.Event()
}

func (wv *WebView) ToolBarEnabled() bool {
	return wv.toolBarEnabled
}

func (wv *WebView) ToolBarEnabledChanged() *Event {
	return wv.toolBarEnabledChangedPublisher.Event()
}

func (wv *WebView) CanGoBack() bool {
	return wv.canGoBack
}

func (wv *WebView) CanGoBackChanged() *Event {
	return wv.canGoBackChangedPublisher.Event()
}

func (wv *WebView) CanGoForward() bool {
	return wv.canGoForward
}

func (wv *WebView) CanGoForwardChanged() *Event {
	return wv.canGoForwardChangedPublisher.Event()
}

func (wv *WebView) ProgressValue() int32 {
	return wv.progressValue
}

func (wv *WebView) ProgressMax() int32 {
	return wv.progressMax
}

func (wv *WebView) ProgressChanged() *Event {
	return wv.progressChangedPublisher.Event()
}

func (wv *WebView) StatusText() string {
	return wv.statusText
}

func (wv *WebView) StatusTextChanged() *Event {
	return wv.statusTextChangedPublisher.Event()
}

func (wv *WebView) DocumentTitle() string {
	return wv.documentTitle
}

func (wv *WebView) DocumentTitleChanged() *Event {
	return wv.documentTitleChangedPublisher.Event()
}

func (wv *WebView) Refresh() error {
	return wv.withWebBrowser2(func(webBrowser2 *shdocvw.IWebBrowser2) error {
		if hr := webBrowser2.Refresh(); win.FAILED(hr) {
			return errs.ErrorFromHRESULT("IWebBrowser2.Refresh", hr)
		}

		return nil
	})
}

func (wv *WebView) withWebBrowser2(f func(webBrowser2 *shdocvw.IWebBrowser2) error) error {
	var webBrowser2Ptr unsafe.Pointer
	if hr := wv.browserObject.QueryInterface(&shdocvw.IID_IWebBrowser2, &webBrowser2Ptr); win.FAILED(hr) {
		return errs.ErrorFromHRESULT("IOleObject.QueryInterface", hr)
	}
	webBrowser2 := (*shdocvw.IWebBrowser2)(webBrowser2Ptr)
	defer webBrowser2.Release()

	return f(webBrowser2)
}

func (wv *WebView) onResize() {
	// FIXME: handle error?
	wv.withWebBrowser2(func(webBrowser2 *shdocvw.IWebBrowser2) error {
		bounds := wv.ClientBoundsPixels()

		webBrowser2.Put_Left(0)
		webBrowser2.Put_Top(0)
		webBrowser2.Put_Width(int32(bounds.Width))
		webBrowser2.Put_Height(int32(bounds.Height))

		return nil
	})
}

func (wv *WebView) withInPlaceActiveObject(f func(activeObject *shdocvw.IOleInPlaceActiveObject) error) error {
	if wv.browserObject == nil {
		return nil
	}
	wv.withWebBrowser2(func(webBrowser2 *shdocvw.IWebBrowser2) error {
		var activeObjectPtr unsafe.Pointer
		if hr := webBrowser2.QueryInterface(&shdocvw.IID_IOleInPlaceActiveObject, &activeObjectPtr); win.FAILED(hr) {
			return errs.ErrorFromHRESULT("WebBowser2.QueryInterface", hr)
		}
		activeObject := (*shdocvw.IOleInPlaceActiveObject)(activeObjectPtr)
		defer activeObject.Release()
		return f(activeObject)
	})
	return nil
}

func (wv *WebView) translateAccelerator(msg *user32.MSG) bool {
	if wv.shortcutsEnabled {
		hr := wv.inPlaceActiveObjectTranslateAccelerator(msg)
		return hr == win.S_OK
	}
	return false
}

func (wv *WebView) inPlaceActiveObjectTranslateAccelerator(msg *user32.MSG) win.HRESULT {
	var ret win.HRESULT
	ret = win.S_FALSE
	wv.withInPlaceActiveObject(func(activeObject *shdocvw.IOleInPlaceActiveObject) error {
		hr := activeObject.TranslateAccelerator(msg)
		if hr == win.S_OK {
			ret = win.S_OK
		}
		return nil
	})
	return ret
}

func (wv *WebView) inPlaceActiveObjectSetFocus() win.HRESULT {
	var ret win.HRESULT
	ret = win.S_FALSE
	wv.withInPlaceActiveObject(func(activeObject *shdocvw.IOleInPlaceActiveObject) error {
		var hWndActive handle.HWND
		hr := activeObject.GetWindow(&hWndActive)
		if hr != win.S_OK {
			return nil
		}
		user32.SetFocus(hWndActive)
		ret = win.S_OK

		return nil
	})
	return ret
}

func (wv *WebView) WndProc(hwnd handle.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case user32.WM_WINDOWPOSCHANGED:
		wp := (*user32.WINDOWPOS)(unsafe.Pointer(lParam))

		if wp.Flags&user32.SWP_NOSIZE != 0 {
			break
		}

		if wv.clientSite.inPlaceSite.inPlaceFrame.webView == nil {
			break
		}

		wv.onResize()

	case user32.WM_MOUSEACTIVATE:
		wv.invalidateBorderInParent()
	}

	return wv.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}

func (wv *WebView) CreateLayoutItem(ctx *LayoutContext) LayoutItem {
	return NewGreedyLayoutItem()
}

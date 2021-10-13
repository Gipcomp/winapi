package winapi

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync/atomic"
	"syscall"
	"unsafe"

	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/winapi/webviewloader"
	"github.com/jchv/go-winloader"
	"golang.org/x/sys/windows"
)

const webView2WindowClass = `\o/ Walk_WebView2_Class \o/`

var errOK = syscall.Errno(0)

func init() {
	runtime.LockOSThread()

	AppendToWalkInit(func() {
		MustRegisterWindowClass(webView2WindowClass)
	})
	//_, _, err := ole32.CoInitializeExCall.Call(0, 2)
	// p := 0
	// _ = ole32.CoInitializeEx(unsafe.Pointer(&p), 2)
	// if err != nil && !errors.Is(err, errOK) {
	// 	log.Printf("warning: CoInitializeEx call failed: %v", err)
	// }
}

type WebView2 struct {
	WidgetBase
	browser *browser
	dll     winloader.Proc
}

func NewWebView2(parent Container) (*WebView2, error) {
	wv := &WebView2{
		browser: &browser{
			config: &browserConfig{
				initialURL:           "about:blank",
				builtInErrorPage:     true,
				defaultContextMenus:  true,
				defaultScriptDialogs: true,
				devtools:             true,
				hostObjects:          true,
				script:               true,
				statusBar:            true,
				webMessage:           true,
				zoomControl:          true,
			},
		},
	}

	for _, s := range []string{"WEBVIEW2_BROWSER_EXECUTABLE_FOLDER", "WEBVIEW2_USER_DATA_FOLDER", "WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS", "WEBVIEW2_RELEASE_CHANNEL_PREFERENCE"} {
		os.Unsetenv(s)
	}

	dll, err := webviewloader.New()
	if err != nil {
		return nil, err
	}

	wv.dll = dll.Proc("CreateCoreWebView2EnvironmentWithOptions")

	if err := InitWidget(
		wv,
		parent,
		webViewWindowClass,
		user32.WS_CLIPCHILDREN|user32.WS_VISIBLE,
		0); err != nil {
		return nil, err
	}

	if err := wv.browser.Navigate(wv.browser.config.initialURL); err != nil {
		return nil, fmt.Errorf("failed at the initial navigation: %w", err)
	}

	return wv, nil
}

func (wv *WebView2) CreateLayoutItem(ctx *LayoutContext) LayoutItem {
	return NewGreedyLayoutItem()
}

func (b *browser) saveSetting(setter uintptr, enabled bool) error {
	var flag uintptr = 0

	if enabled {
		flag = 1
	}

	_, _, err := syscall.Syscall(
		setter, 3,
		uintptr(unsafe.Pointer(b.settings)),
		flag,
		0,
	)

	if !errors.Is(err, errOK) {
		return fmt.Errorf("failed to save a setting: %w", err)
	}

	return nil
}

func (b *browser) saveSettings() error {
	if err := b.saveSetting(b.settings.VTBL.PutIsBuiltInErrorPageEnabled, b.config.builtInErrorPage); err != nil {
		return err
	}

	if err := b.saveSetting(b.settings.VTBL.PutAreDefaultContextMenusEnabled, b.config.defaultContextMenus); err != nil {
		return err
	}

	if err := b.saveSetting(b.settings.VTBL.PutAreDefaultScriptDialogsEnabled, b.config.defaultScriptDialogs); err != nil {
		return err
	}

	if err := b.saveSetting(b.settings.VTBL.PutAreDevToolsEnabled, b.config.devtools); err != nil {
		return err

	}

	if err := b.saveSetting(b.settings.VTBL.PutAreHostObjectsAllowed, b.config.hostObjects); err != nil {
		return err
	}

	if err := b.saveSetting(b.settings.VTBL.PutIsScriptEnabled, b.config.script); err != nil {
		return err
	}

	if err := b.saveSetting(b.settings.VTBL.PutIsStatusBarEnabled, b.config.statusBar); err != nil {
		return err

	}

	if err := b.saveSetting(b.settings.VTBL.PutIsWebMessageEnabled, b.config.webMessage); err != nil {
		return err
	}

	return b.saveSetting(b.settings.VTBL.PutIsZoomControlEnabled, b.config.zoomControl)
}

func (wv *WebView2) environmentCompletedHandler() uintptr {
	h := &ICoreWebView2CreateCoreWebView2EnvironmentCompletedHandler{
		VTBL: &ICoreWebView2CreateCoreWebView2EnvironmentCompletedHandlerVTBL{
			Invoke: windows.NewCallback(func(i uintptr, p uintptr, createdEnvironment *ICoreWebView2Environment) uintptr {
				_, _, _ = syscall.Syscall(createdEnvironment.VTBL.CreateCoreWebView2Controller, 3, uintptr(unsafe.Pointer(createdEnvironment)), uintptr(wv.Handle()), wv.controllerCompletedHandler())
				return 0
			}),
		},
	}

	h.VTBL.BasicVTBL = NewBasicVTBL(&h.Basic)
	return uintptr(unsafe.Pointer(h))
}

func (wv *WebView2) controllerCompletedHandler() uintptr {
	h := &ICoreWebView2CreateCoreWebView2ControllerCompletedHandler{
		VTBL: &ICoreWebView2CreateCoreWebView2ControllerCompletedHandlerVTBL{
			Invoke: windows.NewCallback(func(i *ICoreWebView2CreateCoreWebView2ControllerCompletedHandler, p uintptr, createdController *ICoreWebView2Controller) uintptr {
				_, _, _ = syscall.Syscall(createdController.VTBL.AddRef, 1, uintptr(unsafe.Pointer(createdController)), 0, 0)
				wv.browser.controller = createdController

				createdWebView2 := new(ICoreWebView2)

				_, _, _ = syscall.Syscall(createdController.VTBL.GetCoreWebView2, 2, uintptr(unsafe.Pointer(createdController)), uintptr(unsafe.Pointer(&createdWebView2)), 0)
				wv.browser.view = createdWebView2

				_, _, _ = syscall.Syscall(wv.browser.view.VTBL.AddRef, 1, uintptr(unsafe.Pointer(wv.browser.view)), 0, 0)

				atomic.StoreInt32(&wv.browser.controllerCompleted, 1)

				return 0
			}),
		},
	}

	h.VTBL.BasicVTBL = NewBasicVTBL(&h.Basic)
	return uintptr(unsafe.Pointer(h))
}

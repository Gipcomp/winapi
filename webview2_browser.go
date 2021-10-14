package winapi

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"unsafe"

	"github.com/Gipcomp/win32/gdi32"
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/win"
	"golang.org/x/sys/windows"
)

type browserConfig struct {
	initialURL string

	builtInErrorPage     bool
	defaultContextMenus  bool
	defaultScriptDialogs bool
	devtools             bool
	hostObjects          bool
	script               bool
	statusBar            bool
	webMessage           bool
	zoomControl          bool
}

type browser struct {
	hwnd       handle.HWND
	config     *browserConfig
	view       *ICoreWebView2
	controller *ICoreWebView2Controller
	settings   *ICoreWebView2Settings

	controllerCompleted int32
}

func (wv *WebView2) Browser() *browser {
	return wv.browser
}

func (b *browser) embed(wv *WebView2) error {
	b.hwnd = wv.Handle()

	exePath := make([]uint16, windows.MAX_PATH)

	_, err := windows.GetModuleFileName(windows.Handle(0), &exePath[0], windows.MAX_PATH)
	if err != nil {
		return fmt.Errorf("failed to get module file name: %w", err)
	}

	dataPath := filepath.Join(os.Getenv("AppData"), filepath.Base(windows.UTF16ToString(exePath)))

	r1, _, err := wv.dll.Call(0, uint64(uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(dataPath)))), 0, uint64(wv.environmentCompletedHandler()))
	hr := win.HRESULT(r1)

	if err != nil && err != errOK {
		return fmt.Errorf("failed to call CreateCoreWebView2EnvironmentWithOptions: %w", err)
	}

	if hr > win.S_OK {
		return fmt.Errorf("failed to call CreateCoreWebView2EnvironmentWithOptions: %v", hr)
	}

	for {
		if atomic.LoadInt32(&b.controllerCompleted) != 0 {
			break
		}

		var msg *user32.MSG
		if user32.GetMessage(msg, b.hwnd, 0, 0) == 0 {
			return errors.New("could not retreive msg")
		}

		if msg == nil {
			break
		}

		if !user32.TranslateMessage(msg) {
			return errors.New("could not translate msg")
		}

		user32.DispatchMessage(msg)
	}

	settings := new(ICoreWebView2Settings)

	r, _, err := syscall.Syscall(b.view.VTBL.GetSettings, 2, uintptr(unsafe.Pointer(b.view)), uintptr(unsafe.Pointer(&settings)), 0)
	if !errors.Is(err, errOK) {
		return err
	}

	hr = win.HRESULT(r)
	if hr > win.S_OK {
		return fmt.Errorf("failed to get webview settings: %v", hr)
	}

	b.settings = settings

	return nil
}

func (b *browser) resize() error {
	if b.controller == nil {
		return errors.New("nil controller")
	}
	var rect *gdi32.RECT
	if !user32.GetClientRect(b.hwnd, rect) {
		return errors.New("failed to get client rect")
	}

	_, _, err := syscall.Syscall(
		b.controller.VTBL.PutBounds, 2,
		uintptr(unsafe.Pointer(b.controller)),
		uintptr(unsafe.Pointer(rect)),
		0,
	)

	if !errors.Is(err, errOK) {
		return fmt.Errorf("failed to put rect: %w", err)
	}

	return nil
}

func (b *browser) Navigate(url string) error {
	_, _, err := syscall.Syscall(
		b.view.VTBL.Navigate, 3,
		uintptr(unsafe.Pointer(b.view)),
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(url))),
		0,
	)

	if !errors.Is(err, errOK) {
		return err
	}

	return nil
}

func (b *browser) AddScriptToExecuteOnDocumentCreated(script string) error {
	_, _, err := syscall.Syscall(
		b.view.VTBL.AddScriptToExecuteOnDocumentCreated, 3,
		uintptr(unsafe.Pointer(b.view)),
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(script))),
		0,
	)

	if !errors.Is(err, errOK) {
		return err
	}

	return nil
}

func (b *browser) ExecuteScript(script string) error {
	_, _, err := syscall.Syscall(
		b.view.VTBL.ExecuteScript, 3,
		uintptr(unsafe.Pointer(b.view)),
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(script))),
		0,
	)

	if !errors.Is(err, errOK) {
		return err
	}

	return nil
}

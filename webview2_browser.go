package winapi

import (
	"errors"
	"syscall"
	"unsafe"

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
	config     *browserConfig
	view       *ICoreWebView2
	controller *ICoreWebView2Controller
	settings   *ICoreWebView2Settings

	controllerCompleted int32
}

func (wv *WebView2) Browser() *browser {
	return wv.browser
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

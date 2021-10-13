package winapi

import (
	"sync"

	"golang.org/x/sys/windows"
)

type webviewContextStore struct {
	mu    sync.RWMutex
	store map[windows.Handle]*WebView2
}

var webviewContext = &webviewContextStore{
	store: map[windows.Handle]*WebView2{},
}

func (wcs *webviewContextStore) set(hwnd windows.Handle, wv *WebView2) {
	wcs.mu.Lock()
	defer wcs.mu.Unlock()

	wcs.store[hwnd] = wv
}

func (wcs *webviewContextStore) get(hwnd windows.Handle) (*WebView2, bool) {
	wcs.mu.Lock()
	defer wcs.mu.Unlock()

	wv, ok := wcs.store[hwnd]
	return wv, ok
}

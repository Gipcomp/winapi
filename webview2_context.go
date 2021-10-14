package winapi

import (
	"sync"

	"github.com/Gipcomp/win32/handle"
)

type webviewContextStore struct {
	mu    sync.RWMutex
	store map[handle.HWND]*WebView2
}

var webviewContext = &webviewContextStore{
	store: map[handle.HWND]*WebView2{},
}

func (wcs *webviewContextStore) set(hwnd handle.HWND, wv *WebView2) {
	wcs.mu.Lock()
	defer wcs.mu.Unlock()

	wcs.store[hwnd] = wv
}

func (wcs *webviewContextStore) get(hwnd handle.HWND) (*WebView2, bool) {
	wcs.mu.Lock()
	defer wcs.mu.Unlock()

	wv, ok := wcs.store[hwnd]
	return wv, ok
}

package winapi

import (
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"strconv"
	"sync"

	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/kernel32"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/winapi/edge"
)

const webView2WindowClass = `\o/ Walk_WebView2_Class \o/`

func init() {
	AppendToWalkInit(func() {
		MustRegisterWindowClass(webView2WindowClass)
	})
}

type browser interface {
	Embed(hwnd handle.HWND) bool
	Resize()
	Navigate(url string)
	Init(script string)
	Eval(script string)
	GetDefaultBackgroundColor() (*edge.COREWEBVIEW2_COLOR, error)
	PutDefaultBackgroundColor(backgroundColor edge.COREWEBVIEW2_COLOR) error
}

type WebView2 struct {
	WidgetBase
	hwnd          handle.HWND
	browserObject browser
	mainthread    uint32
	m             sync.Mutex
	bindings      map[string]interface{}
	dispatchq     []func()
}

type rpcMessage struct {
	ID     int               `json:"id"`
	Method string            `json:"method"`
	Params []json.RawMessage `json:"params"`
}

// var (
// 	windowContext     = map[handle.HWND]interface{}{}
// 	windowContextSync sync.RWMutex
// )

func NewWebView2(debug bool, parent Container) (*WebView2, error) {
	wv := &WebView2{}
	if err := InitWidget(
		wv,
		parent,
		webView2WindowClass,
		user32.WS_CLIPCHILDREN|user32.WS_VISIBLE,
		0); err != nil {
		return nil, err
	}
	chromium := edge.NewChromium()
	chromium.MessageCallback = wv.msgcb
	chromium.Debug = debug
	wv.browserObject = chromium

	wv.mainthread = kernel32.GetCurrentThreadId() // w32.Kernel32GetCurrentThreadID.Call()
	if !wv.Create(debug, wv.Handle()) {
		return nil, errors.New("failed to create webview2")
	}
	return wv, nil
}

func jsString(v interface{}) string { b, _ := json.Marshal(v); return string(b) }

// func getWindowContext(wnd handle.HWND) interface{} {
// 	windowContextSync.RLock()
// 	defer windowContextSync.RUnlock()
// 	return windowContext[wnd]
// }

// func setWindowContext(wnd handle.HWND, data interface{}) {
// 	windowContextSync.Lock()
// 	defer windowContextSync.Unlock()
// 	windowContext[wnd] = data
// }

func (wv *WebView2) callbinding(d rpcMessage) (interface{}, error) {
	wv.m.Lock()
	f, ok := wv.bindings[d.Method]
	wv.m.Unlock()
	if !ok {
		return nil, nil
	}

	v := reflect.ValueOf(f)
	isVariadic := v.Type().IsVariadic()
	numIn := v.Type().NumIn()
	if (isVariadic && len(d.Params) < numIn-1) || (!isVariadic && len(d.Params) != numIn) {
		return nil, errors.New("function arguments mismatch")
	}
	args := []reflect.Value{}
	for i := range d.Params {
		var arg reflect.Value
		if isVariadic && i >= numIn-1 {
			arg = reflect.New(v.Type().In(numIn - 1).Elem())
		} else {
			arg = reflect.New(v.Type().In(i))
		}
		if err := json.Unmarshal(d.Params[i], arg.Interface()); err != nil {
			return nil, err
		}
		args = append(args, arg.Elem())
	}

	errorType := reflect.TypeOf((*error)(nil)).Elem()
	res := v.Call(args)
	switch len(res) {
	case 0:
		// No results from the function, just return nil
		return nil, nil

	case 1:
		// One result may be a value, or an error
		if res[0].Type().Implements(errorType) {
			if res[0].Interface() != nil {
				return nil, res[0].Interface().(error)
			}
			return nil, nil
		}
		return res[0].Interface(), nil

	case 2:
		// Two results: first one is value, second is error
		if !res[1].Type().Implements(errorType) {
			return nil, errors.New("second return value must be an error")
		}
		if res[1].Interface() == nil {
			return res[0].Interface(), nil
		}
		return res[0].Interface(), res[1].Interface().(error)

	default:
		return nil, errors.New("unexpected number of return values")
	}
}

func (wv *WebView2) msgcb(msg string) {
	d := rpcMessage{}
	if err := json.Unmarshal([]byte(msg), &d); err != nil {
		log.Printf("invalid RPC message: %v", err)
		return
	}

	id := strconv.Itoa(d.ID)
	if res, err := wv.callbinding(d); err != nil {
		wv.Dispatch(func() {
			wv.Eval("window._rpc[" + id + "].reject(" + jsString(err.Error()) + "); window._rpc[" + id + "] = undefined")
		})
	} else if b, err := json.Marshal(res); err != nil {
		wv.Dispatch(func() {
			wv.Eval("window._rpc[" + id + "].reject(" + jsString(err.Error()) + "); window._rpc[" + id + "] = undefined")
		})
	} else {
		wv.Dispatch(func() {
			wv.Eval("window._rpc[" + id + "].resolve(" + string(b) + "); window._rpc[" + id + "] = undefined")
		})
	}
}

func (wv *WebView2) Create(debug bool, hwnd handle.HWND) bool {
	wv.hwnd = hwnd

	if !wv.browserObject.Embed(wv.hwnd) {
		return false
	}
	wv.browserObject.Resize()
	return true
}

func (wv *WebView2) Eval(js string) {
	wv.browserObject.Eval(js)
}

func (wv *WebView2) Dispatch(f func()) {
	wv.m.Lock()
	wv.dispatchq = append(wv.dispatchq, f)
	wv.m.Unlock()
	kernel32.PostThreadMessageW.Call(uintptr(wv.mainthread), user32.WM_APP, 0, 0)
}

func (w *WebView2) SetURL(url string) {
	w.browserObject.Navigate(url)
}

func (wv *WebView2) CreateLayoutItem(ctx *LayoutContext) LayoutItem {
	return NewGreedyLayoutItem()
}

func (w *WebView2) GetDefaultBackgroundColor() (*edge.COREWEBVIEW2_COLOR, error) {
	return w.browserObject.GetDefaultBackgroundColor()
}

func (w *WebView2) PutDefaultBackgroundColor(bgcolor edge.COREWEBVIEW2_COLOR) error {
	return w.browserObject.PutDefaultBackgroundColor(bgcolor)
}

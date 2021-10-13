package winapi

type Option func(*WebView2)

func WithURL(url string) Option {
	return func(wv *WebView2) {
		wv.browser.config.initialURL = url
	}
}

func WithBuiltinErrorPage(enabled bool) Option {
	return func(wv *WebView2) {
		wv.browser.config.builtInErrorPage = enabled
	}
}

func WithDefaultContextMenus(enabled bool) Option {
	return func(wv *WebView2) {
		wv.browser.config.defaultContextMenus = enabled
	}
}

func WithDefaultScriptDialogs(enabled bool) Option {
	return func(wv *WebView2) {
		wv.browser.config.defaultScriptDialogs = enabled
	}
}

func WithDevtools(enabled bool) Option {
	return func(wv *WebView2) {
		wv.browser.config.devtools = enabled
	}
}

func WithHostObjects(enabled bool) Option {
	return func(wv *WebView2) {
		wv.browser.config.hostObjects = enabled
	}
}

func WithStatusBar(enabled bool) Option {
	return func(wv *WebView2) {
		wv.browser.config.statusBar = enabled
	}
}

func WithScript(enabled bool) Option {
	return func(wv *WebView2) {
		wv.browser.config.script = enabled
	}
}

func WithWebMessage(enabled bool) Option {
	return func(wv *WebView2) {
		wv.browser.config.webMessage = enabled
	}
}

func WithZoomControl(enabled bool) Option {
	return func(wv *WebView2) {
		wv.browser.config.zoomControl = enabled
	}
}

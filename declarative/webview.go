// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import "github.com/Gipcomp/winapi"

type WebView struct {
	// Window

	Accessibility      Accessibility
	Background         Brush
	ContextMenuItems   []MenuItem
	DoubleBuffering    bool
	Enabled            Property
	Font               Font
	MaxSize            Size
	MinSize            Size
	Name               string
	OnBoundsChanged    winapi.EventHandler
	OnKeyDown          winapi.KeyEventHandler
	OnKeyPress         winapi.KeyEventHandler
	OnKeyUp            winapi.KeyEventHandler
	OnMouseDown        winapi.MouseEventHandler
	OnMouseMove        winapi.MouseEventHandler
	OnMouseUp          winapi.MouseEventHandler
	OnSizeChanged      winapi.EventHandler
	Persistent         bool
	RightToLeftReading bool
	ToolTipText        Property
	Visible            Property

	// Widget

	Alignment          Alignment2D
	AlwaysConsumeSpace bool
	Column             int
	ColumnSpan         int
	GraphicsEffects    []winapi.WidgetGraphicsEffect
	Row                int
	RowSpan            int
	StretchFactor      int

	// WebView

	AssignTo                          **winapi.WebView
	NativeContextMenuEnabled          Property
	OnBrowserVisibleChanged           winapi.EventHandler
	OnCanGoBackChanged                winapi.EventHandler
	OnCanGoForwardChanged             winapi.EventHandler
	OnDocumentCompleted               winapi.StringEventHandler
	OnDocumentTitleChanged            winapi.EventHandler
	OnDownloaded                      winapi.EventHandler
	OnDownloading                     winapi.EventHandler
	OnNativeContextMenuEnabledChanged winapi.EventHandler
	OnNavigated                       winapi.StringEventHandler
	OnNavigatedError                  winapi.WebViewNavigatedErrorEventHandler
	OnNavigating                      winapi.WebViewNavigatingEventHandler
	OnNewWindow                       winapi.WebViewNewWindowEventHandler
	OnProgressChanged                 winapi.EventHandler
	OnQuitting                        winapi.EventHandler
	OnShortcutsEnabledChanged         winapi.EventHandler
	OnStatusBarVisibleChanged         winapi.EventHandler
	OnStatusTextChanged               winapi.EventHandler
	OnTheaterModeChanged              winapi.EventHandler
	OnToolBarEnabledChanged           winapi.EventHandler
	OnToolBarVisibleChanged           winapi.EventHandler
	OnURLChanged                      winapi.EventHandler
	OnWindowClosing                   winapi.WebViewWindowClosingEventHandler
	ShortcutsEnabled                  Property
	URL                               Property
}

func (wv WebView) Create(builder *Builder) error {
	w, err := winapi.NewWebView(builder.Parent())
	if err != nil {
		return err
	}

	if wv.AssignTo != nil {
		*wv.AssignTo = w
	}

	return builder.InitWidget(wv, w, func() error {
		if wv.OnBrowserVisibleChanged != nil {
			w.BrowserVisibleChanged().Attach(wv.OnBrowserVisibleChanged)
		}
		if wv.OnCanGoBackChanged != nil {
			w.CanGoBackChanged().Attach(wv.OnCanGoBackChanged)
		}
		if wv.OnCanGoForwardChanged != nil {
			w.CanGoForwardChanged().Attach(wv.OnCanGoForwardChanged)
		}
		if wv.OnDocumentCompleted != nil {
			w.DocumentCompleted().Attach(wv.OnDocumentCompleted)
		}
		if wv.OnDocumentTitleChanged != nil {
			w.DocumentTitleChanged().Attach(wv.OnDocumentTitleChanged)
		}
		if wv.OnDownloaded != nil {
			w.Downloaded().Attach(wv.OnDownloaded)
		}
		if wv.OnDownloading != nil {
			w.Downloading().Attach(wv.OnDownloading)
		}
		if wv.OnNativeContextMenuEnabledChanged != nil {
			w.NativeContextMenuEnabledChanged().Attach(wv.OnNativeContextMenuEnabledChanged)
		}
		if wv.OnNavigated != nil {
			w.Navigated().Attach(wv.OnNavigated)
		}
		if wv.OnNavigatedError != nil {
			w.NavigatedError().Attach(wv.OnNavigatedError)
		}
		if wv.OnNavigating != nil {
			w.Navigating().Attach(wv.OnNavigating)
		}
		if wv.OnNewWindow != nil {
			w.NewWindow().Attach(wv.OnNewWindow)
		}
		if wv.OnProgressChanged != nil {
			w.ProgressChanged().Attach(wv.OnProgressChanged)
		}
		if wv.OnURLChanged != nil {
			w.URLChanged().Attach(wv.OnURLChanged)
		}
		if wv.OnShortcutsEnabledChanged != nil {
			w.ShortcutsEnabledChanged().Attach(wv.OnShortcutsEnabledChanged)
		}
		if wv.OnStatusBarVisibleChanged != nil {
			w.StatusBarVisibleChanged().Attach(wv.OnStatusBarVisibleChanged)
		}
		if wv.OnStatusTextChanged != nil {
			w.StatusTextChanged().Attach(wv.OnStatusTextChanged)
		}
		if wv.OnTheaterModeChanged != nil {
			w.TheaterModeChanged().Attach(wv.OnTheaterModeChanged)
		}
		if wv.OnToolBarEnabledChanged != nil {
			w.ToolBarEnabledChanged().Attach(wv.OnToolBarEnabledChanged)
		}
		if wv.OnToolBarVisibleChanged != nil {
			w.ToolBarVisibleChanged().Attach(wv.OnToolBarVisibleChanged)
		}
		if wv.OnQuitting != nil {
			w.Quitting().Attach(wv.OnQuitting)
		}
		if wv.OnWindowClosing != nil {
			w.WindowClosing().Attach(wv.OnWindowClosing)
		}

		return nil
	})
}

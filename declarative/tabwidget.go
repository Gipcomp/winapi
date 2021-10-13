// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import "github.com/Gipcomp/winapi"

type TabWidget struct {
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

	// TabWidget

	AssignTo              **winapi.TabWidget
	ContentMargins        Margins
	ContentMarginsZero    bool
	OnCurrentIndexChanged winapi.EventHandler
	Pages                 []TabPage
}

func (tw TabWidget) Create(builder *Builder) error {
	w, err := winapi.NewTabWidget(builder.Parent())
	if err != nil {
		return err
	}

	if tw.AssignTo != nil {
		*tw.AssignTo = w
	}

	return builder.InitWidget(tw, w, func() error {
		for _, tp := range tw.Pages {
			var wp *winapi.TabPage
			if tp.AssignTo == nil {
				tp.AssignTo = &wp
			}

			if tp.Content != nil && len(tp.Children) == 0 {
				tp.Layout = HBox{Margins: tw.ContentMargins, MarginsZero: tw.ContentMarginsZero}
			}

			if err := tp.Create(builder); err != nil {
				return err
			}

			if err := w.Pages().Add(*tp.AssignTo); err != nil {
				return err
			}
		}

		if tw.OnCurrentIndexChanged != nil {
			w.CurrentIndexChanged().Attach(tw.OnCurrentIndexChanged)
		}

		return nil
	})
}

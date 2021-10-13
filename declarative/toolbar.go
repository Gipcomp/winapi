// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import "github.com/Gipcomp/winapi"

type ToolBarButtonStyle int

const (
	ToolBarButtonImageOnly ToolBarButtonStyle = iota
	ToolBarButtonTextOnly
	ToolBarButtonImageBeforeText
	ToolBarButtonImageAboveText
)

type ToolBar struct {
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

	// ToolBar

	Actions     []*winapi.Action // Deprecated, use Items instead
	AssignTo    **winapi.ToolBar
	ButtonStyle ToolBarButtonStyle
	Items       []MenuItem
	MaxTextRows int
	Orientation Orientation
}

func (tb ToolBar) Create(builder *Builder) error {
	w, err := winapi.NewToolBarWithOrientationAndButtonStyle(builder.Parent(), winapi.Orientation(tb.Orientation), winapi.ToolBarButtonStyle(tb.ButtonStyle))
	if err != nil {
		return err
	}

	if tb.AssignTo != nil {
		*tb.AssignTo = w
	}

	return builder.InitWidget(tb, w, func() error {
		imageList, err := winapi.NewImageList(winapi.Size{16, 16}, 0)
		if err != nil {
			return err
		}
		w.SetImageList(imageList)

		mtr := tb.MaxTextRows
		if mtr < 1 {
			mtr = 1
		}
		if err := w.SetMaxTextRows(mtr); err != nil {
			return err
		}

		if len(tb.Items) > 0 {
			builder.deferBuildActions(w.Actions(), tb.Items)
		} else {
			if err := addToActionList(w.Actions(), tb.Actions); err != nil {
				return err
			}
		}

		return nil
	})
}

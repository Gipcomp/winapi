// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/winapi"
)

type TextEdit struct {
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

	// TextEdit

	AssignTo      **winapi.TextEdit
	CompactHeight bool
	HScroll       bool
	MaxLength     int
	OnTextChanged winapi.EventHandler
	ReadOnly      Property
	Text          Property
	TextAlignment Alignment1D
	TextColor     winapi.Color
	VScroll       bool
}

func (te TextEdit) Create(builder *Builder) error {
	var style uint32
	if te.HScroll {
		style |= user32.WS_HSCROLL
	}
	if te.VScroll {
		style |= user32.WS_VSCROLL
	}

	w, err := winapi.NewTextEditWithStyle(builder.Parent(), style)
	if err != nil {
		return err
	}

	if te.AssignTo != nil {
		*te.AssignTo = w
	}

	return builder.InitWidget(te, w, func() error {
		w.SetCompactHeight(te.CompactHeight)
		w.SetTextColor(te.TextColor)

		if err := w.SetTextAlignment(winapi.Alignment1D(te.TextAlignment)); err != nil {
			return err
		}

		if te.MaxLength > 0 {
			w.SetMaxLength(te.MaxLength)
		}

		if te.OnTextChanged != nil {
			w.TextChanged().Attach(te.OnTextChanged)
		}

		return nil
	})
}

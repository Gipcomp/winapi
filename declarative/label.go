// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/winapi"
)

type EllipsisMode int

const (
	EllipsisNone = EllipsisMode(winapi.EllipsisNone)
	EllipsisEnd  = EllipsisMode(winapi.EllipsisEnd)
	EllipsisPath = EllipsisMode(winapi.EllipsisPath)
)

type Label struct {
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

	// Label

	AssignTo      **winapi.Label
	EllipsisMode  EllipsisMode
	NoPrefix      bool
	Text          Property
	TextAlignment Alignment1D
	TextColor     winapi.Color
}

func (l Label) Create(builder *Builder) error {
	var style uint32
	if l.NoPrefix {
		style |= user32.SS_NOPREFIX
	}

	w, err := winapi.NewLabelWithStyle(builder.Parent(), style)
	if err != nil {
		return err
	}

	if l.AssignTo != nil {
		*l.AssignTo = w
	}

	return builder.InitWidget(l, w, func() error {
		if err := w.SetEllipsisMode(winapi.EllipsisMode(l.EllipsisMode)); err != nil {
			return err
		}

		if err := w.SetTextAlignment(winapi.Alignment1D(l.TextAlignment)); err != nil {
			return err
		}

		w.SetTextColor(l.TextColor)

		return nil
	})
}

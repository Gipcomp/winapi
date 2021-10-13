// Copyright 2018 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/Gipcomp/winapi"
)

type DateLabel struct {
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

	// static

	TextColor winapi.Color

	// DateLabel

	AssignTo      **winapi.DateLabel
	Date          Property
	Format        Property
	TextAlignment Alignment1D
}

func (dl DateLabel) Create(builder *Builder) error {
	w, err := winapi.NewDateLabel(builder.Parent())
	if err != nil {
		return err
	}

	if dl.AssignTo != nil {
		*dl.AssignTo = w
	}

	return builder.InitWidget(dl, w, func() error {
		if err := w.SetTextAlignment(winapi.Alignment1D(dl.TextAlignment)); err != nil {
			return err
		}

		w.SetTextColor(dl.TextColor)

		return nil
	})
}

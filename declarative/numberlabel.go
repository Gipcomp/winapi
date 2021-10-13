// Copyright 2018 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/Gipcomp/winapi"
)

type NumberLabel struct {
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

	// NumberLabel

	AssignTo      **winapi.NumberLabel
	Decimals      Property
	Suffix        Property
	TextAlignment Alignment1D
	Value         Property
}

func (nl NumberLabel) Create(builder *Builder) error {
	w, err := winapi.NewNumberLabel(builder.Parent())
	if err != nil {
		return err
	}

	if nl.AssignTo != nil {
		*nl.AssignTo = w
	}

	return builder.InitWidget(nl, w, func() error {
		if err := w.SetTextAlignment(winapi.Alignment1D(nl.TextAlignment)); err != nil {
			return err
		}

		w.SetTextColor(nl.TextColor)

		return nil
	})
}

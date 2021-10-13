// Copyright 2018 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/winapi"
)

type Alignment2D uint

const (
	AlignHVDefault      = Alignment2D(winapi.AlignHVDefault)
	AlignHNearVNear     = Alignment2D(winapi.AlignHNearVNear)
	AlignHCenterVNear   = Alignment2D(winapi.AlignHCenterVNear)
	AlignHFarVNear      = Alignment2D(winapi.AlignHFarVNear)
	AlignHNearVCenter   = Alignment2D(winapi.AlignHNearVCenter)
	AlignHCenterVCenter = Alignment2D(winapi.AlignHCenterVCenter)
	AlignHFarVCenter    = Alignment2D(winapi.AlignHFarVCenter)
	AlignHNearVFar      = Alignment2D(winapi.AlignHNearVFar)
	AlignHCenterVFar    = Alignment2D(winapi.AlignHCenterVFar)
	AlignHFarVFar       = Alignment2D(winapi.AlignHFarVFar)
)

type TextLabel struct {
	// Window

	Accessibility      Accessibility
	Background         Brush
	ContextMenuItems   []MenuItem
	DoubleBuffering    bool
	Enabled            Property
	Font               Font
	MaxSize            Size
	MinSize            Size // Set MinSize.Width to a value > 0 to enable dynamic line wrapping.
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

	// Text

	AssignTo      **winapi.TextLabel
	NoPrefix      bool
	TextAlignment Alignment2D
	Text          Property
}

func (tl TextLabel) Create(builder *Builder) error {
	var style uint32
	if tl.NoPrefix {
		style |= user32.SS_NOPREFIX
	}

	w, err := winapi.NewTextLabelWithStyle(builder.Parent(), style)
	if err != nil {
		return err
	}

	if tl.AssignTo != nil {
		*tl.AssignTo = w
	}

	return builder.InitWidget(tl, w, func() error {
		w.SetTextColor(tl.TextColor)

		if err := w.SetTextAlignment(winapi.Alignment2D(tl.TextAlignment)); err != nil {
			return err
		}

		return nil
	})
}

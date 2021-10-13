// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/Gipcomp/winapi"
)

type ProgressBar struct {
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

	// ProgressBar

	AssignTo    **winapi.ProgressBar
	MarqueeMode bool
	MaxValue    int
	MinValue    int
	Value       int
}

func (pb ProgressBar) Create(builder *Builder) error {
	w, err := winapi.NewProgressBar(builder.Parent())
	if err != nil {
		return err
	}

	if pb.AssignTo != nil {
		*pb.AssignTo = w
	}

	return builder.InitWidget(pb, w, func() error {
		if pb.MaxValue > pb.MinValue {
			w.SetRange(pb.MinValue, pb.MaxValue)
		}
		w.SetValue(pb.Value)

		if err := w.SetMarqueeMode(pb.MarqueeMode); err != nil {
			return err
		}

		return nil
	})
}

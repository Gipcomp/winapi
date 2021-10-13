// Copyright 2016 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import "github.com/Gipcomp/winapi"

type Slider struct {
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

	// Slider

	AssignTo       **winapi.Slider
	LineSize       int
	MaxValue       int
	MinValue       int
	Orientation    Orientation
	OnValueChanged winapi.EventHandler
	PageSize       int
	ToolTipsHidden bool
	Tracking       bool
	Value          Property
}

func (sl Slider) Create(builder *Builder) error {
	w, err := winapi.NewSliderWithCfg(builder.Parent(), &winapi.SliderCfg{
		Orientation:    winapi.Orientation(sl.Orientation),
		ToolTipsHidden: sl.ToolTipsHidden,
	})
	if err != nil {
		return err
	}

	if sl.AssignTo != nil {
		*sl.AssignTo = w
	}

	return builder.InitWidget(sl, w, func() error {
		w.SetPersistent(sl.Persistent)
		if sl.LineSize > 0 {
			w.SetLineSize(sl.LineSize)
		}
		if sl.PageSize > 0 {
			w.SetPageSize(sl.PageSize)
		}
		w.SetTracking(sl.Tracking)

		if sl.MaxValue > sl.MinValue {
			w.SetRange(sl.MinValue, sl.MaxValue)
		}

		if sl.OnValueChanged != nil {
			w.ValueChanged().Attach(sl.OnValueChanged)
		}

		return nil
	})
}

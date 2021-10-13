// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/Gipcomp/winapi"
)

type NumberEdit struct {
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

	// NumberEdit

	AssignTo           **winapi.NumberEdit
	Decimals           int
	Increment          float64
	MaxValue           float64
	MinValue           float64
	Prefix             Property
	OnValueChanged     winapi.EventHandler
	ReadOnly           Property
	SpinButtonsVisible bool
	Suffix             Property
	TextColor          winapi.Color
	Value              Property
}

func (ne NumberEdit) Create(builder *Builder) error {
	w, err := winapi.NewNumberEdit(builder.Parent())
	if err != nil {
		return err
	}

	if ne.AssignTo != nil {
		*ne.AssignTo = w
	}

	return builder.InitWidget(ne, w, func() error {
		w.SetTextColor(ne.TextColor)

		if err := w.SetDecimals(ne.Decimals); err != nil {
			return err
		}

		inc := ne.Increment
		if inc == 0 {
			inc = 1
		}

		if err := w.SetIncrement(inc); err != nil {
			return err
		}

		if ne.MinValue != 0 || ne.MaxValue != 0 {
			if err := w.SetRange(ne.MinValue, ne.MaxValue); err != nil {
				return err
			}
		}

		if err := w.SetSpinButtonsVisible(ne.SpinButtonsVisible); err != nil {
			return err
		}

		if ne.OnValueChanged != nil {
			w.ValueChanged().Attach(ne.OnValueChanged)
		}

		return nil
	})
}

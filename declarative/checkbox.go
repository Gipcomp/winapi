// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/Gipcomp/winapi"
)

type CheckBox struct {
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

	// Button

	Checked          Property
	OnCheckedChanged winapi.EventHandler
	OnClicked        winapi.EventHandler
	Text             Property

	// CheckBox

	AssignTo            **winapi.CheckBox
	CheckState          Property
	OnCheckStateChanged winapi.EventHandler
	TextOnLeftSide      bool
	Tristate            bool
}

func (cb CheckBox) Create(builder *Builder) error {
	w, err := winapi.NewCheckBox(builder.Parent())
	if err != nil {
		return err
	}

	if cb.AssignTo != nil {
		*cb.AssignTo = w
	}

	return builder.InitWidget(cb, w, func() error {
		w.SetPersistent(cb.Persistent)

		if err := w.SetTextOnLeftSide(cb.TextOnLeftSide); err != nil {
			return err
		}

		if err := w.SetTristate(cb.Tristate); err != nil {
			return err
		}

		if _, isBindData := cb.CheckState.(bindData); cb.Tristate && (cb.CheckState == nil || isBindData) {
			w.SetCheckState(winapi.CheckIndeterminate)
		}

		if cb.OnClicked != nil {
			w.Clicked().Attach(cb.OnClicked)
		}

		if cb.OnCheckedChanged != nil {
			w.CheckedChanged().Attach(cb.OnCheckedChanged)
		}

		if cb.OnCheckStateChanged != nil {
			w.CheckStateChanged().Attach(cb.OnCheckStateChanged)
		}

		return nil
	})
}

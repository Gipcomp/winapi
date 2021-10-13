// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"errors"

	"github.com/Gipcomp/winapi"
)

type ComboBox struct {
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

	// ComboBox

	AssignTo              **winapi.ComboBox
	BindingMember         string
	CurrentIndex          Property
	DisplayMember         string
	Editable              bool
	Format                string
	MaxLength             int
	Model                 interface{}
	OnCurrentIndexChanged winapi.EventHandler
	OnEditingFinished     winapi.EventHandler
	OnTextChanged         winapi.EventHandler
	Precision             int
	Value                 Property
}

func (cb ComboBox) Create(builder *Builder) error {
	if _, ok := cb.Model.([]string); ok &&
		(cb.BindingMember != "" || cb.DisplayMember != "") {

		return errors.New("ComboBox.Create: BindingMember and DisplayMember must be empty for []string models.")
	}

	var w *winapi.ComboBox
	var err error
	if cb.Editable {
		w, err = winapi.NewComboBox(builder.Parent())
	} else {
		w, err = winapi.NewDropDownBox(builder.Parent())
	}
	if err != nil {
		return err
	}

	if cb.AssignTo != nil {
		*cb.AssignTo = w
	}

	return builder.InitWidget(cb, w, func() error {
		w.SetPersistent(cb.Persistent)
		w.SetFormat(cb.Format)
		w.SetPrecision(cb.Precision)
		w.SetMaxLength(cb.MaxLength)

		if err := w.SetBindingMember(cb.BindingMember); err != nil {
			return err
		}
		if err := w.SetDisplayMember(cb.DisplayMember); err != nil {
			return err
		}

		if err := w.SetModel(cb.Model); err != nil {
			return err
		}

		if cb.OnCurrentIndexChanged != nil {
			w.CurrentIndexChanged().Attach(cb.OnCurrentIndexChanged)
		}
		if cb.OnEditingFinished != nil {
			w.EditingFinished().Attach(cb.OnEditingFinished)
		}
		if cb.OnTextChanged != nil {
			w.TextChanged().Attach(cb.OnTextChanged)
		}

		return nil
	})
}

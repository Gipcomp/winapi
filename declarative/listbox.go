// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"errors"

	"github.com/Gipcomp/win32/winuser"
	"github.com/Gipcomp/winapi"
)

type ListBox struct {
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

	// ListBox

	AssignTo                 **winapi.ListBox
	BindingMember            string
	CurrentIndex             Property
	DisplayMember            string
	Format                   string
	ItemStyler               winapi.ListItemStyler
	Model                    interface{}
	MultiSelection           bool
	OnCurrentIndexChanged    winapi.EventHandler
	OnItemActivated          winapi.EventHandler
	OnSelectedIndexesChanged winapi.EventHandler
	Precision                int
	Value                    Property
}

func (lb ListBox) Create(builder *Builder) error {
	var w *winapi.ListBox
	var err error
	if _, ok := lb.Model.([]string); ok &&
		(lb.BindingMember != "" || lb.DisplayMember != "") {

		return errors.New("ListBox.Create: BindingMember and DisplayMember must be empty for []string models.")
	}

	var style uint32

	if lb.ItemStyler != nil {
		style |= winuser.LBS_OWNERDRAWVARIABLE
	}
	if lb.MultiSelection {
		style |= winuser.LBS_EXTENDEDSEL
	}

	w, err = winapi.NewListBoxWithStyle(builder.Parent(), style)
	if err != nil {
		return err
	}

	if lb.AssignTo != nil {
		*lb.AssignTo = w
	}

	return builder.InitWidget(lb, w, func() error {
		if lb.ItemStyler != nil {
			w.SetItemStyler(lb.ItemStyler)
		}
		w.SetFormat(lb.Format)
		w.SetPrecision(lb.Precision)

		if err := w.SetBindingMember(lb.BindingMember); err != nil {
			return err
		}
		if err := w.SetDisplayMember(lb.DisplayMember); err != nil {
			return err
		}

		if err := w.SetModel(lb.Model); err != nil {
			return err
		}

		if lb.OnCurrentIndexChanged != nil {
			w.CurrentIndexChanged().Attach(lb.OnCurrentIndexChanged)
		}
		if lb.OnSelectedIndexesChanged != nil {
			w.SelectedIndexesChanged().Attach(lb.OnSelectedIndexesChanged)
		}
		if lb.OnItemActivated != nil {
			w.ItemActivated().Attach(lb.OnItemActivated)
		}

		return nil
	})
}

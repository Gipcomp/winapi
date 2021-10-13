// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"time"

	"github.com/Gipcomp/winapi"
)

type DateEdit struct {
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

	// DateEdit

	AssignTo      **winapi.DateEdit
	Date          Property
	Format        string
	MaxDate       time.Time
	MinDate       time.Time
	NoneOption    bool // Deprecated: use Optional instead
	OnDateChanged winapi.EventHandler
	Optional      bool
}

func (de DateEdit) Create(builder *Builder) error {
	var w *winapi.DateEdit
	var err error

	if de.Optional || de.NoneOption {
		w, err = winapi.NewDateEditWithNoneOption(builder.Parent())
	} else {
		w, err = winapi.NewDateEdit(builder.Parent())
	}
	if err != nil {
		return err
	}

	if de.AssignTo != nil {
		*de.AssignTo = w
	}

	return builder.InitWidget(de, w, func() error {
		if err := w.SetFormat(de.Format); err != nil {
			return err
		}

		if err := w.SetRange(de.MinDate, de.MaxDate); err != nil {
			return err
		}

		if de.OnDateChanged != nil {
			w.DateChanged().Attach(de.OnDateChanged)
		}

		return nil
	})
}

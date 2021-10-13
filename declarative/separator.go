// Copyright 2017 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import "github.com/Gipcomp/winapi"

type HSeparator struct {
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

	// Separator

	AssignTo **winapi.Separator
}

func (s HSeparator) Create(builder *Builder) error {
	w, err := winapi.NewHSeparator(builder.Parent())
	if err != nil {
		return err
	}

	if s.AssignTo != nil {
		*s.AssignTo = w
	}

	return builder.InitWidget(s, w, func() error {
		return nil
	})
}

type VSeparator struct {
	// Window

	Accessibility    Accessibility
	ContextMenuItems []MenuItem
	Enabled          Property
	Font             Font
	MaxSize          Size
	MinSize          Size
	Name             string
	OnBoundsChanged  winapi.EventHandler
	OnKeyDown        winapi.KeyEventHandler
	OnKeyPress       winapi.KeyEventHandler
	OnKeyUp          winapi.KeyEventHandler
	OnMouseDown      winapi.MouseEventHandler
	OnMouseMove      winapi.MouseEventHandler
	OnMouseUp        winapi.MouseEventHandler
	OnSizeChanged    winapi.EventHandler
	Persistent       bool
	ToolTipText      Property
	Visible          Property

	// Widget

	AlwaysConsumeSpace bool
	Column             int
	ColumnSpan         int
	Row                int
	RowSpan            int
	StretchFactor      int

	// Separator

	AssignTo **winapi.Separator
}

func (s VSeparator) Create(builder *Builder) error {
	w, err := winapi.NewVSeparator(builder.Parent())
	if err != nil {
		return err
	}

	if s.AssignTo != nil {
		*s.AssignTo = w
	}

	return builder.InitWidget(s, w, func() error {
		return nil
	})
}

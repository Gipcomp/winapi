// Copyright 2017 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/winapi"
)

type GradientComposite struct {
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

	// Container

	Children   []Widget
	Layout     Layout
	DataBinder DataBinder

	// GradientComposite

	AssignTo    **winapi.GradientComposite
	Border      bool
	Color1      Property
	Color2      Property
	Expressions func() map[string]winapi.Expression
	Functions   map[string]func(args ...interface{}) (interface{}, error)
	Vertical    Property
}

func (gc GradientComposite) Create(builder *Builder) error {
	var style uint32
	if gc.Border {
		style |= user32.WS_BORDER
	}
	w, err := winapi.NewGradientCompositeWithStyle(builder.Parent(), style)
	if err != nil {
		return err
	}

	if gc.AssignTo != nil {
		*gc.AssignTo = w
	}

	w.SetSuspended(true)
	builder.Defer(func() error {
		w.SetSuspended(false)
		return nil
	})

	return builder.InitWidget(gc, w, func() error {
		if gc.Expressions != nil {
			for name, expr := range gc.Expressions() {
				builder.expressions[name] = expr
			}
		}
		if gc.Functions != nil {
			for name, fn := range gc.Functions {
				builder.functions[name] = fn
			}
		}

		return nil
	})
}

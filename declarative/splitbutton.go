// Copyright 2016 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import "github.com/Gipcomp/winapi"

type SplitButton struct {
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

	Image     Property
	Text      Property
	OnClicked winapi.EventHandler

	// SplitButton

	AssignTo       **winapi.SplitButton
	ImageAboveText bool
	MenuItems      []MenuItem
}

func (sb SplitButton) Create(builder *Builder) error {
	w, err := winapi.NewSplitButton(builder.Parent())
	if err != nil {
		return err
	}

	if sb.AssignTo != nil {
		*sb.AssignTo = w
	}

	builder.deferBuildMenuActions(w.Menu(), sb.MenuItems)

	return builder.InitWidget(sb, w, func() error {
		if err := w.SetImageAboveText(sb.ImageAboveText); err != nil {
			return err
		}

		if sb.OnClicked != nil {
			w.Clicked().Attach(sb.OnClicked)
		}

		return nil
	})
}

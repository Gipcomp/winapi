// Copyright 2017 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/Gipcomp/winapi"
)

type LinkLabel struct {
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

	// LinkLabel

	AssignTo        **winapi.LinkLabel
	OnLinkActivated winapi.LinkLabelLinkEventHandler
	Text            Property
}

func (ll LinkLabel) Create(builder *Builder) error {
	w, err := winapi.NewLinkLabel(builder.Parent())
	if err != nil {
		return err
	}

	if ll.AssignTo != nil {
		*ll.AssignTo = w
	}

	return builder.InitWidget(ll, w, func() error {
		if ll.OnLinkActivated != nil {
			w.LinkActivated().Attach(ll.OnLinkActivated)
		}

		return nil
	})
}

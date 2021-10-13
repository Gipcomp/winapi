// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/Gipcomp/winapi"
)

type ImageViewMode int

const (
	ImageViewModeIdeal   = ImageViewMode(winapi.ImageViewModeIdeal)
	ImageViewModeCorner  = ImageViewMode(winapi.ImageViewModeCorner)
	ImageViewModeCenter  = ImageViewMode(winapi.ImageViewModeCenter)
	ImageViewModeShrink  = ImageViewMode(winapi.ImageViewModeShrink)
	ImageViewModeZoom    = ImageViewMode(winapi.ImageViewModeZoom)
	ImageViewModeStretch = ImageViewMode(winapi.ImageViewModeStretch)
)

type ImageView struct {
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

	// ImageView

	AssignTo **winapi.ImageView
	Image    Property
	Margin   Property
	Mode     ImageViewMode
}

func (iv ImageView) Create(builder *Builder) error {
	w, err := winapi.NewImageView(builder.Parent())
	if err != nil {
		return err
	}

	if iv.AssignTo != nil {
		*iv.AssignTo = w
	}

	return builder.InitWidget(iv, w, func() error {
		w.SetMode(winapi.ImageViewMode(iv.Mode))

		return nil
	})
}

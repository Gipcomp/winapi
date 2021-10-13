// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/Gipcomp/winapi"
)

type PaintMode int

const (
	PaintNormal   PaintMode = iota // erase background before PaintFunc
	PaintNoErase                   // PaintFunc clears background, single buffered
	PaintBuffered                  // PaintFunc clears background, double buffered
)

type CustomWidget struct {
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

	// CustomWidget

	AssignTo            **winapi.CustomWidget
	ClearsBackground    bool
	InvalidatesOnResize bool
	Paint               winapi.PaintFunc
	PaintPixels         winapi.PaintFunc
	PaintMode           PaintMode
	Style               uint32
}

func (cw CustomWidget) Create(builder *Builder) error {
	var w *winapi.CustomWidget
	var err error
	if cw.PaintPixels != nil {
		w, err = winapi.NewCustomWidgetPixels(builder.Parent(), uint(cw.Style), cw.PaintPixels)
	} else {
		w, err = winapi.NewCustomWidget(builder.Parent(), uint(cw.Style), cw.Paint)
	}
	if err != nil {
		return err
	}

	if cw.AssignTo != nil {
		*cw.AssignTo = w
	}

	return builder.InitWidget(cw, w, func() error {
		if cw.PaintMode != PaintNormal && cw.ClearsBackground {
			panic("PaintMode and ClearsBackground are incompatible")
		}
		w.SetClearsBackground(cw.ClearsBackground)
		w.SetInvalidatesOnResize(cw.InvalidatesOnResize)
		w.SetPaintMode(winapi.PaintMode(cw.PaintMode))

		return nil
	})
}

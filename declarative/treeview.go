// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import "github.com/Gipcomp/winapi"

type TreeView struct {
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

	// TreeView

	AssignTo             **winapi.TreeView
	ItemHeight           int
	Model                winapi.TreeModel
	OnCurrentItemChanged winapi.EventHandler
	OnExpandedChanged    winapi.TreeItemEventHandler
	OnItemActivated      winapi.EventHandler
}

func (tv TreeView) Create(builder *Builder) error {
	w, err := winapi.NewTreeView(builder.Parent())
	if err != nil {
		return err
	}

	if tv.AssignTo != nil {
		*tv.AssignTo = w
	}

	return builder.InitWidget(tv, w, func() error {
		if tv.ItemHeight > 0 {
			w.SetItemHeight(w.IntFrom96DPI(tv.ItemHeight)) // VERIFY: Item height should resize on DPI change.
		}

		if err := w.SetModel(tv.Model); err != nil {
			return err
		}

		if tv.OnCurrentItemChanged != nil {
			w.CurrentItemChanged().Attach(tv.OnCurrentItemChanged)
		}

		if tv.OnExpandedChanged != nil {
			w.ExpandedChanged().Attach(tv.OnExpandedChanged)
		}

		if tv.OnItemActivated != nil {
			w.ItemActivated().Attach(tv.OnItemActivated)
		}

		return nil
	})
}

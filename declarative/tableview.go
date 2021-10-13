// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/Gipcomp/win32/commctrl"
	"github.com/Gipcomp/winapi"
)

type TableView struct {
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

	// TableView

	AlternatingRowBG            bool
	AssignTo                    **winapi.TableView
	CellStyler                  winapi.CellStyler
	CheckBoxes                  bool
	Columns                     []TableViewColumn
	ColumnsOrderable            Property
	ColumnsSizable              Property
	CustomHeaderHeight          int
	CustomRowHeight             int
	ItemStateChangedEventDelay  int
	HeaderHidden                bool
	LastColumnStretched         bool
	Model                       interface{}
	MultiSelection              bool
	NotSortableByHeaderClick    bool
	OnCurrentIndexChanged       winapi.EventHandler
	OnItemActivated             winapi.EventHandler
	OnSelectedIndexesChanged    winapi.EventHandler
	SelectionHiddenWithoutFocus bool
	StyleCell                   func(style *winapi.CellStyle)
}

type tvStyler struct {
	dflt              winapi.CellStyler
	colStyleCellFuncs []func(style *winapi.CellStyle)
}

func (tvs *tvStyler) StyleCell(style *winapi.CellStyle) {
	if tvs.dflt != nil {
		tvs.dflt.StyleCell(style)
	}

	if col := style.Col(); col >= 0 {
		if styleCell := tvs.colStyleCellFuncs[col]; styleCell != nil {
			styleCell(style)
		}
	}
}

type styleCellFunc func(style *winapi.CellStyle)

func (scf styleCellFunc) StyleCell(style *winapi.CellStyle) {
	scf(style)
}

func (tv TableView) Create(builder *Builder) error {
	var w *winapi.TableView
	var err error
	if tv.NotSortableByHeaderClick {
		w, err = winapi.NewTableViewWithStyle(builder.Parent(), commctrl.LVS_NOSORTHEADER)
	} else {
		w, err = winapi.NewTableViewWithCfg(builder.Parent(), &winapi.TableViewCfg{CustomHeaderHeight: tv.CustomHeaderHeight, CustomRowHeight: tv.CustomRowHeight})
	}
	if err != nil {
		return err
	}

	if tv.AssignTo != nil {
		*tv.AssignTo = w
	}

	return builder.InitWidget(tv, w, func() error {
		for i := range tv.Columns {
			if err := tv.Columns[i].Create(w); err != nil {
				return err
			}
		}

		if err := w.SetModel(tv.Model); err != nil {
			return err
		}

		defaultStyler, _ := tv.Model.(winapi.CellStyler)

		if tv.CellStyler != nil {
			defaultStyler = tv.CellStyler
		}

		if tv.StyleCell != nil {
			defaultStyler = styleCellFunc(tv.StyleCell)
		}

		var hasColStyleFunc bool
		for _, c := range tv.Columns {
			if c.StyleCell != nil {
				hasColStyleFunc = true
				break
			}
		}

		if defaultStyler != nil || hasColStyleFunc {
			var styler winapi.CellStyler

			if hasColStyleFunc {
				tvs := &tvStyler{
					dflt:              defaultStyler,
					colStyleCellFuncs: make([]func(style *winapi.CellStyle), len(tv.Columns)),
				}

				styler = tvs

				for i, c := range tv.Columns {
					tvs.colStyleCellFuncs[i] = c.StyleCell
				}
			} else {
				styler = defaultStyler
			}

			w.SetCellStyler(styler)
		}

		w.SetAlternatingRowBG(tv.AlternatingRowBG)
		w.SetCheckBoxes(tv.CheckBoxes)
		w.SetItemStateChangedEventDelay(tv.ItemStateChangedEventDelay)
		if err := w.SetLastColumnStretched(tv.LastColumnStretched); err != nil {
			return err
		}
		if err := w.SetMultiSelection(tv.MultiSelection); err != nil {
			return err
		}
		if err := w.SetSelectionHiddenWithoutFocus(tv.SelectionHiddenWithoutFocus); err != nil {
			return err
		}
		if err := w.SetHeaderHidden(tv.HeaderHidden); err != nil {
			return err
		}

		if tv.OnCurrentIndexChanged != nil {
			w.CurrentIndexChanged().Attach(tv.OnCurrentIndexChanged)
		}
		if tv.OnSelectedIndexesChanged != nil {
			w.SelectedIndexesChanged().Attach(tv.OnSelectedIndexesChanged)
		}
		if tv.OnItemActivated != nil {
			w.ItemActivated().Attach(tv.OnItemActivated)
		}

		return nil
	})
}

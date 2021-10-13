// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/Gipcomp/winapi"
)

type CaseMode uint32

const (
	CaseModeMixed CaseMode = CaseMode(winapi.CaseModeMixed)
	CaseModeUpper CaseMode = CaseMode(winapi.CaseModeUpper)
	CaseModeLower CaseMode = CaseMode(winapi.CaseModeLower)
)

type LineEdit struct {
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

	// LineEdit

	AssignTo          **winapi.LineEdit
	CaseMode          CaseMode
	CueBanner         string
	MaxLength         int
	OnEditingFinished winapi.EventHandler
	OnTextChanged     winapi.EventHandler
	PasswordMode      bool
	ReadOnly          Property
	Text              Property
	TextAlignment     Alignment1D
	TextColor         winapi.Color
}

func (le LineEdit) Create(builder *Builder) error {
	w, err := winapi.NewLineEdit(builder.Parent())
	if err != nil {
		return err
	}

	if le.AssignTo != nil {
		*le.AssignTo = w
	}

	return builder.InitWidget(le, w, func() error {
		w.SetTextColor(le.TextColor)

		if err := w.SetTextAlignment(winapi.Alignment1D(le.TextAlignment)); err != nil {
			return err
		}

		if le.CueBanner != "" {
			if err := w.SetCueBanner(le.CueBanner); err != nil {
				return err
			}
		}
		w.SetMaxLength(le.MaxLength)
		w.SetPasswordMode(le.PasswordMode)

		if err := w.SetCaseMode(winapi.CaseMode(le.CaseMode)); err != nil {
			return err
		}

		if le.OnEditingFinished != nil {
			w.EditingFinished().Attach(le.OnEditingFinished)
		}
		if le.OnTextChanged != nil {
			w.TextChanged().Attach(le.OnTextChanged)
		}

		return nil
	})
}

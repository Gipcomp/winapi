// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"errors"

	"github.com/Gipcomp/winapi"
)

type Orientation byte

const (
	Horizontal Orientation = Orientation(winapi.Horizontal)
	Vertical   Orientation = Orientation(winapi.Vertical)
)

type Margins struct {
	Left   int
	Top    int
	Right  int
	Bottom int
}

func (m Margins) isZero() bool {
	return m.Left == 0 && m.Top == 0 && m.Right == 0 && m.Bottom == 0
}

func (m Margins) toW() winapi.Margins {
	return winapi.Margins{m.Left, m.Top, m.Right, m.Bottom}
}

type Rectangle struct {
	X      int
	Y      int
	Width  int
	Height int
}

func (r Rectangle) toW() winapi.Rectangle {
	return winapi.Rectangle{r.X, r.Y, r.Width, r.Height}
}

type Size struct {
	Width  int
	Height int
}

func (s Size) toW() winapi.Size {
	return winapi.Size{s.Width, s.Height}
}

func setLayoutMargins(layout winapi.Layout, margins Margins, marginsZero bool) error {
	if !marginsZero && margins.isZero() {
		margins = Margins{9, 9, 9, 9}
	}

	return layout.SetMargins(margins.toW())
}

func setLayoutSpacing(layout winapi.Layout, spacing int, spacingZero bool) error {
	if !spacingZero && spacing == 0 {
		spacing = 6
	}

	return layout.SetSpacing(spacing)
}

type HBox struct {
	Margins     Margins
	Alignment   Alignment2D
	Spacing     int
	MarginsZero bool
	SpacingZero bool
}

func (hb HBox) Create() (winapi.Layout, error) {
	l := winapi.NewHBoxLayout()

	if err := setLayoutMargins(l, hb.Margins, hb.MarginsZero); err != nil {
		return nil, err
	}

	if err := setLayoutSpacing(l, hb.Spacing, hb.SpacingZero); err != nil {
		return nil, err
	}

	if err := l.SetAlignment(winapi.Alignment2D(hb.Alignment)); err != nil {
		return nil, err
	}

	return l, nil
}

type VBox struct {
	Margins     Margins
	Alignment   Alignment2D
	Spacing     int
	MarginsZero bool
	SpacingZero bool
}

func (vb VBox) Create() (winapi.Layout, error) {
	l := winapi.NewVBoxLayout()

	if err := setLayoutMargins(l, vb.Margins, vb.MarginsZero); err != nil {
		return nil, err
	}

	if err := setLayoutSpacing(l, vb.Spacing, vb.SpacingZero); err != nil {
		return nil, err
	}

	if err := l.SetAlignment(winapi.Alignment2D(vb.Alignment)); err != nil {
		return nil, err
	}

	return l, nil
}

type Grid struct {
	Rows        int
	Columns     int
	Margins     Margins
	Alignment   Alignment2D
	Spacing     int
	MarginsZero bool
	SpacingZero bool
}

func (g Grid) Create() (winapi.Layout, error) {
	if g.Rows > 0 && g.Columns > 0 {
		return nil, errors.New("only one of Rows and Columns may be > 0")
	}

	l := winapi.NewGridLayout()

	if err := setLayoutMargins(l, g.Margins, g.MarginsZero); err != nil {
		return nil, err
	}

	if err := setLayoutSpacing(l, g.Spacing, g.SpacingZero); err != nil {
		return nil, err
	}

	if err := l.SetAlignment(winapi.Alignment2D(g.Alignment)); err != nil {
		return nil, err
	}

	return l, nil
}

type Flow struct {
	Margins     Margins
	Alignment   Alignment2D
	Spacing     int
	MarginsZero bool
	SpacingZero bool
}

func (f Flow) Create() (winapi.Layout, error) {
	l := winapi.NewFlowLayout()

	if err := setLayoutMargins(l, f.Margins, f.MarginsZero); err != nil {
		return nil, err
	}

	if err := setLayoutSpacing(l, f.Spacing, f.SpacingZero); err != nil {
		return nil, err
	}

	if err := l.SetAlignment(winapi.Alignment2D(f.Alignment)); err != nil {
		return nil, err
	}

	return l, nil
}

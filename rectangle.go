// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import "github.com/Gipcomp/win32/gdi32"

// Rectangle defines upper left corner with width and height region in 1/96" units, or native
// pixels, or grid rows and columns.
type Rectangle struct {
	X, Y, Width, Height int
}

func (r Rectangle) IsZero() bool {
	return r.X == 0 && r.Y == 0 && r.Width == 0 && r.Height == 0
}

func rectangleFromRECT(r gdi32.RECT) Rectangle {
	return Rectangle{
		X:      int(r.Left),
		Y:      int(r.Top),
		Width:  int(r.Right - r.Left),
		Height: int(r.Bottom - r.Top),
	}
}

func (r Rectangle) Left() int {
	return r.X
}

func (r Rectangle) Top() int {
	return r.Y
}

func (r Rectangle) Right() int {
	return r.X + r.Width - 1
}

func (r Rectangle) Bottom() int {
	return r.Y + r.Height - 1
}

func (r Rectangle) Location() Point {
	return Point{r.X, r.Y}
}

func (r *Rectangle) SetLocation(p Point) Rectangle {
	r.X = p.X
	r.Y = p.Y

	return *r
}

func (r Rectangle) Size() Size {
	return Size{r.Width, r.Height}
}

func (r *Rectangle) SetSize(s Size) Rectangle {
	r.Width = s.Width
	r.Height = s.Height

	return *r
}

func (r Rectangle) toRECT() gdi32.RECT {
	return gdi32.RECT{
		Left:   int32(r.X),
		Top:    int32(r.Y),
		Right:  int32(r.X + r.Width),
		Bottom: int32(r.Y + r.Height),
	}
}

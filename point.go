// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import "github.com/Gipcomp/win32/gdi32"

// Point defines 2D coordinate in 1/96" units ot native pixels.
type Point struct {
	X, Y int
}

func (p Point) toPOINT() gdi32.POINT {
	return gdi32.POINT{
		X: int32(p.X),
		Y: int32(p.Y),
	}
}

func pointPixelsFromPOINT(p gdi32.POINT) Point {
	return Point{
		X: int(p.X),
		Y: int(p.Y),
	}
}

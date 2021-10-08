// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"github.com/Gipcomp/win32/gdi32"
)

// Size defines width and height in 1/96" units or native pixels, or dialog base units.
//
// When Size is used for DPI metrics, it defines a 1"x1" rectangle in native pixels.
type Size struct {
	Width, Height int
}

func (s Size) IsZero() bool {
	return s.Width == 0 && s.Height == 0
}

func (s Size) toSIZE() gdi32.SIZE {
	return gdi32.SIZE{
		CX: int32(s.Width),
		CY: int32(s.Height),
	}
}

func minSize(a, b Size) Size {
	var s Size

	if a.Width < b.Width {
		s.Width = a.Width
	} else {
		s.Width = b.Width
	}

	if a.Height < b.Height {
		s.Height = a.Height
	} else {
		s.Height = b.Height
	}

	return s
}

func maxSize(a, b Size) Size {
	var s Size

	if a.Width > b.Width {
		s.Width = a.Width
	} else {
		s.Width = b.Width
	}

	if a.Height > b.Height {
		s.Height = a.Height
	} else {
		s.Height = b.Height
	}

	return s
}

func sizeFromSIZE(s gdi32.SIZE) Size {
	return Size{
		Width:  int(s.CX),
		Height: int(s.CY),
	}
}

func sizeFromRECT(r gdi32.RECT) Size {
	return Size{
		Width:  int(r.Right - r.Left),
		Height: int(r.Bottom - r.Top),
	}
}

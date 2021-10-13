// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/Gipcomp/winapi"
)

type Font struct {
	Family    string
	PointSize int
	Bold      bool
	Italic    bool
	Underline bool
	StrikeOut bool
}

func (f Font) Create() (*winapi.Font, error) {
	if f.Family == "" && f.PointSize == 0 {
		return nil, nil
	}

	var fs winapi.FontStyle

	if f.Bold {
		fs |= winapi.FontBold
	}
	if f.Italic {
		fs |= winapi.FontItalic
	}
	if f.Underline {
		fs |= winapi.FontUnderline
	}
	if f.StrikeOut {
		fs |= winapi.FontStrikeOut
	}

	return winapi.NewFont(f.Family, f.PointSize, fs)
}

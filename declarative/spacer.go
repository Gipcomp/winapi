// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import "github.com/Gipcomp/winapi"

type HSpacer struct {
	// Window

	MaxSize Size
	MinSize Size
	Name    string

	// Widget

	Column        int
	ColumnSpan    int
	Row           int
	RowSpan       int
	StretchFactor int

	// Spacer

	GreedyLocallyOnly bool
	Size              int
}

func (hs HSpacer) Create(builder *Builder) (err error) {
	var flags winapi.LayoutFlags
	if hs.Size == 0 {
		flags = winapi.ShrinkableHorz | winapi.GrowableHorz | winapi.GreedyHorz
	}

	var w *winapi.Spacer
	if w, err = winapi.NewSpacerWithCfg(builder.Parent(), &winapi.SpacerCfg{
		LayoutFlags:       flags,
		SizeHint:          Size{Width: hs.Size}.toW(),
		GreedyLocallyOnly: hs.GreedyLocallyOnly,
	}); err != nil {
		return
	}

	return builder.InitWidget(hs, w, nil)
}

type VSpacer struct {
	// Window

	MaxSize Size
	MinSize Size
	Name    string

	// Widget

	Column        int
	ColumnSpan    int
	Row           int
	RowSpan       int
	StretchFactor int

	// Spacer

	GreedyLocallyOnly bool
	Size              int
}

func (vs VSpacer) Create(builder *Builder) (err error) {
	var flags winapi.LayoutFlags
	if vs.Size == 0 {
		flags = winapi.ShrinkableVert | winapi.GrowableVert | winapi.GreedyVert
	}

	var w *winapi.Spacer
	if w, err = winapi.NewSpacerWithCfg(builder.Parent(), &winapi.SpacerCfg{
		LayoutFlags:       flags,
		SizeHint:          Size{Height: vs.Size}.toW(),
		GreedyLocallyOnly: vs.GreedyLocallyOnly,
	}); err != nil {
		return
	}

	return builder.InitWidget(vs, w, nil)
}

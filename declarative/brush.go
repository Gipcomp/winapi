// Copyright 2017 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"strconv"

	"github.com/Gipcomp/winapi"
)

type TransparentBrush struct {
}

func (TransparentBrush) Create() (winapi.Brush, error) {
	return winapi.NullBrush(), nil
}

type SolidColorBrush struct {
	Color winapi.Color
}

func (scb SolidColorBrush) Create() (winapi.Brush, error) {
	return winapi.NewSolidColorBrush(scb.Color)
}

type SystemColorBrush struct {
	Color winapi.SystemColor
}

func (scb SystemColorBrush) Create() (winapi.Brush, error) {
	return winapi.NewSystemColorBrush(scb.Color)
}

type BitmapBrush struct {
	Image interface{}
}

func (bb BitmapBrush) Create() (winapi.Brush, error) {
	var bmp *winapi.Bitmap
	var err error

	switch img := bb.Image.(type) {
	case *winapi.Bitmap:
		bmp = img

	case string:
		if bmp, err = winapi.Resources.Bitmap(img); err != nil {
			return nil, err
		}

	case int:
		if bmp, err = winapi.Resources.Bitmap(strconv.Itoa(img)); err != nil {
			return nil, err
		}

	default:
		return nil, winapi.ErrInvalidType
	}

	return winapi.NewBitmapBrush(bmp)
}

type GradientBrush struct {
	Vertexes  []winapi.GradientVertex
	Triangles []winapi.GradientTriangle
}

func (gb GradientBrush) Create() (winapi.Brush, error) {
	return winapi.NewGradientBrush(gb.Vertexes, gb.Triangles)
}

type HorizontalGradientBrush struct {
	Stops []winapi.GradientStop
}

func (hgb HorizontalGradientBrush) Create() (winapi.Brush, error) {
	return winapi.NewHorizontalGradientBrush(hgb.Stops)
}

type VerticalGradientBrush struct {
	Stops []winapi.GradientStop
}

func (vgb VerticalGradientBrush) Create() (winapi.Brush, error) {
	return winapi.NewVerticalGradientBrush(vgb.Stops)
}

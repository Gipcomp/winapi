// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"github.com/Gipcomp/win32/gdi32"
)

type PenStyle int

// Pen styles
const (
	PenSolid       PenStyle = gdi32.PS_SOLID
	PenDash        PenStyle = gdi32.PS_DASH
	PenDot         PenStyle = gdi32.PS_DOT
	PenDashDot     PenStyle = gdi32.PS_DASHDOT
	PenDashDotDot  PenStyle = gdi32.PS_DASHDOTDOT
	PenNull        PenStyle = gdi32.PS_NULL
	PenInsideFrame PenStyle = gdi32.PS_INSIDEFRAME
	PenUserStyle   PenStyle = gdi32.PS_USERSTYLE
	PenAlternate   PenStyle = gdi32.PS_ALTERNATE
)

// Pen cap styles (geometric pens only)
const (
	PenCapRound  PenStyle = gdi32.PS_ENDCAP_ROUND
	PenCapSquare PenStyle = gdi32.PS_ENDCAP_SQUARE
	PenCapFlat   PenStyle = gdi32.PS_ENDCAP_FLAT
)

// Pen join styles (geometric pens only)
const (
	PenJoinBevel PenStyle = gdi32.PS_JOIN_BEVEL
	PenJoinMiter PenStyle = gdi32.PS_JOIN_MITER
	PenJoinRound PenStyle = gdi32.PS_JOIN_ROUND
)

type Pen interface {
	handleForDPI(dpi int) gdi32.HPEN
	Dispose()
	Style() PenStyle

	// Width returns pen width in 1/96" units.
	Width() int
}

type nullPen struct {
	hPen gdi32.HPEN
}

func newNullPen() *nullPen {
	lb := &gdi32.LOGBRUSH{LbStyle: gdi32.BS_NULL}

	hPen := gdi32.ExtCreatePen(gdi32.PS_COSMETIC|gdi32.PS_NULL, 1, lb, 0, nil)
	if hPen == 0 {
		panic("failed to create null brush")
	}

	return &nullPen{hPen: hPen}
}

func (p *nullPen) Dispose() {
	if p.hPen != 0 {
		gdi32.DeleteObject(gdi32.HGDIOBJ(p.hPen))

		p.hPen = 0
	}
}

func (p *nullPen) handleForDPI(dpi int) gdi32.HPEN {
	return p.hPen
}

func (p *nullPen) Style() PenStyle {
	return PenNull
}

func (p *nullPen) Width() int {
	return 0
}

var nullPenSingleton Pen

func init() {
	AppendToWalkInit(func() {
		nullPenSingleton = newNullPen()
	})
}

func NullPen() Pen {
	return nullPenSingleton
}

type CosmeticPen struct {
	hPen  gdi32.HPEN
	style PenStyle
	color Color
}

func NewCosmeticPen(style PenStyle, color Color) (*CosmeticPen, error) {
	lb := &gdi32.LOGBRUSH{LbStyle: gdi32.BS_SOLID, LbColor: gdi32.COLORREF(color)}

	style |= gdi32.PS_COSMETIC

	hPen := gdi32.ExtCreatePen(uint32(style), 1, lb, 0, nil)
	if hPen == 0 {
		return nil, newError("ExtCreatePen failed")
	}

	return &CosmeticPen{hPen: hPen, style: style, color: color}, nil
}

func (p *CosmeticPen) Dispose() {
	if p.hPen != 0 {
		gdi32.DeleteObject(gdi32.HGDIOBJ(p.hPen))

		p.hPen = 0
	}
}

func (p *CosmeticPen) handleForDPI(dpi int) gdi32.HPEN {
	return p.hPen
}

func (p *CosmeticPen) Style() PenStyle {
	return p.style
}

func (p *CosmeticPen) Color() Color {
	return p.color
}

func (p *CosmeticPen) Width() int {
	return 1
}

type GeometricPen struct {
	dpi2hPen   map[int]gdi32.HPEN
	style      PenStyle
	brush      Brush
	width96dpi int
}

// NewGeometricPen prepares new geometric pen. width parameter is specified in 1/96" units.
func NewGeometricPen(style PenStyle, width int, brush Brush) (*GeometricPen, error) {
	if brush == nil {
		return nil, newError("brush cannot be nil")
	}

	style |= gdi32.PS_GEOMETRIC

	return &GeometricPen{
		style:      style,
		width96dpi: width,
		brush:      brush,
	}, nil
}

func (p *GeometricPen) Dispose() {
	if len(p.dpi2hPen) == 0 {
		return
	}

	for dpi, hPen := range p.dpi2hPen {
		gdi32.DeleteObject(gdi32.HGDIOBJ(hPen))
		delete(p.dpi2hPen, dpi)
	}
}

func (p *GeometricPen) handleForDPI(dpi int) gdi32.HPEN {
	hPen, _ := p.handleForDPIWithError(dpi)
	return hPen
}

func (p *GeometricPen) handleForDPIWithError(dpi int) (gdi32.HPEN, error) {
	if p.dpi2hPen == nil {
		p.dpi2hPen = make(map[int]gdi32.HPEN)
	} else if handle, ok := p.dpi2hPen[dpi]; ok {
		return handle, nil
	}

	hPen := gdi32.ExtCreatePen(
		uint32(p.style),
		uint32(IntFrom96DPI(p.width96dpi, dpi)),
		p.brush.logbrush(), 0, nil)
	if hPen == 0 {
		return 0, newError("ExtCreatePen failed")
	}

	p.dpi2hPen[dpi] = hPen

	return hPen, nil
}

func (p *GeometricPen) Style() PenStyle {
	return p.style
}

// Width returns pen width in 1/96" units.
func (p *GeometricPen) Width() int {
	return p.width96dpi
}

func (p *GeometricPen) Brush() Brush {
	return p.brush
}

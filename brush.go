// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"unsafe"

	"github.com/Gipcomp/win32/gdi32"
	"github.com/Gipcomp/win32/user32"
)

type HatchStyle int

const (
	HatchHorizontal       HatchStyle = gdi32.HS_HORIZONTAL
	HatchVertical         HatchStyle = gdi32.HS_VERTICAL
	HatchForwardDiagonal  HatchStyle = gdi32.HS_FDIAGONAL
	HatchBackwardDiagonal HatchStyle = gdi32.HS_BDIAGONAL
	HatchCross            HatchStyle = gdi32.HS_CROSS
	HatchDiagonalCross    HatchStyle = gdi32.HS_DIAGCROSS
)

type SystemColor int

const (
	SysColor3DDkShadow              SystemColor = user32.COLOR_3DDKSHADOW
	SysColor3DFace                  SystemColor = user32.COLOR_3DFACE
	SysColor3DHighlight             SystemColor = user32.COLOR_3DHIGHLIGHT
	SysColor3DLight                 SystemColor = user32.COLOR_3DLIGHT
	SysColor3DShadow                SystemColor = user32.COLOR_3DSHADOW
	SysColorActiveBorder            SystemColor = user32.COLOR_ACTIVEBORDER
	SysColorActiveCaption           SystemColor = user32.COLOR_ACTIVECAPTION
	SysColorAppWorkspace            SystemColor = user32.COLOR_APPWORKSPACE
	SysColorBackground              SystemColor = user32.COLOR_BACKGROUND
	SysColorDesktop                 SystemColor = user32.COLOR_DESKTOP
	SysColorBtnFace                 SystemColor = user32.COLOR_BTNFACE
	SysColorBtnHighlight            SystemColor = user32.COLOR_BTNHIGHLIGHT
	SysColorBtnShadow               SystemColor = user32.COLOR_BTNSHADOW
	SysColorBtnText                 SystemColor = user32.COLOR_BTNTEXT
	SysColorCaptionText             SystemColor = user32.COLOR_CAPTIONTEXT
	SysColorGrayText                SystemColor = user32.COLOR_GRAYTEXT
	SysColorHighlight               SystemColor = user32.COLOR_HIGHLIGHT
	SysColorHighlightText           SystemColor = user32.COLOR_HIGHLIGHTTEXT
	SysColorInactiveBorder          SystemColor = user32.COLOR_INACTIVEBORDER
	SysColorInactiveCaption         SystemColor = user32.COLOR_INACTIVECAPTION
	SysColorInactiveCaptionText     SystemColor = user32.COLOR_INACTIVECAPTIONTEXT
	SysColorInfoBk                  SystemColor = user32.COLOR_INFOBK
	SysColorInfoText                SystemColor = user32.COLOR_INFOTEXT
	SysColorMenu                    SystemColor = user32.COLOR_MENU
	SysColorMenuText                SystemColor = user32.COLOR_MENUTEXT
	SysColorScrollBar               SystemColor = user32.COLOR_SCROLLBAR
	SysColorWindow                  SystemColor = user32.COLOR_WINDOW
	SysColorWindowFrame             SystemColor = user32.COLOR_WINDOWFRAME
	SysColorWindowText              SystemColor = user32.COLOR_WINDOWTEXT
	SysColorHotLight                SystemColor = user32.COLOR_HOTLIGHT
	SysColorGradientActiveCaption   SystemColor = user32.COLOR_GRADIENTACTIVECAPTION
	SysColorGradientInactiveCaption SystemColor = user32.COLOR_GRADIENTINACTIVECAPTION
)

type Brush interface {
	Dispose()
	handle() gdi32.HBRUSH
	logbrush() *gdi32.LOGBRUSH
	attachWindow(wb *WindowBase)
	detachWindow(wb *WindowBase)
	simple() bool
}

type perWindowBrush interface {
	Brush
	delegateForWindow(wb *WindowBase) Brush
}

type windowBrushInfo struct {
	SizeChangedHandle int
	Delegate          *BitmapBrush
}

type brushBase struct {
	hBrush  gdi32.HBRUSH
	wb2info map[*WindowBase]*windowBrushInfo
}

func (bb *brushBase) Dispose() {
	if bb.hBrush != 0 {
		gdi32.DeleteObject(gdi32.HGDIOBJ(bb.hBrush))

		bb.hBrush = 0
	}
}

func (bb *brushBase) handle() gdi32.HBRUSH {
	return bb.hBrush
}

func (bb *brushBase) attachWindow(wb *WindowBase) {
	if wb == nil {
		return
	}

	if bb.wb2info == nil {
		bb.wb2info = make(map[*WindowBase]*windowBrushInfo)
	}

	bb.wb2info[wb] = nil
}

func (bb *brushBase) detachWindow(wb *WindowBase) {
	if bb.wb2info == nil || wb == nil {
		return
	}

	delete(bb.wb2info, wb)

	if len(bb.wb2info) == 0 {
		bb.Dispose()
	}
}

type nullBrush struct {
	brushBase
}

func newNullBrush() *nullBrush {
	lb := &gdi32.LOGBRUSH{LbStyle: gdi32.BS_NULL}

	hBrush := gdi32.CreateBrushIndirect(lb)
	if hBrush == 0 {
		panic("failed to create null brush")
	}

	return &nullBrush{brushBase: brushBase{hBrush: hBrush}}
}

func (b *nullBrush) Dispose() {
	if b == nullBrushSingleton {
		return
	}

	b.brushBase.Dispose()
}

func (*nullBrush) logbrush() *gdi32.LOGBRUSH {
	return &gdi32.LOGBRUSH{LbStyle: gdi32.BS_NULL}
}

func (*nullBrush) simple() bool {
	return true
}

var (
	nullBrushSingleton   Brush
	sysColorBtnFaceBrush *SystemColorBrush
)

func NullBrush() Brush {
	return nullBrushSingleton
}

type SystemColorBrush struct {
	brushBase
	sysColor SystemColor
}

func init() {
	AppendToWalkInit(func() {
		nullBrushSingleton = newNullBrush()
		sysColorBtnFaceBrush, _ = NewSystemColorBrush(SysColorBtnFace)
	})
}

func NewSystemColorBrush(sysColor SystemColor) (*SystemColorBrush, error) {
	hBrush := user32.GetSysColorBrush(int(sysColor))
	if hBrush == 0 {
		return nil, newError("GetSysColorBrush failed")
	}

	return &SystemColorBrush{brushBase: brushBase{hBrush: hBrush}, sysColor: sysColor}, nil
}

func (b *SystemColorBrush) Color() Color {
	return Color(user32.GetSysColor(int(b.sysColor)))
}

func (b *SystemColorBrush) SystemColor() SystemColor {
	return b.sysColor
}

func (*SystemColorBrush) Dispose() {
	// nop
}

func (b *SystemColorBrush) logbrush() *gdi32.LOGBRUSH {
	return &gdi32.LOGBRUSH{
		LbStyle: gdi32.BS_SOLID,
		LbColor: gdi32.COLORREF(user32.GetSysColor(int(b.sysColor))),
	}
}

func (*SystemColorBrush) simple() bool {
	return true
}

type SolidColorBrush struct {
	brushBase
	color Color
}

func NewSolidColorBrush(color Color) (*SolidColorBrush, error) {
	lb := &gdi32.LOGBRUSH{LbStyle: gdi32.BS_SOLID, LbColor: gdi32.COLORREF(color)}

	hBrush := gdi32.CreateBrushIndirect(lb)
	if hBrush == 0 {
		return nil, newError("CreateBrushIndirect failed")
	}

	return &SolidColorBrush{brushBase: brushBase{hBrush: hBrush}, color: color}, nil
}

func (b *SolidColorBrush) Color() Color {
	return b.color
}

func (b *SolidColorBrush) logbrush() *gdi32.LOGBRUSH {
	return &gdi32.LOGBRUSH{LbStyle: gdi32.BS_SOLID, LbColor: gdi32.COLORREF(b.color)}
}

func (*SolidColorBrush) simple() bool {
	return true
}

type HatchBrush struct {
	brushBase
	color Color
	style HatchStyle
}

func NewHatchBrush(color Color, style HatchStyle) (*HatchBrush, error) {
	lb := &gdi32.LOGBRUSH{LbStyle: gdi32.BS_HATCHED, LbColor: gdi32.COLORREF(color), LbHatch: uintptr(style)}

	hBrush := gdi32.CreateBrushIndirect(lb)
	if hBrush == 0 {
		return nil, newError("CreateBrushIndirect failed")
	}

	return &HatchBrush{brushBase: brushBase{hBrush: hBrush}, color: color, style: style}, nil
}

func (b *HatchBrush) Color() Color {
	return b.color
}

func (b *HatchBrush) logbrush() *gdi32.LOGBRUSH {
	return &gdi32.LOGBRUSH{LbStyle: gdi32.BS_HATCHED, LbColor: gdi32.COLORREF(b.color), LbHatch: uintptr(b.style)}
}

func (b *HatchBrush) Style() HatchStyle {
	return b.style
}

func (b *HatchBrush) simple() bool {
	return false
}

type BitmapBrush struct {
	brushBase
	bitmap *Bitmap
}

func NewBitmapBrush(bitmap *Bitmap) (*BitmapBrush, error) {
	if bitmap == nil {
		return nil, newError("bitmap cannot be nil")
	}

	hBrush := gdi32.CreatePatternBrush(bitmap.hBmp)
	if hBrush == 0 {
		return nil, newError("CreatePatternBrush failed")
	}

	return &BitmapBrush{brushBase: brushBase{hBrush: hBrush}, bitmap: bitmap}, nil
}

func (b *BitmapBrush) logbrush() *gdi32.LOGBRUSH {
	return &gdi32.LOGBRUSH{LbStyle: gdi32.BS_DIBPATTERN, LbColor: gdi32.DIB_RGB_COLORS, LbHatch: uintptr(b.bitmap.hPackedDIB)}
}

func (b *BitmapBrush) Bitmap() *Bitmap {
	return b.bitmap
}

func (b *BitmapBrush) simple() bool {
	return false
}

type GradientStop struct {
	Offset float64
	Color  Color
}

type GradientVertex struct {
	X     float64
	Y     float64
	Color Color
}

type GradientTriangle struct {
	Vertex1 int
	Vertex2 int
	Vertex3 int
}

type GradientBrush struct {
	brushBase
	mainDelegate *BitmapBrush
	vertexes     []GradientVertex
	triangles    []GradientTriangle
	orientation  gradientOrientation
	absolute     bool
}

type gradientOrientation int

const (
	gradientOrientationNone gradientOrientation = iota
	gradientOrientationHorizontal
	gradientOrientationVertical
)

func NewHorizontalGradientBrush(stops []GradientStop) (*GradientBrush, error) {
	return newGradientBrushWithOrientation(stops, gradientOrientationHorizontal)
}

func NewVerticalGradientBrush(stops []GradientStop) (*GradientBrush, error) {
	return newGradientBrushWithOrientation(stops, gradientOrientationVertical)
}

func newGradientBrushWithOrientation(stops []GradientStop, orientation gradientOrientation) (*GradientBrush, error) {
	if len(stops) < 2 {
		return nil, newError("at least 2 stops are required")
	}

	var vertexes []GradientVertex
	var triangles []GradientTriangle

	for i, stop := range stops {
		var x0, y0, x1, y1 float64
		if orientation == gradientOrientationHorizontal {
			x0 = stop.Offset
			x1 = stop.Offset
			y1 = 1.0
		} else {
			y0 = stop.Offset
			x1 = 1.0
			y1 = stop.Offset
		}

		vertexes = append(vertexes, GradientVertex{X: x0, Y: y0, Color: stop.Color})
		vertexes = append(vertexes, GradientVertex{X: x1, Y: y1, Color: stop.Color})

		if i > 0 {
			triangles = append(triangles, GradientTriangle{Vertex1: i*2 - 2, Vertex2: i*2 + 1, Vertex3: i*2 - 1})
			triangles = append(triangles, GradientTriangle{Vertex1: i*2 - 2, Vertex2: i * 2, Vertex3: i*2 + 1})
		}
	}

	return newGradientBrush(vertexes, triangles, orientation)
}

func NewGradientBrush(vertexes []GradientVertex, triangles []GradientTriangle) (*GradientBrush, error) {
	if len(vertexes) < 3 {
		return nil, newError("at least 3 vertexes are required")
	}

	if len(triangles) < 1 {
		return nil, newError("at least 1 triangle is required")
	}

	return newGradientBrush(vertexes, triangles, gradientOrientationNone)
}

func newGradientBrush(vertexes []GradientVertex, triangles []GradientTriangle, orientation gradientOrientation) (*GradientBrush, error) {
	var size Size
	for _, v := range vertexes {
		size = maxSize(size, Size{int(v.X), int(v.Y)})
	}

	gb := &GradientBrush{vertexes: vertexes, triangles: triangles, orientation: orientation, absolute: size.Width > 1 || size.Height > 1}

	if gb.absolute {
		bb, err := gb.create(size)
		if err != nil {
			return nil, err
		}

		gb.mainDelegate = bb
		gb.hBrush = bb.hBrush
	}

	return gb, nil
}

func (b *GradientBrush) logbrush() *gdi32.LOGBRUSH {
	if b.mainDelegate == nil {
		return nil
	}

	return b.mainDelegate.logbrush()
}

func (*GradientBrush) simple() bool {
	return false
}

// create creates a gradient brush at given size in native pixels.
func (b *GradientBrush) create(size Size) (*BitmapBrush, error) {
	var disposables Disposables
	defer disposables.Treat()

	switch b.orientation {
	case gradientOrientationHorizontal:
		size.Height = 1

	case gradientOrientationVertical:
		size.Width = 1
	}

	bitmap, err := NewBitmapForDPI(size, 96) // Size is in native pixels and bitmap is used for brush only => DPI is not used anywhere.
	if err != nil {
		return nil, err
	}
	disposables.Add(bitmap)

	canvas, err := NewCanvasFromImage(bitmap)
	if err != nil {
		return nil, err
	}
	defer canvas.Dispose()

	var scaleX, scaleY float64
	if b.absolute {
		scaleX, scaleY = 1, 1
	} else {
		scaleX, scaleY = float64(size.Width), float64(size.Height)
	}

	vertexes := make([]gdi32.TRIVERTEX, len(b.vertexes))
	for i, src := range b.vertexes {
		dst := &vertexes[i]

		dst.X = int32(src.X * scaleX)
		dst.Y = int32(src.Y * scaleY)
		dst.Red = uint16(src.Color.R()) * 256
		dst.Green = uint16(src.Color.G()) * 256
		dst.Blue = uint16(src.Color.B()) * 256
	}

	triangles := make([]gdi32.GRADIENT_TRIANGLE, len(b.triangles))
	for i, src := range b.triangles {
		dst := &triangles[i]

		dst.Vertex1 = uint32(src.Vertex1)
		dst.Vertex2 = uint32(src.Vertex2)
		dst.Vertex3 = uint32(src.Vertex3)
	}

	if !gdi32.GradientFill(canvas.hdc, &vertexes[0], uint32(len(vertexes)), unsafe.Pointer(&triangles[0]), uint32(len(triangles)), gdi32.GRADIENT_FILL_TRIANGLE) {
		return nil, newError("GradientFill failed")
	}

	disposables.Spare()

	return NewBitmapBrush(bitmap)
}

func (b *GradientBrush) attachWindow(wb *WindowBase) {
	b.brushBase.attachWindow(wb)

	if b.absolute {
		return
	}

	var info *windowBrushInfo

	update := func() {
		if bb, err := b.create(wb.window.ClientBoundsPixels().Size()); err == nil {
			if info.Delegate != nil {
				info.Delegate.bitmap.Dispose()
				info.Delegate.Dispose()
			}

			info.Delegate = bb

			wb.Invalidate()
		}
	}

	info = &windowBrushInfo{
		SizeChangedHandle: wb.SizeChanged().Attach(update),
	}

	update()

	b.wb2info[wb] = info
}

func (b *GradientBrush) detachWindow(wb *WindowBase) {
	if !b.absolute {
		if info, ok := b.wb2info[wb]; ok {
			if info.Delegate != nil {
				info.Delegate.bitmap.Dispose()
				info.Delegate.Dispose()
			}

			wb.SizeChanged().Detach(info.SizeChangedHandle)
		}
	}

	b.brushBase.detachWindow(wb)
}

func (b *GradientBrush) delegateForWindow(wb *WindowBase) Brush {
	if b.absolute {
		return b.mainDelegate
	}

	if info, ok := b.wb2info[wb]; ok && info.Delegate != nil {
		return info.Delegate
	}

	return nil
}

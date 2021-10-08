// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"syscall"
	"unsafe"

	"github.com/Gipcomp/win32/gdi32"
	"github.com/Gipcomp/win32/gdiplus"
	"github.com/Gipcomp/win32/kernel32"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/win"
)

const inchesPerMeter float64 = 39.37007874

type Bitmap struct {
	hBmp               gdi32.HBITMAP
	hPackedDIB         kernel32.HGLOBAL
	size               Size // in native pixels
	dpi                int
	transparencyStatus transparencyStatus
}

type transparencyStatus byte

const (
	transparencyUnknown transparencyStatus = iota
	transparencyOpaque
	transparencyTransparent
)

func BitmapFrom(src interface{}, dpi int) (*Bitmap, error) {
	if src == nil {
		return nil, nil
	}

	img, err := ImageFrom(src)
	if err != nil {
		return nil, err
	}

	return iconCache.Bitmap(img, dpi)
}

// NewBitmap creates an opaque bitmap with given size in 1/96" units at screen DPI.
//
// Deprecated: Newer applications should use NewBitmapForDPI.
func NewBitmap(size Size) (*Bitmap, error) {
	dpi := screenDPI()
	return newBitmap(SizeFrom96DPI(size, dpi), false, dpi)
}

// NewBitmapForDPI creates an opaque bitmap with given size in native pixels and DPI.
func NewBitmapForDPI(size Size, dpi int) (*Bitmap, error) {
	return newBitmap(size, false, dpi)
}

// NewBitmapWithTransparentPixels creates a transparent bitmap with given size in 1/96" units at screen DPI.
//
// Deprecated: Newer applications should use NewBitmapWithTransparentPixelsForDPI.
func NewBitmapWithTransparentPixels(size Size) (*Bitmap, error) {
	dpi := screenDPI()
	return newBitmap(SizeFrom96DPI(size, dpi), true, dpi)
}

// NewBitmapWithTransparentPixelsForDPI creates a transparent bitmap with given size in native pixels and DPI.
func NewBitmapWithTransparentPixelsForDPI(size Size, dpi int) (*Bitmap, error) {
	return newBitmap(size, true, dpi)
}

// newBitmap creates a bitmap with given size in native pixels and DPI.
func newBitmap(size Size, transparent bool, dpi int) (bmp *Bitmap, err error) {
	err = withCompatibleDC(func(hdc gdi32.HDC) error {
		bufSize := int(size.Width * size.Height * 4)

		var hdr gdi32.BITMAPINFOHEADER
		hdr.BiSize = uint32(unsafe.Sizeof(hdr))
		hdr.BiBitCount = 32
		hdr.BiCompression = gdi32.BI_RGB
		hdr.BiPlanes = 1
		hdr.BiWidth = int32(size.Width)
		hdr.BiHeight = int32(size.Height)
		hdr.BiSizeImage = uint32(bufSize)
		dpm := int32(math.Round(float64(dpi) * inchesPerMeter))
		hdr.BiXPelsPerMeter = dpm
		hdr.BiYPelsPerMeter = dpm

		var bitsPtr unsafe.Pointer

		hBmp := gdi32.CreateDIBSection(hdc, &hdr, gdi32.DIB_RGB_COLORS, &bitsPtr, 0, 0)
		switch hBmp {
		case 0, kernel32.ERROR_INVALID_PARAMETER:
			return newError("CreateDIBSection failed")
		}

		if transparent {
			gdi32.GdiFlush()

			bits := (*[1 << 24]byte)(bitsPtr)

			for i := 0; i < bufSize; i += 4 {
				// Mark pixel as not drawn to by GDI.
				bits[i+3] = 0x01
			}
		}

		bmp, err = newBitmapFromHBITMAP(hBmp, dpi)
		return err
	})

	return
}

// NewBitmapFromFile creates new bitmap from a bitmap file at 96dpi.
//
// Deprecated: Newer applications should use NewBitmapFromFileForDPI.
func NewBitmapFromFile(filePath string) (*Bitmap, error) {
	return NewBitmapFromFileForDPI(filePath, 96)
}

// NewBitmapFromFileForDPI creates new bitmap from a bitmap file at given DPI.
func NewBitmapFromFileForDPI(filePath string, dpi int) (*Bitmap, error) {
	var si gdiplus.GdiplusStartupInput
	si.GdiplusVersion = 1
	if status := gdiplus.GdiplusStartup(&si, nil); status != gdiplus.Ok {
		return nil, newError(fmt.Sprintf("GdiplusStartup failed with status '%s'", status))
	}
	defer gdiplus.GdiplusShutdown()

	var gpBmp *gdiplus.GpBitmap
	fiLePathPtr, err := syscall.UTF16PtrFromString(filePath)
	if err != nil {
		newError(err.Error())
	}
	if status := gdiplus.GdipCreateBitmapFromFile(fiLePathPtr, &gpBmp); status != gdiplus.Ok {
		return nil, newError(fmt.Sprintf("GdipCreateBitmapFromFile failed with status '%s' for file '%s'", status, filePath))
	}
	defer gdiplus.GdipDisposeImage((*gdiplus.GpImage)(gpBmp))

	var hBmp gdi32.HBITMAP
	if status := gdiplus.GdipCreateHBITMAPFromBitmap(gpBmp, &hBmp, 0); status != gdiplus.Ok {
		return nil, newError(fmt.Sprintf("GdipCreateHBITMAPFromBitmap failed with status '%s' for file '%s'", status, filePath))
	}

	return newBitmapFromHBITMAP(hBmp, dpi)
}

// NewBitmapFromImage creates a Bitmap from image.Image at 96dpi.
//
// Deprecated: Newer applications should use NewBitmapFromImageForDPI.
func NewBitmapFromImage(im image.Image) (*Bitmap, error) {
	return NewBitmapFromImageForDPI(im, 96)
}

// NewBitmapFromImageForDPI creates a Bitmap from image.Image at given DPI.
func NewBitmapFromImageForDPI(im image.Image, dpi int) (*Bitmap, error) {
	hBmp, err := hBitmapFromImage(im, dpi)
	if err != nil {
		return nil, err
	}

	return newBitmapFromHBITMAP(hBmp, dpi)
}

// NewBitmapFromResource creates a Bitmap at 96dpi from resource by name.
//
// Deprecated: Newer applications should use NewBitmapFromResourceForDPI.
func NewBitmapFromResource(name string) (*Bitmap, error) {
	strPtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		newError(err.Error())
	}
	return newBitmapFromResource(strPtr, 96)
}

// NewBitmapFromResourceForDPI creates a Bitmap at given DPI from resource by name.
func NewBitmapFromResourceForDPI(name string, dpi int) (*Bitmap, error) {
	strPtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		newError(err.Error())
	}
	return newBitmapFromResource(strPtr, dpi)
}

// NewBitmapFromResourceId creates a Bitmap at 96dpi from resource by ID.
//
// Deprecated: Newer applications should use NewBitmapFromResourceIdForDPI.
func NewBitmapFromResourceId(id int) (*Bitmap, error) {
	return newBitmapFromResource(win.MAKEINTRESOURCE(uintptr(id)), 96)
}

// NewBitmapFromResourceIdForDPI creates a Bitmap at given DPI from resource by ID.
func NewBitmapFromResourceIdForDPI(id int, dpi int) (*Bitmap, error) {
	return newBitmapFromResource(win.MAKEINTRESOURCE(uintptr(id)), dpi)
}

func newBitmapFromResource(res *uint16, dpi int) (bm *Bitmap, err error) {
	hInst := kernel32.GetModuleHandle(nil)
	if hInst == 0 {
		err = lastError("GetModuleHandle")
		return
	}

	if hBmp := user32.LoadImage(hInst, res, user32.IMAGE_BITMAP, 0, 0, user32.LR_CREATEDIBSECTION); hBmp == 0 {
		err = lastError("LoadImage")
	} else {
		bm, err = newBitmapFromHBITMAP(gdi32.HBITMAP(hBmp), dpi)
	}

	return
}

// NewBitmapFromImageWithSize creates a bitmap with given size in native units and paints the image on it streched.
func NewBitmapFromImageWithSize(image Image, size Size) (*Bitmap, error) {
	var disposables Disposables
	defer disposables.Treat()

	dpi := int(math.Round(float64(size.Width) / float64(image.Size().Width) * 96.0))
	bmp, err := NewBitmapWithTransparentPixelsForDPI(size, dpi)
	if err != nil {
		return nil, err
	}
	disposables.Add(bmp)

	canvas, err := NewCanvasFromImage(bmp)
	if err != nil {
		return nil, err
	}
	defer canvas.Dispose()

	canvas.dpi = dpi

	if err := canvas.DrawImageStretchedPixels(image, Rectangle{0, 0, size.Width, size.Height}); err != nil {
		return nil, err
	}

	disposables.Spare()

	return bmp, nil
}

func NewBitmapFromWindow(window Window) (*Bitmap, error) {
	hBmp, err := hBitmapFromWindow(window)
	if err != nil {
		return nil, err
	}

	return newBitmapFromHBITMAP(hBmp, window.DPI())
}

// NewBitmapFromIcon creates a new bitmap with given size in native pixels and 96dpi and paints the
// icon on it.
//
// Deprecated: Newer applications should use NewBitmapFromIconForDPI.
func NewBitmapFromIcon(icon *Icon, size Size) (*Bitmap, error) {
	return NewBitmapFromIconForDPI(icon, size, 96)
}

// NewBitmapFromIconForDPI creates a new bitmap with given size in native pixels and DPI and paints
// the icon on it.
func NewBitmapFromIconForDPI(icon *Icon, size Size, dpi int) (*Bitmap, error) {
	hBmp, err := hBitmapFromIcon(icon, size, dpi)
	if err != nil {
		return nil, err
	}

	return newBitmapFromHBITMAP(hBmp, dpi)
}

func (bmp *Bitmap) ToImage() (*image.RGBA, error) {
	var bi gdi32.BITMAPINFO
	bi.BmiHeader.BiSize = uint32(unsafe.Sizeof(bi.BmiHeader))
	hdc := user32.GetDC(0)
	if ret := gdi32.GetDIBits(hdc, bmp.hBmp, 0, 0, nil, &bi, gdi32.DIB_RGB_COLORS); ret == 0 {
		return nil, newError("GetDIBits get bitmapinfo failed")
	}

	buf := make([]byte, bi.BmiHeader.BiSizeImage)
	bi.BmiHeader.BiCompression = gdi32.BI_RGB
	if ret := gdi32.GetDIBits(hdc, bmp.hBmp, 0, uint32(bi.BmiHeader.BiHeight), &buf[0], &bi, gdi32.DIB_RGB_COLORS); ret == 0 {
		return nil, newError("GetDIBits failed")
	}

	width := int(bi.BmiHeader.BiWidth)
	height := int(bi.BmiHeader.BiHeight)
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	n := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			a := buf[n+3]
			r := buf[n+2]
			g := buf[n+1]
			b := buf[n+0]
			n += int(bi.BmiHeader.BiBitCount) / 8
			img.Set(x, height-y-1, color.RGBA{r, g, b, a})
		}
	}

	return img, nil
}

func (bmp *Bitmap) hasTransparency() (bool, error) {
	if bmp.transparencyStatus == transparencyUnknown {
		if err := bmp.withPixels(func(bi *gdi32.BITMAPINFO, hdc gdi32.HDC, pixels *[maxPixels]bgraPixel, pixelsLen int) error {
			for i := 0; i < pixelsLen; i++ {
				if pixels[i].A == 0x00 {
					bmp.transparencyStatus = transparencyTransparent
					break
				}
			}

			return nil
		}); err != nil {
			return false, err
		}

		if bmp.transparencyStatus == transparencyUnknown {
			bmp.transparencyStatus = transparencyOpaque
		}
	}

	return bmp.transparencyStatus == transparencyTransparent, nil
}

func (bmp *Bitmap) postProcess() error {
	return bmp.withPixels(func(bi *gdi32.BITMAPINFO, hdc gdi32.HDC, pixels *[maxPixels]bgraPixel, pixelsLen int) error {
		for i := 0; i < pixelsLen; i++ {
			switch pixels[i].A {
			case 0x00:
				// The pixel has been drawn to by GDI, so we make it fully opaque.
				pixels[i].A = 0xff

			case 0x01:
				// The pixel has not been drawn to by GDI, so we make it fully transparent.
				pixels[i].A = 0x00
				bmp.transparencyStatus = transparencyTransparent
			}
		}

		if gdi32.SetDIBits(hdc, bmp.hBmp, 0, uint32(bi.BmiHeader.BiHeight), &pixels[0].B, bi, gdi32.DIB_RGB_COLORS) == 0 {
			return newError("SetDIBits")
		}

		return nil
	})
}

type bgraPixel struct {
	B byte
	G byte
	R byte
	A byte
}

const maxPixels = 2 << 27

func (bmp *Bitmap) withPixels(f func(bi *gdi32.BITMAPINFO, hdc gdi32.HDC, pixels *[maxPixels]bgraPixel, pixelsLen int) error) error {
	var bi gdi32.BITMAPINFO
	bi.BmiHeader.BiSize = uint32(unsafe.Sizeof(bi.BmiHeader))

	hdc := user32.GetDC(0)
	if hdc == 0 {
		return newError("GetDC")
	}
	defer user32.ReleaseDC(0, hdc)

	if ret := gdi32.GetDIBits(hdc, bmp.hBmp, 0, 0, nil, &bi, gdi32.DIB_RGB_COLORS); ret == 0 {
		return newError("GetDIBits #1")
	}

	hPixels := kernel32.GlobalAlloc(kernel32.GMEM_FIXED, uintptr(bi.BmiHeader.BiSizeImage))
	defer kernel32.GlobalFree(hPixels)

	pixels := (*[maxPixels]bgraPixel)(unsafe.Pointer(uintptr(hPixels)))

	bi.BmiHeader.BiCompression = gdi32.BI_RGB
	if ret := gdi32.GetDIBits(hdc, bmp.hBmp, 0, uint32(bi.BmiHeader.BiHeight), &pixels[0].B, &bi, gdi32.DIB_RGB_COLORS); ret == 0 {
		return newError("GetDIBits #2")
	}

	gdi32.GdiFlush()

	return f(&bi, hdc, pixels, int(bi.BmiHeader.BiSizeImage)/4)
}

func (bmp *Bitmap) Dispose() {
	if bmp.hBmp != 0 {
		gdi32.DeleteObject(gdi32.HGDIOBJ(bmp.hBmp))

		kernel32.GlobalUnlock(bmp.hPackedDIB)
		kernel32.GlobalFree(bmp.hPackedDIB)

		bmp.hPackedDIB = 0
		bmp.hBmp = 0
	}
}

// Size returns bitmap size in 1/96" units.
func (bmp *Bitmap) Size() Size {
	return SizeTo96DPI(bmp.size, bmp.dpi)
}

func (bmp *Bitmap) handle() gdi32.HBITMAP {
	return bmp.hBmp
}

func (bmp *Bitmap) draw(hdc gdi32.HDC, location Point) error {
	return bmp.drawStretched(hdc, Rectangle{X: location.X, Y: location.Y, Width: bmp.size.Width, Height: bmp.size.Height})
}

func (bmp *Bitmap) drawStretched(hdc gdi32.HDC, bounds Rectangle) error {
	return bmp.alphaBlend(hdc, bounds, 255)
}

// alphaBlend displays bitmaps that have transparent or semitransparent pixels. bounds is represented in native pixels.
func (bmp *Bitmap) alphaBlend(hdc gdi32.HDC, bounds Rectangle, opacity byte) error {
	return bmp.alphaBlendPart(hdc, bounds, Rectangle{0, 0, bmp.size.Width, bmp.size.Height}, opacity)
}

// alphaBlendPart displays bitmaps that have transparent or semitransparent pixels. dst and src are
// represented in native pixels.
func (bmp *Bitmap) alphaBlendPart(hdc gdi32.HDC, dst, src Rectangle, opacity byte) error {
	return bmp.withSelectedIntoMemDC(func(hdcMem gdi32.HDC) error {
		if opacity == 255 && (dst.Width != src.Width || dst.Height != src.Height) {
			transparent, err := bmp.hasTransparency()
			if err != nil {
				return err
			}

			if !transparent {
				if gdi32.SetStretchBltMode(hdc, gdi32.HALFTONE) == 0 {
					return newError("SetStretchBltMode")
				}

				if !gdi32.StretchBlt(
					hdc,
					int32(dst.X),
					int32(dst.Y),
					int32(dst.Width),
					int32(dst.Height),
					hdcMem,
					int32(src.X),
					int32(src.Y),
					int32(src.Width),
					int32(src.Height),
					gdi32.SRCCOPY,
				) {
					return newError("StretchBlt failed")
				}

				return nil
			}
		}

		if !gdi32.AlphaBlend(
			hdc,
			int32(dst.X),
			int32(dst.Y),
			int32(dst.Width),
			int32(dst.Height),
			hdcMem,
			int32(src.X),
			int32(src.Y),
			int32(src.Width),
			int32(src.Height),
			gdi32.BLENDFUNCTION{AlphaFormat: gdi32.AC_SRC_ALPHA, SourceConstantAlpha: opacity},
		) {
			return newError("AlphaBlend failed")
		}

		return nil
	})
}

func (bmp *Bitmap) withSelectedIntoMemDC(f func(hdcMem gdi32.HDC) error) error {
	return withCompatibleDC(func(hdcMem gdi32.HDC) error {
		hBmpOld := gdi32.SelectObject(hdcMem, gdi32.HGDIOBJ(bmp.hBmp))
		if hBmpOld == 0 {
			return newError("SelectObject failed")
		}
		defer gdi32.SelectObject(hdcMem, hBmpOld)

		return f(hdcMem)
	})
}

// newBitmapFromHBITMAP creates Bitmap from gdi32.HBITMAP.
//
// The BiXPelsPerMeter and BiYPelsPerMeter fields of gdi32.BITMAPINFOHEADER are unreliable (for
// loaded PNG they are both unset). Therefore, we require caller to specify DPI explicitly.
func newBitmapFromHBITMAP(hBmp gdi32.HBITMAP, dpi int) (bmp *Bitmap, err error) {
	var dib gdi32.DIBSECTION
	if gdi32.GetObject(gdi32.HGDIOBJ(hBmp), unsafe.Sizeof(dib), unsafe.Pointer(&dib)) == 0 {
		return nil, newError("GetObject failed")
	}

	bmih := &dib.DsBmih

	bmihSize := uintptr(unsafe.Sizeof(*bmih))
	pixelsSize := uintptr(int32(bmih.BiBitCount)*bmih.BiWidth*bmih.BiHeight) / 8

	totalSize := uintptr(bmihSize + pixelsSize)

	hPackedDIB := kernel32.GlobalAlloc(kernel32.GHND, totalSize)
	dest := kernel32.GlobalLock(hPackedDIB)
	defer kernel32.GlobalUnlock(hPackedDIB)

	src := unsafe.Pointer(&dib.DsBmih)

	kernel32.MoveMemory(dest, src, bmihSize)

	dest = unsafe.Pointer(uintptr(dest) + bmihSize)
	src = dib.DsBm.BmBits

	kernel32.MoveMemory(dest, src, pixelsSize)

	return &Bitmap{
		hBmp:       hBmp,
		hPackedDIB: hPackedDIB,
		size: Size{
			int(bmih.BiWidth),
			int(bmih.BiHeight),
		},
		dpi: dpi,
	}, nil
}

func hBitmapFromImage(im image.Image, dpi int) (gdi32.HBITMAP, error) {
	var bi gdi32.BITMAPV5HEADER
	bi.BiSize = uint32(unsafe.Sizeof(bi))
	bi.BiWidth = int32(im.Bounds().Dx())
	bi.BiHeight = -int32(im.Bounds().Dy())
	bi.BiPlanes = 1
	bi.BiBitCount = 32
	bi.BiCompression = gdi32.BI_BITFIELDS
	dpm := int32(math.Round(float64(dpi) * inchesPerMeter))
	bi.BiXPelsPerMeter = dpm
	bi.BiYPelsPerMeter = dpm
	// The following mask specification specifies a supported 32 BPP
	// alpha format for Windows XP.
	bi.BV4RedMask = 0x00FF0000
	bi.BV4GreenMask = 0x0000FF00
	bi.BV4BlueMask = 0x000000FF
	bi.BV4AlphaMask = 0xFF000000

	hdc := user32.GetDC(0)
	defer user32.ReleaseDC(0, hdc)

	var lpBits unsafe.Pointer

	// Create the DIB section with an alpha channel.
	hBitmap := gdi32.CreateDIBSection(hdc, &bi.BITMAPINFOHEADER, gdi32.DIB_RGB_COLORS, &lpBits, 0, 0)
	switch hBitmap {
	case 0, kernel32.ERROR_INVALID_PARAMETER:
		return 0, newError("CreateDIBSection failed")
	}

	// Fill the image
	bitmap_array := (*[1 << 30]byte)(unsafe.Pointer(lpBits))
	i := 0
	for y := im.Bounds().Min.Y; y != im.Bounds().Max.Y; y++ {
		for x := im.Bounds().Min.X; x != im.Bounds().Max.X; x++ {
			r, g, b, a := im.At(x, y).RGBA()
			bitmap_array[i+3] = byte(a >> 8)
			bitmap_array[i+2] = byte(r >> 8)
			bitmap_array[i+1] = byte(g >> 8)
			bitmap_array[i+0] = byte(b >> 8)
			i += 4
		}
	}

	return hBitmap, nil
}

func hBitmapFromWindow(window Window) (gdi32.HBITMAP, error) {
	hdcMem := gdi32.CreateCompatibleDC(0)
	if hdcMem == 0 {
		return 0, newError("CreateCompatibleDC failed")
	}
	defer gdi32.DeleteDC(hdcMem)

	var r gdi32.RECT
	if !user32.GetWindowRect(window.Handle(), &r) {
		return 0, newError("GetWindowRect failed")
	}

	hdc := user32.GetDC(window.Handle())
	width, height := r.Right-r.Left, r.Bottom-r.Top
	hBmp := gdi32.CreateCompatibleBitmap(hdc, width, height)
	user32.ReleaseDC(window.Handle(), hdc)

	hOld := gdi32.SelectObject(hdcMem, gdi32.HGDIOBJ(hBmp))
	flags := gdi32.PRF_CHILDREN | gdi32.PRF_CLIENT | gdi32.PRF_ERASEBKGND | gdi32.PRF_NONCLIENT | gdi32.PRF_OWNED
	window.SendMessage(user32.WM_PRINT, uintptr(hdcMem), uintptr(flags))

	gdi32.SelectObject(hdcMem, hOld)

	return hBmp, nil
}

// hBitmapFromIcon creates a new gdi32.HBITMAP with given size in native pixels and DPI, and paints
// the icon on it stretched.
func hBitmapFromIcon(icon *Icon, size Size, dpi int) (gdi32.HBITMAP, error) {
	hdc := user32.GetDC(0)
	defer user32.ReleaseDC(0, hdc)

	hdcMem := gdi32.CreateCompatibleDC(hdc)
	if hdcMem == 0 {
		return 0, newError("CreateCompatibleDC failed")
	}
	defer gdi32.DeleteDC(hdcMem)

	var bi gdi32.BITMAPV5HEADER
	bi.BiSize = uint32(unsafe.Sizeof(bi))
	bi.BiWidth = int32(size.Width)
	bi.BiHeight = int32(size.Height)
	bi.BiPlanes = 1
	bi.BiBitCount = 32
	bi.BiCompression = gdi32.BI_RGB
	dpm := int32(math.Round(float64(dpi) * inchesPerMeter))
	bi.BiXPelsPerMeter = dpm
	bi.BiYPelsPerMeter = dpm
	// The following mask specification specifies a supported 32 BPP
	// alpha format for Windows XP.
	bi.BV4RedMask = 0x00FF0000
	bi.BV4GreenMask = 0x0000FF00
	bi.BV4BlueMask = 0x000000FF
	bi.BV4AlphaMask = 0xFF000000

	hBmp := gdi32.CreateDIBSection(hdcMem, &bi.BITMAPINFOHEADER, gdi32.DIB_RGB_COLORS, nil, 0, 0)
	switch hBmp {
	case 0, kernel32.ERROR_INVALID_PARAMETER:
		return 0, newError("CreateDIBSection failed")
	}

	hOld := gdi32.SelectObject(hdcMem, gdi32.HGDIOBJ(hBmp))
	defer gdi32.SelectObject(hdcMem, hOld)

	err := icon.drawStretched(hdcMem, Rectangle{Width: size.Width, Height: size.Height})
	if err != nil {
		return 0, err
	}

	return hBmp, nil
}

func withCompatibleDC(f func(hdc gdi32.HDC) error) error {
	hdc := gdi32.CreateCompatibleDC(0)
	if hdc == 0 {
		return newError("CreateCompatibleDC failed")
	}
	defer gdi32.DeleteDC(hdc)

	return f(hdc)
}

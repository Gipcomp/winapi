// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"image"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/Gipcomp/win32/comctl32"
	"github.com/Gipcomp/win32/gdi32"
	"github.com/Gipcomp/win32/kernel32"
	"github.com/Gipcomp/win32/shell32"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/win"
	"golang.org/x/sys/windows"
)

// Icon is a bitmap that supports transparency and combining multiple
// variants of an image in different resolutions.
type Icon struct {
	filePath  string
	index     int
	res       *uint16
	dpi2hIcon map[int]user32.HICON
	size96dpi Size
	isStock   bool
	hasIndex  bool
}

type ExtractableIcon interface {
	FilePath_() string
	Index_() int
	Size_() int
}

func IconFrom(src interface{}, dpi int) (*Icon, error) {
	if src == nil {
		return nil, nil
	}

	img, err := ImageFrom(src)
	if err != nil {
		return nil, err
	}

	return iconCache.Icon(img, dpi)
}

func IconApplication() *Icon {
	return stockIcon(user32.IDI_APPLICATION)
}

func IconError() *Icon {
	return stockIcon(user32.IDI_ERROR)
}

func IconQuestion() *Icon {
	return stockIcon(user32.IDI_QUESTION)
}

func IconWarning() *Icon {
	return stockIcon(user32.IDI_WARNING)
}

func IconInformation() *Icon {
	return stockIcon(user32.IDI_INFORMATION)
}

func IconWinLogo() *Icon {
	return stockIcon(user32.IDI_WINLOGO)
}

func IconShield() *Icon {
	return stockIcon(user32.IDI_SHIELD)
}

func stockIcon(id uintptr) *Icon {
	return &Icon{res: win.MAKEINTRESOURCE(id), size96dpi: defaultIconSize(), isStock: true}
}

// NewIconFromFile returns a new Icon, using the specified icon image file and default size.
func NewIconFromFile(filePath string) (*Icon, error) {
	return NewIconFromFileWithSize(filePath, Size{})
}

// NewIconFromFileWithSize returns a new Icon, using the specified icon image file and size.
func NewIconFromFileWithSize(filePath string, size Size) (*Icon, error) {
	if size.Width == 0 || size.Height == 0 {
		size = defaultIconSize()
	}

	return checkNewIcon(&Icon{filePath: filePath, size96dpi: size})
}

// NewIconFromResource returns a new Icon of default size, using the specified icon resource.
func NewIconFromResource(name string) (*Icon, error) {
	return NewIconFromResourceWithSize(name, Size{})
}

// NewIconFromResourceWithSize returns a new Icon of size size, using the specified icon resource.
func NewIconFromResourceWithSize(name string, size Size) (*Icon, error) {
	strPtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		newError(err.Error())
	}
	return newIconFromResource(strPtr, size)
}

// NewIconFromResourceId returns a new Icon of default size, using the specified icon resource.
func NewIconFromResourceId(id int) (*Icon, error) {
	return NewIconFromResourceIdWithSize(id, Size{})
}

// NewIconFromResourceIdWithSize returns a new Icon of size size, using the specified icon resource.
func NewIconFromResourceIdWithSize(id int, size Size) (*Icon, error) {
	return newIconFromResource(win.MAKEINTRESOURCE(uintptr(id)), size)
}

func newIconFromResource(res *uint16, size Size) (*Icon, error) {
	if size.Width == 0 || size.Height == 0 {
		size = defaultIconSize()
	}

	return checkNewIcon(&Icon{res: res, size96dpi: size})
}

// NewIconFromSysDLL returns a new Icon, as identified by index of
// size 16x16 from the system DLL identified by dllBaseName.
func NewIconFromSysDLL(dllBaseName string, index int) (*Icon, error) {
	return NewIconFromSysDLLWithSize(dllBaseName, index, 16)
}

// NewIconFromSysDLLWithSize returns a new Icon, as identified by
// index of the desired size from the system DLL identified by dllBaseName.
func NewIconFromSysDLLWithSize(dllBaseName string, index, size int) (*Icon, error) {
	system32, err := windows.GetSystemDirectory()
	if err != nil {
		return nil, err
	}

	return checkNewIcon(&Icon{filePath: filepath.Join(system32, dllBaseName+".dll"), index: index, hasIndex: true, size96dpi: Size{size, size}})
}

// NewIconExtractedFromFile returns a new Icon, as identified by index of size 16x16 from filePath.
func NewIconExtractedFromFile(filePath string, index, _ int) (*Icon, error) {
	return checkNewIcon(&Icon{filePath: filePath, index: index, hasIndex: true, size96dpi: Size{16, 16}})
}

// NewIconExtractedFromFileWithSize returns a new Icon, as identified by index of the desired size from filePath.
func NewIconExtractedFromFileWithSize(filePath string, index, size int) (*Icon, error) {
	return checkNewIcon(&Icon{filePath: filePath, index: index, hasIndex: true, size96dpi: Size{size, size}})
}

// NewIconFromImage returns a new Icon at 96dpi, using the specified image.Image as source.
//
// Deprecated: Newer applications should use NewIconFromImageForDPI.
func NewIconFromImage(im image.Image) (ic *Icon, err error) {
	return NewIconFromImageForDPI(im, 96)
}

// NewIconFromImageForDPI returns a new Icon at given DPI, using the specified image.Image as source.
func NewIconFromImageForDPI(im image.Image, dpi int) (ic *Icon, err error) {
	hIcon, err := createAlphaCursorOrIconFromImage(im, image.Pt(0, 0), true)
	if err != nil {
		return nil, err
	}
	b := im.Bounds()
	return newIconFromHICONAndSize(hIcon, SizeTo96DPI(Size{b.Dx(), b.Dy()}, dpi), dpi), nil
}

// NewIconFromImageWithSize returns a new Icon of the given size in native pixels, using the
// specified Image as source.
func NewIconFromImageWithSize(image Image, size Size) (*Icon, error) {
	bmp, err := NewBitmapFromImageWithSize(image, size)
	if err != nil {
		return nil, err
	}

	return NewIconFromBitmap(bmp)
}

func newIconFromImageForDPI(image Image, dpi int) (*Icon, error) {
	size96dpi := image.Size()
	size := SizeFrom96DPI(size96dpi, dpi)

	bmp, err := NewBitmapFromImageWithSize(image, size)
	if err != nil {
		return nil, err
	}

	hIcon, err := createAlphaCursorOrIconFromBitmap(bmp, Point{}, true)
	if err != nil {
		return nil, err
	}

	return &Icon{dpi2hIcon: map[int]user32.HICON{dpi: hIcon}, size96dpi: size96dpi}, nil
}

// NewIconFromBitmap returns a new Icon, using the specified Bitmap as source.
func NewIconFromBitmap(bmp *Bitmap) (ic *Icon, err error) {
	hIcon, err := createAlphaCursorOrIconFromBitmap(bmp, Point{}, true)
	if err != nil {
		return nil, err
	}
	return newIconFromHICONAndSize(hIcon, bmp.Size(), bmp.dpi), nil
}

// NewIconFromHICON returns a new Icon at 96dpi, using the specified user32.HICON as source.
//
// Deprecated: Newer applications should use NewIconFromHICONForDPI.
func NewIconFromHICON(hIcon user32.HICON) (ic *Icon, err error) {
	return NewIconFromHICONForDPI(hIcon, 96)
}

// NewIconFromHICONForDPI returns a new Icon at given DPI, using the specified user32.HICON as source.
func NewIconFromHICONForDPI(hIcon user32.HICON, dpi int) (ic *Icon, err error) {
	s, err := sizeFromHICON(hIcon)
	if err != nil {
		return nil, err
	}

	return newIconFromHICONAndSize(hIcon, SizeTo96DPI(s, dpi), dpi), nil
}

func newIconFromHICONAndSize(hIcon user32.HICON, size Size, dpi int) *Icon {
	return &Icon{dpi2hIcon: map[int]user32.HICON{dpi: hIcon}, size96dpi: size}
}

func checkNewIcon(icon *Icon) (*Icon, error) {
	if _, err := icon.handleForDPIWithError(96); err != nil {
		return nil, err
	}

	return icon, nil
}

func (i *Icon) handleForDPI(dpi int) user32.HICON {
	hIcon, _ := i.handleForDPIWithError(dpi)
	return hIcon
}

func (i *Icon) handleForDPIWithError(dpi int) (user32.HICON, error) {
	if i.dpi2hIcon == nil {
		i.dpi2hIcon = make(map[int]user32.HICON)
	} else if handle, ok := i.dpi2hIcon[dpi]; ok {
		return handle, nil
	}

	var hInst kernel32.HINSTANCE
	var name *uint16
	if i.filePath != "" {
		absFilePath, err := filepath.Abs(i.filePath)
		if err != nil {
			return 0, err
		}
		strPtr, err := syscall.UTF16PtrFromString(absFilePath)
		if err != nil {
			newError(err.Error())
		}
		name = strPtr
	} else {
		if !i.isStock {
			if hInst = kernel32.GetModuleHandle(nil); hInst == 0 {
				return 0, lastError("GetModuleHandle")
			}
		}

		name = i.res
	}

	var size Size
	if i.size96dpi.Width == 0 || i.size96dpi.Height == 0 {
		size = SizeFrom96DPI(defaultIconSize(), dpi)
	} else {
		size = SizeFrom96DPI(i.size96dpi, dpi)
	}

	var hIcon user32.HICON

	if i.hasIndex {
		shell32.SHDefExtractIcon(
			name,
			int32(i.index),
			0,
			nil,
			&hIcon,
			win.MAKELONG(0, uint16(size.Width)))
		if hIcon == 0 {
			return 0, newError("SHDefExtractIcon")
		}
	} else {
		hr := user32.HICON(comctl32.LoadIconWithScaleDown(
			hInst,
			name,
			int32(size.Width),
			int32(size.Height),
			&hIcon))

		if hr < 0 || hIcon == 0 {
			return 0, lastError("LoadIconWithScaleDown")
		}
	}

	i.dpi2hIcon[dpi] = hIcon

	return hIcon, nil
}

// Dispose releases the operating system resources associated with the Icon.
func (i *Icon) Dispose() {
	if i.isStock || len(i.dpi2hIcon) == 0 {
		return
	}

	for dpi, hIcon := range i.dpi2hIcon {
		user32.DestroyIcon(hIcon)
		delete(i.dpi2hIcon, dpi)
	}
}

func (i *Icon) draw(hdc gdi32.HDC, location Point) error {
	dpi := dpiForHDC(hdc)
	size := SizeFrom96DPI(i.size96dpi, dpi)

	return i.drawStretched(hdc, Rectangle{location.X, location.Y, size.Width, size.Height})
}

func (i *Icon) drawStretched(hdc gdi32.HDC, bounds Rectangle) error {
	dpi := int(float64(bounds.Width) / float64(i.size96dpi.Width) * 96.0)

	hIcon := i.handleForDPI(dpi)
	if hIcon == 0 {
		var dpiAvailMax int
		for dpiAvail, handle := range i.dpi2hIcon {
			if dpiAvail > dpiAvailMax {
				hIcon = handle
				dpiAvailMax = dpiAvail
			}
			if dpiAvail > dpi {
				break
			}
		}
	}

	if !user32.DrawIconEx(hdc, int32(bounds.X), int32(bounds.Y), hIcon, int32(bounds.Width), int32(bounds.Height), 0, 0, user32.DI_NORMAL) {
		return lastError("DrawIconEx")
	}

	return nil
}

// Size returns icon size in 1/96" units.
func (i *Icon) Size() Size {
	return i.size96dpi
}

// create an Alpha Icon or Cursor from an Image
// http://support.microsoft.com/kb/318876
func createAlphaCursorOrIconFromImage(im image.Image, hotspot image.Point, fIcon bool) (user32.HICON, error) {
	bmp, err := NewBitmapFromImage(im)
	if err != nil {
		return 0, err
	}
	defer bmp.Dispose()

	return createAlphaCursorOrIconFromBitmap(bmp, Point{hotspot.X, hotspot.Y}, fIcon)
}

// createAlphaCursorOrIconFromBitmap creates a cursor/icon from a bitmap. hotspot coordinates are in native pixels.
func createAlphaCursorOrIconFromBitmap(bmp *Bitmap, hotspot Point, fIcon bool) (user32.HICON, error) {
	// Create an empty mask bitmap.
	hMonoBitmap := gdi32.CreateBitmap(int32(bmp.size.Width), int32(bmp.size.Height), 1, 1, nil)
	if hMonoBitmap == 0 {
		return 0, newError("CreateBitmap failed")
	}
	defer gdi32.DeleteObject(gdi32.HGDIOBJ(hMonoBitmap))

	var ii user32.ICONINFO
	if fIcon {
		ii.FIcon = win.TRUE
	}
	ii.XHotspot = uint32(hotspot.X)
	ii.YHotspot = uint32(hotspot.Y)
	ii.HbmMask = hMonoBitmap
	ii.HbmColor = bmp.hBmp

	// Create the alpha cursor with the alpha DIB section.
	hIconOrCursor := user32.CreateIconIndirect(&ii)

	return hIconOrCursor, nil
}

// sizeFromHICON returns icon size in native pixels.
func sizeFromHICON(hIcon user32.HICON) (Size, error) {
	var ii user32.ICONINFO
	var bi gdi32.BITMAPINFO

	if !user32.GetIconInfo(hIcon, &ii) {
		return Size{}, lastError("GetIconInfo")
	}
	defer gdi32.DeleteObject(gdi32.HGDIOBJ(ii.HbmMask))

	var hBmp gdi32.HBITMAP
	if ii.HbmColor != 0 {
		hBmp = ii.HbmColor

		defer gdi32.DeleteObject(gdi32.HGDIOBJ(ii.HbmColor))
	} else {
		hBmp = ii.HbmMask
	}

	if 0 == gdi32.GetObject(gdi32.HGDIOBJ(hBmp), unsafe.Sizeof(bi), unsafe.Pointer(&bi)) {
		return Size{}, newError("GetObject")
	}

	return Size{int(bi.BmiHeader.BiWidth), int(bi.BmiHeader.BiHeight)}, nil
}

// defaultIconSize returns default small icon size in 1/92" units.
func defaultIconSize() Size {
	return Size{int(user32.GetSystemMetricsForDpi(user32.SM_CXSMICON, 96)), int(user32.GetSystemMetricsForDpi(user32.SM_CYSMICON, 96))}
}

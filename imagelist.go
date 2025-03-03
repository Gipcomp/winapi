// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"syscall"
	"unsafe"

	"github.com/Gipcomp/win32/comctl32"
	"github.com/Gipcomp/win32/gdi32"
	"github.com/Gipcomp/win32/shell32"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/winapi/errs"
)

type ImageList struct {
	hIml                     comctl32.HIMAGELIST
	dpi                      int
	maskColor                Color
	imageSize96dpi           Size
	colorMaskedBitmap2Index  map[*Bitmap]int
	bitmapMaskedBitmap2Index map[bitmapMaskedBitmap]int
	icon2Index               map[*Icon]int32
}

type bitmapMaskedBitmap struct {
	bitmap *Bitmap
	mask   *Bitmap
}

// NewImageList creates an empty image list at 96dpi. imageSize parameter is specified in 1/96"
// units.
//
// Deprecated: Newer applications should use NewImageListForDPI.
func NewImageList(imageSize Size, maskColor Color) (*ImageList, error) {
	return NewImageListForDPI(SizeFrom96DPI(imageSize, 96), maskColor, 96)
}

// NewImageListForDPI creates an empty image list for image size at given DPI. imageSize is
// specified in native pixels.
func NewImageListForDPI(imageSize Size, maskColor Color, dpi int) (*ImageList, error) {
	hIml := comctl32.ImageList_Create(
		int32(imageSize.Width),
		int32(imageSize.Height),
		comctl32.ILC_MASK|comctl32.ILC_COLOR32,
		8,
		8)
	if hIml == 0 {
		return nil, errs.NewError("ImageList_Create failed")
	}

	return &ImageList{
		hIml:                     hIml,
		dpi:                      dpi,
		maskColor:                maskColor,
		imageSize96dpi:           SizeTo96DPI(imageSize, dpi),
		colorMaskedBitmap2Index:  make(map[*Bitmap]int),
		bitmapMaskedBitmap2Index: make(map[bitmapMaskedBitmap]int),
		icon2Index:               make(map[*Icon]int32),
	}, nil
}

func (il *ImageList) Handle() comctl32.HIMAGELIST {
	return il.hIml
}

func (il *ImageList) Add(bitmap, maskBitmap *Bitmap) (int, error) {
	if bitmap == nil {
		return 0, errs.NewError("bitmap cannot be nil")
	}

	key := bitmapMaskedBitmap{bitmap: bitmap, mask: maskBitmap}

	if index, ok := il.bitmapMaskedBitmap2Index[key]; ok {
		return index, nil
	}

	var maskHandle gdi32.HBITMAP
	if maskBitmap != nil {
		maskHandle = maskBitmap.handle()
	}

	index := int(comctl32.ImageList_Add(il.hIml, bitmap.handle(), maskHandle))
	if index == -1 {
		return 0, errs.NewError("ImageList_Add failed")
	}

	il.bitmapMaskedBitmap2Index[key] = index

	return index, nil
}

func (il *ImageList) AddMasked(bitmap *Bitmap) (int32, error) {
	if bitmap == nil {
		return 0, errs.NewError("bitmap cannot be nil")
	}

	if index, ok := il.colorMaskedBitmap2Index[bitmap]; ok {
		return int32(index), nil
	}

	index := comctl32.ImageList_AddMasked(
		il.hIml,
		bitmap.handle(),
		gdi32.COLORREF(il.maskColor))
	if index == -1 {
		return 0, errs.NewError("ImageList_AddMasked failed")
	}

	il.colorMaskedBitmap2Index[bitmap] = int(index)

	return index, nil
}

func (il *ImageList) AddIcon(icon *Icon) (int32, error) {
	if icon == nil {
		return 0, errs.NewError("icon cannot be nil")
	}

	if index, ok := il.icon2Index[icon]; ok {
		return index, nil
	}

	index := comctl32.ImageList_ReplaceIcon(il.hIml, -1, icon.handleForDPI(il.dpi))
	if index == -1 {
		return 0, errs.NewError("ImageList_ReplaceIcon failed")
	}

	il.icon2Index[icon] = index

	return index, nil
}

func (il *ImageList) AddImage(image interface{}) (int32, error) {
	switch image.(type) {
	case ExtractableIcon, *Icon:
		icon, err := IconFrom(image, il.dpi)
		if err != nil {
			return 0, err
		}

		return il.AddIcon(icon)

	default:
		bmp, err := BitmapFrom(image, il.dpi)
		if err != nil {
			return 0, err
		}

		return il.AddMasked(bmp)
	}
}

func (il *ImageList) DrawPixels(canvas *Canvas, index int, bounds Rectangle) error {
	if !comctl32.ImageList_DrawEx(il.hIml, int32(index), canvas.hdc, int32(bounds.X), int32(bounds.Y), int32(bounds.Width), int32(bounds.Height), gdi32.CLR_DEFAULT, gdi32.CLR_DEFAULT, comctl32.ILD_NORMAL) {
		return errs.NewError("ImageList_DrawEx")
	}

	return nil
}

func (il *ImageList) Dispose() {
	if il.hIml != 0 {
		comctl32.ImageList_Destroy(il.hIml)
		il.hIml = 0
	}
}

func (il *ImageList) MaskColor() Color {
	return il.maskColor
}

func imageListForImage(image interface{}, dpi int) (hIml comctl32.HIMAGELIST, isSysIml bool, err error) {
	if name, ok := image.(string); ok {
		if img, err := Resources.Image(name); err == nil {
			image = img
		}
	}

	if filePath, ok := image.(string); ok {
		_, hIml = iconIndexAndHImlForFilePath(filePath)
		isSysIml = hIml != 0
	} else {
		w := int32(user32.GetSystemMetricsForDpi(user32.SM_CXSMICON, uint32(dpi)))
		h := int32(user32.GetSystemMetricsForDpi(user32.SM_CYSMICON, uint32(dpi)))

		hIml = comctl32.ImageList_Create(w, h, comctl32.ILC_MASK|comctl32.ILC_COLOR32, 8, 8)
		if hIml == 0 {
			return 0, false, errs.NewError("ImageList_Create failed")
		}
	}

	return
}

func iconIndexAndHImlForFilePath(filePath string) (int32, comctl32.HIMAGELIST) {
	var shfi shell32.SHFILEINFO
	strPtr, err := syscall.UTF16PtrFromString(filePath)
	if err != nil {
		errs.NewError(err.Error())
	}
	if hIml := comctl32.HIMAGELIST(shell32.SHGetFileInfo(
		strPtr,
		0,
		&shfi,
		uint32(unsafe.Sizeof(shfi)),
		shell32.SHGFI_SYSICONINDEX|shell32.SHGFI_SMALLICON)); hIml != 0 {

		return shfi.IIcon, hIml
	}

	return -1, 0
}

func imageIndexMaybeAdd(image interface{}, hIml comctl32.HIMAGELIST, isSysIml bool, imageUintptr2Index map[uintptr]int32, filePath2IconIndex map[string]int32, dpi int) int32 {
	if !isSysIml {
		return imageIndexAddIfNotExists(image, hIml, imageUintptr2Index, dpi)
	} else if filePath, ok := image.(string); ok {
		if iIcon, ok := filePath2IconIndex[filePath]; ok {
			return iIcon
		}

		if iIcon, _ := iconIndexAndHImlForFilePath(filePath); iIcon != -1 {
			filePath2IconIndex[filePath] = iIcon
			return iIcon
		}
	}

	return -1
}

func imageIndexAddIfNotExists(image interface{}, hIml comctl32.HIMAGELIST, imageUintptr2Index map[uintptr]int32, dpi int) int32 {
	imageIndex := int32(-1)

	if image != nil {
		if name, ok := image.(string); ok {
			image, _ = Resources.Image(name)
		}

		var ptr uintptr
		switch img := image.(type) {
		case *Bitmap:
			ptr = uintptr(unsafe.Pointer(img))

		case *Icon:
			ptr = uintptr(unsafe.Pointer(img))
		}

		if ptr == 0 {
			return -1
		}

		if imageIndex, ok := imageUintptr2Index[ptr]; ok {
			return imageIndex
		}

		switch img := image.(type) {
		case *Bitmap:
			imageIndex = comctl32.ImageList_AddMasked(hIml, img.hBmp, 0)

		case *Icon:
			imageIndex = comctl32.ImageList_ReplaceIcon(hIml, -1, img.handleForDPI(dpi))
		}

		if imageIndex > -1 {
			imageUintptr2Index[ptr] = imageIndex
		}
	}

	return imageIndex
}

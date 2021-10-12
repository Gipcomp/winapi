// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"syscall"

	"github.com/Gipcomp/win32/gdi32"
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/kernel32"
	"github.com/Gipcomp/win32/win"
	"github.com/Gipcomp/winapi/errs"
)

// FontMemResource represents a font resource loaded into memory from
// the application's resources.
type FontMemResource struct {
	hFontResource handle.HANDLE
}

func newFontMemResource(resourceName *uint16) (*FontMemResource, error) {
	hModule := kernel32.HMODULE(kernel32.GetModuleHandle(nil))
	if hModule == kernel32.HMODULE(0) {
		return nil, errs.LastError("GetModuleHandle")
	}

	hres := kernel32.FindResource(hModule, resourceName, win.MAKEINTRESOURCE(8) /*RT_FONT*/)
	if hres == kernel32.HRSRC(0) {
		return nil, errs.LastError("FindResource")
	}

	size := kernel32.SizeofResource(hModule, hres)
	if size == 0 {
		return nil, errs.LastError("SizeofResource")
	}

	hResLoad := kernel32.LoadResource(hModule, hres)
	if hResLoad == kernel32.HGLOBAL(0) {
		return nil, errs.LastError("LoadResource")
	}

	ptr := kernel32.LockResource(hResLoad)
	if ptr == 0 {
		return nil, errs.LastError("LockResource")
	}

	numFonts := uint32(0)
	hFontResource := gdi32.AddFontMemResourceEx(ptr, size, nil, &numFonts)

	if hFontResource == handle.HANDLE(0) || numFonts == 0 {
		return nil, errs.LastError("AddFontMemResource")
	}

	return &FontMemResource{hFontResource: hFontResource}, nil
}

// NewFontMemResourceByName function loads a font resource from the executable's resources
// using the resource name.
// The font must be embedded into resources using corresponding operator in the
// application's RC script.
func NewFontMemResourceByName(name string) (*FontMemResource, error) {
	lpstr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return nil, err
	}

	return newFontMemResource(lpstr)
}

// NewFontMemResourceById function loads a font resource from the executable's resources
// using the resource ID.
// The font must be embedded into resources using corresponding operator in the
// application's RC script.
func NewFontMemResourceById(id int) (*FontMemResource, error) {
	return newFontMemResource(win.MAKEINTRESOURCE(uintptr(id)))
}

// Dispose removes the font resource from memory
func (fmr *FontMemResource) Dispose() {
	if fmr.hFontResource != 0 {
		gdi32.RemoveFontMemResourceEx(fmr.hFontResource)
		fmr.hFontResource = 0
	}
}

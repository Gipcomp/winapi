// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/Gipcomp/win32/comdlg32"
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/kernel32"
	"github.com/Gipcomp/win32/ole32"
	"github.com/Gipcomp/win32/shell32"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/win"
)

type FileDialog struct {
	Title          string
	FilePath       string
	FilePaths      []string
	InitialDirPath string
	Filter         string
	FilterIndex    int
	Flags          uint32
	ShowReadOnlyCB bool
}

func (dlg *FileDialog) show(owner Form, fun func(ofn *comdlg32.OPENFILENAME) bool, flags uint32) (accepted bool, err error) {
	ofn := new(comdlg32.OPENFILENAME)

	ofn.LStructSize = uint32(unsafe.Sizeof(*ofn))
	if owner != nil {
		ofn.HwndOwner = owner.Handle()
	}

	filter := make([]uint16, len(dlg.Filter)+2)
	strUtf, err := syscall.UTF16FromString(dlg.Filter)
	if err != nil {
		newError(err.Error())
	}
	copy(filter, strUtf)
	// Replace '|' with the expected '\0'.
	for i, c := range filter {
		if byte(c) == '|' {
			filter[i] = uint16(0)
		}
	}
	ofn.LpstrFilter = &filter[0]
	ofn.NFilterIndex = uint32(dlg.FilterIndex)

	ofn.LpstrInitialDir, err = syscall.UTF16PtrFromString(dlg.InitialDirPath)
	if err != nil {
		newError(err.Error())
	}
	ofn.LpstrTitle, err = syscall.UTF16PtrFromString(dlg.Title)
	if err != nil {
		newError(err.Error())
	}
	ofn.Flags = comdlg32.OFN_FILEMUSTEXIST | flags | dlg.Flags

	if !dlg.ShowReadOnlyCB {
		ofn.Flags |= comdlg32.OFN_HIDEREADONLY
	}

	var fileBuf []uint16
	if flags&comdlg32.OFN_ALLOWMULTISELECT > 0 {
		fileBuf = make([]uint16, 65536)
	} else {
		fileBuf = make([]uint16, 1024)
		strUtf, err := syscall.UTF16FromString(dlg.FilePath)
		if err != nil {
			newError(err.Error())
		}
		copy(fileBuf, strUtf)
	}
	ofn.LpstrFile = &fileBuf[0]
	ofn.NMaxFile = uint32(len(fileBuf))

	if !fun(ofn) {
		errno := comdlg32.CommDlgExtendedError()
		if errno != 0 {
			err = newError(fmt.Sprintf("Error %d", errno))
		}
		return
	}

	dlg.FilterIndex = int(ofn.NFilterIndex)

	if flags&comdlg32.OFN_ALLOWMULTISELECT > 0 {
		split := func() [][]uint16 {
			var parts [][]uint16

			from := 0
			for i, c := range fileBuf {
				if c == 0 {
					if i == from {
						return parts
					}

					parts = append(parts, fileBuf[from:i])
					from = i + 1
				}
			}

			return parts
		}

		parts := split()

		if len(parts) == 1 {
			dlg.FilePaths = []string{syscall.UTF16ToString(parts[0])}
		} else {
			dirPath := syscall.UTF16ToString(parts[0])
			dlg.FilePaths = make([]string, len(parts)-1)

			for i, fp := range parts[1:] {
				dlg.FilePaths[i] = filepath.Join(dirPath, syscall.UTF16ToString(fp))
			}
		}
	} else {
		dlg.FilePath = syscall.UTF16ToString(fileBuf)
	}

	accepted = true

	return
}

func (dlg *FileDialog) ShowOpen(owner Form) (accepted bool, err error) {
	return dlg.show(owner, comdlg32.GetOpenFileName, comdlg32.OFN_NOCHANGEDIR)
}

func (dlg *FileDialog) ShowOpenMultiple(owner Form) (accepted bool, err error) {
	return dlg.show(owner, comdlg32.GetOpenFileName, comdlg32.OFN_ALLOWMULTISELECT|comdlg32.OFN_EXPLORER|comdlg32.OFN_NOCHANGEDIR)
}

func (dlg *FileDialog) ShowSave(owner Form) (accepted bool, err error) {
	return dlg.show(owner, comdlg32.GetSaveFileName, comdlg32.OFN_NOCHANGEDIR)
}

func pathFromPIDL(pidl uintptr) (string, error) {
	var path [kernel32.MAX_PATH]uint16
	if !shell32.SHGetPathFromIDList(pidl, &path[0]) {
		return "", newError("SHGetPathFromIDList failed")
	}

	return syscall.UTF16ToString(path[:]), nil
}

// We use this callback to disable the OK button in case of "invalid" selections.
func browseFolderCallback(hwnd handle.HWND, msg uint32, lp, wp uintptr) uintptr {
	const BFFM_SELCHANGED = 2
	if msg == BFFM_SELCHANGED {
		_, err := pathFromPIDL(lp)
		var enabled uintptr
		if err == nil {
			enabled = 1
		}

		const BFFM_ENABLEOK = user32.WM_USER + 101

		user32.SendMessage(hwnd, BFFM_ENABLEOK, 0, enabled)
	}

	return 0
}

var browseFolderCallbackPtr uintptr

func init() {
	AppendToWalkInit(func() {
		browseFolderCallbackPtr = syscall.NewCallback(browseFolderCallback)
	})
}

func (dlg *FileDialog) ShowBrowseFolder(owner Form) (accepted bool, err error) {
	// Calling OleInitialize (or similar) is required for BIF_NEWDIALOGSTYLE.
	if hr := ole32.OleInitialize(); hr != win.S_OK && hr != win.S_FALSE {
		return false, newError(fmt.Sprint("OleInitialize Error: ", hr))
	}
	defer ole32.OleUninitialize()

	var ownerHwnd handle.HWND
	if owner != nil {
		ownerHwnd = owner.Handle()
	}

	// We need to put the initial path into a buffer of at least MAX_LENGTH
	// length, or we may get random crashes.
	var buf [kernel32.MAX_PATH]uint16
	strUtf, err := syscall.UTF16FromString(dlg.InitialDirPath)
	if err != nil {
		newError(err.Error())
	}
	copy(buf[:], strUtf)

	const BIF_NEWDIALOGSTYLE = 0x00000040
	strPtr, err := syscall.UTF16PtrFromString(dlg.Title)
	if err != nil {
		newError(err.Error())
	}
	bi := shell32.BROWSEINFO{
		HwndOwner: ownerHwnd,
		LpszTitle: strPtr,
		UlFlags:   BIF_NEWDIALOGSTYLE,
		Lpfn:      browseFolderCallbackPtr,
	}

	shell32.SHParseDisplayName(&buf[0], 0, &bi.PidlRoot, 0, nil)

	pidl := shell32.SHBrowseForFolder(&bi)
	if pidl == 0 {
		return false, nil
	}
	defer ole32.CoTaskMemFree(pidl)

	dlg.FilePath, err = pathFromPIDL(pidl)
	accepted = dlg.FilePath != ""
	return
}

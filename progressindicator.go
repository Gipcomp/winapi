// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"syscall"
	"unsafe"

	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/ole32"
	"github.com/Gipcomp/win32/shobj"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/win"
)

type ProgressIndicator struct {
	hwnd                   handle.HWND
	taskbarList3           *shobj.ITaskbarList3
	completed              uint32
	total                  uint32
	state                  PIState
	overlayIcon            *Icon
	overlayIconDescription string
}

type PIState int

const (
	PINoProgress    PIState = shobj.TBPF_NOPROGRESS
	PIIndeterminate PIState = shobj.TBPF_INDETERMINATE
	PINormal        PIState = shobj.TBPF_NORMAL
	PIError         PIState = shobj.TBPF_ERROR
	PIPaused        PIState = shobj.TBPF_PAUSED
)

//newTaskbarList3 precondition: Windows version is at least 6.1 (yes, Win 7 is version 6.1).
func newTaskbarList3(hwnd handle.HWND) (*ProgressIndicator, error) {
	var classFactoryPtr unsafe.Pointer
	if hr := ole32.CoGetClassObject(&shobj.CLSID_TaskbarList, ole32.CLSCTX_ALL, nil, &ole32.IID_IClassFactory, &classFactoryPtr); win.FAILED(hr) {
		return nil, errorFromHRESULT("CoGetClassObject", hr)
	}

	var taskbarList3ObjectPtr unsafe.Pointer
	classFactory := (*ole32.IClassFactory)(classFactoryPtr)
	defer classFactory.Release()

	if hr := classFactory.CreateInstance(nil, &shobj.IID_ITaskbarList3, &taskbarList3ObjectPtr); win.FAILED(hr) {
		return nil, errorFromHRESULT("IClassFactory.CreateInstance", hr)
	}

	return &ProgressIndicator{taskbarList3: (*shobj.ITaskbarList3)(taskbarList3ObjectPtr), hwnd: hwnd}, nil
}

func (pi *ProgressIndicator) SetState(state PIState) error {
	if hr := pi.taskbarList3.SetProgressState(pi.hwnd, (int)(state)); win.FAILED(hr) {
		return errorFromHRESULT("ITaskbarList3.setprogressState", hr)
	}
	pi.state = state
	return nil
}

func (pi *ProgressIndicator) State() PIState {
	return pi.state
}

func (pi *ProgressIndicator) SetTotal(total uint32) {
	pi.total = total
}

func (pi *ProgressIndicator) Total() uint32 {
	return pi.total
}

func (pi *ProgressIndicator) SetCompleted(completed uint32) error {
	if hr := pi.taskbarList3.SetProgressValue(pi.hwnd, completed, pi.total); win.FAILED(hr) {
		return errorFromHRESULT("ITaskbarList3.SetProgressValue", hr)
	}
	pi.completed = completed
	return nil
}

func (pi *ProgressIndicator) Completed() uint32 {
	return pi.completed
}

func (pi *ProgressIndicator) SetOverlayIcon(icon *Icon, description string) error {
	handle := user32.HICON(0)
	if icon != nil {
		handle = icon.handleForDPI(int(user32.GetDpiForWindow(pi.hwnd)))
	}
	description16, err := syscall.UTF16PtrFromString(description)
	if err != nil {
		description16 = &[]uint16{0}[0]
	}
	if hr := pi.taskbarList3.SetOverlayIcon(pi.hwnd, handle, description16); win.FAILED(hr) {
		return errorFromHRESULT("ITaskbarList3.SetOverlayIcon", hr)
	}
	pi.overlayIcon = icon
	pi.overlayIconDescription = description
	return nil
}

// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"strings"
	"syscall"

	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/user32"
)

type MsgBoxStyle uint

const (
	MsgBoxOK                  MsgBoxStyle = user32.MB_OK
	MsgBoxOKCancel            MsgBoxStyle = user32.MB_OKCANCEL
	MsgBoxAbortRetryIgnore    MsgBoxStyle = user32.MB_ABORTRETRYIGNORE
	MsgBoxYesNoCancel         MsgBoxStyle = user32.MB_YESNOCANCEL
	MsgBoxYesNo               MsgBoxStyle = user32.MB_YESNO
	MsgBoxRetryCancel         MsgBoxStyle = user32.MB_RETRYCANCEL
	MsgBoxCancelTryContinue   MsgBoxStyle = user32.MB_CANCELTRYCONTINUE
	MsgBoxIconHand            MsgBoxStyle = user32.MB_ICONHAND
	MsgBoxIconQuestion        MsgBoxStyle = user32.MB_ICONQUESTION
	MsgBoxIconExclamation     MsgBoxStyle = user32.MB_ICONEXCLAMATION
	MsgBoxIconAsterisk        MsgBoxStyle = user32.MB_ICONASTERISK
	MsgBoxUserIcon            MsgBoxStyle = user32.MB_USERICON
	MsgBoxIconWarning         MsgBoxStyle = user32.MB_ICONWARNING
	MsgBoxIconError           MsgBoxStyle = user32.MB_ICONERROR
	MsgBoxIconInformation     MsgBoxStyle = user32.MB_ICONINFORMATION
	MsgBoxIconStop            MsgBoxStyle = user32.MB_ICONSTOP
	MsgBoxDefButton1          MsgBoxStyle = user32.MB_DEFBUTTON1
	MsgBoxDefButton2          MsgBoxStyle = user32.MB_DEFBUTTON2
	MsgBoxDefButton3          MsgBoxStyle = user32.MB_DEFBUTTON3
	MsgBoxDefButton4          MsgBoxStyle = user32.MB_DEFBUTTON4
	MsgBoxApplModal           MsgBoxStyle = user32.MB_APPLMODAL
	MsgBoxSystemModal         MsgBoxStyle = user32.MB_SYSTEMMODAL
	MsgBoxTaskModal           MsgBoxStyle = user32.MB_TASKMODAL
	MsgBoxHelp                MsgBoxStyle = user32.MB_HELP
	MsgBoxSetForeground       MsgBoxStyle = user32.MB_SETFOREGROUND
	MsgBoxDefaultDesktopOnly  MsgBoxStyle = user32.MB_DEFAULT_DESKTOP_ONLY
	MsgBoxTopMost             MsgBoxStyle = user32.MB_TOPMOST
	MsgBoxRight               MsgBoxStyle = user32.MB_RIGHT
	MsgBoxRTLReading          MsgBoxStyle = user32.MB_RTLREADING
	MsgBoxServiceNotification MsgBoxStyle = user32.MB_SERVICE_NOTIFICATION
)

func MsgBox(owner Form, title, message string, style MsgBoxStyle) int {
	var ownerHWnd handle.HWND

	if owner != nil {
		ownerHWnd = owner.Handle()
	}
	messagePtr, err := syscall.UTF16PtrFromString(strings.ReplaceAll(message, "\x00", "␀"))
	if err != nil {
		newError(err.Error())
	}
	titlePtr, err := syscall.UTF16PtrFromString(strings.ReplaceAll(title, "\x00", "␀"))
	if err != nil {
		newError(err.Error())
	}
	return int(user32.MessageBox(
		ownerHWnd,
		messagePtr,
		titlePtr,
		uint32(style)))
}

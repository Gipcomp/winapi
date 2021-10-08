// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"syscall"

	"github.com/Gipcomp/win32/kernel32"
	"github.com/Gipcomp/win32/shell32"
)

func knownFolderPath(id shell32.CSIDL) (string, error) {
	var buf [kernel32.MAX_PATH]uint16

	if !shell32.SHGetSpecialFolderPath(0, &buf[0], id, false) {
		return "", newError("SHGetSpecialFolderPath failed")
	}

	return syscall.UTF16ToString(buf[0:]), nil
}

func AppDataPath() (string, error) {
	return knownFolderPath(shell32.CSIDL_APPDATA)
}

func CommonAppDataPath() (string, error) {
	return knownFolderPath(shell32.CSIDL_COMMON_APPDATA)
}

func LocalAppDataPath() (string, error) {
	return knownFolderPath(shell32.CSIDL_LOCAL_APPDATA)
}

func PersonalPath() (string, error) {
	return knownFolderPath(shell32.CSIDL_PERSONAL)
}

func SystemPath() (string, error) {
	return knownFolderPath(shell32.CSIDL_SYSTEM)
}

func DriveNames() ([]string, error) {
	bufLen := kernel32.GetLogicalDriveStrings(0, nil)
	if bufLen == 0 {
		return nil, lastError("GetLogicalDriveStrings")
	}
	buf := make([]uint16, bufLen+1)

	bufLen = kernel32.GetLogicalDriveStrings(bufLen+1, &buf[0])
	if bufLen == 0 {
		return nil, lastError("GetLogicalDriveStrings")
	}

	var names []string

	for i := 0; i < len(buf)-2; {
		name := syscall.UTF16ToString(buf[i:])
		names = append(names, name)
		i += len(name) + 1
	}

	return names, nil
}

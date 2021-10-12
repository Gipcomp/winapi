// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"syscall"
	"unsafe"

	"github.com/Gipcomp/win32/advapi32"
	"github.com/Gipcomp/win32/kernel32"
	"github.com/Gipcomp/winapi/errs"
)

type RegistryKey struct {
	hKey advapi32.HKEY
}

func ClassesRootKey() *RegistryKey {
	return &RegistryKey{advapi32.HKEY_CLASSES_ROOT}
}

func CurrentUserKey() *RegistryKey {
	return &RegistryKey{advapi32.HKEY_CURRENT_USER}
}

func LocalMachineKey() *RegistryKey {
	return &RegistryKey{advapi32.HKEY_LOCAL_MACHINE}
}

func RegistryKeyString(rootKey *RegistryKey, subKeyPath, valueName string) (value string, err error) {
	var hKey advapi32.HKEY
	strPtr, err := syscall.UTF16PtrFromString(subKeyPath)
	if err != nil {
		return "", err
	}
	if advapi32.RegOpenKeyEx(
		rootKey.hKey,
		strPtr,
		0,
		advapi32.KEY_READ,
		&hKey) != kernel32.ERROR_SUCCESS {

		return "", errs.NewError("RegistryKeyString: Failed to open subkey.")
	}
	defer advapi32.RegCloseKey(hKey)

	var typ uint32
	var data []uint16
	var bufSize uint32
	strPtr, err = syscall.UTF16PtrFromString(valueName)
	if err != nil {
		return "", err
	}
	if kernel32.ERROR_SUCCESS != advapi32.RegQueryValueEx(
		hKey,
		strPtr,
		nil,
		&typ,
		nil,
		&bufSize) {

		return "", errs.NewError("RegQueryValueEx #1")
	}

	data = make([]uint16, bufSize/2+1)

	if kernel32.ERROR_SUCCESS != advapi32.RegQueryValueEx(
		hKey,
		strPtr,
		nil,
		&typ,
		(*byte)(unsafe.Pointer(&data[0])),
		&bufSize) {

		return "", errs.NewError("RegQueryValueEx #2")
	}

	return syscall.UTF16ToString(data), nil
}

func RegistryKeyUint32(rootKey *RegistryKey, subKeyPath, valueName string) (value uint32, err error) {
	var hKey advapi32.HKEY
	strPtr, err := syscall.UTF16PtrFromString(subKeyPath)
	if err != nil {
		return 0, err
	}
	if advapi32.RegOpenKeyEx(
		rootKey.hKey,
		strPtr,
		0,
		advapi32.KEY_READ,
		&hKey) != kernel32.ERROR_SUCCESS {

		return 0, errs.NewError("RegistryKeyUint32: Failed to open subkey.")
	}
	defer advapi32.RegCloseKey(hKey)

	bufSize := uint32(4)
	strPtr, err = syscall.UTF16PtrFromString(valueName)
	if err != nil {
		return 0, err
	}
	if kernel32.ERROR_SUCCESS != advapi32.RegQueryValueEx(
		hKey,
		strPtr,
		nil,
		nil,
		(*byte)(unsafe.Pointer(&value)),
		&bufSize) {

		return 0, errs.NewError("RegQueryValueEx")
	}

	return
}

// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package errs

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/Gipcomp/win32/kernel32"
	"github.com/Gipcomp/win32/win"
)

var (
	logErrors    bool
	panicOnError bool
)

type Error struct {
	inner   error
	message string
	stack   []byte
}

func LogErrors() bool {
	return logErrors
}

func SetLogErrors(v bool) {
	logErrors = v
}

func PanicOnError() bool {
	return panicOnError
}

func SetPanicOnError(v bool) {
	panicOnError = v
}

func (err *Error) Inner() error {
	return err.inner
}

func (err *Error) Message() string {
	if err.message != "" {
		return err.message
	}

	if err.inner != nil {
		if walkErr, ok := err.inner.(*Error); ok {
			return walkErr.Message()
		} else {
			return err.inner.Error()
		}
	}

	return ""
}

func (err *Error) Stack() []byte {
	return err.stack
}

func (err *Error) Error() string {
	return fmt.Sprintf("%s\n\nStack:\n%s", err.Message(), err.stack)
}

func processErrorNoPanic(err error) error {
	if logErrors {
		if walkErr, ok := err.(*Error); ok {
			log.Print(walkErr.Error())
		} else {
			log.Printf("%s\n\nStack:\n%s", err, debug.Stack())
		}
	}

	return err
}

func processError(err error) error {
	processErrorNoPanic(err)

	if panicOnError {
		panic(err)
	}

	return err
}

func newErr(message string) error {
	return &Error{message: message, stack: debug.Stack()}
}

func NewError(message string) error {
	return processError(newErr(message))
}

func NewErrorNoPanic(message string) error {
	return processErrorNoPanic(newErr(message))
}

func LastError(win32FuncName string) error {
	if errno := kernel32.GetLastError(); errno != kernel32.ERROR_SUCCESS {
		return NewError(fmt.Sprintf("%s: Error %d", win32FuncName, errno))
	}

	return NewError(win32FuncName)
}

func ErrorFromHRESULT(funcName string, hr win.HRESULT) error {
	return NewError(fmt.Sprintf("%s: Error %d", funcName, hr))
}

func wrapErr(err error) error {
	if _, ok := err.(*Error); ok {
		return err
	}

	return &Error{inner: err, stack: debug.Stack()}
}

func WrapErrorNoPanic(err error) error {
	return processErrorNoPanic(wrapErr(err))
}

func WrapError(err error) error {
	return processError(wrapErr(err))
}

func toErrorNoPanic(x interface{}) error {
	switch x := x.(type) {
	case *Error:
		return x

	case error:
		return WrapErrorNoPanic(x)

	case string:
		return NewErrorNoPanic(x)
	}

	return NewErrorNoPanic(fmt.Sprintf("Error: %v", x))
}

func toError(x interface{}) error {
	err := toErrorNoPanic(x)

	if panicOnError {
		panic(err)
	}

	return err
}

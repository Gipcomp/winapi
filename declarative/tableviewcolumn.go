// Copyright 2013 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"github.com/Gipcomp/winapi"
)

type Alignment1D uint

const (
	AlignDefault = Alignment1D(winapi.AlignDefault)
	AlignNear    = Alignment1D(winapi.AlignNear)
	AlignCenter  = Alignment1D(winapi.AlignCenter)
	AlignFar     = Alignment1D(winapi.AlignFar)
)

type TableViewColumn struct {
	Name       string
	DataMember string
	Format     string
	Title      string
	Alignment  Alignment1D
	Precision  int
	Width      int
	Hidden     bool
	Frozen     bool
	StyleCell  func(style *winapi.CellStyle)
	LessFunc   func(i, j int) bool
	FormatFunc func(value interface{}) string
}

func (tvc TableViewColumn) Create(tv *winapi.TableView) error {
	w := winapi.NewTableViewColumn()

	if err := w.SetAlignment(winapi.Alignment1D(tvc.Alignment)); err != nil {
		return err
	}
	w.SetDataMember(tvc.DataMember)
	if tvc.Format != "" {
		if err := w.SetFormat(tvc.Format); err != nil {
			return err
		}
	}
	if err := w.SetPrecision(tvc.Precision); err != nil {
		return err
	}
	w.SetName(tvc.Name)
	if err := w.SetTitle(tvc.Title); err != nil {
		return err
	}
	if err := w.SetVisible(!tvc.Hidden); err != nil {
		return err
	}
	if err := w.SetFrozen(tvc.Frozen); err != nil {
		return err
	}
	if err := w.SetWidth(tvc.Width); err != nil {
		return err
	}
	w.SetLessFunc(tvc.LessFunc)
	w.SetFormatFunc(tvc.FormatFunc)

	return tv.Columns().Add(w)
}

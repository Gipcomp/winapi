// Copyright 2013 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"syscall"
	"unsafe"

	"github.com/Gipcomp/win32/commctrl"
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/win"
	"github.com/Gipcomp/winapi/errs"
)

// TableViewColumn represents a column in a TableView.
type TableViewColumn struct {
	tv            *TableView
	name          string
	dataMember    string
	alignment     Alignment1D
	format        string
	precision     int
	title         string
	titleOverride string
	width         int
	lessFunc      func(i, j int) bool
	formatFunc    func(value interface{}) string
	visible       bool
	frozen        bool
}

// NewTableViewColumn returns a new TableViewColumn.
func NewTableViewColumn() *TableViewColumn {
	return &TableViewColumn{
		format:  "%v",
		visible: true,
		width:   50,
	}
}

// Alignment returns the alignment of the TableViewColumn.
func (tvc *TableViewColumn) Alignment() Alignment1D {
	return tvc.alignment
}

// SetAlignment sets the alignment of the TableViewColumn.
func (tvc *TableViewColumn) SetAlignment(alignment Alignment1D) (err error) {
	if alignment == AlignDefault {
		alignment = AlignNear
	}

	if alignment == tvc.alignment {
		return nil
	}

	old := tvc.alignment
	defer func() {
		if err != nil {
			tvc.alignment = old
		}
	}()

	tvc.alignment = alignment

	return tvc.update()
}

// DataMember returns the data member this TableViewColumn is bound against.
func (tvc *TableViewColumn) DataMember() string {
	return tvc.dataMember
}

// DataMemberEffective returns the effective data member this TableViewColumn is
// bound against.
func (tvc *TableViewColumn) DataMemberEffective() string {
	if tvc.dataMember != "" {
		return tvc.dataMember
	}

	return tvc.name
}

// SetDataMember sets the data member this TableViewColumn is bound against.
func (tvc *TableViewColumn) SetDataMember(dataMember string) {
	tvc.dataMember = dataMember
}

// Format returns the format string for converting a value into a string.
func (tvc *TableViewColumn) Format() string {
	return tvc.format
}

// SetFormat sets the format string for converting a value into a string.
func (tvc *TableViewColumn) SetFormat(format string) (err error) {
	if format == tvc.format {
		return nil
	}

	old := tvc.format
	defer func() {
		if err != nil {
			tvc.format = old
		}
	}()

	tvc.format = format

	if tvc.tv == nil {
		return nil
	}

	return tvc.tv.Invalidate()
}

// Name returns the name of this TableViewColumn.
func (tvc *TableViewColumn) Name() string {
	return tvc.name
}

// SetName sets the name of this TableViewColumn.
func (tvc *TableViewColumn) SetName(name string) {
	tvc.name = name
}

// Precision returns the number of decimal places for formatting float32,
// float64 or big.Rat values.
func (tvc *TableViewColumn) Precision() int {
	return tvc.precision
}

// SetPrecision sets the number of decimal places for formatting float32,
// float64 or big.Rat values.
func (tvc *TableViewColumn) SetPrecision(precision int) (err error) {
	if precision == tvc.precision {
		return nil
	}

	old := tvc.precision
	defer func() {
		if err != nil {
			tvc.precision = old
		}
	}()

	tvc.precision = precision

	if tvc.tv == nil {
		return nil
	}

	return tvc.tv.Invalidate()
}

// Title returns the (default) text to display in the column header.
func (tvc *TableViewColumn) Title() string {
	return tvc.title
}

// SetTitle sets the (default) text to display in the column header.
func (tvc *TableViewColumn) SetTitle(title string) (err error) {
	if title == tvc.title {
		return nil
	}

	old := tvc.title
	defer func() {
		if err != nil {
			tvc.title = old
		}
	}()

	tvc.title = title

	return tvc.update()
}

// TitleOverride returns the (overridden by user) text to display in the column
// header.
func (tvc *TableViewColumn) TitleOverride() string {
	return tvc.titleOverride
}

// SetTitleOverride sets the (overridden by user) text to display in the column
// header.
func (tvc *TableViewColumn) SetTitleOverride(titleOverride string) (err error) {
	if titleOverride == tvc.titleOverride {
		return nil
	}

	old := tvc.titleOverride
	defer func() {
		if err != nil {
			tvc.titleOverride = old
		}
	}()

	tvc.titleOverride = titleOverride

	return tvc.update()
}

// TitleEffective returns the effective text to display in the column header.
func (tvc *TableViewColumn) TitleEffective() string {
	if tvc.titleOverride != "" {
		return tvc.titleOverride
	}

	if tvc.title != "" {
		return tvc.title
	}

	return tvc.DataMemberEffective()
}

// Visible returns if the column is visible.
func (tvc *TableViewColumn) Visible() bool {
	return tvc.visible
}

// SetVisible sets if the column is visible.
func (tvc *TableViewColumn) SetVisible(visible bool) (err error) {
	if visible == tvc.visible {
		return nil
	}

	old := tvc.visible
	defer func() {
		if err != nil {
			tvc.visible = old
		}
	}()

	tvc.visible = visible

	if tvc.tv == nil {
		return nil
	}

	if visible {
		return tvc.create()
	}

	return tvc.destroy()
}

// Frozen returns if the column is frozen.
func (tvc *TableViewColumn) Frozen() bool {
	return tvc.frozen
}

// SetFrozen sets if the column is frozen.
func (tvc *TableViewColumn) SetFrozen(frozen bool) (err error) {
	if frozen == tvc.frozen {
		return nil
	}

	var checkBoxes bool
	if tvc.tv != nil {
		checkBoxes = tvc.tv.CheckBoxes()
	}

	old := tvc.frozen
	defer func() {
		if err != nil {
			tvc.frozen = old

			if tvc.tv != nil {
				tvc.create()
			}
		}

		if tvc.tv != nil {
			tvc.tv.hasFrozenColumn = tvc.tv.visibleFrozenColumnCount() > 0
			tvc.tv.SetCheckBoxes(checkBoxes)
			tvc.tv.applyImageList()
		}
	}()

	if tvc.tv != nil && tvc.visible {
		if err = tvc.destroy(); err != nil {
			return
		}
	}

	tvc.frozen = frozen

	if tvc.tv != nil && tvc.visible {
		return tvc.create()
	}

	return nil
}

// Width returns the width of the column in pixels.
func (tvc *TableViewColumn) Width() int {
	if tvc.tv == nil || !tvc.visible {
		return tvc.width
	}

	// We call user32.SendMessage instead of tvc.sendMessage here, because some
	// call inside the latter interferes with scrolling via scroll bar button
	// when *TableViewColumn.Width is called from *TableView.StretchLastColumn.
	var hwnd handle.HWND
	if tvc.frozen {
		hwnd = tvc.tv.hwndFrozenLV
	} else {
		hwnd = tvc.tv.hwndNormalLV
	}

	return tvc.tv.IntTo96DPI(int(user32.SendMessage(hwnd, commctrl.LVM_GETCOLUMNWIDTH, uintptr(tvc.indexInListView()), 0)))
}

// SetWidth sets the width of the column in pixels.
func (tvc *TableViewColumn) SetWidth(width int) (err error) {
	if width == tvc.width {
		return nil
	}

	old := tvc.width
	defer func() {
		if err != nil {
			tvc.width = old
		}
	}()

	tvc.width = width

	return tvc.update()
}

// LessFunc returns the less func of this TableViewColumn.
//
// This function is used to provide custom sorting for models based on ReflectTableModel only.
func (tvc *TableViewColumn) LessFunc() func(i, j int) bool {
	return tvc.lessFunc
}

// SetLessFunc sets the less func of this TableViewColumn.
//
// This function is used to provide custom sorting for models based on ReflectTableModel only.
func (tvc *TableViewColumn) SetLessFunc(lessFunc func(i, j int) bool) {
	tvc.lessFunc = lessFunc
}

// FormatFunc returns the custom format func of this TableViewColumn.
func (tvc *TableViewColumn) FormatFunc() func(value interface{}) string {
	return tvc.formatFunc
}

// FormatFunc sets the custom format func of this TableViewColumn.
func (tvc *TableViewColumn) SetFormatFunc(formatFunc func(value interface{}) string) {
	tvc.formatFunc = formatFunc
}

func (tvc *TableViewColumn) indexInListView() int32 {
	if tvc.tv == nil {
		return -1
	}

	var idx int32

	for _, c := range tvc.tv.columns.items {
		if c.frozen != tvc.frozen {
			continue
		}

		if c == tvc {
			break
		}

		if c.visible {
			idx++
		}
	}

	return idx
}

func (tvc *TableViewColumn) create() error {
	var lvc commctrl.LVCOLUMN

	index := tvc.indexInListView()

	dpi := tvc.tv.DPI()
	lvc.Mask = commctrl.LVCF_FMT | commctrl.LVCF_WIDTH | commctrl.LVCF_TEXT | commctrl.LVCF_SUBITEM
	lvc.ISubItem = index
	var err error
	lvc.PszText, err = syscall.UTF16PtrFromString(tvc.TitleEffective())
	if err != nil {
		return err
	}
	if tvc.width > 0 {
		lvc.Cx = int32(IntFrom96DPI(tvc.width, dpi))
	} else {
		lvc.Cx = int32(IntFrom96DPI(100, dpi))
	}

	switch tvc.alignment {
	case AlignCenter:
		lvc.Fmt = 2

	case AlignFar:
		lvc.Fmt = 1
	}

	if -1 == int(tvc.sendMessage(commctrl.LVM_INSERTCOLUMN, uintptr(index), uintptr(unsafe.Pointer(&lvc)))) {
		return errs.NewError("LVM_INSERTCOLUMN")
	}

	tvc.tv.updateLVSizes()

	return nil
}

func (tvc *TableViewColumn) destroy() error {
	width := tvc.Width()

	if win.FALSE == tvc.sendMessage(commctrl.LVM_DELETECOLUMN, uintptr(tvc.indexInListView()), 0) {
		return errs.NewError("LVM_DELETECOLUMN")
	}

	tvc.width = width

	tvc.tv.updateLVSizes()

	return nil
}

func (tvc *TableViewColumn) update() error {
	if tvc.tv == nil || !tvc.visible {
		return nil
	}

	lvc := tvc.getLVCOLUMN()

	if win.FALSE == tvc.sendMessage(commctrl.LVM_SETCOLUMN, uintptr(tvc.indexInListView()), uintptr(unsafe.Pointer(lvc))) {
		return errs.NewError("LVM_SETCOLUMN")
	}

	tvc.tv.updateLVSizes()

	return nil
}

func (tvc *TableViewColumn) getLVCOLUMN() *commctrl.LVCOLUMN {
	var lvc commctrl.LVCOLUMN

	dpi := 96
	if tvc.tv != nil {
		dpi = tvc.tv.DPI()
	} else {
		dpi = screenDPI()
	}
	width := IntFrom96DPI(tvc.width, dpi)

	lvc.Mask = commctrl.LVCF_FMT | commctrl.LVCF_WIDTH | commctrl.LVCF_TEXT | commctrl.LVCF_SUBITEM
	lvc.ISubItem = int32(tvc.indexInListView())
	var err error
	lvc.PszText, err = syscall.UTF16PtrFromString(tvc.TitleEffective())
	if err != nil {
		errs.NewError(err.Error())
	}
	lvc.Cx = int32(width)

	switch tvc.alignment {
	case AlignCenter:
		lvc.Fmt = 2

	case AlignFar:
		lvc.Fmt = 1
	}

	return &lvc
}

func (tvc *TableViewColumn) sendMessage(msg uint32, wp, lp uintptr) uintptr {
	if tvc.tv == nil {
		return 0
	}

	tvc.tv.hasFrozenColumn = tvc.tv.visibleFrozenColumnCount() > 0
	tvc.tv.SetCheckBoxes(tvc.tv.CheckBoxes())
	tvc.tv.applyImageList()

	var hwnd handle.HWND
	if tvc.frozen {
		hwnd = tvc.tv.hwndFrozenLV
	} else {
		hwnd = tvc.tv.hwndNormalLV
	}

	return user32.SendMessage(hwnd, msg, wp, lp)
}

// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/Gipcomp/win32/commctrl"
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/kernel32"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/winapi/errs"
)

type DateEdit struct {
	WidgetBase
	dateChangedPublisher EventPublisher
	format               string
}

func newDateEdit(parent Container, style uint32) (*DateEdit, error) {
	de := new(DateEdit)

	if err := InitWidget(
		de,
		parent,
		"SysDateTimePick32",
		user32.WS_TABSTOP|user32.WS_VISIBLE|commctrl.DTS_SHORTDATEFORMAT|style,
		0); err != nil {
		return nil, err
	}

	if style&commctrl.DTS_SHOWNONE != 0 {
		de.setSystemTime(nil)
	}

	de.GraphicsEffects().Add(InteractionEffect)
	de.GraphicsEffects().Add(FocusEffect)

	de.MustRegisterProperty("Date", NewProperty(
		func() interface{} {
			return de.Date()
		},
		func(v interface{}) error {
			return de.SetDate(assertTimeOr(v, time.Time{}))
		},
		de.dateChangedPublisher.Event()))

	return de, nil
}

func NewDateEdit(parent Container) (*DateEdit, error) {
	return newDateEdit(parent, 0)
}

func NewDateEditWithNoneOption(parent Container) (*DateEdit, error) {
	return newDateEdit(parent, commctrl.DTS_SHOWNONE)
}

func (de *DateEdit) systemTimeToTime(st *kernel32.SYSTEMTIME) time.Time {
	if st == nil || !de.hasStyleBits(commctrl.DTS_SHOWNONE) && st.WYear == 1601 && st.WMonth == 1 && st.WDay == 1 {
		return time.Time{}
	}

	var hour, minute, second int
	if de.timeOfDayDisplayed() {
		hour = int(st.WHour)
		minute = int(st.WMinute)
		second = int(st.WSecond)
	}

	return time.Date(int(st.WYear), time.Month(st.WMonth), int(st.WDay), hour, minute, second, 0, time.Local)
}

func (de *DateEdit) timeToSystemTime(t time.Time) *kernel32.SYSTEMTIME {
	if t.Year() < 1601 {
		if de.hasStyleBits(commctrl.DTS_SHOWNONE) {
			return nil
		} else {
			return &kernel32.SYSTEMTIME{
				WYear:  uint16(1601),
				WMonth: uint16(1),
				WDay:   uint16(1),
			}
		}
	}

	st := &kernel32.SYSTEMTIME{
		WYear:  uint16(t.Year()),
		WMonth: uint16(t.Month()),
		WDay:   uint16(t.Day()),
	}

	if de.timeOfDayDisplayed() {
		st.WHour = uint16(t.Hour())
		st.WMinute = uint16(t.Minute())
		st.WSecond = uint16(t.Second())
	}

	return st
}

func (de *DateEdit) systemTime() (*kernel32.SYSTEMTIME, error) {
	var st kernel32.SYSTEMTIME

	switch de.SendMessage(commctrl.DTM_GETSYSTEMTIME, 0, uintptr(unsafe.Pointer(&st))) {
	case commctrl.GDT_VALID:
		return &st, nil

	case commctrl.GDT_NONE:
		return nil, nil
	}

	return nil, errs.NewError("SendMessage(DTM_GETSYSTEMTIME)")
}

func (de *DateEdit) setSystemTime(st *kernel32.SYSTEMTIME) error {
	var wParam uintptr

	if st != nil {
		wParam = commctrl.GDT_VALID
	} else {
		// Ensure today's date is displayed.
		de.setSystemTime(de.timeToSystemTime(time.Now()))

		wParam = commctrl.GDT_NONE
	}

	if de.SendMessage(commctrl.DTM_SETSYSTEMTIME, wParam, uintptr(unsafe.Pointer(st))) == 0 {
		return errs.NewError("SendMessage(DTM_SETSYSTEMTIME)")
	}

	return nil
}

func (de *DateEdit) timeOfDayDisplayed() bool {
	return strings.ContainsAny(de.format, "Hhms")
}

func (de *DateEdit) Format() string {
	return de.format
}

func (de *DateEdit) SetFormat(format string) error {
	strPtr, err := syscall.UTF16PtrFromString(format)
	if err != nil {
		errs.NewError(err.Error())
	}
	lp := uintptr(unsafe.Pointer(strPtr))

	if de.SendMessage(commctrl.DTM_SETFORMAT, 0, lp) == 0 {
		return errs.NewError("DTM_SETFORMAT failed")
	}

	de.format = format

	return nil
}

func (de *DateEdit) Range() (min, max time.Time) {
	var st [2]kernel32.SYSTEMTIME

	ret := de.SendMessage(commctrl.DTM_GETRANGE, 0, uintptr(unsafe.Pointer(&st[0])))

	if ret&commctrl.GDTR_MIN > 0 {
		min = de.systemTimeToTime(&st[0])
	}

	if ret&commctrl.GDTR_MAX > 0 {
		max = de.systemTimeToTime(&st[1])
	}

	return
}

func (de *DateEdit) SetRange(min, max time.Time) error {
	if !min.IsZero() && !max.IsZero() {
		if min.Year() > max.Year() ||
			min.Year() == max.Year() && min.Month() > max.Month() ||
			min.Year() == max.Year() && min.Month() == max.Month() && min.Day() > max.Day() {
			return errs.NewError("invalid range")
		}
	}

	var st [2]kernel32.SYSTEMTIME
	var wParam uintptr

	if !min.IsZero() {
		wParam |= commctrl.GDTR_MIN
		st[0] = *de.timeToSystemTime(min)
	}

	if !max.IsZero() {
		wParam |= commctrl.GDTR_MAX
		st[1] = *de.timeToSystemTime(max)
	}

	if de.SendMessage(commctrl.DTM_SETRANGE, wParam, uintptr(unsafe.Pointer(&st[0]))) == 0 {
		return errs.NewError("SendMessage(DTM_SETRANGE)")
	}

	return nil
}

func (de *DateEdit) Date() time.Time {
	st, err := de.systemTime()
	if err != nil || st == nil {
		return time.Time{}
	}

	return de.systemTimeToTime(st)
}

func (de *DateEdit) SetDate(date time.Time) error {
	stNew := de.timeToSystemTime(date)
	stOld, err := de.systemTime()
	if err != nil {
		return err
	} else if stNew == stOld || stNew != nil && stOld != nil && *stNew == *stOld {
		return nil
	}

	if err := de.setSystemTime(stNew); err != nil {
		return err
	}

	de.dateChangedPublisher.Publish()

	return nil
}

func (de *DateEdit) DateChanged() *Event {
	return de.dateChangedPublisher.Event()
}

func (de *DateEdit) WndProc(hwnd handle.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case user32.WM_NOTIFY:
		switch uint32(((*user32.NMHDR)(unsafe.Pointer(lParam))).Code) {
		case commctrl.DTN_DATETIMECHANGE:
			de.dateChangedPublisher.Publish()
		}
	}

	return de.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}

func (*DateEdit) NeedsWmSize() bool {
	return true
}

func (de *DateEdit) CreateLayoutItem(ctx *LayoutContext) LayoutItem {
	return &dateEditLayoutItem{
		idealSize: de.dialogBaseUnitsToPixels(Size{80, 12}),
	}
}

type dateEditLayoutItem struct {
	LayoutItemBase
	idealSize Size // in native pixels
}

func (*dateEditLayoutItem) LayoutFlags() LayoutFlags {
	return GrowableHorz
}

func (li *dateEditLayoutItem) IdealSize() Size {
	return li.idealSize
}

func (li *dateEditLayoutItem) MinSize() Size {
	return li.idealSize
}

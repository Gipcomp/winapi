// Copyright 2013 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"syscall"
	"unsafe"

	"github.com/Gipcomp/win32/comctl32"
	"github.com/Gipcomp/win32/commctrl"
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/win"
	"github.com/Gipcomp/winapi/errs"
)

// StatusBar is a widget that displays status messages.
type StatusBar struct {
	WidgetBase
	items *StatusBarItemList
}

// NewStatusBar returns a new StatusBar as child of container parent.
func NewStatusBar(parent Container) (*StatusBar, error) {
	sb := new(StatusBar)

	if err := InitWidget(
		sb,
		parent,
		"msctls_statusbar32",
		commctrl.SBARS_SIZEGRIP|commctrl.SBARS_TOOLTIPS,
		0); err != nil {
		return nil, err
	}

	sb.items = newStatusBarItemList(sb)

	return sb, nil
}

// Items returns the list of items in the StatusBar.
func (sb *StatusBar) Items() *StatusBarItemList {
	return sb.items
}

// SetVisible sets whether the StatusBar is visible.
func (sb *StatusBar) SetVisible(visible bool) {
	sb.WidgetBase.SetVisible(visible)

	sb.RequestLayout()
}

func (sb *StatusBar) ApplyDPI(dpi int) {
	sb.WidgetBase.ApplyDPI(dpi)

	sb.update()
}

func (sb *StatusBar) update() error {
	if err := sb.updateParts(); err != nil {
		return err
	}

	for i, item := range sb.items.items {
		if err := item.update(i); err != nil {
			return err
		}
	}

	sb.SetVisible(sb.items.Len() > 0)

	return nil
}

func (sb *StatusBar) updateParts() error {
	items := sb.items.items

	dpi := sb.DPI()

	rightEdges := make([]int32, len(items))
	var right int32
	for i, item := range items {
		right += int32(IntFrom96DPI(item.width, dpi))
		rightEdges[i] = right
	}
	var rep *int32
	if len(rightEdges) > 0 {
		rep = &rightEdges[0]
	}

	if len(rightEdges) == 1 {
		rightEdges[0] = -1
	}

	if 0 == sb.SendMessage(
		commctrl.SB_SETPARTS,
		uintptr(len(items)),
		uintptr(unsafe.Pointer(rep))) {

		return errs.NewError("SB_SETPARTS")
	}

	return nil
}

func (sb *StatusBar) WndProc(hwnd handle.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case user32.WM_NOTIFY:
		nmhdr := (*user32.NMHDR)(unsafe.Pointer(lParam))

		switch nmhdr.Code {
		case comctl32.NM_CLICK:
			lpnm := (*commctrl.NMMOUSE)(unsafe.Pointer(lParam))
			if n := int(lpnm.DwItemSpec); n >= 0 && n < sb.items.Len() {
				sb.items.At(n).raiseClicked()
			}
		}
	}

	return sb.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}

func (*StatusBar) CreateLayoutItem(ctx *LayoutContext) LayoutItem {
	return new(statusBarLayoutItem)
}

type statusBarLayoutItem struct {
	LayoutItemBase
}

func (*statusBarLayoutItem) LayoutFlags() LayoutFlags {
	return 0
}

func (*statusBarLayoutItem) IdealSize() Size {
	return Size{}
}

// StatusBarItem represents a section of a StatusBar that can have its own icon,
// text, tool tip text and width.
type StatusBarItem struct {
	sb               *StatusBar
	icon             *Icon
	text             string
	toolTipText      string
	width            int
	clickedPublisher EventPublisher
}

// NewStatusBarItem returns a new StatusBarItem.
func NewStatusBarItem() *StatusBarItem {
	return &StatusBarItem{width: 100}
}

// Icon returns the Icon of the StatusBarItem.
func (sbi *StatusBarItem) Icon() *Icon {
	return sbi.icon
}

// SetIcon sets the Icon of the StatusBarItem.
func (sbi *StatusBarItem) SetIcon(icon *Icon) error {
	if icon == sbi.icon {
		return nil
	}

	old := sbi.icon
	sbi.icon = icon

	return sbi.maybeTry(sbi.updateIcon, func() { sbi.icon = old })
}

// Text returns the text of the StatusBarItem.
func (sbi *StatusBarItem) Text() string {
	return sbi.text
}

// SetText sets the text of the StatusBarItem.
func (sbi *StatusBarItem) SetText(text string) error {
	if text == sbi.text {
		return nil
	}

	old := sbi.text
	sbi.text = text

	return sbi.maybeTry(sbi.updateText, func() { sbi.text = old })
}

// ToolTipText returns the tool tip text of the StatusBarItem.
func (sbi *StatusBarItem) ToolTipText() string {
	return sbi.toolTipText
}

// SetToolTipText sets the tool tip text of the StatusBarItem.
func (sbi *StatusBarItem) SetToolTipText(toolTipText string) error {
	if toolTipText == sbi.toolTipText {
		return nil
	}

	old := sbi.toolTipText
	sbi.toolTipText = toolTipText

	return sbi.maybeTry(sbi.updateToolTipText, func() { sbi.toolTipText = old })
}

// Width returns the width of the StatusBarItem.
func (sbi *StatusBarItem) Width() int {
	return sbi.width
}

// SetWidth sets the width of the StatusBarItem.
func (sbi *StatusBarItem) SetWidth(width int) error {
	if width == sbi.width {
		return nil
	}

	old := sbi.width
	sbi.width = width

	if sbi.sb != nil {
		succeeded := false
		defer func() {
			if !succeeded {
				sbi.width = old
			}
		}()

		if err := sbi.sb.updateParts(); err != nil {
			return err
		}

		succeeded = true
	}

	return nil
}

func (sbi *StatusBarItem) Clicked() *Event {
	return sbi.clickedPublisher.Event()
}

func (sbi *StatusBarItem) raiseClicked() {
	sbi.clickedPublisher.Publish()
}

func (sbi *StatusBarItem) maybeTry(f func(index int) error, rollback func()) error {
	if sbi.sb != nil {
		succeeded := false
		defer func() {
			if !succeeded {
				rollback()
			}
		}()

		if err := f(sbi.sb.items.Index(sbi)); err != nil {
			return err
		}

		succeeded = true
	}

	return nil
}

func (sbi *StatusBarItem) update(index int) error {
	if err := sbi.updateIcon(index); err != nil {
		return err
	}
	if err := sbi.updateText(index); err != nil {
		return err
	}
	if err := sbi.updateToolTipText(index); err != nil {
		return err
	}

	return nil
}

func (sbi *StatusBarItem) updateIcon(index int) error {
	var hIcon user32.HICON
	if sbi.icon != nil {
		hIcon = sbi.icon.handleForDPI(sbi.sb.DPI())
	}

	if 0 == sbi.sb.SendMessage(
		commctrl.SB_SETICON,
		uintptr(index),
		uintptr(hIcon)) {

		return errs.NewError("SB_SETICON")
	}

	return nil
}

func (sbi *StatusBarItem) updateText(index int) error {
	utf16, err := syscall.UTF16PtrFromString(sbi.text)
	if err != nil {
		return err
	}

	if 0 == sbi.sb.SendMessage(
		commctrl.SB_SETTEXT,
		uintptr(win.MAKEWORD(byte(index), 0)),
		uintptr(unsafe.Pointer(utf16))) {

		return errs.NewError("SB_SETTEXT")
	}

	return nil
}

func (sbi *StatusBarItem) updateToolTipText(index int) error {
	utf16, err := syscall.UTF16PtrFromString(sbi.toolTipText)
	if err != nil {
		return err
	}

	sbi.sb.SendMessage(
		commctrl.SB_SETTIPTEXT,
		uintptr(index),
		uintptr(unsafe.Pointer(utf16)))

	return nil
}

type StatusBarItemList struct {
	sb    *StatusBar
	items []*StatusBarItem
}

func newStatusBarItemList(statusBar *StatusBar) *StatusBarItemList {
	return &StatusBarItemList{sb: statusBar}
}

func (l *StatusBarItemList) Add(item *StatusBarItem) error {
	return l.Insert(len(l.items), item)
}

func (l *StatusBarItemList) At(index int) *StatusBarItem {
	return l.items[index]
}

func (l *StatusBarItemList) Clear() error {
	old := l.items
	l.items = l.items[:0]

	succeeded := false
	defer func() {
		if !succeeded {
			l.items = old

			l.sb.update()
		}
	}()

	if err := l.sb.update(); err != nil {
		return err
	}

	succeeded = true

	return nil
}

func (l *StatusBarItemList) Index(item *StatusBarItem) int {
	for i, it := range l.items {
		if it == item {
			return i
		}
	}

	return -1
}

func (l *StatusBarItemList) Contains(item *StatusBarItem) bool {
	return l.Index(item) > -1
}

func (l *StatusBarItemList) Insert(index int, item *StatusBarItem) error {
	if item.sb != nil {
		return errs.NewError("item already contained in a StatusBar")
	}

	l.items = append(l.items, nil)
	copy(l.items[index+1:], l.items[index:])
	l.items[index] = item

	item.sb = l.sb

	succeeded := false
	defer func() {
		if !succeeded {
			item.sb = nil
			l.items = append(l.items[:index], l.items[index+1:]...)

			l.sb.update()
		}
	}()

	if err := l.sb.update(); err != nil {
		return err
	}

	succeeded = true

	return nil
}

func (l *StatusBarItemList) Len() int {
	return len(l.items)
}

func (l *StatusBarItemList) Remove(item *StatusBarItem) error {
	index := l.Index(item)
	if index == -1 {
		return nil
	}

	return l.RemoveAt(index)
}

func (l *StatusBarItemList) RemoveAt(index int) error {
	item := l.items[index]
	item.sb = nil

	l.items = append(l.items[:index], l.items[index+1:]...)

	succeeded := false
	defer func() {
		if !succeeded {
			l.items = append(l.items, nil)
			copy(l.items[index+1:], l.items[index:])
			l.items[index] = item

			item.sb = l.sb

			l.sb.update()
		}
	}()

	if err := l.sb.update(); err != nil {
		return err
	}

	succeeded = true

	return nil
}

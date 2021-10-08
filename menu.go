// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/winuser"
)

type Menu struct {
	hMenu   winuser.HMENU
	window  Window
	actions *ActionList
	getDPI  func() int
}

func newMenuBar(window Window) (*Menu, error) {
	hMenu := user32.CreateMenu()
	if hMenu == 0 {
		return nil, lastError("CreateMenu")
	}

	m := &Menu{
		hMenu:  hMenu,
		window: window,
	}
	m.actions = newActionList(m)

	return m, nil
}

func NewMenu() (*Menu, error) {
	hMenu := user32.CreatePopupMenu()
	if hMenu == 0 {
		return nil, lastError("CreatePopupMenu")
	}

	var mi winuser.MENUINFO
	mi.CbSize = uint32(unsafe.Sizeof(mi))

	if !user32.GetMenuInfo(hMenu, &mi) {
		return nil, lastError("GetMenuInfo")
	}

	mi.FMask |= winuser.MIM_STYLE
	mi.DwStyle = winuser.MNS_CHECKORBMP

	if !user32.SetMenuInfo(hMenu, &mi) {
		return nil, lastError("SetMenuInfo")
	}

	m := &Menu{
		hMenu: hMenu,
	}
	m.actions = newActionList(m)

	return m, nil
}

func (m *Menu) Dispose() {
	m.actions.Clear()

	if m.hMenu != 0 {
		user32.DestroyMenu(m.hMenu)
		m.hMenu = 0
	}
}

func (m *Menu) IsDisposed() bool {
	return m.hMenu == 0
}

func (m *Menu) Actions() *ActionList {
	return m.actions
}

func (m *Menu) updateItemsWithImageForWindow(window Window) {
	if m.window == nil {
		m.window = window
		defer func() {
			m.window = nil
		}()
	}

	for _, action := range m.actions.actions {
		if action.image != nil {
			m.onActionChanged(action)
		}
		if action.menu != nil {
			action.menu.updateItemsWithImageForWindow(window)
		}
	}
}

func (m *Menu) initMenuItemInfoFromAction(mii *winuser.MENUITEMINFO, action *Action) {
	mii.CbSize = uint32(unsafe.Sizeof(*mii))
	mii.FMask = winuser.MIIM_FTYPE | winuser.MIIM_ID | winuser.MIIM_STATE | winuser.MIIM_STRING
	if action.image != nil {
		mii.FMask |= winuser.MIIM_BITMAP
		dpi := 96
		if m.getDPI != nil {
			dpi = m.getDPI()
		} else if m.window != nil {
			dpi = m.window.DPI()
		} else {
			dpi = screenDPI()
		}
		if bmp, err := iconCache.Bitmap(action.image, dpi); err == nil {
			mii.HbmpItem = bmp.hBmp
		}
	}
	if action.IsSeparator() {
		mii.FType |= winuser.MFT_SEPARATOR
	} else {
		mii.FType |= winuser.MFT_STRING
		var text string
		if s := action.shortcut; s.Key != 0 {
			text = fmt.Sprintf("%s\t%s", action.text, s.String())
		} else {
			text = action.text
		}

		var err error
		mii.DwTypeData, err = syscall.UTF16PtrFromString(text)
		if err != nil {
			newError(err.Error())
		}
		mii.Cch = uint32(len([]rune(action.text)))
	}
	mii.WID = uint32(action.id)

	if action.Enabled() {
		mii.FState &^= winuser.MFS_DISABLED
	} else {
		mii.FState |= winuser.MFS_DISABLED
	}

	if action.Checkable() {
		mii.FMask |= winuser.MIIM_CHECKMARKS
	}
	if action.Checked() {
		mii.FState |= winuser.MFS_CHECKED
	}
	if action.Exclusive() {
		mii.FType |= winuser.MFT_RADIOCHECK
	}

	menu := action.menu
	if menu != nil {
		mii.FMask |= winuser.MIIM_SUBMENU
		mii.HSubMenu = menu.hMenu
	}
}

func (m *Menu) handleDefaultState(action *Action) {
	if action.Default() {
		// Unset other default actions before we set this one. Otherwise insertion fails.
		user32.SetMenuDefaultItem(m.hMenu, ^uint32(0), false)
		for _, otherAction := range m.actions.actions {
			if otherAction != action {
				otherAction.SetDefault(false)
			}
		}
	}
}

func (m *Menu) onActionChanged(action *Action) error {
	m.handleDefaultState(action)

	if !action.Visible() {
		return nil
	}

	var mii winuser.MENUITEMINFO

	m.initMenuItemInfoFromAction(&mii, action)

	if !user32.SetMenuItemInfo(m.hMenu, uint32(m.actions.indexInObserver(action)), true, &mii) {
		return newError("SetMenuItemInfo failed")
	}

	if action.Default() {
		user32.SetMenuDefaultItem(m.hMenu, uint32(m.actions.indexInObserver(action)), true)
	}

	if action.Exclusive() && action.Checked() {
		var first, last int

		index := m.actions.Index(action)

		for i := index; i >= 0; i-- {
			first = i
			if !m.actions.At(i).Exclusive() {
				break
			}
		}

		for i := index; i < m.actions.Len(); i++ {
			last = i
			if !m.actions.At(i).Exclusive() {
				break
			}
		}

		if !user32.CheckMenuRadioItem(m.hMenu, uint32(first), uint32(last), uint32(index), winuser.MF_BYPOSITION) {
			return newError("CheckMenuRadioItem failed")
		}
	}

	return nil
}

func (m *Menu) onActionVisibleChanged(action *Action) error {
	if !action.IsSeparator() {
		defer m.actions.updateSeparatorVisibility()
	}

	if action.Visible() {
		return m.insertAction(action, true)
	}

	return m.removeAction(action, true)
}

func (m *Menu) insertAction(action *Action, visibleChanged bool) (err error) {
	m.handleDefaultState(action)

	if !visibleChanged {
		action.addChangedHandler(m)
		defer func() {
			if err != nil {
				action.removeChangedHandler(m)
			}
		}()
	}

	if !action.Visible() {
		return
	}

	index := m.actions.indexInObserver(action)

	var mii winuser.MENUITEMINFO

	m.initMenuItemInfoFromAction(&mii, action)

	if !user32.InsertMenuItem(m.hMenu, uint32(index), true, &mii) {
		return newError("InsertMenuItem failed")
	}

	if action.Default() {
		user32.SetMenuDefaultItem(m.hMenu, uint32(m.actions.indexInObserver(action)), true)
	}

	menu := action.menu
	if menu != nil {
		menu.window = m.window
	}

	m.ensureMenuBarRedrawn()

	return
}

func (m *Menu) removeAction(action *Action, visibleChanged bool) error {
	index := m.actions.indexInObserver(action)

	if !user32.RemoveMenu(m.hMenu, uint32(index), winuser.MF_BYPOSITION) {
		return lastError("RemoveMenu")
	}

	if !visibleChanged {
		action.removeChangedHandler(m)
	}

	m.ensureMenuBarRedrawn()

	return nil
}

func (m *Menu) ensureMenuBarRedrawn() {
	if m.window != nil {
		if mw, ok := m.window.(*MainWindow); ok && mw != nil {
			user32.DrawMenuBar(mw.Handle())
		}
	}
}

func (m *Menu) onInsertedAction(action *Action) error {
	return m.insertAction(action, false)
}

func (m *Menu) onRemovingAction(action *Action) error {
	return m.removeAction(action, false)
}

func (m *Menu) onClearingActions() error {
	for i := m.actions.Len() - 1; i >= 0; i-- {
		if action := m.actions.At(i); action.Visible() {
			if err := m.onRemovingAction(action); err != nil {
				return err
			}
		}
	}

	return nil
}

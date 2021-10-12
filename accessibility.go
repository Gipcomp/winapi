// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/oleacc"
	"github.com/Gipcomp/win32/oleaut32"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/win"
	"github.com/Gipcomp/winapi/errs"
)

// AccState enum defines the state of the window/control
type AccState int32

// Window/control states
const (
	AccStateNormal          AccState = oleacc.STATE_SYSTEM_NORMAL
	AccStateUnavailable     AccState = oleacc.STATE_SYSTEM_UNAVAILABLE
	AccStateSelected        AccState = oleacc.STATE_SYSTEM_SELECTED
	AccStateFocused         AccState = oleacc.STATE_SYSTEM_FOCUSED
	AccStatePressed         AccState = oleacc.STATE_SYSTEM_PRESSED
	AccStateChecked         AccState = oleacc.STATE_SYSTEM_CHECKED
	AccStateMixed           AccState = oleacc.STATE_SYSTEM_MIXED
	AccStateIndeterminate   AccState = oleacc.STATE_SYSTEM_INDETERMINATE
	AccStateReadonly        AccState = oleacc.STATE_SYSTEM_READONLY
	AccStateHotTracked      AccState = oleacc.STATE_SYSTEM_HOTTRACKED
	AccStateDefault         AccState = oleacc.STATE_SYSTEM_DEFAULT
	AccStateExpanded        AccState = oleacc.STATE_SYSTEM_EXPANDED
	AccStateCollapsed       AccState = oleacc.STATE_SYSTEM_COLLAPSED
	AccStateBusy            AccState = oleacc.STATE_SYSTEM_BUSY
	AccStateFloating        AccState = oleacc.STATE_SYSTEM_FLOATING
	AccStateMarqueed        AccState = oleacc.STATE_SYSTEM_MARQUEED
	AccStateAnimated        AccState = oleacc.STATE_SYSTEM_ANIMATED
	AccStateInvisible       AccState = oleacc.STATE_SYSTEM_INVISIBLE
	AccStateOffscreen       AccState = oleacc.STATE_SYSTEM_OFFSCREEN
	AccStateSizeable        AccState = oleacc.STATE_SYSTEM_SIZEABLE
	AccStateMoveable        AccState = oleacc.STATE_SYSTEM_MOVEABLE
	AccStateSelfVoicing     AccState = oleacc.STATE_SYSTEM_SELFVOICING
	AccStateFocusable       AccState = oleacc.STATE_SYSTEM_FOCUSABLE
	AccStateSelectable      AccState = oleacc.STATE_SYSTEM_SELECTABLE
	AccStateLinked          AccState = oleacc.STATE_SYSTEM_LINKED
	AccStateTraversed       AccState = oleacc.STATE_SYSTEM_TRAVERSED
	AccStateMultiselectable AccState = oleacc.STATE_SYSTEM_MULTISELECTABLE
	AccStateExtselectable   AccState = oleacc.STATE_SYSTEM_EXTSELECTABLE
	AccStateAlertLow        AccState = oleacc.STATE_SYSTEM_ALERT_LOW
	AccStateAlertMedium     AccState = oleacc.STATE_SYSTEM_ALERT_MEDIUM
	AccStateAlertHigh       AccState = oleacc.STATE_SYSTEM_ALERT_HIGH
	AccStateProtected       AccState = oleacc.STATE_SYSTEM_PROTECTED
	AccStateHasPopup        AccState = oleacc.STATE_SYSTEM_HASPOPUP
	AccStateValid           AccState = oleacc.STATE_SYSTEM_VALID
)

// AccRole enum defines the role of the window/control in UI.
type AccRole int32

// Window/control system roles
const (
	AccRoleTitlebar           AccRole = oleacc.ROLE_SYSTEM_TITLEBAR
	AccRoleMenubar            AccRole = oleacc.ROLE_SYSTEM_MENUBAR
	AccRoleScrollbar          AccRole = oleacc.ROLE_SYSTEM_SCROLLBAR
	AccRoleGrip               AccRole = oleacc.ROLE_SYSTEM_GRIP
	AccRoleSound              AccRole = oleacc.ROLE_SYSTEM_SOUND
	AccRoleCursor             AccRole = oleacc.ROLE_SYSTEM_CURSOR
	AccRoleCaret              AccRole = oleacc.ROLE_SYSTEM_CARET
	AccRoleAlert              AccRole = oleacc.ROLE_SYSTEM_ALERT
	AccRoleWindow             AccRole = oleacc.ROLE_SYSTEM_WINDOW
	AccRoleClient             AccRole = oleacc.ROLE_SYSTEM_CLIENT
	AccRoleMenuPopup          AccRole = oleacc.ROLE_SYSTEM_MENUPOPUP
	AccRoleMenuItem           AccRole = oleacc.ROLE_SYSTEM_MENUITEM
	AccRoleTooltip            AccRole = oleacc.ROLE_SYSTEM_TOOLTIP
	AccRoleApplication        AccRole = oleacc.ROLE_SYSTEM_APPLICATION
	AccRoleDocument           AccRole = oleacc.ROLE_SYSTEM_DOCUMENT
	AccRolePane               AccRole = oleacc.ROLE_SYSTEM_PANE
	AccRoleChart              AccRole = oleacc.ROLE_SYSTEM_CHART
	AccRoleDialog             AccRole = oleacc.ROLE_SYSTEM_DIALOG
	AccRoleBorder             AccRole = oleacc.ROLE_SYSTEM_BORDER
	AccRoleGrouping           AccRole = oleacc.ROLE_SYSTEM_GROUPING
	AccRoleSeparator          AccRole = oleacc.ROLE_SYSTEM_SEPARATOR
	AccRoleToolbar            AccRole = oleacc.ROLE_SYSTEM_TOOLBAR
	AccRoleStatusbar          AccRole = oleacc.ROLE_SYSTEM_STATUSBAR
	AccRoleTable              AccRole = oleacc.ROLE_SYSTEM_TABLE
	AccRoleColumnHeader       AccRole = oleacc.ROLE_SYSTEM_COLUMNHEADER
	AccRoleRowHeader          AccRole = oleacc.ROLE_SYSTEM_ROWHEADER
	AccRoleColumn             AccRole = oleacc.ROLE_SYSTEM_COLUMN
	AccRoleRow                AccRole = oleacc.ROLE_SYSTEM_ROW
	AccRoleCell               AccRole = oleacc.ROLE_SYSTEM_CELL
	AccRoleLink               AccRole = oleacc.ROLE_SYSTEM_LINK
	AccRoleHelpBalloon        AccRole = oleacc.ROLE_SYSTEM_HELPBALLOON
	AccRoleCharacter          AccRole = oleacc.ROLE_SYSTEM_CHARACTER
	AccRoleList               AccRole = oleacc.ROLE_SYSTEM_LIST
	AccRoleListItem           AccRole = oleacc.ROLE_SYSTEM_LISTITEM
	AccRoleOutline            AccRole = oleacc.ROLE_SYSTEM_OUTLINE
	AccRoleOutlineItem        AccRole = oleacc.ROLE_SYSTEM_OUTLINEITEM
	AccRolePagetab            AccRole = oleacc.ROLE_SYSTEM_PAGETAB
	AccRolePropertyPage       AccRole = oleacc.ROLE_SYSTEM_PROPERTYPAGE
	AccRoleIndicator          AccRole = oleacc.ROLE_SYSTEM_INDICATOR
	AccRoleGraphic            AccRole = oleacc.ROLE_SYSTEM_GRAPHIC
	AccRoleStatictext         AccRole = oleacc.ROLE_SYSTEM_STATICTEXT
	AccRoleText               AccRole = oleacc.ROLE_SYSTEM_TEXT
	AccRolePushbutton         AccRole = oleacc.ROLE_SYSTEM_PUSHBUTTON
	AccRoleCheckbutton        AccRole = oleacc.ROLE_SYSTEM_CHECKBUTTON
	AccRoleRadiobutton        AccRole = oleacc.ROLE_SYSTEM_RADIOBUTTON
	AccRoleCombobox           AccRole = oleacc.ROLE_SYSTEM_COMBOBOX
	AccRoleDroplist           AccRole = oleacc.ROLE_SYSTEM_DROPLIST
	AccRoleProgressbar        AccRole = oleacc.ROLE_SYSTEM_PROGRESSBAR
	AccRoleDial               AccRole = oleacc.ROLE_SYSTEM_DIAL
	AccRoleHotkeyfield        AccRole = oleacc.ROLE_SYSTEM_HOTKEYFIELD
	AccRoleSlider             AccRole = oleacc.ROLE_SYSTEM_SLIDER
	AccRoleSpinbutton         AccRole = oleacc.ROLE_SYSTEM_SPINBUTTON
	AccRoleDiagram            AccRole = oleacc.ROLE_SYSTEM_DIAGRAM
	AccRoleAnimation          AccRole = oleacc.ROLE_SYSTEM_ANIMATION
	AccRoleEquation           AccRole = oleacc.ROLE_SYSTEM_EQUATION
	AccRoleButtonDropdown     AccRole = oleacc.ROLE_SYSTEM_BUTTONDROPDOWN
	AccRoleButtonMenu         AccRole = oleacc.ROLE_SYSTEM_BUTTONMENU
	AccRoleButtonDropdownGrid AccRole = oleacc.ROLE_SYSTEM_BUTTONDROPDOWNGRID
	AccRoleWhitespace         AccRole = oleacc.ROLE_SYSTEM_WHITESPACE
	AccRolePageTabList        AccRole = oleacc.ROLE_SYSTEM_PAGETABLIST
	AccRoleClock              AccRole = oleacc.ROLE_SYSTEM_CLOCK
	AccRoleSplitButton        AccRole = oleacc.ROLE_SYSTEM_SPLITBUTTON
	AccRoleIPAddress          AccRole = oleacc.ROLE_SYSTEM_IPADDRESS
	AccRoleOutlineButton      AccRole = oleacc.ROLE_SYSTEM_OUTLINEBUTTON
)

// Accessibility provides basic Dynamic Annotation of windows and controls.
type Accessibility struct {
	wb *WindowBase
}

// SetAccelerator sets window accelerator name using Dynamic Annotation.
func (a *Accessibility) SetAccelerator(acc string) error {
	return a.accSetPropertyStr(a.wb.hWnd, &oleacc.PROPID_ACC_KEYBOARDSHORTCUT, user32.EVENT_OBJECT_ACCELERATORCHANGE, acc)
}

// SetDefaultAction sets window default action using Dynamic Annotation.
func (a *Accessibility) SetDefaultAction(defAction string) error {
	return a.accSetPropertyStr(a.wb.hWnd, &oleacc.PROPID_ACC_DEFAULTACTION, user32.EVENT_OBJECT_DEFACTIONCHANGE, defAction)
}

// SetDescription sets window description using Dynamic Annotation.
func (a *Accessibility) SetDescription(acc string) error {
	return a.accSetPropertyStr(a.wb.hWnd, &oleacc.PROPID_ACC_DESCRIPTION, user32.EVENT_OBJECT_DESCRIPTIONCHANGE, acc)
}

// SetHelp sets window help using Dynamic Annotation.
func (a *Accessibility) SetHelp(help string) error {
	return a.accSetPropertyStr(a.wb.hWnd, &oleacc.PROPID_ACC_HELP, user32.EVENT_OBJECT_HELPCHANGE, help)
}

// SetName sets window name using Dynamic Annotation.
func (a *Accessibility) SetName(name string) error {
	return a.accSetPropertyStr(a.wb.hWnd, &oleacc.PROPID_ACC_NAME, user32.EVENT_OBJECT_NAMECHANGE, name)
}

// SetRole sets window role using Dynamic Annotation. The role must be set when the window is
// created and is not to be modified later.
func (a *Accessibility) SetRole(role AccRole) error {
	return a.accSetPropertyInt(a.wb.hWnd, &oleacc.PROPID_ACC_ROLE, 0, int32(role))
}

// SetRoleMap sets window role map using Dynamic Annotation. The role map must be set when the
// window is created and is not to be modified later.
func (a *Accessibility) SetRoleMap(roleMap string) error {
	return a.accSetPropertyStr(a.wb.hWnd, &oleacc.PROPID_ACC_ROLEMAP, 0, roleMap)
}

// SetState sets window state using Dynamic Annotation.
func (a *Accessibility) SetState(state AccState) error {
	return a.accSetPropertyInt(a.wb.hWnd, &oleacc.PROPID_ACC_STATE, user32.EVENT_OBJECT_STATECHANGE, int32(state))
}

// SetStateMap sets window state map using Dynamic Annotation. The state map must be set when
// the window is created and is not to be modified later.
func (a *Accessibility) SetStateMap(stateMap string) error {
	return a.accSetPropertyStr(a.wb.hWnd, &oleacc.PROPID_ACC_STATEMAP, 0, stateMap)
}

// SetValueMap sets window value map using Dynamic Annotation. The value map must be set when
// the window is created and is not to be modified later.
func (a *Accessibility) SetValueMap(valueMap string) error {
	return a.accSetPropertyStr(a.wb.hWnd, &oleacc.PROPID_ACC_VALUEMAP, 0, valueMap)
}

// accSetPropertyInt sets integer window property for Dynamic Annotation.
func (a *Accessibility) accSetPropertyInt(hwnd handle.HWND, idProp *oleacc.MSAAPROPID, event uint32, value int32) error {
	accPropServices := a.wb.group.accessibilityServices()
	if accPropServices == nil {
		return errs.NewError("Dynamic Annotation not available")
	}
	var v oleaut32.VARIANT
	v.SetLong(value)
	hr := accPropServices.SetHwndProp(hwnd, user32.OBJID_CLIENT, user32.CHILDID_SELF, idProp, &v)
	if win.FAILED(hr) {
		return errs.ErrorFromHRESULT("IAccPropServices.SetHwndProp", hr)
	}
	if user32.EVENT_OBJECT_CREATE <= event && event <= user32.EVENT_OBJECT_END {
		user32.NotifyWinEvent(event, hwnd, user32.OBJID_CLIENT, user32.CHILDID_SELF)
	}
	return nil
}

// accSetPropertyStr sets string window property for Dynamic Annotation.
func (a *Accessibility) accSetPropertyStr(hwnd handle.HWND, idProp *oleacc.MSAAPROPID, event uint32, value string) error {
	accPropServices := a.wb.group.accessibilityServices()
	if accPropServices == nil {
		return errs.NewError("Dynamic Annotation not available")
	}
	hr := accPropServices.SetHwndPropStr(hwnd, user32.OBJID_CLIENT, user32.CHILDID_SELF, idProp, value)
	if win.FAILED(hr) {
		return errs.ErrorFromHRESULT("IAccPropServices.SetHwndPropStr", hr)
	}
	if user32.EVENT_OBJECT_CREATE <= event && event <= user32.EVENT_OBJECT_END {
		user32.NotifyWinEvent(event, hwnd, user32.OBJID_CLIENT, user32.CHILDID_SELF)
	}
	return nil
}

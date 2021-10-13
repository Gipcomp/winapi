// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import "github.com/Gipcomp/winapi"

// AccState enum defines the state of the window/control
type AccState int32

// Window/control states
const (
	AccStateNormal          = AccState(winapi.AccStateNormal)
	AccStateUnavailable     = AccState(winapi.AccStateUnavailable)
	AccStateSelected        = AccState(winapi.AccStateSelected)
	AccStateFocused         = AccState(winapi.AccStateFocused)
	AccStatePressed         = AccState(winapi.AccStatePressed)
	AccStateChecked         = AccState(winapi.AccStateChecked)
	AccStateMixed           = AccState(winapi.AccStateMixed)
	AccStateIndeterminate   = AccState(winapi.AccStateIndeterminate)
	AccStateReadonly        = AccState(winapi.AccStateReadonly)
	AccStateHotTracked      = AccState(winapi.AccStateHotTracked)
	AccStateDefault         = AccState(winapi.AccStateDefault)
	AccStateExpanded        = AccState(winapi.AccStateExpanded)
	AccStateCollapsed       = AccState(winapi.AccStateCollapsed)
	AccStateBusy            = AccState(winapi.AccStateBusy)
	AccStateFloating        = AccState(winapi.AccStateFloating)
	AccStateMarqueed        = AccState(winapi.AccStateMarqueed)
	AccStateAnimated        = AccState(winapi.AccStateAnimated)
	AccStateInvisible       = AccState(winapi.AccStateInvisible)
	AccStateOffscreen       = AccState(winapi.AccStateOffscreen)
	AccStateSizeable        = AccState(winapi.AccStateSizeable)
	AccStateMoveable        = AccState(winapi.AccStateMoveable)
	AccStateSelfVoicing     = AccState(winapi.AccStateSelfVoicing)
	AccStateFocusable       = AccState(winapi.AccStateFocusable)
	AccStateSelectable      = AccState(winapi.AccStateSelectable)
	AccStateLinked          = AccState(winapi.AccStateLinked)
	AccStateTraversed       = AccState(winapi.AccStateTraversed)
	AccStateMultiselectable = AccState(winapi.AccStateMultiselectable)
	AccStateExtselectable   = AccState(winapi.AccStateExtselectable)
	AccStateAlertLow        = AccState(winapi.AccStateAlertLow)
	AccStateAlertMedium     = AccState(winapi.AccStateAlertMedium)
	AccStateAlertHigh       = AccState(winapi.AccStateAlertHigh)
	AccStateProtected       = AccState(winapi.AccStateProtected)
	AccStateHasPopup        = AccState(winapi.AccStateHasPopup)
	AccStateValid           = AccState(winapi.AccStateValid)
)

// AccRole enum defines the role of the window/control in UI.
type AccRole int32

// Window/control system roles
const (
	AccRoleTitlebar           = AccRole(winapi.AccRoleTitlebar)
	AccRoleMenubar            = AccRole(winapi.AccRoleMenubar)
	AccRoleScrollbar          = AccRole(winapi.AccRoleScrollbar)
	AccRoleGrip               = AccRole(winapi.AccRoleGrip)
	AccRoleSound              = AccRole(winapi.AccRoleSound)
	AccRoleCursor             = AccRole(winapi.AccRoleCursor)
	AccRoleCaret              = AccRole(winapi.AccRoleCaret)
	AccRoleAlert              = AccRole(winapi.AccRoleAlert)
	AccRoleWindow             = AccRole(winapi.AccRoleWindow)
	AccRoleClient             = AccRole(winapi.AccRoleClient)
	AccRoleMenuPopup          = AccRole(winapi.AccRoleMenuPopup)
	AccRoleMenuItem           = AccRole(winapi.AccRoleMenuItem)
	AccRoleTooltip            = AccRole(winapi.AccRoleTooltip)
	AccRoleApplication        = AccRole(winapi.AccRoleApplication)
	AccRoleDocument           = AccRole(winapi.AccRoleDocument)
	AccRolePane               = AccRole(winapi.AccRolePane)
	AccRoleChart              = AccRole(winapi.AccRoleChart)
	AccRoleDialog             = AccRole(winapi.AccRoleDialog)
	AccRoleBorder             = AccRole(winapi.AccRoleBorder)
	AccRoleGrouping           = AccRole(winapi.AccRoleGrouping)
	AccRoleSeparator          = AccRole(winapi.AccRoleSeparator)
	AccRoleToolbar            = AccRole(winapi.AccRoleToolbar)
	AccRoleStatusbar          = AccRole(winapi.AccRoleStatusbar)
	AccRoleTable              = AccRole(winapi.AccRoleTable)
	AccRoleColumnHeader       = AccRole(winapi.AccRoleColumnHeader)
	AccRoleRowHeader          = AccRole(winapi.AccRoleRowHeader)
	AccRoleColumn             = AccRole(winapi.AccRoleColumn)
	AccRoleRow                = AccRole(winapi.AccRoleRow)
	AccRoleCell               = AccRole(winapi.AccRoleCell)
	AccRoleLink               = AccRole(winapi.AccRoleLink)
	AccRoleHelpBalloon        = AccRole(winapi.AccRoleHelpBalloon)
	AccRoleCharacter          = AccRole(winapi.AccRoleCharacter)
	AccRoleList               = AccRole(winapi.AccRoleList)
	AccRoleListItem           = AccRole(winapi.AccRoleListItem)
	AccRoleOutline            = AccRole(winapi.AccRoleOutline)
	AccRoleOutlineItem        = AccRole(winapi.AccRoleOutlineItem)
	AccRolePagetab            = AccRole(winapi.AccRolePagetab)
	AccRolePropertyPage       = AccRole(winapi.AccRolePropertyPage)
	AccRoleIndicator          = AccRole(winapi.AccRoleIndicator)
	AccRoleGraphic            = AccRole(winapi.AccRoleGraphic)
	AccRoleStatictext         = AccRole(winapi.AccRoleStatictext)
	AccRoleText               = AccRole(winapi.AccRoleText)
	AccRolePushbutton         = AccRole(winapi.AccRolePushbutton)
	AccRoleCheckbutton        = AccRole(winapi.AccRoleCheckbutton)
	AccRoleRadiobutton        = AccRole(winapi.AccRoleRadiobutton)
	AccRoleCombobox           = AccRole(winapi.AccRoleCombobox)
	AccRoleDroplist           = AccRole(winapi.AccRoleDroplist)
	AccRoleProgressbar        = AccRole(winapi.AccRoleProgressbar)
	AccRoleDial               = AccRole(winapi.AccRoleDial)
	AccRoleHotkeyfield        = AccRole(winapi.AccRoleHotkeyfield)
	AccRoleSlider             = AccRole(winapi.AccRoleSlider)
	AccRoleSpinbutton         = AccRole(winapi.AccRoleSpinbutton)
	AccRoleDiagram            = AccRole(winapi.AccRoleDiagram)
	AccRoleAnimation          = AccRole(winapi.AccRoleAnimation)
	AccRoleEquation           = AccRole(winapi.AccRoleEquation)
	AccRoleButtonDropdown     = AccRole(winapi.AccRoleButtonDropdown)
	AccRoleButtonMenu         = AccRole(winapi.AccRoleButtonMenu)
	AccRoleButtonDropdownGrid = AccRole(winapi.AccRoleButtonDropdownGrid)
	AccRoleWhitespace         = AccRole(winapi.AccRoleWhitespace)
	AccRolePageTabList        = AccRole(winapi.AccRolePageTabList)
	AccRoleClock              = AccRole(winapi.AccRoleClock)
	AccRoleSplitButton        = AccRole(winapi.AccRoleSplitButton)
	AccRoleIPAddress          = AccRole(winapi.AccRoleIPAddress)
	AccRoleOutlineButton      = AccRole(winapi.AccRoleOutlineButton)
)

// Accessibility properties
type Accessibility struct {
	Accelerator   string
	DefaultAction string
	Description   string
	Help          string
	Name          string
	Role          AccRole
	RoleMap       string
	State         AccState
	StateMap      string
	ValueMap      string
}

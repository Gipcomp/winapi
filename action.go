// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import "github.com/Gipcomp/winapi/errs"

type actionChangedHandler interface {
	onActionChanged(action *Action) error
	onActionVisibleChanged(action *Action) error
}

var (
	// ISSUE: When pressing enter resp. escape,
	// WM_COMMAND with wParam=1 resp. 2 is sent.
	// Maybe there is more to consider.
	nextActionId    uint16 = 3
	actionsById            = make(map[uint16]*Action)
	shortcut2Action        = make(map[Shortcut]*Action)
)

type Action struct {
	menu                          *Menu
	triggeredPublisher            EventPublisher
	changedHandlers               []actionChangedHandler
	text                          string
	toolTip                       string
	image                         Image
	checkedCondition              Condition
	checkedConditionChangedHandle int
	defaultCondition              Condition
	defaultConditionChangedHandle int
	enabledCondition              Condition
	enabledConditionChangedHandle int
	visibleCondition              Condition
	visibleConditionChangedHandle int
	refCount                      int
	shortcut                      Shortcut
	enabled                       bool
	visible                       bool
	checkable                     bool
	checked                       bool
	defawlt                       bool
	exclusive                     bool
	id                            uint16
}

func NewAction() *Action {
	a := &Action{
		enabled: true,
		id:      nextActionId,
		visible: true,
	}

	actionsById[a.id] = a

	nextActionId++

	return a
}

func NewMenuAction(menu *Menu) *Action {
	a := NewAction()
	a.menu = menu

	return a
}

func NewSeparatorAction() *Action {
	return &Action{
		enabled: true,
		visible: true,
	}
}

func (a *Action) addRef() {
	a.refCount++
}

func (a *Action) release() {
	a.refCount--

	if a.refCount == 0 {
		a.SetEnabledCondition(nil)
		a.SetVisibleCondition(nil)

		if a.menu != nil {
			a.menu.actions.Clear()
			a.menu.Dispose()
		}

		delete(actionsById, a.id)
		delete(shortcut2Action, a.shortcut)
	}
}

func (a *Action) Menu() *Menu {
	return a.menu
}

func (a *Action) Checkable() bool {
	return a.checkable
}

func (a *Action) SetCheckable(value bool) (err error) {
	if value != a.checkable {
		old := a.checkable

		a.checkable = value

		if err = a.raiseChanged(); err != nil {
			a.checkable = old
			a.raiseChanged()
		}
	}

	return
}

func (a *Action) Checked() bool {
	return a.checked
}

func (a *Action) SetChecked(value bool) (err error) {
	if a.checkedCondition != nil {
		if bp, ok := a.checkedCondition.(*boolProperty); ok {
			if err := bp.Set(value); err != nil {
				return err
			}
		} else {
			return errs.NewError("CheckedCondition != nil")
		}
	}

	if value != a.checked {
		old := a.checked

		a.checked = value

		if err = a.raiseChanged(); err != nil {
			a.checked = old
			a.raiseChanged()
		}
	}

	return
}

func (a *Action) CheckedCondition() Condition {
	return a.checkedCondition
}

func (a *Action) SetCheckedCondition(c Condition) {
	if a.checkedCondition != nil {
		a.checkedCondition.Changed().Detach(a.checkedConditionChangedHandle)
	}

	a.checkedCondition = c

	if c != nil {
		a.checked = c.Satisfied()

		a.checkedConditionChangedHandle = c.Changed().Attach(func() {
			if a.checked != c.Satisfied() {
				a.checked = !a.checked

				a.raiseChanged()
			}
		})
	}

	a.raiseChanged()
}

func (a *Action) Default() bool {
	return a.defawlt
}

func (a *Action) SetDefault(value bool) (err error) {
	if a.defaultCondition != nil {
		if bp, ok := a.defaultCondition.(*boolProperty); ok {
			if err := bp.Set(value); err != nil {
				return err
			}
		} else {
			return errs.NewError("DefaultCondition != nil")
		}
	}

	if value != a.defawlt {
		old := a.defawlt

		a.defawlt = value

		if err = a.raiseChanged(); err != nil {
			a.defawlt = old
			a.raiseChanged()
		}
	}

	return
}

func (a *Action) DefaultCondition() Condition {
	return a.defaultCondition
}

func (a *Action) SetDefaultCondition(c Condition) {
	if a.defaultCondition != nil {
		a.defaultCondition.Changed().Detach(a.defaultConditionChangedHandle)
	}

	a.defaultCondition = c

	if c != nil {
		a.defawlt = c.Satisfied()

		a.defaultConditionChangedHandle = c.Changed().Attach(func() {
			if a.defawlt != c.Satisfied() {
				a.defawlt = !a.defawlt

				a.raiseChanged()
			}
		})
	}

	a.raiseChanged()
}

func (a *Action) Enabled() bool {
	return a.enabled
}

func (a *Action) SetEnabled(value bool) (err error) {
	if a.enabledCondition != nil {
		return errs.NewError("EnabledCondition != nil")
	}

	if value != a.enabled {
		old := a.enabled

		a.enabled = value

		if err = a.raiseChanged(); err != nil {
			a.enabled = old
			a.raiseChanged()
		}
	}

	return
}

func (a *Action) EnabledCondition() Condition {
	return a.enabledCondition
}

func (a *Action) SetEnabledCondition(c Condition) {
	if a.enabledCondition != nil {
		a.enabledCondition.Changed().Detach(a.enabledConditionChangedHandle)
	}

	a.enabledCondition = c

	if c != nil {
		a.enabled = c.Satisfied()

		a.enabledConditionChangedHandle = c.Changed().Attach(func() {
			if a.enabled != c.Satisfied() {
				a.enabled = !a.enabled

				a.raiseChanged()
			}
		})
	}

	a.raiseChanged()
}

func (a *Action) Exclusive() bool {
	return a.exclusive
}

func (a *Action) SetExclusive(value bool) (err error) {
	if value != a.exclusive {
		old := a.exclusive

		a.exclusive = value

		if err = a.raiseChanged(); err != nil {
			a.exclusive = old
			a.raiseChanged()
		}
	}

	return
}

func (a *Action) Image() Image {
	return a.image
}

func (a *Action) SetImage(value Image) (err error) {
	if value != a.image {
		old := a.image

		a.image = value

		if err = a.raiseChanged(); err != nil {
			a.image = old
			a.raiseChanged()
		}
	}

	return
}

func (a *Action) Shortcut() Shortcut {
	return a.shortcut
}

func (a *Action) SetShortcut(shortcut Shortcut) (err error) {
	if shortcut != a.shortcut {
		old := a.shortcut

		a.shortcut = shortcut
		defer func() {
			if err != nil {
				a.shortcut = old
			}
		}()

		if err = a.raiseChanged(); err != nil {
			a.shortcut = old
			a.raiseChanged()
		} else {
			if shortcut.Key == 0 {
				delete(shortcut2Action, old)
			} else {
				shortcut2Action[shortcut] = a
			}
		}
	}

	return
}

func (a *Action) Text() string {
	return a.text
}

func (a *Action) SetText(value string) (err error) {
	if value != a.text {
		old := a.text

		a.text = value

		if err = a.raiseChanged(); err != nil {
			a.text = old
			a.raiseChanged()
		}
	}

	return
}

func (a *Action) IsSeparator() bool {
	return a.id == 0 || a.text == "-"
}

func (a *Action) ToolTip() string {
	return a.toolTip
}

func (a *Action) SetToolTip(value string) (err error) {
	if value != a.toolTip {
		old := a.toolTip

		a.toolTip = value

		if err = a.raiseChanged(); err != nil {
			a.toolTip = old
			a.raiseChanged()
		}
	}

	return
}

func (a *Action) Visible() bool {
	return a.visible
}

func (a *Action) SetVisible(value bool) (err error) {
	if a.visibleCondition != nil {
		return errs.NewError("VisibleCondition != nil")
	}

	if value != a.visible {
		old := a.visible

		a.visible = value

		if err = a.raiseVisibleChanged(); err != nil {
			a.visible = old
			a.raiseVisibleChanged()
		}
	}

	return
}

func (a *Action) VisibleCondition() Condition {
	return a.visibleCondition
}

func (a *Action) SetVisibleCondition(c Condition) {
	if a.visibleCondition != nil {
		a.visibleCondition.Changed().Detach(a.visibleConditionChangedHandle)
	}

	a.visibleCondition = c

	if c != nil {
		a.visible = c.Satisfied()

		a.visibleConditionChangedHandle = c.Changed().Attach(func() {
			if a.visible != c.Satisfied() {
				a.visible = !a.visible

				a.raiseVisibleChanged()
			}
		})
	}

	a.raiseChanged()
}

func (a *Action) Triggered() *Event {
	return a.triggeredPublisher.Event()
}

func (a *Action) raiseTriggered() {
	if a.Checkable() {
		a.SetChecked(!a.Checked())
	}

	a.triggeredPublisher.Publish()
}

func (a *Action) addChangedHandler(handler actionChangedHandler) {
	a.changedHandlers = append(a.changedHandlers, handler)
}

func (a *Action) removeChangedHandler(handler actionChangedHandler) {
	for i, h := range a.changedHandlers {
		if h == handler {
			a.changedHandlers = append(a.changedHandlers[:i], a.changedHandlers[i+1:]...)
			break
		}
	}
}

func (a *Action) raiseChanged() error {
	for _, handler := range a.changedHandlers {
		if err := handler.onActionChanged(a); err != nil {
			return err
		}
	}

	return nil
}

func (a *Action) raiseVisibleChanged() error {
	for _, handler := range a.changedHandlers {
		if err := handler.onActionVisibleChanged(a); err != nil {
			return err
		}
	}

	return nil
}

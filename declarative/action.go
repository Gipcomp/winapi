// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"fmt"

	"github.com/Gipcomp/winapi"
)

type Shortcut struct {
	Modifiers winapi.Modifiers
	Key       winapi.Key
}

type Action struct {
	AssignTo    **winapi.Action
	Text        string
	Image       interface{}
	Checked     Property
	Enabled     Property
	Visible     Property
	Shortcut    Shortcut
	OnTriggered winapi.EventHandler
	Checkable   bool
}

func (a Action) createAction(builder *Builder, menu *winapi.Menu) (*winapi.Action, error) {
	action := winapi.NewAction()

	if a.AssignTo != nil {
		*a.AssignTo = action
	}

	if err := action.SetText(a.Text); err != nil {
		return nil, err
	}
	if err := setActionImage(action, a.Image, builder.dpi); err != nil {
		return nil, err
	}

	if err := setActionBoolOrCondition(action.SetChecked, action.SetCheckedCondition, a.Checked, "Action.Checked", builder); err != nil {
		return nil, err
	}
	if err := setActionBoolOrCondition(action.SetEnabled, action.SetEnabledCondition, a.Enabled, "Action.Enabled", builder); err != nil {
		return nil, err
	}
	if err := setActionBoolOrCondition(action.SetVisible, action.SetVisibleCondition, a.Visible, "Action.Visible", builder); err != nil {
		return nil, err
	}

	if err := action.SetCheckable(a.Checkable || action.CheckedCondition() != nil); err != nil {
		return nil, err
	}

	s := a.Shortcut
	if err := action.SetShortcut(winapi.Shortcut{s.Modifiers, s.Key}); err != nil {
		return nil, err
	}

	if a.OnTriggered != nil {
		action.Triggered().Attach(a.OnTriggered)
	}

	if menu != nil {
		if err := menu.Actions().Add(action); err != nil {
			return nil, err
		}
	}

	return action, nil
}

type ActionRef struct {
	Action **winapi.Action
}

func (ar ActionRef) createAction(builder *Builder, menu *winapi.Menu) (*winapi.Action, error) {
	if menu != nil {
		if err := menu.Actions().Add(*ar.Action); err != nil {
			return nil, err
		}
	}

	return *ar.Action, nil
}

type Menu struct {
	AssignTo       **winapi.Menu
	AssignActionTo **winapi.Action
	Text           string
	Image          interface{}
	Enabled        Property
	Visible        Property
	Items          []MenuItem
	OnTriggered    winapi.EventHandler
}

func (m Menu) createAction(builder *Builder, menu *winapi.Menu) (*winapi.Action, error) {
	subMenu, err := winapi.NewMenu()
	if err != nil {
		return nil, err
	}

	var action *winapi.Action
	if menu == nil {
		action = winapi.NewMenuAction(subMenu)
	} else if action, err = menu.Actions().AddMenu(subMenu); err != nil {
		return nil, err
	}

	if err := action.SetText(m.Text); err != nil {
		return nil, err
	}
	if err := setActionImage(action, m.Image, builder.dpi); err != nil {
		return nil, err
	}

	if err := setActionBoolOrCondition(action.SetEnabled, action.SetEnabledCondition, m.Enabled, "Menu.Enabled", builder); err != nil {
		return nil, err
	}
	if err := setActionBoolOrCondition(action.SetVisible, action.SetVisibleCondition, m.Visible, "Menu.Visible", builder); err != nil {
		return nil, err
	}

	for _, item := range m.Items {
		if _, err := item.createAction(builder, subMenu); err != nil {
			return nil, err
		}
	}

	if m.OnTriggered != nil {
		action.Triggered().Attach(m.OnTriggered)
	}

	if m.AssignActionTo != nil {
		*m.AssignActionTo = action
	}
	if m.AssignTo != nil {
		*m.AssignTo = subMenu
	}

	return action, nil
}

type Separator struct {
}

func (s Separator) createAction(builder *Builder, menu *winapi.Menu) (*winapi.Action, error) {
	action := winapi.NewSeparatorAction()

	if menu != nil {
		if err := menu.Actions().Add(action); err != nil {
			return nil, err
		}
	}

	return action, nil
}

func addToActionList(list *winapi.ActionList, actions []*winapi.Action) error {
	for _, a := range actions {
		if err := list.Add(a); err != nil {
			return err
		}
	}

	return nil
}

func setActionImage(action *winapi.Action, image interface{}, dpi int) (err error) {
	var img winapi.Image

	switch image.(type) {
	case *winapi.Bitmap:
		if img, err = winapi.BitmapFrom(image, dpi); err != nil {
			return
		}

	case winapi.ExtractableIcon, *winapi.Icon:
		if img, err = winapi.IconFrom(image, dpi); err != nil {
			return
		}

	default:
		if img, err = winapi.ImageFrom(image); err != nil {
			return
		}
	}

	return action.SetImage(img)
}

func setActionBoolOrCondition(setBool func(bool) error, setCond func(winapi.Condition), value Property, path string, builder *Builder) error {
	if value != nil {
		if b, ok := value.(bool); ok {
			if err := setBool(b); err != nil {
				return err
			}
		} else if s := builder.conditionOrProperty(value); s != nil {
			if c, ok := s.(winapi.Condition); ok {
				setCond(c)
			} else {
				return fmt.Errorf("value of invalid type bound to %s: %T", path, s)
			}
		}
	}

	return nil
}

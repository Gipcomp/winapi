// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import "github.com/Gipcomp/win32/user32"

const compositeWindowClass = `\o/ Walk_Composite_Class \o/`

func init() {
	AppendToWalkInit(func() {
		MustRegisterWindowClass(compositeWindowClass)
	})
}

type Composite struct {
	ContainerBase
}

func NewCompositeWithStyle(parent Window, style uint32) (*Composite, error) {
	c := new(Composite)
	c.children = newWidgetList(c)
	c.SetPersistent(true)

	if err := InitWidget(
		c,
		parent,
		compositeWindowClass,
		user32.WS_CHILD|user32.WS_VISIBLE|style,
		user32.WS_EX_CONTROLPARENT); err != nil {
		return nil, err
	}

	c.SetBackground(NullBrush())

	return c, nil
}

func NewComposite(parent Container) (*Composite, error) {
	return NewCompositeWithStyle(parent, 0)
}

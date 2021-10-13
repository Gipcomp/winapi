// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"path/filepath"

	"github.com/Gipcomp/winapi"
)

func tr(source string, context ...string) string {
	if translation := winapi.TranslationFunc(); translation != nil {
		return translation(source, context...)
	}

	return source
}

type Property interface{}

type bindData struct {
	expression string
	validator  Validator
}

func Bind(expression string, validators ...Validator) Property {
	bd := bindData{expression: expression}
	switch len(validators) {
	case 0:
		// nop

	case 1:
		bd.validator = validators[0]

	default:
		bd.validator = dMultiValidator{validators}
	}

	return bd
}

type SysDLLIcon struct {
	FileName string
	Index    int
	Size     int
}

func (sdi SysDLLIcon) FilePath_() string {
	root, _ := winapi.SystemPath()

	name := sdi.FileName
	if filepath.Ext(name) == "" {
		name += ".dll"
	}

	return filepath.Join(root, name)
}

func (sdi SysDLLIcon) Index_() int {
	return sdi.Index
}

func (sdi SysDLLIcon) Size_() int {
	if sdi.Size == 0 {
		return 16
	}

	return sdi.Size
}

type Brush interface {
	Create() (winapi.Brush, error)
}

type Layout interface {
	Create() (winapi.Layout, error)
}

type Widget interface {
	Create(builder *Builder) error
}

type MenuItem interface {
	createAction(builder *Builder, menu *winapi.Menu) (*winapi.Action, error)
}

type Validator interface {
	Create() (winapi.Validator, error)
}

type ErrorPresenter interface {
	Create() (winapi.ErrorPresenter, error)
}

type ErrorPresenterRef struct {
	ErrorPresenter *winapi.ErrorPresenter
}

func (epr ErrorPresenterRef) Create() (winapi.ErrorPresenter, error) {
	if epr.ErrorPresenter != nil {
		return *epr.ErrorPresenter, nil
	}

	return nil, nil
}

type ToolTipErrorPresenter struct {
}

func (ToolTipErrorPresenter) Create() (winapi.ErrorPresenter, error) {
	return winapi.NewToolTipErrorPresenter()
}

type formInfo struct {
	// Window

	Accessibility      Accessibility
	Background         Brush
	ContextMenuItems   []MenuItem
	DoubleBuffering    bool
	Enabled            Property
	Font               Font
	MaxSize            Size
	MinSize            Size
	Name               string
	OnBoundsChanged    winapi.EventHandler
	OnKeyDown          winapi.KeyEventHandler
	OnKeyPress         winapi.KeyEventHandler
	OnKeyUp            winapi.KeyEventHandler
	OnMouseDown        winapi.MouseEventHandler
	OnMouseMove        winapi.MouseEventHandler
	OnMouseUp          winapi.MouseEventHandler
	OnSizeChanged      winapi.EventHandler
	RightToLeftReading bool
	ToolTipText        string
	Visible            Property

	// Container

	Children   []Widget
	DataBinder DataBinder
	Layout     Layout

	// Form

	Icon  Property
	Title Property
}

func (formInfo) Create(builder *Builder) error {
	return nil
}

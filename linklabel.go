// Copyright 2017 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"syscall"
	"unsafe"

	"github.com/Gipcomp/win32/comctl32"
	"github.com/Gipcomp/win32/commctrl"
	"github.com/Gipcomp/win32/gdi32"
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/win"
)

type LinkLabel struct {
	WidgetBase
	textChangedPublisher   EventPublisher
	linkActivatedPublisher LinkLabelLinkEventPublisher
}

func NewLinkLabel(parent Container) (*LinkLabel, error) {
	ll := new(LinkLabel)

	if err := InitWidget(
		ll,
		parent,
		"SysLink",
		user32.WS_TABSTOP|user32.WS_VISIBLE,
		0); err != nil {
		return nil, err
	}

	ll.SetBackground(nullBrushSingleton)

	ll.MustRegisterProperty("Text", NewProperty(
		func() interface{} {
			return ll.Text()
		},
		func(v interface{}) error {
			return ll.SetText(assertStringOr(v, ""))
		},
		ll.textChangedPublisher.Event()))

	return ll, nil
}

func (ll *LinkLabel) Text() string {
	return ll.text()
}

func (ll *LinkLabel) SetText(value string) error {
	if value == ll.Text() {
		return nil
	}

	if err := ll.setText(value); err != nil {
		return err
	}

	ll.RequestLayout()

	return nil
}

func (ll *LinkLabel) LinkActivated() *LinkLabelLinkEvent {
	return ll.linkActivatedPublisher.Event()
}

func (ll *LinkLabel) WndProc(hwnd handle.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case user32.WM_NOTIFY:
		nml := (*commctrl.NMLINK)(unsafe.Pointer(lParam))

		switch nml.Hdr.Code {
		case comctl32.NM_CLICK, comctl32.NM_RETURN:
			link := &LinkLabelLink{
				ll:    ll,
				index: int(nml.Item.ILink),
				id:    syscall.UTF16ToString(nml.Item.SzID[:]),
				url:   syscall.UTF16ToString(nml.Item.SzUrl[:]),
			}

			ll.linkActivatedPublisher.Publish(link)
		}

	case user32.WM_KILLFOCUS:
		ll.ensureStyleBits(user32.WS_TABSTOP, true)

	case user32.WM_SETTEXT:
		ll.textChangedPublisher.Publish()

	case user32.WM_WINDOWPOSCHANGED:
		wp := (*user32.WINDOWPOS)(unsafe.Pointer(lParam))

		if wp.Flags&user32.SWP_NOSIZE != 0 {
			break
		}

		ll.Invalidate()
	}

	return ll.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}

type LinkLabelLinkEventHandler func(link *LinkLabelLink)

type LinkLabelLinkEvent struct {
	handlers []LinkLabelLinkEventHandler
}

func (e *LinkLabelLinkEvent) Attach(handler LinkLabelLinkEventHandler) int {
	for i, h := range e.handlers {
		if h == nil {
			e.handlers[i] = handler
			return i
		}
	}

	e.handlers = append(e.handlers, handler)
	return len(e.handlers) - 1
}

func (e *LinkLabelLinkEvent) Detach(handle int) {
	e.handlers[handle] = nil
}

type LinkLabelLinkEventPublisher struct {
	event LinkLabelLinkEvent
}

func (p *LinkLabelLinkEventPublisher) Event() *LinkLabelLinkEvent {
	return &p.event
}

func (p *LinkLabelLinkEventPublisher) Publish(link *LinkLabelLink) {
	for _, handler := range p.event.handlers {
		if handler != nil {
			handler(link)
		}
	}
}

type LinkLabelLink struct {
	ll    *LinkLabel
	index int
	id    string
	url   string
}

func (lll *LinkLabelLink) Index() int {
	return lll.index
}

func (lll *LinkLabelLink) Id() string {
	return lll.id
}

func (lll *LinkLabelLink) URL() string {
	return lll.url
}

func (lll *LinkLabelLink) Enabled() (bool, error) {
	return lll.hasState(commctrl.LIS_ENABLED)
}

func (lll *LinkLabelLink) SetEnabled(enabled bool) error {
	return lll.setState(commctrl.LIS_ENABLED, enabled)
}

func (lll *LinkLabelLink) Focused() (bool, error) {
	return lll.hasState(commctrl.LIS_FOCUSED)
}

func (lll *LinkLabelLink) SetFocused(focused bool) error {
	return lll.setState(commctrl.LIS_FOCUSED, focused)
}

func (lll *LinkLabelLink) Visited() (bool, error) {
	return lll.hasState(commctrl.LIS_VISITED)
}

func (lll *LinkLabelLink) SetVisited(visited bool) error {
	return lll.setState(commctrl.LIS_VISITED, visited)
}

func (lll *LinkLabelLink) hasState(state uint32) (bool, error) {
	li := commctrl.LITEM{
		ILink:     int32(lll.index),
		Mask:      commctrl.LIF_ITEMINDEX | commctrl.LIF_STATE,
		StateMask: state,
	}

	if win.TRUE != lll.ll.SendMessage(commctrl.LM_GETITEM, 0, uintptr(unsafe.Pointer(&li))) {
		return false, newError("LM_GETITEM")
	}

	return li.State&state == state, nil
}

func (lll *LinkLabelLink) setState(state uint32, set bool) error {
	li := commctrl.LITEM{
		Mask:      commctrl.LIF_STATE,
		StateMask: state,
	}

	if set {
		li.State = state
	}

	li.Mask |= commctrl.LIF_ITEMINDEX
	li.ILink = int32(lll.index)

	if win.TRUE != lll.ll.SendMessage(commctrl.LM_SETITEM, 0, uintptr(unsafe.Pointer(&li))) {
		return newError("LM_SETITEM")
	}

	return nil
}

func (ll *LinkLabel) CreateLayoutItem(ctx *LayoutContext) LayoutItem {
	var s gdi32.SIZE
	ll.SendMessage(commctrl.LM_GETIDEALSIZE, uintptr(ll.IntFrom96DPI(ll.maxSize96dpi.Width)), uintptr(unsafe.Pointer(&s)))

	return &linkLabelLayoutItem{
		idealSize: sizeFromSIZE(s),
	}
}

type linkLabelLayoutItem struct {
	LayoutItemBase
	idealSize Size // in native pixels
}

func (*linkLabelLayoutItem) LayoutFlags() LayoutFlags {
	return 0
}

func (li *linkLabelLayoutItem) IdealSize() Size {
	return li.idealSize
}

func (li *linkLabelLayoutItem) MinSize() Size {
	return li.idealSize
}

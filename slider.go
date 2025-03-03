// Copyright 2016 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"strconv"

	"github.com/Gipcomp/win32/comctl32"
	"github.com/Gipcomp/win32/commctrl"
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/win"
)

type Slider struct {
	WidgetBase
	valueChangedPublisher EventPublisher
	layoutFlags           LayoutFlags
	tracking              bool
	persistent            bool
}

type SliderCfg struct {
	Orientation    Orientation
	ToolTipsHidden bool
}

func NewSlider(parent Container) (*Slider, error) {
	return NewSliderWithOrientation(parent, Horizontal)
}

func NewSliderWithOrientation(parent Container, orientation Orientation) (*Slider, error) {
	return NewSliderWithCfg(parent, &SliderCfg{Orientation: orientation})
}

func NewSliderWithCfg(parent Container, cfg *SliderCfg) (*Slider, error) {
	sl := new(Slider)

	var style uint32 = user32.WS_TABSTOP | user32.WS_VISIBLE
	if cfg.Orientation == Vertical {
		style |= comctl32.TBS_VERT
		sl.layoutFlags = ShrinkableVert | GrowableVert
	} else {
		sl.layoutFlags = ShrinkableHorz | GrowableHorz
	}
	if !cfg.ToolTipsHidden {
		style |= comctl32.TBS_TOOLTIPS
	}

	if err := InitWidget(
		sl,
		parent,
		"msctls_trackbar32",
		style,
		0); err != nil {
		return nil, err
	}

	sl.SetBackground(nullBrushSingleton)

	sl.GraphicsEffects().Add(InteractionEffect)
	sl.GraphicsEffects().Add(FocusEffect)

	sl.MustRegisterProperty("Value", NewProperty(
		func() interface{} {
			return sl.Value()
		},
		func(v interface{}) error {
			sl.SetValue(assertIntOr(v, 0))
			return nil
		},
		sl.valueChangedPublisher.Event()))

	return sl, nil
}

func (sl *Slider) MinValue() int {
	return int(sl.SendMessage(comctl32.TBM_GETRANGEMIN, 0, 0))
}

func (sl *Slider) MaxValue() int {
	return int(sl.SendMessage(comctl32.TBM_GETRANGEMAX, 0, 0))
}

func (sl *Slider) SetRange(min, max int) {
	sl.SendMessage(comctl32.TBM_SETRANGEMIN, 0, uintptr(min))
	sl.SendMessage(comctl32.TBM_SETRANGEMAX, 1, uintptr(max))
}

func (sl *Slider) Value() int {
	return int(sl.SendMessage(comctl32.TBM_GETPOS, 0, 0))
}

func (sl *Slider) SetValue(value int) {
	sl.SendMessage(comctl32.TBM_SETPOS, 1, uintptr(value))
	sl.valueChangedPublisher.Publish()
}

// ValueChanged returns an Event that can be used to track changes to Value.
func (sl *Slider) ValueChanged() *Event {
	return sl.valueChangedPublisher.Event()
}

func (sl *Slider) Persistent() bool {
	return sl.persistent
}

func (sl *Slider) SetPersistent(value bool) {
	sl.persistent = value
}

func (sl *Slider) SaveState() error {
	return sl.WriteState(strconv.Itoa(sl.Value()))
}

func (sl *Slider) RestoreState() error {
	s, err := sl.ReadState()
	if err != nil {
		return err
	}

	value, err := strconv.Atoi(s)
	if err != nil {
		return err
	}

	sl.SetValue(value)

	return nil
}

func (sl *Slider) LineSize() int {
	return int(sl.SendMessage(comctl32.TBM_GETLINESIZE, 0, 0))
}

func (sl *Slider) SetLineSize(lineSize int) {
	sl.SendMessage(comctl32.TBM_SETLINESIZE, 0, uintptr(lineSize))
}

func (sl *Slider) PageSize() int {
	return int(sl.SendMessage(comctl32.TBM_GETPAGESIZE, 0, 0))
}

func (sl *Slider) SetPageSize(pageSize int) {
	sl.SendMessage(comctl32.TBM_SETPAGESIZE, 0, uintptr(pageSize))
}

func (sl *Slider) Tracking() bool {
	return sl.tracking
}

func (sl *Slider) SetTracking(tracking bool) {
	sl.tracking = tracking
}

func (sl *Slider) WndProc(hwnd handle.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case user32.WM_HSCROLL, user32.WM_VSCROLL:
		switch win.LOWORD(uint32(wParam)) {
		case commctrl.TB_THUMBPOSITION, commctrl.TB_ENDTRACK:
			sl.valueChangedPublisher.Publish()

		case commctrl.TB_THUMBTRACK:
			if sl.tracking {
				sl.valueChangedPublisher.Publish()
			}
		}
		return 0
	}
	return sl.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}

func (*Slider) NeedsWmSize() bool {
	return true
}

func (sl *Slider) CreateLayoutItem(ctx *LayoutContext) LayoutItem {
	return &sliderLayoutItem{
		layoutFlags: sl.layoutFlags,
		idealSize:   sl.dialogBaseUnitsToPixels(Size{15, 15}),
	}
}

type sliderLayoutItem struct {
	LayoutItemBase
	layoutFlags LayoutFlags
	idealSize   Size // in native pixels
}

func (li *sliderLayoutItem) LayoutFlags() LayoutFlags {
	return li.layoutFlags
}

func (li *sliderLayoutItem) IdealSize() Size {
	return li.idealSize
}

func (li *sliderLayoutItem) MinSize() Size {
	return li.idealSize
}

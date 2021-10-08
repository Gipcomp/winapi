// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"github.com/Gipcomp/win32/comctl32"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/win"
)

type ProgressBar struct {
	WidgetBase
}

func NewProgressBar(parent Container) (*ProgressBar, error) {
	pb := new(ProgressBar)

	if err := InitWidget(
		pb,
		parent,
		"msctls_progress32",
		user32.WS_VISIBLE,
		0); err != nil {
		return nil, err
	}

	return pb, nil
}

func (pb *ProgressBar) MinValue() int {
	return int(pb.SendMessage(comctl32.PBM_GETRANGE, 1, 0))
}

func (pb *ProgressBar) MaxValue() int {
	return int(pb.SendMessage(comctl32.PBM_GETRANGE, 0, 0))
}

func (pb *ProgressBar) SetRange(min, max int) {
	pb.SendMessage(comctl32.PBM_SETRANGE32, uintptr(min), uintptr(max))
}

func (pb *ProgressBar) Value() int {
	return int(pb.SendMessage(comctl32.PBM_GETPOS, 0, 0))
}

func (pb *ProgressBar) SetValue(value int) {
	pb.SendMessage(comctl32.PBM_SETPOS, uintptr(value), 0)
}

func (pb *ProgressBar) MarqueeMode() bool {
	return pb.hasStyleBits(comctl32.PBS_MARQUEE)
}

func (pb *ProgressBar) SetMarqueeMode(marqueeMode bool) error {
	if err := pb.ensureStyleBits(comctl32.PBS_MARQUEE, marqueeMode); err != nil {
		return err
	}

	pb.SendMessage(comctl32.PBM_SETMARQUEE, uintptr(win.BoolToBOOL(marqueeMode)), 0)

	return nil
}

func (pb *ProgressBar) CreateLayoutItem(ctx *LayoutContext) LayoutItem {
	return &progressBarLayoutItem{
		idealSize: pb.dialogBaseUnitsToPixels(Size{50, 14}),
		minSize:   pb.dialogBaseUnitsToPixels(Size{10, 14}),
	}
}

type progressBarLayoutItem struct {
	LayoutItemBase
	idealSize Size // in native pixels
	minSize   Size // in native pixels
}

func (*progressBarLayoutItem) LayoutFlags() LayoutFlags {
	return ShrinkableHorz | GrowableHorz | GreedyHorz
}

func (li *progressBarLayoutItem) IdealSize() Size {
	return li.idealSize
}

func (li *progressBarLayoutItem) MinSize() Size {
	return li.minSize
}

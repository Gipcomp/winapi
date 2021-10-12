// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/Gipcomp/win32/comctl32"
	"github.com/Gipcomp/win32/commctrl"
	"github.com/Gipcomp/win32/gdi32"
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/win"
	"github.com/Gipcomp/winapi/errs"
)

type ToolBarButtonStyle int

const (
	ToolBarButtonImageOnly ToolBarButtonStyle = iota
	ToolBarButtonTextOnly
	ToolBarButtonImageBeforeText
	ToolBarButtonImageAboveText
)

type ToolBar struct {
	WidgetBase
	imageList          *ImageList
	actions            *ActionList
	defaultButtonWidth int
	maxTextRows        int
	buttonStyle        ToolBarButtonStyle
}

func NewToolBarWithOrientationAndButtonStyle(parent Container, orientation Orientation, buttonStyle ToolBarButtonStyle) (*ToolBar, error) {
	var style uint32
	if orientation == Vertical {
		style = comctl32.CCS_VERT | comctl32.CCS_NORESIZE
	} else {
		style = commctrl.TBSTYLE_WRAPABLE
	}

	if buttonStyle != ToolBarButtonImageAboveText {
		style |= commctrl.TBSTYLE_LIST
	}

	tb := &ToolBar{
		buttonStyle: buttonStyle,
	}
	tb.actions = newActionList(tb)

	if orientation == Vertical {
		tb.defaultButtonWidth = 100
	}

	if err := InitWidget(
		tb,
		parent,
		"ToolbarWindow32",
		comctl32.CCS_NODIVIDER|commctrl.TBSTYLE_FLAT|commctrl.TBSTYLE_TOOLTIPS|style,
		0); err != nil {
		return nil, err
	}

	exStyle := tb.SendMessage(commctrl.TB_GETEXTENDEDSTYLE, 0, 0)
	exStyle |= commctrl.TBSTYLE_EX_DRAWDDARROWS | commctrl.TBSTYLE_EX_MIXEDBUTTONS
	tb.SendMessage(commctrl.TB_SETEXTENDEDSTYLE, 0, exStyle)

	return tb, nil
}

func NewToolBar(parent Container) (*ToolBar, error) {
	return NewToolBarWithOrientationAndButtonStyle(parent, Horizontal, ToolBarButtonImageOnly)
}

func NewVerticalToolBar(parent Container) (*ToolBar, error) {
	return NewToolBarWithOrientationAndButtonStyle(parent, Vertical, ToolBarButtonImageAboveText)
}

func (tb *ToolBar) Dispose() {
	tb.WidgetBase.Dispose()

	tb.actions.Clear()

	if tb.imageList != nil {
		tb.imageList.Dispose()
		tb.imageList = nil
	}
}

func (tb *ToolBar) applyFont(font *Font) {
	tb.WidgetBase.applyFont(font)

	tb.applyDefaultButtonWidth()

	tb.RequestLayout()
}

func (tb *ToolBar) ApplyDPI(dpi int) {
	tb.WidgetBase.ApplyDPI(dpi)

	var maskColor Color
	var size Size
	if tb.imageList != nil {
		maskColor = tb.imageList.maskColor
		size = SizeFrom96DPI(tb.imageList.imageSize96dpi, dpi)
	} else {
		size = SizeFrom96DPI(Size{16, 16}, dpi)
	}

	iml, err := NewImageListForDPI(size, maskColor, dpi)
	if err != nil {
		return
	}

	tb.SendMessage(commctrl.TB_SETIMAGELIST, 0, uintptr(iml.hIml))

	if tb.imageList != nil {
		tb.imageList.Dispose()
	}

	tb.imageList = iml

	for _, action := range tb.actions.actions {
		if action.image != nil {
			tb.onActionChanged(action)
		}
	}

	tb.hFont = tb.Font().handleForDPI(tb.DPI())
	setWindowFont(tb.hWnd, tb.hFont)
}

func (tb *ToolBar) Orientation() Orientation {
	style := user32.GetWindowLong(tb.hWnd, user32.GWL_STYLE)

	if style&comctl32.CCS_VERT > 0 {
		return Vertical
	}

	return Horizontal
}

func (tb *ToolBar) ButtonStyle() ToolBarButtonStyle {
	return tb.buttonStyle
}

func (tb *ToolBar) applyDefaultButtonWidth() error {
	if tb.defaultButtonWidth == 0 {
		return nil
	}

	dpi := tb.DPI()
	width := IntFrom96DPI(tb.defaultButtonWidth, dpi)

	lParam := uintptr(win.MAKELONG(uint16(width), uint16(width)))
	if tb.SendMessage(commctrl.TB_SETBUTTONWIDTH, 0, lParam) == 0 {
		return errs.NewError("SendMessage(TB_SETBUTTONWIDTH)")
	}

	size := uint32(tb.SendMessage(commctrl.TB_GETBUTTONSIZE, 0, 0))
	height := win.HIWORD(size)

	lParam = uintptr(win.MAKELONG(uint16(width), height))
	if win.FALSE == tb.SendMessage(commctrl.TB_SETBUTTONSIZE, 0, lParam) {
		return errs.NewError("SendMessage(TB_SETBUTTONSIZE)")
	}

	return nil
}

// DefaultButtonWidth returns the default button width of the ToolBar.
//
// The default value for a horizontal ToolBar is 0, resulting in automatic
// sizing behavior. For a vertical ToolBar, the default is 100 pixels.
func (tb *ToolBar) DefaultButtonWidth() int {
	return tb.defaultButtonWidth
}

// SetDefaultButtonWidth sets the default button width of the ToolBar.
//
// Calling this method affects all buttons in the ToolBar, no matter if they are
// added before or after the call. A width of 0 results in automatic sizing
// behavior. Negative values are not allowed.
func (tb *ToolBar) SetDefaultButtonWidth(width int) error {
	if width == tb.defaultButtonWidth {
		return nil
	}

	if width < 0 {
		return errs.NewError("width must be >= 0")
	}

	old := tb.defaultButtonWidth

	tb.defaultButtonWidth = width

	for _, action := range tb.actions.actions {
		if err := tb.onActionChanged(action); err != nil {
			tb.defaultButtonWidth = old

			return err
		}
	}

	return tb.applyDefaultButtonWidth()
}

func (tb *ToolBar) MaxTextRows() int {
	return tb.maxTextRows
}

func (tb *ToolBar) SetMaxTextRows(maxTextRows int) error {
	if tb.SendMessage(commctrl.TB_SETMAXTEXTROWS, uintptr(maxTextRows), 0) == 0 {
		return errs.NewError("SendMessage(TB_SETMAXTEXTROWS)")
	}

	tb.maxTextRows = maxTextRows

	return nil
}

func (tb *ToolBar) Actions() *ActionList {
	return tb.actions
}

func (tb *ToolBar) ImageList() *ImageList {
	return tb.imageList
}

func (tb *ToolBar) SetImageList(value *ImageList) {
	var hIml comctl32.HIMAGELIST

	if tb.buttonStyle != ToolBarButtonTextOnly && value != nil {
		hIml = value.hIml
	}

	tb.SendMessage(commctrl.TB_SETIMAGELIST, 0, uintptr(hIml))

	tb.imageList = value
}

func (tb *ToolBar) imageIndex(image Image) (imageIndex int32, err error) {
	if tb.imageList == nil {
		dpi := tb.DPI()
		iml, err := NewImageListForDPI(SizeFrom96DPI(Size{16, 16}, dpi), 0, dpi)
		if err != nil {
			return 0, err
		}

		tb.SetImageList(iml)
	}

	imageIndex = -1
	if image != nil {
		if imageIndex, err = tb.imageList.AddImage(image); err != nil {
			return
		}
	}

	return
}

func (tb *ToolBar) WndProc(hwnd handle.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case user32.WM_MOUSEMOVE, user32.WM_MOUSELEAVE, user32.WM_LBUTTONDOWN:
		tb.Invalidate()

	case user32.WM_COMMAND:
		switch win.HIWORD(uint32(wParam)) {
		case user32.BN_CLICKED:
			actionId := uint16(win.LOWORD(uint32(wParam)))
			if action, ok := actionsById[actionId]; ok {
				action.raiseTriggered()
				return 0
			}
		}

	case user32.WM_NOTIFY:
		nmhdr := (*user32.NMHDR)(unsafe.Pointer(lParam))

		switch int32(nmhdr.Code) {
		case commctrl.TBN_DROPDOWN:
			nmtb := (*commctrl.NMTOOLBAR)(unsafe.Pointer(lParam))
			actionId := uint16(nmtb.IItem)
			if action := actionsById[actionId]; action != nil {
				var r gdi32.RECT
				if tb.SendMessage(commctrl.TB_GETRECT, uintptr(actionId), uintptr(unsafe.Pointer(&r))) == 0 {
					break
				}

				p := gdi32.POINT{X: r.Left, Y: r.Bottom}

				if !user32.ClientToScreen(tb.hWnd, &p) {
					break
				}

				action.menu.updateItemsWithImageForWindow(tb)

				user32.TrackPopupMenuEx(
					action.menu.hMenu,
					user32.TPM_NOANIMATION,
					p.X,
					p.Y,
					tb.hWnd,
					nil)

				return commctrl.TBDDRET_DEFAULT
			}
		}

	case user32.WM_WINDOWPOSCHANGED:
		wp := (*user32.WINDOWPOS)(unsafe.Pointer(lParam))

		if wp.Flags&user32.SWP_NOSIZE != 0 {
			break
		}

		tb.SendMessage(commctrl.TB_AUTOSIZE, 0, 0)
	}

	return tb.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}

func (tb *ToolBar) initButtonForAction(action *Action, state, style *byte, image *int32, text *uintptr) (err error) {
	if tb.hasStyleBits(comctl32.CCS_VERT) {
		*state |= commctrl.TBSTATE_WRAP
	} else if tb.defaultButtonWidth == 0 {
		*style |= commctrl.BTNS_AUTOSIZE
	}

	if action.checked {
		*state |= commctrl.TBSTATE_CHECKED
	}

	if action.enabled {
		*state |= commctrl.TBSTATE_ENABLED
	}

	if action.checkable {
		*style |= commctrl.BTNS_CHECK
	}

	if action.exclusive {
		*style |= commctrl.BTNS_GROUP
	}

	if tb.buttonStyle != ToolBarButtonImageOnly && len(action.text) > 0 {
		*style |= commctrl.BTNS_SHOWTEXT
	}

	if action.menu != nil {
		if len(action.Triggered().handlers) > 0 {
			*style |= commctrl.BTNS_DROPDOWN
		} else {
			*style |= commctrl.BTNS_WHOLEDROPDOWN
		}
	}

	if action.IsSeparator() {
		*style = commctrl.BTNS_SEP
	}

	if tb.buttonStyle != ToolBarButtonTextOnly {
		if *image, err = tb.imageIndex(action.image); err != nil {
			return err
		}
	}

	var actionText string
	if s := action.shortcut; tb.buttonStyle == ToolBarButtonImageOnly && s.Key != 0 {
		actionText = fmt.Sprintf("%s (%s)", action.Text(), s.String())
	} else {
		actionText = action.Text()
	}

	if len(actionText) != 0 {
		strPtr, err := syscall.UTF16PtrFromString(actionText)
		if err != nil {
			return err
		}
		*text = uintptr(unsafe.Pointer(strPtr))
	} else if len(action.toolTip) != 0 {
		strPtr, err := syscall.UTF16PtrFromString(action.toolTip)
		if err != nil {
			return err
		}
		*text = uintptr(unsafe.Pointer(strPtr))
	}

	return
}

func (tb *ToolBar) onActionChanged(action *Action) error {
	tbbi := commctrl.TBBUTTONINFO{
		DwMask: commctrl.TBIF_IMAGE | commctrl.TBIF_STATE | commctrl.TBIF_STYLE | commctrl.TBIF_TEXT,
		IImage: comctl32.I_IMAGENONE,
	}

	tbbi.CbSize = uint32(unsafe.Sizeof(tbbi))

	if err := tb.initButtonForAction(
		action,
		&tbbi.FsState,
		&tbbi.FsStyle,
		&tbbi.IImage,
		&tbbi.PszText); err != nil {

		return err
	}

	if tb.SendMessage(
		commctrl.TB_SETBUTTONINFO,
		uintptr(action.id),
		uintptr(unsafe.Pointer(&tbbi))) == 0 {

		return errs.NewError("SendMessage(TB_SETBUTTONINFO) failed")
	}

	tb.RequestLayout()

	return nil
}

func (tb *ToolBar) onActionVisibleChanged(action *Action) error {
	if !action.IsSeparator() {
		defer tb.actions.updateSeparatorVisibility()
	}

	if action.Visible() {
		return tb.insertAction(action, true)
	}

	return tb.removeAction(action, true)
}

func (tb *ToolBar) insertAction(action *Action, visibleChanged bool) (err error) {
	if !visibleChanged {
		action.addChangedHandler(tb)
		defer func() {
			if err != nil {
				action.removeChangedHandler(tb)
			}
		}()
	}

	if !action.Visible() {
		return
	}

	index := tb.actions.indexInObserver(action)

	tbb := commctrl.TBBUTTON{
		IdCommand: int32(action.id),
	}

	if err = tb.initButtonForAction(
		action,
		&tbb.FsState,
		&tbb.FsStyle,
		&tbb.IBitmap,
		&tbb.IString); err != nil {

		return
	}

	tb.SetVisible(true)

	tb.SendMessage(commctrl.TB_BUTTONSTRUCTSIZE, uintptr(unsafe.Sizeof(tbb)), 0)

	if win.FALSE == tb.SendMessage(commctrl.TB_INSERTBUTTON, uintptr(index), uintptr(unsafe.Pointer(&tbb))) {
		return errs.NewError("SendMessage(TB_ADDBUTTONS)")
	}

	if err = tb.applyDefaultButtonWidth(); err != nil {
		return
	}

	tb.SendMessage(commctrl.TB_AUTOSIZE, 0, 0)

	tb.RequestLayout()

	return
}

func (tb *ToolBar) removeAction(action *Action, visibleChanged bool) error {
	index := tb.actions.indexInObserver(action)

	if !visibleChanged {
		action.removeChangedHandler(tb)
	}

	if tb.SendMessage(commctrl.TB_DELETEBUTTON, uintptr(index), 0) == 0 {
		return errs.NewError("SendMessage(TB_DELETEBUTTON) failed")
	}

	tb.RequestLayout()

	return nil
}

func (tb *ToolBar) onInsertedAction(action *Action) error {
	return tb.insertAction(action, false)
}

func (tb *ToolBar) onRemovingAction(action *Action) error {
	return tb.removeAction(action, false)
}

func (tb *ToolBar) onClearingActions() error {
	for i := tb.actions.Len() - 1; i >= 0; i-- {
		if action := tb.actions.At(i); action.Visible() {
			if err := tb.onRemovingAction(action); err != nil {
				return err
			}
		}
	}

	return nil
}

func (tb *ToolBar) CreateLayoutItem(ctx *LayoutContext) LayoutItem {
	buttonSize := uint32(tb.SendMessage(commctrl.TB_GETBUTTONSIZE, 0, 0))

	dpi := tb.DPI()
	width := IntFrom96DPI(tb.defaultButtonWidth, dpi)
	if width == 0 {
		width = int(win.LOWORD(buttonSize))
	}

	height := int(win.HIWORD(buttonSize))

	var size gdi32.SIZE
	var wp uintptr
	var layoutFlags LayoutFlags

	if tb.Orientation() == Vertical {
		wp = win.TRUE
		layoutFlags = ShrinkableVert | GrowableVert | GreedyVert
	} else {
		wp = win.FALSE
		// FIXME: Since reimplementation of BoxLayout we must use 0 here,
		// otherwise the ToolBar contained in MainWindow will eat half the space.
		//layoutFlags = ShrinkableHorz | GrowableHorz
	}

	if win.FALSE != tb.SendMessage(commctrl.TB_GETIDEALSIZE, wp, uintptr(unsafe.Pointer(&size))) {
		if wp == win.TRUE {
			height = int(size.CY)
		} else {
			width = int(size.CX)
		}
	}

	return &toolBarLayoutItem{
		layoutFlags: layoutFlags,
		idealSize:   Size{width, height},
	}
}

type toolBarLayoutItem struct {
	LayoutItemBase
	layoutFlags LayoutFlags
	idealSize   Size // in native pixels
}

func (li *toolBarLayoutItem) LayoutFlags() LayoutFlags {
	return li.layoutFlags
}

func (li *toolBarLayoutItem) IdealSize() Size {
	return li.idealSize
}

func (li *toolBarLayoutItem) MinSize() Size {
	return li.idealSize
}

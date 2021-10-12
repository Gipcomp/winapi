// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"syscall"
	"time"
	"unsafe"

	"github.com/Gipcomp/win32/comctl32"
	"github.com/Gipcomp/win32/commctrl"
	"github.com/Gipcomp/win32/gdi32"
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/uxtheme"
	"github.com/Gipcomp/win32/win"
	"github.com/Gipcomp/winapi/errs"
)

const tableViewWindowClass = `\o/ Walk_TableView_Class \o/`

var (
	white                       = gdi32.COLORREF(RGB(255, 255, 255))
	checkmark                   = string([]byte{0xE2, 0x9C, 0x94})
	tableViewFrozenLVWndProcPtr uintptr
	tableViewNormalLVWndProcPtr uintptr
	tableViewHdrWndProcPtr      uintptr
)

func init() {
	AppendToWalkInit(func() {
		MustRegisterWindowClass(tableViewWindowClass)
		tableViewFrozenLVWndProcPtr = syscall.NewCallback(tableViewFrozenLVWndProc)
		tableViewNormalLVWndProcPtr = syscall.NewCallback(tableViewNormalLVWndProc)
		tableViewHdrWndProcPtr = syscall.NewCallback(tableViewHdrWndProc)
	})
}

const (
	tableViewCurrentIndexChangedTimerId = 1 + iota
	tableViewSelectedIndexesChangedTimerId
)

type TableViewCfg struct {
	Style              uint32
	CustomHeaderHeight int // in native pixels?
	CustomRowHeight    int // in native pixels?
}

// TableView is a model based widget for record centric, tabular data.
//
// TableView is implemented as a virtual mode list view to support quite large
// amounts of data.
type TableView struct {
	WidgetBase
	hwndFrozenLV                       handle.HWND
	hwndFrozenHdr                      handle.HWND
	frozenLVOrigWndProcPtr             uintptr
	frozenHdrOrigWndProcPtr            uintptr
	hwndNormalLV                       handle.HWND
	hwndNormalHdr                      handle.HWND
	normalLVOrigWndProcPtr             uintptr
	normalHdrOrigWndProcPtr            uintptr
	state                              *tableViewState
	columns                            *TableViewColumnList
	model                              TableModel
	providedModel                      interface{}
	itemChecker                        ItemChecker
	imageProvider                      ImageProvider
	styler                             CellStyler
	style                              CellStyle
	itemFont                           *Font
	hIml                               comctl32.HIMAGELIST
	usingSysIml                        bool
	imageUintptr2Index                 map[uintptr]int32
	filePath2IconIndex                 map[string]int32
	rowsResetHandlerHandle             int
	rowChangedHandlerHandle            int
	rowsChangedHandlerHandle           int
	rowsInsertedHandlerHandle          int
	rowsRemovedHandlerHandle           int
	sortChangedHandlerHandle           int
	selectedIndexes                    []int
	prevIndex                          int
	currentIndex                       int
	itemIndexOfLastMouseButtonDown     int
	hwndItemChanged                    handle.HWND
	currentIndexChangedPublisher       EventPublisher
	selectedIndexesChangedPublisher    EventPublisher
	itemActivatedPublisher             EventPublisher
	columnClickedPublisher             IntEventPublisher
	columnsOrderableChangedPublisher   EventPublisher
	columnsSizableChangedPublisher     EventPublisher
	itemCountChangedPublisher          EventPublisher
	publishNextSelClear                bool
	inSetSelectedIndexes               bool
	lastColumnStretched                bool
	persistent                         bool
	itemStateChangedEventDelay         int
	themeNormalBGColor                 Color
	themeNormalTextColor               Color
	themeSelectedBGColor               Color
	themeSelectedTextColor             Color
	themeSelectedNotFocusedBGColor     Color
	itemBGColor                        Color
	itemTextColor                      Color
	alternatingRowBGColor              Color
	alternatingRowTextColor            Color
	alternatingRowBG                   bool
	delayedCurrentIndexChangedCanceled bool
	sortedColumnIndex                  int
	sortOrder                          SortOrder
	formActivatingHandle               int
	customHeaderHeight                 int // in native pixels?
	customRowHeight                    int // in native pixels?
	dpiOfPrevStretchLastColumn         int
	scrolling                          bool
	inSetCurrentIndex                  bool
	inMouseEvent                       bool
	hasFrozenColumn                    bool
	busyStretchingLastColumn           bool
	focused                            bool
	ignoreNowhere                      bool
	updateLVSizesNeedsSpecialCare      bool
	scrollbarOrientation               Orientation
	currentItemChangedPublisher        EventPublisher
	currentItemID                      interface{}
	restoringCurrentItemOnReset        bool
}

// NewTableView creates and returns a *TableView as child of the specified
// Container.
func NewTableView(parent Container) (*TableView, error) {
	return NewTableViewWithStyle(parent, commctrl.LVS_SHOWSELALWAYS)
}

// NewTableViewWithStyle creates and returns a *TableView as child of the specified
// Container and with the provided additional style bits set.
func NewTableViewWithStyle(parent Container, style uint32) (*TableView, error) {
	return NewTableViewWithCfg(parent, &TableViewCfg{Style: style})
}

// NewTableViewWithCfg creates and returns a *TableView as child of the specified
// Container and with the provided additional configuration.
func NewTableViewWithCfg(parent Container, cfg *TableViewCfg) (*TableView, error) {
	tv := &TableView{
		imageUintptr2Index:          make(map[uintptr]int32),
		filePath2IconIndex:          make(map[string]int32),
		formActivatingHandle:        -1,
		customHeaderHeight:          cfg.CustomHeaderHeight,
		customRowHeight:             cfg.CustomRowHeight,
		scrollbarOrientation:        Horizontal | Vertical,
		restoringCurrentItemOnReset: true,
	}

	tv.columns = newTableViewColumnList(tv)

	if err := InitWidget(
		tv,
		parent,
		tableViewWindowClass,
		user32.WS_BORDER|user32.WS_VISIBLE,
		user32.WS_EX_CONTROLPARENT); err != nil {
		return nil, err
	}

	succeeded := false
	defer func() {
		if !succeeded {
			tv.Dispose()
		}
	}()

	var rowHeightStyle uint32
	if cfg.CustomRowHeight > 0 {
		rowHeightStyle = commctrl.LVS_OWNERDRAWFIXED
	}
	strPtr, err := syscall.UTF16PtrFromString("SysListView32")
	if err != nil {
		errs.NewError(err.Error())
	}
	if tv.hwndFrozenLV = user32.CreateWindowEx(
		0,
		strPtr,
		nil,
		user32.WS_CHILD|user32.WS_CLIPSIBLINGS|user32.WS_TABSTOP|user32.WS_VISIBLE|commctrl.LVS_OWNERDATA|commctrl.LVS_REPORT|cfg.Style|rowHeightStyle,
		user32.CW_USEDEFAULT,
		user32.CW_USEDEFAULT,
		user32.CW_USEDEFAULT,
		user32.CW_USEDEFAULT,
		tv.hWnd,
		0,
		0,
		nil,
	); tv.hwndFrozenLV == 0 {
		return nil, errs.NewError("creating frozen lv failed")
	}

	tv.frozenLVOrigWndProcPtr = user32.SetWindowLongPtr(tv.hwndFrozenLV, user32.GWLP_WNDPROC, tableViewFrozenLVWndProcPtr)
	if tv.frozenLVOrigWndProcPtr == 0 {
		return nil, errs.LastError("SetWindowLongPtr")
	}

	tv.hwndFrozenHdr = handle.HWND(user32.SendMessage(tv.hwndFrozenLV, commctrl.LVM_GETHEADER, 0, 0))
	tv.frozenHdrOrigWndProcPtr = user32.SetWindowLongPtr(tv.hwndFrozenHdr, user32.GWLP_WNDPROC, tableViewHdrWndProcPtr)
	if tv.frozenHdrOrigWndProcPtr == 0 {
		return nil, errs.LastError("SetWindowLongPtr")
	}

	if tv.hwndNormalLV = user32.CreateWindowEx(
		0,
		strPtr,
		nil,
		user32.WS_CHILD|user32.WS_CLIPSIBLINGS|user32.WS_TABSTOP|user32.WS_VISIBLE|commctrl.LVS_OWNERDATA|commctrl.LVS_REPORT|cfg.Style|rowHeightStyle,
		user32.CW_USEDEFAULT,
		user32.CW_USEDEFAULT,
		user32.CW_USEDEFAULT,
		user32.CW_USEDEFAULT,
		tv.hWnd,
		0,
		0,
		nil,
	); tv.hwndNormalLV == 0 {
		return nil, errs.NewError("creating normal lv failed")
	}

	tv.normalLVOrigWndProcPtr = user32.SetWindowLongPtr(tv.hwndNormalLV, user32.GWLP_WNDPROC, tableViewNormalLVWndProcPtr)
	if tv.normalLVOrigWndProcPtr == 0 {
		return nil, errs.LastError("SetWindowLongPtr")
	}

	tv.hwndNormalHdr = handle.HWND(user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_GETHEADER, 0, 0))
	tv.normalHdrOrigWndProcPtr = user32.SetWindowLongPtr(tv.hwndNormalHdr, user32.GWLP_WNDPROC, tableViewHdrWndProcPtr)
	if tv.normalHdrOrigWndProcPtr == 0 {
		return nil, errs.LastError("SetWindowLongPtr")
	}

	tv.SetPersistent(true)

	exStyle := user32.SendMessage(tv.hwndFrozenLV, commctrl.LVM_GETEXTENDEDLISTVIEWSTYLE, 0, 0)
	exStyle |= commctrl.LVS_EX_DOUBLEBUFFER | commctrl.LVS_EX_FULLROWSELECT | commctrl.LVS_EX_HEADERDRAGDROP | commctrl.LVS_EX_LABELTIP | commctrl.LVS_EX_SUBITEMIMAGES
	user32.SendMessage(tv.hwndFrozenLV, commctrl.LVM_SETEXTENDEDLISTVIEWSTYLE, 0, exStyle)
	user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_SETEXTENDEDLISTVIEWSTYLE, 0, exStyle)
	explorer, err := syscall.UTF16PtrFromString("Explorer")
	if err != nil {
		errs.NewError(err.Error())
	}
	if hr := uxtheme.SetWindowTheme(tv.hwndFrozenLV, explorer, nil); win.FAILED(hr) {
		return nil, errs.ErrorFromHRESULT("SetWindowTheme", hr)
	}
	if hr := uxtheme.SetWindowTheme(tv.hwndNormalLV, explorer, nil); win.FAILED(hr) {
		return nil, errs.ErrorFromHRESULT("SetWindowTheme", hr)
	}

	user32.SendMessage(tv.hwndFrozenLV, user32.WM_CHANGEUISTATE, uintptr(win.MAKELONG(user32.UIS_SET, user32.UISF_HIDEFOCUS)), 0)
	user32.SendMessage(tv.hwndNormalLV, user32.WM_CHANGEUISTATE, uintptr(win.MAKELONG(user32.UIS_SET, user32.UISF_HIDEFOCUS)), 0)

	tv.group.toolTip.addTool(tv.hwndFrozenHdr, false)
	tv.group.toolTip.addTool(tv.hwndNormalHdr, false)

	tv.applyFont(parent.Font())

	tv.style.dpi = tv.DPI()
	tv.ApplySysColors()

	tv.currentIndex = -1

	tv.GraphicsEffects().Add(InteractionEffect)
	tv.GraphicsEffects().Add(FocusEffect)

	tv.MustRegisterProperty("ColumnsOrderable", NewBoolProperty(
		func() bool {
			return tv.ColumnsOrderable()
		},
		func(b bool) error {
			tv.SetColumnsOrderable(b)
			return nil
		},
		tv.columnsOrderableChangedPublisher.Event()))

	tv.MustRegisterProperty("ColumnsSizable", NewBoolProperty(
		func() bool {
			return tv.ColumnsSizable()
		},
		func(b bool) error {
			return tv.SetColumnsSizable(b)
		},
		tv.columnsSizableChangedPublisher.Event()))

	tv.MustRegisterProperty("CurrentIndex", NewProperty(
		func() interface{} {
			return tv.CurrentIndex()
		},
		func(v interface{}) error {
			return tv.SetCurrentIndex(assertIntOr(v, -1))
		},
		tv.CurrentIndexChanged()))

	tv.MustRegisterProperty("CurrentItem", NewReadOnlyProperty(
		func() interface{} {
			if i := tv.CurrentIndex(); i > -1 {
				if rm, ok := tv.providedModel.(reflectModel); ok {
					return reflect.ValueOf(rm.Items()).Index(i).Interface()
				}
			}

			return nil
		},
		tv.CurrentIndexChanged()))

	tv.MustRegisterProperty("HasCurrentItem", NewReadOnlyBoolProperty(
		func() bool {
			return tv.CurrentIndex() != -1
		},
		tv.CurrentIndexChanged()))

	tv.MustRegisterProperty("ItemCount", NewReadOnlyProperty(
		func() interface{} {
			if tv.model == nil {
				return 0
			}
			return tv.model.RowCount()
		},
		tv.itemCountChangedPublisher.Event()))

	tv.MustRegisterProperty("SelectedCount", NewReadOnlyProperty(
		func() interface{} {
			return len(tv.selectedIndexes)
		},
		tv.SelectedIndexesChanged()))

	succeeded = true

	return tv, nil
}

func (tv *TableView) asTableView() *TableView {
	return tv
}

// Dispose releases the operating system resources, associated with the
// *TableView.
func (tv *TableView) Dispose() {
	tv.columns.unsetColumnsTV()

	tv.disposeImageListAndCaches()

	if tv.hWnd != 0 {
		if !user32.KillTimer(tv.hWnd, tableViewCurrentIndexChangedTimerId) {
			errs.LastError("KillTimer")
		}
		if !user32.KillTimer(tv.hWnd, tableViewSelectedIndexesChangedTimerId) {
			errs.LastError("KillTimer")
		}
	}

	if tv.hwndFrozenLV != 0 {
		tv.group.toolTip.removeTool(tv.hwndFrozenHdr)
		user32.DestroyWindow(tv.hwndFrozenLV)
		tv.hwndFrozenLV = 0
	}

	if tv.hwndNormalLV != 0 {
		tv.group.toolTip.removeTool(tv.hwndNormalHdr)
		user32.DestroyWindow(tv.hwndNormalLV)
		tv.hwndNormalLV = 0
	}

	if tv.formActivatingHandle > -1 {
		if form := tv.Form(); form != nil {
			form.Activating().Detach(tv.formActivatingHandle)
		}
		tv.formActivatingHandle = -1
	}

	tv.WidgetBase.Dispose()
}

func (tv *TableView) applyEnabled(enabled bool) {
	tv.WidgetBase.applyEnabled(enabled)

	user32.EnableWindow(tv.hwndFrozenLV, enabled)
	user32.EnableWindow(tv.hwndNormalLV, enabled)
}

func (tv *TableView) applyFont(font *Font) {
	if tv.customHeaderHeight > 0 || tv.customRowHeight > 0 {
		return
	}

	tv.WidgetBase.applyFont(font)

	hFont := uintptr(font.handleForDPI(tv.DPI()))

	user32.SendMessage(tv.hwndFrozenLV, user32.WM_SETFONT, hFont, 0)
	user32.SendMessage(tv.hwndNormalLV, user32.WM_SETFONT, hFont, 0)
}

func (tv *TableView) ApplyDPI(dpi int) {
	tv.style.dpi = dpi
	if tv.style.canvas != nil {
		tv.style.canvas.dpi = dpi
	}

	tv.WidgetBase.ApplyDPI(dpi)

	for _, column := range tv.columns.items {
		column.update()
	}

	if tv.hIml != 0 {
		tv.disposeImageListAndCaches()

		if bmp, err := NewBitmapForDPI(SizeFrom96DPI(Size{16, 16}, dpi), dpi); err == nil {
			tv.applyImageListForImage(bmp)
			bmp.Dispose()
		}
	}
}

func (tv *TableView) ApplySysColors() {
	tv.WidgetBase.ApplySysColors()

	// As some combinations of property and state may be invalid for any theme,
	// we set some defaults here.
	tv.themeNormalBGColor = Color(user32.GetSysColor(user32.COLOR_WINDOW))
	tv.themeNormalTextColor = Color(user32.GetSysColor(user32.COLOR_WINDOWTEXT))
	tv.themeSelectedBGColor = tv.themeNormalBGColor
	tv.themeSelectedTextColor = tv.themeNormalTextColor
	tv.themeSelectedNotFocusedBGColor = tv.themeNormalBGColor
	tv.alternatingRowBGColor = Color(user32.GetSysColor(user32.COLOR_BTNFACE))
	tv.alternatingRowTextColor = Color(user32.GetSysColor(user32.COLOR_BTNTEXT))

	type item struct {
		stateID    int32
		propertyID int32
		color      *Color
	}

	getThemeColor := func(theme uxtheme.HTHEME, partId int32, items []item) {
		for _, item := range items {
			var c gdi32.COLORREF
			if result := uxtheme.GetThemeColor(theme, partId, item.stateID, item.propertyID, &c); !win.FAILED(result) {
				(*item.color) = Color(c)
			}
		}
	}
	listView, err := syscall.UTF16PtrFromString("Listview")
	if err != nil {
		errs.NewError(err.Error())
	}
	if hThemeListView := uxtheme.OpenThemeData(tv.hwndNormalLV, listView); hThemeListView != 0 {
		defer uxtheme.CloseThemeData(hThemeListView)

		getThemeColor(hThemeListView, uxtheme.LVP_LISTITEM, []item{
			{uxtheme.LISS_NORMAL, uxtheme.TMT_FILLCOLOR, &tv.themeNormalBGColor},
			{uxtheme.LISS_NORMAL, uxtheme.TMT_TEXTCOLOR, &tv.themeNormalTextColor},
			{uxtheme.LISS_SELECTED, uxtheme.TMT_FILLCOLOR, &tv.themeSelectedBGColor},
			{uxtheme.LISS_SELECTED, uxtheme.TMT_TEXTCOLOR, &tv.themeSelectedTextColor},
			{uxtheme.LISS_SELECTEDNOTFOCUS, uxtheme.TMT_FILLCOLOR, &tv.themeSelectedNotFocusedBGColor},
		})
	} else {
		// The others already have been retrieved above.
		tv.themeSelectedBGColor = Color(user32.GetSysColor(user32.COLOR_HIGHLIGHT))
		tv.themeSelectedTextColor = Color(user32.GetSysColor(user32.COLOR_HIGHLIGHTTEXT))
		tv.themeSelectedNotFocusedBGColor = Color(user32.GetSysColor(user32.COLOR_BTNFACE))
	}
	button, err := syscall.UTF16PtrFromString("BUTTON")
	if err != nil {
		errs.NewError(err.Error())
	}
	if hThemeButton := uxtheme.OpenThemeData(tv.hwndNormalLV, button); hThemeButton != 0 {
		defer uxtheme.CloseThemeData(hThemeButton)

		getThemeColor(hThemeButton, uxtheme.BP_PUSHBUTTON, []item{
			{uxtheme.PBS_NORMAL, uxtheme.TMT_FILLCOLOR, &tv.alternatingRowBGColor},
			{uxtheme.PBS_NORMAL, uxtheme.TMT_TEXTCOLOR, &tv.alternatingRowTextColor},
		})
	}

	user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_SETBKCOLOR, 0, uintptr(tv.themeNormalBGColor))
	user32.SendMessage(tv.hwndFrozenLV, commctrl.LVM_SETBKCOLOR, 0, uintptr(tv.themeNormalBGColor))
}

// ColumnsOrderable returns if the user can reorder columns by dragging and
// dropping column headers.
func (tv *TableView) ColumnsOrderable() bool {
	exStyle := user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_GETEXTENDEDLISTVIEWSTYLE, 0, 0)
	return exStyle&commctrl.LVS_EX_HEADERDRAGDROP > 0
}

// SetColumnsOrderable sets if the user can reorder columns by dragging and
// dropping column headers.
func (tv *TableView) SetColumnsOrderable(enabled bool) {
	var hwnd handle.HWND
	if tv.hasFrozenColumn {
		hwnd = tv.hwndFrozenLV
	} else {
		hwnd = tv.hwndNormalLV
	}

	exStyle := user32.SendMessage(hwnd, commctrl.LVM_GETEXTENDEDLISTVIEWSTYLE, 0, 0)
	if enabled {
		exStyle |= commctrl.LVS_EX_HEADERDRAGDROP
	} else {
		exStyle &^= commctrl.LVS_EX_HEADERDRAGDROP
	}
	user32.SendMessage(tv.hwndFrozenLV, commctrl.LVM_SETEXTENDEDLISTVIEWSTYLE, 0, exStyle)
	user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_SETEXTENDEDLISTVIEWSTYLE, 0, exStyle)

	tv.columnsOrderableChangedPublisher.Publish()
}

// ColumnsSizable returns if the user can change column widths by dragging
// dividers in the header.
func (tv *TableView) ColumnsSizable() bool {
	style := user32.GetWindowLong(tv.hwndNormalHdr, user32.GWL_STYLE)

	return style&commctrl.HDS_NOSIZING == 0
}

// SetColumnsSizable sets if the user can change column widths by dragging
// dividers in the header.
func (tv *TableView) SetColumnsSizable(b bool) error {
	updateStyle := func(headerHWnd handle.HWND) error {
		style := user32.GetWindowLong(headerHWnd, user32.GWL_STYLE)

		if b {
			style &^= commctrl.HDS_NOSIZING
		} else {
			style |= commctrl.HDS_NOSIZING
		}

		if user32.SetWindowLong(headerHWnd, user32.GWL_STYLE, style) == 0 {
			return errs.LastError("SetWindowLong(GWL_STYLE)")
		}

		return nil
	}

	if err := updateStyle(tv.hwndFrozenHdr); err != nil {
		return err
	}
	if err := updateStyle(tv.hwndNormalHdr); err != nil {
		return err
	}

	tv.columnsSizableChangedPublisher.Publish()

	return nil
}

// ContextMenuLocation returns selected item position in screen coordinates in native pixels.
func (tv *TableView) ContextMenuLocation() Point {
	idx := user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_GETSELECTIONMARK, 0, 0)
	rc := gdi32.RECT{Left: comctl32.LVIR_BOUNDS}
	if user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_GETITEMRECT, idx, uintptr(unsafe.Pointer(&rc))) == 0 {
		return tv.WidgetBase.ContextMenuLocation()
	}
	var pt gdi32.POINT
	if tv.RightToLeftReading() {
		pt.X = rc.Right
	} else {
		pt.X = rc.Left
	}
	pt.X = rc.Bottom
	windowTrimToClientBounds(tv.hwndNormalLV, &pt)
	user32.ClientToScreen(tv.hwndNormalLV, &pt)
	return pointPixelsFromPOINT(pt)
}

// SortableByHeaderClick returns if the user can change sorting by clicking the header.
func (tv *TableView) SortableByHeaderClick() bool {
	return !hasWindowLongBits(tv.hwndFrozenLV, user32.GWL_STYLE, commctrl.LVS_NOSORTHEADER) ||
		!hasWindowLongBits(tv.hwndNormalLV, user32.GWL_STYLE, commctrl.LVS_NOSORTHEADER)
}

// HeaderHidden returns whether the column header is hidden.
func (tv *TableView) HeaderHidden() bool {
	style := user32.GetWindowLong(tv.hwndNormalLV, user32.GWL_STYLE)

	return style&commctrl.LVS_NOCOLUMNHEADER != 0
}

// SetHeaderHidden sets whether the column header is hidden.
func (tv *TableView) SetHeaderHidden(hidden bool) error {
	if err := ensureWindowLongBits(tv.hwndFrozenLV, user32.GWL_STYLE, commctrl.LVS_NOCOLUMNHEADER, hidden); err != nil {
		return err
	}

	return ensureWindowLongBits(tv.hwndNormalLV, user32.GWL_STYLE, commctrl.LVS_NOCOLUMNHEADER, hidden)
}

// AlternatingRowBG returns the alternating row background.
func (tv *TableView) AlternatingRowBG() bool {
	return tv.alternatingRowBG
}

// SetAlternatingRowBG sets the alternating row background.
func (tv *TableView) SetAlternatingRowBG(enabled bool) {
	tv.alternatingRowBG = enabled

	tv.Invalidate()
}

// Gridlines returns if the rows are separated by grid lines.
func (tv *TableView) Gridlines() bool {
	exStyle := user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_GETEXTENDEDLISTVIEWSTYLE, 0, 0)
	return exStyle&commctrl.LVS_EX_GRIDLINES > 0
}

// SetGridlines sets if the rows are separated by grid lines.
func (tv *TableView) SetGridlines(enabled bool) {
	var hwnd handle.HWND
	if tv.hasFrozenColumn {
		hwnd = tv.hwndFrozenLV
	} else {
		hwnd = tv.hwndNormalLV
	}

	exStyle := user32.SendMessage(hwnd, commctrl.LVM_GETEXTENDEDLISTVIEWSTYLE, 0, 0)
	if enabled {
		exStyle |= commctrl.LVS_EX_GRIDLINES
	} else {
		exStyle &^= commctrl.LVS_EX_GRIDLINES
	}
	user32.SendMessage(tv.hwndFrozenLV, commctrl.LVM_SETEXTENDEDLISTVIEWSTYLE, 0, exStyle)
	user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_SETEXTENDEDLISTVIEWSTYLE, 0, exStyle)
}

// Columns returns the list of columns.
func (tv *TableView) Columns() *TableViewColumnList {
	return tv.columns
}

// VisibleColumnsInDisplayOrder returns a slice of visible columns in display
// order.
func (tv *TableView) VisibleColumnsInDisplayOrder() []*TableViewColumn {
	visibleCols := tv.visibleColumns()
	indices := make([]int32, len(visibleCols))

	frozenCount := tv.visibleFrozenColumnCount()
	normalCount := len(visibleCols) - frozenCount

	if frozenCount > 0 {
		if win.FALSE == user32.SendMessage(tv.hwndFrozenLV, commctrl.LVM_GETCOLUMNORDERARRAY, uintptr(frozenCount), uintptr(unsafe.Pointer(&indices[0]))) {
			errs.NewError("LVM_GETCOLUMNORDERARRAY")
			return nil
		}
	}
	if normalCount > 0 {
		if win.FALSE == user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_GETCOLUMNORDERARRAY, uintptr(normalCount), uintptr(unsafe.Pointer(&indices[frozenCount]))) {
			errs.NewError("LVM_GETCOLUMNORDERARRAY")
			return nil
		}
	}

	orderedCols := make([]*TableViewColumn, len(visibleCols))

	for i, j := range indices {
		if i >= frozenCount {
			j += int32(frozenCount)
		}
		orderedCols[i] = visibleCols[j]
	}

	return orderedCols
}

// RowsPerPage returns the number of fully visible rows.
func (tv *TableView) RowsPerPage() int {
	return int(user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_GETCOUNTPERPAGE, 0, 0))
}

func (tv *TableView) Invalidate() error {
	user32.InvalidateRect(tv.hwndFrozenLV, nil, true)
	user32.InvalidateRect(tv.hwndNormalLV, nil, true)

	return tv.WidgetBase.Invalidate()
}

func (tv *TableView) redrawItems() {
	first := user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_GETTOPINDEX, 0, 0)
	last := first + user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_GETCOUNTPERPAGE, 0, 0) + 1
	user32.SendMessage(tv.hwndFrozenLV, commctrl.LVM_REDRAWITEMS, first, last)
	user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_REDRAWITEMS, first, last)
}

// UpdateItem ensures the item at index will be redrawn.
//
// If the model supports sorting, it will be resorted.
func (tv *TableView) UpdateItem(index int) error {
	if s, ok := tv.model.(Sorter); ok {
		if err := s.Sort(s.SortedColumn(), s.SortOrder()); err != nil {
			return err
		}
	} else {
		if win.FALSE == user32.SendMessage(tv.hwndFrozenLV, commctrl.LVM_UPDATE, uintptr(index), 0) {
			return errs.NewError("LVM_UPDATE")
		}
		if win.FALSE == user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_UPDATE, uintptr(index), 0) {
			return errs.NewError("LVM_UPDATE")
		}
	}

	return nil
}

func (tv *TableView) attachModel() {
	restoreCurrentItemOrFallbackToFirst := func(ip IDProvider) {
		if tv.itemStateChangedEventDelay == 0 {
			defer tv.currentItemChangedPublisher.Publish()
		} else {
			if user32.SetTimer(
				tv.hWnd,
				tableViewCurrentIndexChangedTimerId,
				uint32(tv.itemStateChangedEventDelay),
				0,
			) == 0 {
				errs.LastError("SetTimer")
			}
		}

		count := tv.model.RowCount()
		for i := 0; i < count; i++ {
			if ip.ID(i) == tv.currentItemID {
				tv.SetCurrentIndex(i)
				return
			}
		}

		tv.SetCurrentIndex(0)
	}

	tv.rowsResetHandlerHandle = tv.model.RowsReset().Attach(func() {
		tv.setItemCount()

		if ip, ok := tv.providedModel.(IDProvider); ok && tv.restoringCurrentItemOnReset {
			if _, ok := tv.model.(Sorter); !ok {
				restoreCurrentItemOrFallbackToFirst(ip)
			}
		} else {
			tv.SetCurrentIndex(-1)
		}

		tv.itemCountChangedPublisher.Publish()
	})

	tv.rowChangedHandlerHandle = tv.model.RowChanged().Attach(func(row int) {
		tv.UpdateItem(row)
	})

	tv.rowsChangedHandlerHandle = tv.model.RowsChanged().Attach(func(from, to int) {
		if s, ok := tv.model.(Sorter); ok {
			s.Sort(s.SortedColumn(), s.SortOrder())
		} else {
			first, last := uintptr(from), uintptr(to)
			user32.SendMessage(tv.hwndFrozenLV, commctrl.LVM_REDRAWITEMS, first, last)
			user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_REDRAWITEMS, first, last)
		}
	})

	tv.rowsInsertedHandlerHandle = tv.model.RowsInserted().Attach(func(from, to int) {
		i := tv.currentIndex

		tv.setItemCount()

		if from <= i {
			i += 1 + to - from

			tv.SetCurrentIndex(i)
		}

		tv.itemCountChangedPublisher.Publish()
	})

	tv.rowsRemovedHandlerHandle = tv.model.RowsRemoved().Attach(func(from, to int) {
		i := tv.currentIndex

		tv.setItemCount()

		index := i

		if from <= i && i <= to {
			index = -1
		} else if from < i {
			index -= 1 + to - from
		}

		if index != i {
			tv.SetCurrentIndex(index)
		}

		tv.itemCountChangedPublisher.Publish()
	})

	if sorter, ok := tv.model.(Sorter); ok {
		tv.sortChangedHandlerHandle = sorter.SortChanged().Attach(func() {
			if ip, ok := tv.providedModel.(IDProvider); ok && tv.restoringCurrentItemOnReset {
				restoreCurrentItemOrFallbackToFirst(ip)
			}

			col := sorter.SortedColumn()
			tv.setSortIcon(col, sorter.SortOrder())

			tv.redrawItems()
		})
	}
}

func (tv *TableView) detachModel() {
	tv.model.RowsReset().Detach(tv.rowsResetHandlerHandle)
	tv.model.RowChanged().Detach(tv.rowChangedHandlerHandle)
	tv.model.RowsInserted().Detach(tv.rowsInsertedHandlerHandle)
	tv.model.RowsRemoved().Detach(tv.rowsRemovedHandlerHandle)
	if sorter, ok := tv.model.(Sorter); ok {
		sorter.SortChanged().Detach(tv.sortChangedHandlerHandle)
	}
}

// ItemCountChanged returns the event that is published when the number of items
// in the model of the TableView changed.
func (tv *TableView) ItemCountChanged() *Event {
	return tv.itemCountChangedPublisher.Event()
}

// Model returns the model of the TableView.
func (tv *TableView) Model() interface{} {
	return tv.providedModel
}

// SetModel sets the model of the TableView.
//
// It is required that mdl either implements walk.TableModel,
// walk.ReflectTableModel or be a slice of pointers to struct or a
// []map[string]interface{}. A walk.TableModel implementation must also
// implement walk.Sorter to support sorting, all other options get sorting for
// free. To support item check boxes and icons, mdl must implement
// walk.ItemChecker and walk.ImageProvider, respectively. On-demand model
// population for a walk.ReflectTableModel or slice requires mdl to implement
// walk.Populator.
func (tv *TableView) SetModel(mdl interface{}) error {
	model, ok := mdl.(TableModel)
	if !ok && mdl != nil {
		var err error
		if model, err = newReflectTableModel(mdl); err != nil {
			if model, err = newMapTableModel(mdl); err != nil {
				return err
			}
		}
	}

	tv.SetSuspended(true)
	defer tv.SetSuspended(false)

	if tv.model != nil {
		tv.detachModel()

		tv.disposeImageListAndCaches()
	}

	oldProvidedModelStyler, _ := tv.providedModel.(CellStyler)
	if styler, ok := mdl.(CellStyler); ok || tv.styler == oldProvidedModelStyler {
		tv.styler = styler
	}

	tv.providedModel = mdl
	tv.model = model

	tv.itemChecker, _ = model.(ItemChecker)
	tv.imageProvider, _ = model.(ImageProvider)

	if model != nil {
		tv.attachModel()

		if dms, ok := model.(dataMembersSetter); ok {
			// FIXME: This depends on columns to be initialized before
			// calling this method.
			dataMembers := make([]string, len(tv.columns.items))

			for i, col := range tv.columns.items {
				dataMembers[i] = col.DataMemberEffective()
			}

			dms.setDataMembers(dataMembers)
		}

		if lfs, ok := model.(lessFuncsSetter); ok {
			lessFuncs := make([]func(i, j int) bool, tv.columns.Len())
			for i, c := range tv.columns.items {
				lessFuncs[i] = c.lessFunc
			}
			lfs.setLessFuncs(lessFuncs)
		}

		if sorter, ok := tv.model.(Sorter); ok {
			if tv.sortedColumnIndex >= tv.visibleColumnCount() {
				tv.sortedColumnIndex = maxi(-1, mini(0, tv.visibleColumnCount()-1))
				tv.sortOrder = SortAscending
			}

			sorter.Sort(tv.sortedColumnIndex, tv.sortOrder)
		}
	}

	tv.SetCurrentIndex(-1)

	tv.setItemCount()

	tv.itemCountChangedPublisher.Publish()

	return nil
}

// TableModel returns the TableModel of the TableView.
func (tv *TableView) TableModel() TableModel {
	return tv.model
}

// ItemChecker returns the ItemChecker of the TableView.
func (tv *TableView) ItemChecker() ItemChecker {
	return tv.itemChecker
}

// SetItemChecker sets the ItemChecker of the TableView.
func (tv *TableView) SetItemChecker(itemChecker ItemChecker) {
	tv.itemChecker = itemChecker
}

// CellStyler returns the CellStyler of the TableView.
func (tv *TableView) CellStyler() CellStyler {
	return tv.styler
}

// SetCellStyler sets the CellStyler of the TableView.
func (tv *TableView) SetCellStyler(styler CellStyler) {
	tv.styler = styler
}

func (tv *TableView) setItemCount() error {
	var count int

	if tv.model != nil {
		count = tv.model.RowCount()
	}

	if user32.SendMessage(tv.hwndFrozenLV, commctrl.LVM_SETITEMCOUNT, uintptr(count), commctrl.LVSICF_NOINVALIDATEALL|commctrl.LVSICF_NOSCROLL) == 0 {
		return errs.NewError("SendMessage(LVM_SETITEMCOUNT)")
	}
	if user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_SETITEMCOUNT, uintptr(count), commctrl.LVSICF_NOINVALIDATEALL|commctrl.LVSICF_NOSCROLL) == 0 {
		return errs.NewError("SendMessage(LVM_SETITEMCOUNT)")
	}

	return nil
}

// CheckBoxes returns if the *TableView has check boxes.
func (tv *TableView) CheckBoxes() bool {
	var hwnd handle.HWND
	if tv.hasFrozenColumn {
		hwnd = tv.hwndFrozenLV
	} else {
		hwnd = tv.hwndNormalLV
	}

	return user32.SendMessage(hwnd, commctrl.LVM_GETEXTENDEDLISTVIEWSTYLE, 0, 0)&commctrl.LVS_EX_CHECKBOXES > 0
}

// SetCheckBoxes sets if the *TableView has check boxes.
func (tv *TableView) SetCheckBoxes(checkBoxes bool) {
	var hwnd, hwndOther handle.HWND
	if tv.hasFrozenColumn {
		hwnd, hwndOther = tv.hwndFrozenLV, tv.hwndNormalLV
	} else {
		hwnd, hwndOther = tv.hwndNormalLV, tv.hwndFrozenLV
	}

	exStyle := user32.SendMessage(hwnd, commctrl.LVM_GETEXTENDEDLISTVIEWSTYLE, 0, 0)
	oldStyle := exStyle
	if checkBoxes {
		exStyle |= commctrl.LVS_EX_CHECKBOXES
	} else {
		exStyle &^= commctrl.LVS_EX_CHECKBOXES
	}
	if exStyle != oldStyle {
		user32.SendMessage(hwnd, commctrl.LVM_SETEXTENDEDLISTVIEWSTYLE, 0, exStyle)
	}

	user32.SendMessage(hwndOther, commctrl.LVM_SETEXTENDEDLISTVIEWSTYLE, 0, exStyle&^commctrl.LVS_EX_CHECKBOXES)

	mask := user32.SendMessage(hwnd, commctrl.LVM_GETCALLBACKMASK, 0, 0)

	if checkBoxes {
		mask |= commctrl.LVIS_STATEIMAGEMASK
	} else {
		mask &^= commctrl.LVIS_STATEIMAGEMASK
	}

	if win.FALSE == user32.SendMessage(hwnd, commctrl.LVM_SETCALLBACKMASK, mask, 0) {
		errs.NewError("SendMessage(LVM_SETCALLBACKMASK)")
	}
}

func (tv *TableView) fromLVColIdx(frozen bool, index int32) int {
	var idx int32

	for i, tvc := range tv.columns.items {
		if frozen == tvc.frozen && tvc.visible {
			if idx == index {
				return i
			}

			idx++
		}
	}

	return -1
}

func (tv *TableView) toLVColIdx(index int) int32 {
	var idx int32

	for i, tvc := range tv.columns.items {
		if tvc.visible {
			if i == index {
				return idx
			}

			idx++
		}
	}

	return -1
}

func (tv *TableView) visibleFrozenColumnCount() int {
	var count int

	for _, tvc := range tv.columns.items {
		if tvc.frozen && tvc.visible {
			count++
		}
	}

	return count
}

func (tv *TableView) visibleColumnCount() int {
	var count int

	for _, tvc := range tv.columns.items {
		if tvc.visible {
			count++
		}
	}

	return count
}

func (tv *TableView) visibleColumns() []*TableViewColumn {
	var cols []*TableViewColumn

	for _, tvc := range tv.columns.items {
		if tvc.visible {
			cols = append(cols, tvc)
		}
	}

	return cols
}

/*func (tv *TableView) selectedColumnIndex() int {
	return tv.fromLVColIdx(tv.SendMessage(LVM_GETSELECTEDCOLUMN, 0, 0))
}*/

// func (tv *TableView) setSelectedColumnIndex(index int) {
// 	tv.SendMessage(commctrl.LVM_SETSELECTEDCOLUMN, uintptr(tv.toLVColIdx(index)), 0)
// }

func (tv *TableView) setSortIcon(index int, order SortOrder) error {
	idx := int(tv.toLVColIdx(index))

	frozenCount := tv.visibleFrozenColumnCount()

	for i, col := range tv.visibleColumns() {
		item := commctrl.HDITEM{
			Mask: commctrl.HDI_FORMAT,
		}

		var headerHwnd handle.HWND
		var offset int
		if col.frozen {
			headerHwnd = tv.hwndFrozenHdr
		} else {
			headerHwnd = tv.hwndNormalHdr
			offset = -frozenCount
		}

		iPtr := uintptr(offset + i)
		itemPtr := uintptr(unsafe.Pointer(&item))

		if user32.SendMessage(headerHwnd, commctrl.HDM_GETITEM, iPtr, itemPtr) == 0 {
			return errs.NewError("SendMessage(HDM_GETITEM)")
		}

		if i == idx {
			switch order {
			case SortAscending:
				item.Fmt &^= commctrl.HDF_SORTDOWN
				item.Fmt |= commctrl.HDF_SORTUP

			case SortDescending:
				item.Fmt &^= commctrl.HDF_SORTUP
				item.Fmt |= commctrl.HDF_SORTDOWN
			}
		} else {
			item.Fmt &^= commctrl.HDF_SORTDOWN | commctrl.HDF_SORTUP
		}

		if user32.SendMessage(headerHwnd, commctrl.HDM_SETITEM, iPtr, itemPtr) == 0 {
			return errs.NewError("SendMessage(HDM_SETITEM)")
		}
	}

	return nil
}

// ColumnClicked returns the event that is published after a column header was
// clicked.
func (tv *TableView) ColumnClicked() *IntEvent {
	return tv.columnClickedPublisher.Event()
}

// ItemActivated returns the event that is published after an item was
// activated.
//
// An item is activated when it is double clicked or the enter key is pressed
// when the item is selected.
func (tv *TableView) ItemActivated() *Event {
	return tv.itemActivatedPublisher.Event()
}

// RestoringCurrentItemOnReset returns whether the TableView after its model
// has been reset should attempt to restore CurrentIndex to the item that was
// current before the reset.
//
// For this to work, the model must implement the IDProvider interface.
func (tv *TableView) RestoringCurrentItemOnReset() bool {
	return tv.restoringCurrentItemOnReset
}

// SetRestoringCurrentItemOnReset sets whether the TableView after its model
// has been reset should attempt to restore CurrentIndex to the item that was
// current before the reset.
//
// For this to work, the model must implement the IDProvider interface.
func (tv *TableView) SetRestoringCurrentItemOnReset(restoring bool) {
	tv.restoringCurrentItemOnReset = restoring
}

// CurrentItemChanged returns the event that is published after the current
// item has changed.
//
// For this to work, the model must implement the IDProvider interface.
func (tv *TableView) CurrentItemChanged() *Event {
	return tv.currentItemChangedPublisher.Event()
}

// CurrentIndex returns the index of the current item, or -1 if there is no
// current item.
func (tv *TableView) CurrentIndex() int {
	return tv.currentIndex
}

// SetCurrentIndex sets the index of the current item.
//
// Call this with a value of -1 to have no current item.
func (tv *TableView) SetCurrentIndex(index int) error {
	if tv.inSetCurrentIndex {
		return nil
	}
	tv.inSetCurrentIndex = true
	defer func() {
		tv.inSetCurrentIndex = false
	}()

	var lvi commctrl.LVITEM

	lvi.StateMask = commctrl.LVIS_FOCUSED | commctrl.LVIS_SELECTED

	if tv.MultiSelection() {
		if win.FALSE == user32.SendMessage(tv.hwndFrozenLV, commctrl.LVM_SETITEMSTATE, ^uintptr(0), uintptr(unsafe.Pointer(&lvi))) {
			return errs.NewError("SendMessage(LVM_SETITEMSTATE)")
		}
		if win.FALSE == user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_SETITEMSTATE, ^uintptr(0), uintptr(unsafe.Pointer(&lvi))) {
			return errs.NewError("SendMessage(LVM_SETITEMSTATE)")
		}
	}

	if index > -1 {
		lvi.State = commctrl.LVIS_FOCUSED | commctrl.LVIS_SELECTED
	}

	if win.FALSE == user32.SendMessage(tv.hwndFrozenLV, commctrl.LVM_SETITEMSTATE, uintptr(index), uintptr(unsafe.Pointer(&lvi))) {
		return errs.NewError("SendMessage(LVM_SETITEMSTATE)")
	}
	if win.FALSE == user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_SETITEMSTATE, uintptr(index), uintptr(unsafe.Pointer(&lvi))) {
		return errs.NewError("SendMessage(LVM_SETITEMSTATE)")
	}

	if index > -1 {
		if win.FALSE == user32.SendMessage(tv.hwndFrozenLV, commctrl.LVM_ENSUREVISIBLE, uintptr(index), uintptr(0)) {
			return errs.NewError("SendMessage(LVM_ENSUREVISIBLE)")
		}
		// Windows bug? Sometimes a second LVM_ENSUREVISIBLE is required.
		if win.FALSE == user32.SendMessage(tv.hwndFrozenLV, commctrl.LVM_ENSUREVISIBLE, uintptr(index), uintptr(0)) {
			return errs.NewError("SendMessage(LVM_ENSUREVISIBLE)")
		}
		if win.FALSE == user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_ENSUREVISIBLE, uintptr(index), uintptr(0)) {
			return errs.NewError("SendMessage(LVM_ENSUREVISIBLE)")
		}
		// Windows bug? Sometimes a second LVM_ENSUREVISIBLE is required.
		if win.FALSE == user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_ENSUREVISIBLE, uintptr(index), uintptr(0)) {
			return errs.NewError("SendMessage(LVM_ENSUREVISIBLE)")
		}

		if ip, ok := tv.providedModel.(IDProvider); ok && tv.restoringCurrentItemOnReset {
			if id := ip.ID(index); id != tv.currentItemID {
				tv.currentItemID = id
				if tv.itemStateChangedEventDelay == 0 {
					defer tv.currentItemChangedPublisher.Publish()
				}
			}
		}
	} else {
		tv.currentItemID = nil
		if tv.itemStateChangedEventDelay == 0 {
			defer tv.currentItemChangedPublisher.Publish()
		}
	}

	tv.currentIndex = index

	if index == -1 || tv.itemStateChangedEventDelay == 0 {
		tv.currentIndexChangedPublisher.Publish()
	}

	if tv.MultiSelection() {
		tv.updateSelectedIndexes()
	}

	return nil
}

// CurrentIndexChanged is the event that is published after CurrentIndex has
// changed.
func (tv *TableView) CurrentIndexChanged() *Event {
	return tv.currentIndexChangedPublisher.Event()
}

// IndexAt returns the item index at coordinates x, y of the
// TableView or -1, if that point is not inside any item.
func (tv *TableView) IndexAt(x, y int) int {
	var hti commctrl.LVHITTESTINFO

	var rc gdi32.RECT
	if !user32.GetWindowRect(tv.hwndFrozenLV, &rc) {
		return -1
	}

	var hwnd handle.HWND
	if x < int(rc.Right-rc.Left) {
		hwnd = tv.hwndFrozenLV
	} else {
		hwnd = tv.hwndNormalLV
	}

	hti.Pt.X = int32(x)
	hti.Pt.Y = int32(y)

	user32.SendMessage(hwnd, commctrl.LVM_HITTEST, 0, uintptr(unsafe.Pointer(&hti)))

	return int(hti.IItem)
}

// ItemVisible returns whether the item at position index is visible.
func (tv *TableView) ItemVisible(index int) bool {
	return user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_ISITEMVISIBLE, uintptr(index), 0) != 0
}

// EnsureItemVisible ensures the item at position index is visible, scrolling if necessary.
func (tv *TableView) EnsureItemVisible(index int) {
	user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_ENSUREVISIBLE, uintptr(index), 0)
}

// SelectionHiddenWithoutFocus returns whether selection indicators are hidden
// when the TableView does not have the keyboard input focus.
func (tv *TableView) SelectionHiddenWithoutFocus() bool {
	style := uint(user32.GetWindowLong(tv.hwndNormalLV, user32.GWL_STYLE))
	if style == 0 {
		errs.LastError("GetWindowLong")
		return false
	}

	return style&commctrl.LVS_SHOWSELALWAYS == 0
}

// SetSelectionHiddenWithoutFocus sets whether selection indicators are visible when the TableView does not have the keyboard input focus.
func (tv *TableView) SetSelectionHiddenWithoutFocus(hidden bool) error {
	if err := ensureWindowLongBits(tv.hwndFrozenLV, user32.GWL_STYLE, commctrl.LVS_SHOWSELALWAYS, !hidden); err != nil {
		return err
	}

	return ensureWindowLongBits(tv.hwndNormalLV, user32.GWL_STYLE, commctrl.LVS_SHOWSELALWAYS, !hidden)
}

// MultiSelection returns whether multiple items can be selected at once.
//
// By default only a single item can be selected at once.
func (tv *TableView) MultiSelection() bool {
	style := uint(user32.GetWindowLong(tv.hwndNormalLV, user32.GWL_STYLE))
	if style == 0 {
		errs.LastError("GetWindowLong")
		return false
	}

	return style&commctrl.LVS_SINGLESEL == 0
}

// SetMultiSelection sets whether multiple items can be selected at once.
func (tv *TableView) SetMultiSelection(multiSel bool) error {
	if err := ensureWindowLongBits(tv.hwndFrozenLV, user32.GWL_STYLE, commctrl.LVS_SINGLESEL, !multiSel); err != nil {
		return err
	}

	return ensureWindowLongBits(tv.hwndNormalLV, user32.GWL_STYLE, commctrl.LVS_SINGLESEL, !multiSel)
}

// SelectedIndexes returns the indexes of the currently selected items.
func (tv *TableView) SelectedIndexes() []int {
	indexes := make([]int, len(tv.selectedIndexes))

	for i, j := range tv.selectedIndexes {
		indexes[i] = j
	}

	return indexes
}

// SetSelectedIndexes sets the indexes of the currently selected items.
func (tv *TableView) SetSelectedIndexes(indexes []int) error {
	tv.inSetSelectedIndexes = true
	defer func() {
		tv.inSetSelectedIndexes = false
		tv.publishSelectedIndexesChanged()
	}()

	lvi := &commctrl.LVITEM{StateMask: commctrl.LVIS_FOCUSED | commctrl.LVIS_SELECTED}
	lp := uintptr(unsafe.Pointer(lvi))

	if win.FALSE == user32.SendMessage(tv.hwndFrozenLV, commctrl.LVM_SETITEMSTATE, ^uintptr(0), lp) {
		return errs.NewError("SendMessage(LVM_SETITEMSTATE)")
	}
	if win.FALSE == user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_SETITEMSTATE, ^uintptr(0), lp) {
		return errs.NewError("SendMessage(LVM_SETITEMSTATE)")
	}

	selectAll := false
	lvi.State = commctrl.LVIS_FOCUSED | commctrl.LVIS_SELECTED
	for _, i := range indexes {
		val := uintptr(i)
		if i == -1 {
			selectAll = true
			val = ^uintptr(0)
		}
		if win.FALSE == user32.SendMessage(tv.hwndFrozenLV, commctrl.LVM_SETITEMSTATE, val, lp) && i != -1 {
			return errs.NewError("SendMessage(LVM_SETITEMSTATE)")
		}
		if win.FALSE == user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_SETITEMSTATE, val, lp) && i != -1 {
			return errs.NewError("SendMessage(LVM_SETITEMSTATE)")
		}
	}

	if !selectAll {
		idxs := make([]int, len(indexes))

		copy(idxs, indexes)
		// for i, j := range indexes {
		// 	idxs[i] = j
		// }

		tv.selectedIndexes = idxs
	} else {
		count := int(user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_GETSELECTEDCOUNT, 0, 0))
		idxs := make([]int, count)
		for i := range idxs {
			idxs[i] = i
		}
		tv.selectedIndexes = idxs
	}

	return nil
}

func (tv *TableView) updateSelectedIndexes() {
	count := int(user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_GETSELECTEDCOUNT, 0, 0))
	indexes := make([]int, count)

	j := -1
	for i := 0; i < count; i++ {
		j = int(user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_GETNEXTITEM, uintptr(j), commctrl.LVNI_SELECTED))
		indexes[i] = j
	}

	changed := len(indexes) != len(tv.selectedIndexes)
	if !changed {
		for i := 0; i < len(indexes); i++ {
			if indexes[i] != tv.selectedIndexes[i] {
				changed = true
				break
			}
		}
	}

	if changed {
		tv.selectedIndexes = indexes
		tv.publishSelectedIndexesChanged()
	}
}

func (tv *TableView) copySelectedIndexes(hwndTo, hwndFrom handle.HWND) error {
	count := int(user32.SendMessage(hwndFrom, commctrl.LVM_GETSELECTEDCOUNT, 0, 0))

	lvi := &commctrl.LVITEM{StateMask: commctrl.LVIS_FOCUSED | commctrl.LVIS_SELECTED}
	lp := uintptr(unsafe.Pointer(lvi))

	if win.FALSE == user32.SendMessage(hwndTo, commctrl.LVM_SETITEMSTATE, ^uintptr(0), lp) {
		return errs.NewError("SendMessage(LVM_SETITEMSTATE)")
	}

	lvi.StateMask = commctrl.LVIS_SELECTED
	lvi.State = commctrl.LVIS_SELECTED

	j := -1
	for i := 0; i < count; i++ {
		j = int(user32.SendMessage(hwndFrom, commctrl.LVM_GETNEXTITEM, uintptr(j), commctrl.LVNI_SELECTED))

		if win.FALSE == user32.SendMessage(hwndTo, commctrl.LVM_SETITEMSTATE, uintptr(j), lp) {
			return errs.NewError("SendMessage(LVM_SETITEMSTATE)")
		}
	}

	return nil
}

// ItemStateChangedEventDelay returns the delay in milliseconds, between the
// moment the state of an item in the *TableView changes and the moment the
// associated event is published.
//
// By default there is no delay.
func (tv *TableView) ItemStateChangedEventDelay() int {
	return tv.itemStateChangedEventDelay
}

// SetItemStateChangedEventDelay sets the delay in milliseconds, between the
// moment the state of an item in the *TableView changes and the moment the
// associated event is published.
//
// An example where this may be useful is a master-details scenario. If the
// master TableView is configured to delay the event, you can avoid pointless
// updates of the details TableView, if the user uses arrow keys to rapidly
// navigate the master view.
func (tv *TableView) SetItemStateChangedEventDelay(delay int) {
	tv.itemStateChangedEventDelay = delay
}

// SelectedIndexesChanged returns the event that is published when the list of
// selected item indexes changed.
func (tv *TableView) SelectedIndexesChanged() *Event {
	return tv.selectedIndexesChangedPublisher.Event()
}

func (tv *TableView) publishSelectedIndexesChanged() {
	if tv.itemStateChangedEventDelay > 0 {
		if user32.SetTimer(
			tv.hWnd,
			tableViewSelectedIndexesChangedTimerId,
			uint32(tv.itemStateChangedEventDelay),
			0) == 0 {
			errs.LastError("SetTimer")
		}
	} else {
		tv.selectedIndexesChangedPublisher.Publish()
	}
}

// LastColumnStretched returns if the last column should take up all remaining
// horizontal space of the *TableView.
func (tv *TableView) LastColumnStretched() bool {
	return tv.lastColumnStretched
}

// SetLastColumnStretched sets if the last column should take up all remaining
// horizontal space of the *TableView.
//
// The effect of setting this is persistent.
func (tv *TableView) SetLastColumnStretched(value bool) error {
	if value {
		if err := tv.StretchLastColumn(); err != nil {
			return err
		}
	}

	tv.lastColumnStretched = value

	return nil
}

// StretchLastColumn makes the last column take up all remaining horizontal
// space of the *TableView.
//
// The effect of this is not persistent.
func (tv *TableView) StretchLastColumn() error {
	colCount := tv.visibleColumnCount()
	if colCount == 0 {
		return nil
	}

	var hwnd handle.HWND
	frozenColCount := tv.visibleFrozenColumnCount()
	if colCount-frozenColCount == 0 {
		hwnd = tv.hwndFrozenLV
		colCount = frozenColCount
	} else {
		hwnd = tv.hwndNormalLV
		colCount -= frozenColCount
	}

	var lp uintptr
	if tv.scrollbarOrientation&Horizontal != 0 {
		lp = commctrl.LVSCW_AUTOSIZE_USEHEADER
	} else {
		width := tv.ClientBoundsPixels().Width

		lastIndexInLV := -1
		var lastIndexInLVWidth int

		for _, tvc := range tv.columns.items {
			var offset int
			if !tvc.Frozen() {
				offset = frozenColCount
			}

			colWidth := tv.IntFrom96DPI(tvc.Width())
			width -= colWidth

			if index := int32(offset) + tvc.indexInListView(); int(index) > lastIndexInLV {
				lastIndexInLV = int(index)
				lastIndexInLVWidth = colWidth
			}
		}

		width += lastIndexInLVWidth

		if hasWindowLongBits(tv.hwndNormalLV, user32.GWL_STYLE, user32.WS_VSCROLL) {
			width -= int(user32.GetSystemMetricsForDpi(user32.SM_CXVSCROLL, uint32(tv.DPI())))
		}

		lp = uintptr(maxi(0, width))
	}

	if lp > 0 {
		if user32.SendMessage(hwnd, commctrl.LVM_SETCOLUMNWIDTH, uintptr(colCount-1), lp) == 0 {
			return errs.NewError("LVM_SETCOLUMNWIDTH failed")
		}

		if dpi := tv.DPI(); dpi != tv.dpiOfPrevStretchLastColumn {
			tv.dpiOfPrevStretchLastColumn = dpi

			tv.Invalidate()
		}
	}

	return nil
}

// Persistent returns if the *TableView should persist its UI state, like column
// widths. See *App.Settings for details.
func (tv *TableView) Persistent() bool {
	return tv.persistent
}

// SetPersistent sets if the *TableView should persist its UI state, like column
// widths. See *App.Settings for details.
func (tv *TableView) SetPersistent(value bool) {
	tv.persistent = value
}

// IgnoreNowhere returns if the *TableView should ignore left mouse clicks in the
// empty space. It forbids the user from unselecting the current index, or when
// multi selection is enabled, disables click drag selection.
func (tv *TableView) IgnoreNowhere() bool {
	return tv.ignoreNowhere
}

// IgnoreNowhere sets if the *TableView should ignore left mouse clicks in the
// empty space. It forbids the user from unselecting the current index, or when
// multi selection is enabled, disables click drag selection.
func (tv *TableView) SetIgnoreNowhere(value bool) {
	tv.ignoreNowhere = value
}

type tableViewState struct {
	SortColumnName     string
	SortOrder          SortOrder
	ColumnDisplayOrder []string
	Columns            []*tableViewColumnState
}

type tableViewColumnState struct {
	Name         string
	Title        string
	Width        int
	Visible      bool
	Frozen       bool
	LastSeenDate string
}

// SaveState writes the UI state of the *TableView to the settings.
func (tv *TableView) SaveState() error {
	if tv.columns.Len() == 0 {
		return nil
	}

	if tv.state == nil {
		tv.state = new(tableViewState)
	}

	tvs := tv.state

	tvs.SortColumnName = tv.columns.items[tv.sortedColumnIndex].name
	tvs.SortOrder = tv.sortOrder

	// tvs.Columns = make([]tableViewColumnState, tv.columns.Len())

	for _, tvc := range tv.columns.items {
		var tvcs *tableViewColumnState
		for _, cur := range tvs.Columns {
			if cur.Name == tvc.name {
				tvcs = cur
				break
			}
		}

		// tvcs := &tvs.Columns[i]

		if tvcs == nil {
			tvs.Columns = append(tvs.Columns, new(tableViewColumnState))
			tvcs = tvs.Columns[len(tvs.Columns)-1]
		}

		tvcs.Name = tvc.name
		tvcs.Title = tvc.titleOverride
		tvcs.Width = tvc.Width()
		tvcs.Visible = tvc.Visible()
		tvcs.Frozen = tvc.Frozen()
		tvcs.LastSeenDate = time.Now().Format("2006-01-02")
	}

	visibleCols := tv.visibleColumns()
	frozenCount := tv.visibleFrozenColumnCount()
	normalCount := len(visibleCols) - frozenCount
	indices := make([]int32, len(visibleCols))
	var lp uintptr
	if frozenCount > 0 {
		lp = uintptr(unsafe.Pointer(&indices[0]))

		if user32.SendMessage(tv.hwndFrozenLV, commctrl.LVM_GETCOLUMNORDERARRAY, uintptr(frozenCount), lp) == 0 {
			return errs.NewError("LVM_GETCOLUMNORDERARRAY")
		}
	}
	if normalCount > 0 {
		lp = uintptr(unsafe.Pointer(&indices[frozenCount]))

		if user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_GETCOLUMNORDERARRAY, uintptr(normalCount), lp) == 0 {
			return errs.NewError("LVM_GETCOLUMNORDERARRAY")
		}
	}

	tvs.ColumnDisplayOrder = make([]string, len(visibleCols))
	for i, j := range indices {
		if i >= frozenCount {
			j += int32(frozenCount)
		}
		tvs.ColumnDisplayOrder[i] = visibleCols[j].name
	}

	state, err := json.Marshal(tvs)
	if err != nil {
		return err
	}

	return tv.WriteState(string(state))
}

// RestoreState restores the UI state of the *TableView from the settings.
func (tv *TableView) RestoreState() error {
	state, err := tv.ReadState()
	if err != nil {
		return err
	}
	if state == "" {
		return nil
	}

	tv.SetSuspended(true)
	defer tv.SetSuspended(false)

	if tv.state == nil {
		tv.state = new(tableViewState)
	}

	tvs := tv.state

	if err := json.Unmarshal(([]byte)(state), tvs); err != nil {
		return err
	}

	name2tvc := make(map[string]*TableViewColumn)

	for _, tvc := range tv.columns.items {
		name2tvc[tvc.name] = tvc
	}

	name2tvcs := make(map[string]*tableViewColumnState)

	tvcsRetained := make([]*tableViewColumnState, 0, len(tvs.Columns))
	for _, tvcs := range tvs.Columns {
		if tvcs.LastSeenDate != "" {
			if lastSeen, err := time.Parse("2006-02-01", tvcs.LastSeenDate); err != nil {
				tvcs.LastSeenDate = ""
			} else if name2tvc[tvcs.Name] == nil && lastSeen.Add(time.Hour*24*90).Before(time.Now()) {
				continue
			}
		}
		tvcsRetained = append(tvcsRetained, tvcs)

		name2tvcs[tvcs.Name] = tvcsRetained[len(tvcsRetained)-1]

		if tvc := name2tvc[tvcs.Name]; tvc != nil {
			if err := tvc.SetFrozen(tvcs.Frozen); err != nil {
				return err
			}
			var visible bool
			for _, name := range tvs.ColumnDisplayOrder {
				if name == tvc.name {
					visible = true
					break
				}
			}
			if err := tvc.SetVisible(tvc.visible && (visible || tvcs.Visible)); err != nil {
				return err
			}
			if err := tvc.SetTitleOverride(tvcs.Title); err != nil {
				return err
			}
			if err := tvc.SetWidth(tvcs.Width); err != nil {
				return err
			}
		}
	}
	tvs.Columns = tvcsRetained

	visibleCount := tv.visibleColumnCount()
	frozenCount := tv.visibleFrozenColumnCount()
	normalCount := visibleCount - frozenCount

	indices := make([]int32, visibleCount)

	knownNames := make(map[string]struct{})

	displayOrder := make([]string, 0, visibleCount)
	for _, name := range tvs.ColumnDisplayOrder {
		knownNames[name] = struct{}{}
		if tvc, ok := name2tvc[name]; ok && tvc.visible {
			displayOrder = append(displayOrder, name)
		}
	}
	for _, tvc := range tv.visibleColumns() {
		if _, ok := knownNames[tvc.name]; !ok {
			displayOrder = append(displayOrder, tvc.name)
		}
	}

	for i, tvc := range tv.visibleColumns() {
		for j, name := range displayOrder {
			if tvc.name == name && j < visibleCount {
				idx := i
				if j >= frozenCount {
					idx -= frozenCount
				}
				indices[j] = int32(idx)
				break
			}
		}
	}

	var lp uintptr
	if frozenCount > 0 {
		lp = uintptr(unsafe.Pointer(&indices[0]))

		if user32.SendMessage(tv.hwndFrozenLV, commctrl.LVM_SETCOLUMNORDERARRAY, uintptr(frozenCount), lp) == 0 {
			return errs.NewError("LVM_SETCOLUMNORDERARRAY")
		}
	}
	if normalCount > 0 {
		lp = uintptr(unsafe.Pointer(&indices[frozenCount]))

		if user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_SETCOLUMNORDERARRAY, uintptr(normalCount), lp) == 0 {
			return errs.NewError("LVM_SETCOLUMNORDERARRAY")
		}
	}

	for i, c := range tvs.Columns {
		if c.Name == tvs.SortColumnName && i < visibleCount {
			tv.sortedColumnIndex = i
			tv.sortOrder = tvs.SortOrder
			break
		}
	}

	if sorter, ok := tv.model.(Sorter); ok {
		if !sorter.ColumnSortable(tv.sortedColumnIndex) {
			for i := range tvs.Columns {
				if sorter.ColumnSortable(i) {
					tv.sortedColumnIndex = i
					break
				}
			}
		}

		sorter.Sort(tv.sortedColumnIndex, tvs.SortOrder)
	}

	return nil
}

func (tv *TableView) toggleItemChecked(index int) error {
	checked := tv.itemChecker.Checked(index)

	if err := tv.itemChecker.SetChecked(index, !checked); err != nil {
		return errs.WrapError(err)
	}

	if win.FALSE == user32.SendMessage(tv.hwndFrozenLV, commctrl.LVM_UPDATE, uintptr(index), 0) {
		return errs.NewError("SendMessage(LVM_UPDATE)")
	}
	if win.FALSE == user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_UPDATE, uintptr(index), 0) {
		return errs.NewError("SendMessage(LVM_UPDATE)")
	}

	return nil
}

func (tv *TableView) applyImageListForImage(image interface{}) {
	tv.hIml, tv.usingSysIml, _ = imageListForImage(image, tv.DPI())

	tv.applyImageList()

	tv.imageUintptr2Index = make(map[uintptr]int32)
	tv.filePath2IconIndex = make(map[string]int32)
}

func (tv *TableView) applyImageList() {
	user32.SendMessage(tv.hwndFrozenLV, commctrl.LVM_SETIMAGELIST, commctrl.LVSIL_SMALL, uintptr(tv.hIml))
	user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_SETIMAGELIST, commctrl.LVSIL_SMALL, uintptr(tv.hIml))
}

func (tv *TableView) disposeImageListAndCaches() {
	if tv.hIml != 0 && !tv.usingSysIml {
		user32.SendMessage(tv.hwndFrozenLV, commctrl.LVM_SETIMAGELIST, commctrl.LVSIL_SMALL, 0)
		user32.SendMessage(tv.hwndNormalLV, commctrl.LVM_SETIMAGELIST, commctrl.LVSIL_SMALL, 0)

		comctl32.ImageList_Destroy(tv.hIml)
	}
	tv.hIml = 0

	tv.imageUintptr2Index = nil
	tv.filePath2IconIndex = nil
}

func (tv *TableView) Focused() bool {
	focused := user32.GetFocus()

	return focused == tv.hwndFrozenLV || focused == tv.hwndNormalLV
}

func (tv *TableView) maybePublishFocusChanged(hwnd handle.HWND, msg uint32, wp uintptr) {
	focused := msg == user32.WM_SETFOCUS

	if focused != tv.focused && wp != uintptr(tv.hwndFrozenLV) && wp != uintptr(tv.hwndNormalLV) {
		tv.focused = focused
		tv.focusedChangedPublisher.Publish()
	}
}

func tableViewFrozenLVWndProc(hwnd handle.HWND, msg uint32, wp, lp uintptr) uintptr {
	tv := (*TableView)(unsafe.Pointer(windowFromHandle(user32.GetParent(hwnd)).AsWindowBase()))

	switch msg {
	case user32.WM_NCCALCSIZE:
		ensureWindowLongBits(hwnd, user32.GWL_STYLE, user32.WS_HSCROLL|user32.WS_VSCROLL, false)

	case user32.WM_SETFOCUS:
		user32.SetFocus(tv.hwndNormalLV)
		tv.maybePublishFocusChanged(hwnd, msg, wp)

	case user32.WM_KILLFOCUS:
		tv.maybePublishFocusChanged(hwnd, msg, wp)

	case user32.WM_MOUSEWHEEL:
		tableViewNormalLVWndProc(tv.hwndNormalLV, msg, wp, lp)
	}

	return tv.lvWndProc(tv.frozenLVOrigWndProcPtr, hwnd, msg, wp, lp)
}

func tableViewNormalLVWndProc(hwnd handle.HWND, msg uint32, wp, lp uintptr) uintptr {
	tv := (*TableView)(unsafe.Pointer(windowFromHandle(user32.GetParent(hwnd)).AsWindowBase()))

	switch msg {
	case user32.WM_LBUTTONDOWN, user32.WM_RBUTTONDOWN:
		user32.SetFocus(tv.hwndFrozenLV)

	case user32.WM_SETFOCUS:
		tv.invalidateBorderInParent()
		tv.maybePublishFocusChanged(hwnd, msg, wp)

	case user32.WM_KILLFOCUS:
		user32.SendMessage(tv.hwndFrozenLV, msg, wp, lp)
		tv.WndProc(tv.hWnd, msg, wp, lp)
		tv.maybePublishFocusChanged(hwnd, msg, wp)
	}

	result := tv.lvWndProc(tv.normalLVOrigWndProcPtr, hwnd, msg, wp, lp)

	var off uint32 = user32.WS_HSCROLL | user32.WS_VSCROLL
	if tv.scrollbarOrientation&Horizontal != 0 {
		off &^= user32.WS_HSCROLL
	}
	if tv.scrollbarOrientation&Vertical != 0 {
		off &^= user32.WS_VSCROLL
	}
	if off != 0 {
		ensureWindowLongBits(hwnd, user32.GWL_STYLE, off, false)
	}

	return result
}

func (tv *TableView) lvWndProc(origWndProcPtr uintptr, hwnd handle.HWND, msg uint32, wp, lp uintptr) uintptr {
	var hwndOther handle.HWND
	if hwnd == tv.hwndFrozenLV {
		hwndOther = tv.hwndNormalLV
	} else {
		hwndOther = tv.hwndFrozenLV
	}

	var maybeStretchLastColumn bool

	switch msg {
	case user32.WM_ERASEBKGND:
		maybeStretchLastColumn = true

	case user32.WM_WINDOWPOSCHANGED:
		wp := (*user32.WINDOWPOS)(unsafe.Pointer(lp))

		if wp.Flags&user32.SWP_NOSIZE != 0 {
			break
		}

		maybeStretchLastColumn = int(wp.Cx) < tv.WidthPixels()

	case user32.WM_GETDLGCODE:
		if wp == user32.VK_RETURN {
			return user32.DLGC_WANTALLKEYS
		}

	case user32.WM_LBUTTONDOWN, user32.WM_RBUTTONDOWN, user32.WM_LBUTTONDBLCLK, user32.WM_RBUTTONDBLCLK:
		var hti commctrl.LVHITTESTINFO
		hti.Pt = gdi32.POINT{X: user32.GET_X_LPARAM(lp), Y: user32.GET_Y_LPARAM(lp)}
		user32.SendMessage(hwnd, commctrl.LVM_HITTEST, 0, uintptr(unsafe.Pointer(&hti)))

		tv.itemIndexOfLastMouseButtonDown = int(hti.IItem)

		if hti.Flags == commctrl.LVHT_NOWHERE {
			if tv.MultiSelection() {
				tv.publishNextSelClear = true
			} else {
				if tv.CheckBoxes() {
					if tv.currentIndex > -1 {
						tv.SetCurrentIndex(-1)
					}
				} else {
					// We keep the current item, if in single item selection mode without check boxes.
					user32.SetFocus(tv.hwndFrozenLV)
					return 0
				}
			}

			if tv.IgnoreNowhere() {
				return 0
			}
		}

		switch msg {
		case user32.WM_LBUTTONDOWN, user32.WM_RBUTTONDOWN:
			if hti.Flags == commctrl.LVHT_ONITEMSTATEICON &&
				tv.itemChecker != nil &&
				tv.CheckBoxes() {

				tv.toggleItemChecked(int(hti.IItem))
			}

		case user32.WM_LBUTTONDBLCLK, user32.WM_RBUTTONDBLCLK:
			if tv.currentIndex != tv.prevIndex && tv.itemStateChangedEventDelay > 0 {
				tv.prevIndex = tv.currentIndex
				tv.currentIndexChangedPublisher.Publish()
				tv.currentItemChangedPublisher.Publish()
			}
		}

	case user32.WM_LBUTTONUP, user32.WM_RBUTTONUP:
		tv.itemIndexOfLastMouseButtonDown = -1

	case user32.WM_MOUSEMOVE, user32.WM_MOUSELEAVE:
		if tv.inMouseEvent {
			break
		}
		tv.inMouseEvent = true
		defer func() {
			tv.inMouseEvent = false
		}()

		if msg == user32.WM_MOUSEMOVE {
			y := int(user32.GET_Y_LPARAM(lp))
			lp = uintptr(win.MAKELONG(0, uint16(y)))
		}

		user32.SendMessage(hwndOther, msg, wp, lp)

	case user32.WM_KEYDOWN:
		if wp == user32.VK_SPACE &&
			tv.currentIndex > -1 &&
			tv.itemChecker != nil &&
			tv.CheckBoxes() {

			tv.toggleItemChecked(tv.currentIndex)
		}

		tv.handleKeyDown(wp, lp)

	case user32.WM_KEYUP:
		tv.handleKeyUp(wp, lp)

	case user32.WM_NOTIFY:
		nmh := ((*user32.NMHDR)(unsafe.Pointer(lp)))
		switch nmh.HwndFrom {
		case tv.hwndFrozenHdr, tv.hwndNormalHdr:
			if nmh.Code == comctl32.NM_CUSTOMDRAW {
				return tableViewHdrWndProc(nmh.HwndFrom, msg, wp, lp)
			}
		}

		switch nmh.Code {
		case commctrl.LVN_GETDISPINFO:
			di := (*commctrl.NMLVDISPINFO)(unsafe.Pointer(lp))

			row := int(di.Item.IItem)
			col := tv.fromLVColIdx(hwnd == tv.hwndFrozenLV, di.Item.ISubItem)
			if col == -1 {
				break
			}

			if di.Item.Mask&commctrl.LVIF_TEXT > 0 {
				value := tv.model.Value(row, col)
				var text string
				if format := tv.columns.items[col].formatFunc; format != nil {
					text = format(value)
				} else {
					switch val := value.(type) {
					case string:
						text = val

					case float32:
						prec := tv.columns.items[col].precision
						if prec == 0 {
							prec = 2
						}
						text = FormatFloatGrouped(float64(val), prec)

					case float64:
						prec := tv.columns.items[col].precision
						if prec == 0 {
							prec = 2
						}
						text = FormatFloatGrouped(val, prec)

					case time.Time:
						if val.Year() > 1601 {
							text = val.Format(tv.columns.items[col].format)
						}

					case bool:
						if val {
							text = checkmark
						}

					case *big.Rat:
						prec := tv.columns.items[col].precision
						if prec == 0 {
							prec = 2
						}
						text = formatBigRatGrouped(val, prec)

					default:
						text = fmt.Sprintf(tv.columns.items[col].format, val)
					}
				}
				utf16, err := syscall.UTF16FromString(text)
				if err != nil {
					errs.NewError(err.Error())
				}
				buf := (*[264]uint16)(unsafe.Pointer(di.Item.PszText))
				max := mini(len(utf16), int(di.Item.CchTextMax))
				copy((*buf)[:], utf16[:max])
				(*buf)[max-1] = 0
			}

			if (tv.imageProvider != nil || tv.styler != nil) && di.Item.Mask&commctrl.LVIF_IMAGE > 0 {
				var image interface{}
				if di.Item.ISubItem == 0 {
					if ip := tv.imageProvider; ip != nil && image == nil {
						image = ip.Image(row)
					}
				}
				if styler := tv.styler; styler != nil && image == nil {
					tv.style.row = row
					tv.style.col = col
					tv.style.bounds = Rectangle{}
					tv.style.dpi = tv.DPI()
					tv.style.Image = nil

					styler.StyleCell(&tv.style)

					image = tv.style.Image
				}

				if image != nil {
					if tv.hIml == 0 {
						tv.applyImageListForImage(image)
					}

					di.Item.IImage = imageIndexMaybeAdd(
						image,
						tv.hIml,
						tv.usingSysIml,
						tv.imageUintptr2Index,
						tv.filePath2IconIndex,
						tv.DPI())
				}
			}

			if di.Item.ISubItem == 0 && di.Item.StateMask&commctrl.LVIS_STATEIMAGEMASK > 0 &&
				tv.itemChecker != nil {
				checked := tv.itemChecker.Checked(row)

				if checked {
					di.Item.State = 0x2000
				} else {
					di.Item.State = 0x1000
				}
			}

		case comctl32.NM_CUSTOMDRAW:
			nmlvcd := (*commctrl.NMLVCUSTOMDRAW)(unsafe.Pointer(lp))

			if nmlvcd.IIconPhase == 0 {
				row := int(nmlvcd.Nmcd.DwItemSpec)
				col := tv.fromLVColIdx(hwnd == tv.hwndFrozenLV, nmlvcd.ISubItem)
				if col == -1 {
					break
				}

				applyCellStyle := func() int {
					if tv.styler != nil {
						dpi := tv.DPI()

						tv.style.row = row
						tv.style.col = col
						tv.style.bounds = rectangleFromRECT(nmlvcd.Nmcd.Rc)
						tv.style.dpi = dpi
						tv.style.hdc = nmlvcd.Nmcd.Hdc
						tv.style.BackgroundColor = tv.itemBGColor
						tv.style.TextColor = tv.itemTextColor
						tv.style.Font = nil
						tv.style.Image = nil

						tv.styler.StyleCell(&tv.style)

						defer func() {
							tv.style.bounds = Rectangle{}
							if tv.style.canvas != nil {
								tv.style.canvas.Dispose()
								tv.style.canvas = nil
							}
							tv.style.hdc = 0
						}()

						if tv.style.canvas != nil {
							return comctl32.CDRF_SKIPDEFAULT
						}

						nmlvcd.ClrTextBk = gdi32.COLORREF(tv.style.BackgroundColor)
						nmlvcd.ClrText = gdi32.COLORREF(tv.style.TextColor)

						font := tv.style.Font
						if font == nil {
							font = tv.Font()
						}
						gdi32.SelectObject(nmlvcd.Nmcd.Hdc, gdi32.HGDIOBJ(font.handleForDPI(dpi)))
					}

					return 0
				}

				switch nmlvcd.Nmcd.DwDrawStage {
				case comctl32.CDDS_PREPAINT:
					return comctl32.CDRF_NOTIFYITEMDRAW

				case comctl32.CDDS_ITEMPREPAINT:
					var selected bool
					if itemState := user32.SendMessage(hwnd, commctrl.LVM_GETITEMSTATE, nmlvcd.Nmcd.DwItemSpec, commctrl.LVIS_SELECTED); itemState&commctrl.LVIS_SELECTED != 0 {
						selected = true

						tv.itemBGColor = tv.themeSelectedBGColor
						tv.itemTextColor = tv.themeSelectedTextColor
					} else {
						tv.itemBGColor = tv.themeNormalBGColor
						tv.itemTextColor = tv.themeNormalTextColor
					}

					if !selected && tv.alternatingRowBG && row%2 == 1 {
						tv.itemBGColor = tv.alternatingRowBGColor
						tv.itemTextColor = tv.alternatingRowTextColor
					}

					tv.style.BackgroundColor = tv.itemBGColor
					tv.style.TextColor = tv.itemTextColor

					if tv.styler != nil {
						tv.style.row = row
						tv.style.col = -1
						tv.style.bounds = rectangleFromRECT(nmlvcd.Nmcd.Rc)
						tv.style.dpi = tv.DPI()
						tv.style.hdc = 0
						tv.style.Font = nil
						tv.style.Image = nil

						tv.styler.StyleCell(&tv.style)

						tv.itemFont = tv.style.Font
					}

					if selected {
						tv.style.BackgroundColor = tv.itemBGColor
						tv.style.TextColor = tv.itemTextColor
					} else {
						tv.itemBGColor = tv.style.BackgroundColor
						tv.itemTextColor = tv.style.TextColor
					}

					if tv.style.BackgroundColor != tv.themeNormalBGColor {
						var color Color
						if selected && !tv.Focused() {
							color = tv.themeSelectedNotFocusedBGColor
						} else {
							color = tv.style.BackgroundColor
						}

						if brush, _ := NewSolidColorBrush(color); brush != nil {
							defer brush.Dispose()

							canvas, _ := newCanvasFromHDC(nmlvcd.Nmcd.Hdc)
							canvas.FillRectanglePixels(brush, rectangleFromRECT(nmlvcd.Nmcd.Rc))
						}
					}

					nmlvcd.ClrText = gdi32.COLORREF(tv.style.TextColor)
					nmlvcd.ClrTextBk = gdi32.COLORREF(tv.style.BackgroundColor)

					return comctl32.CDRF_NOTIFYSUBITEMDRAW

				case comctl32.CDDS_ITEMPREPAINT | comctl32.CDDS_SUBITEM:
					if tv.itemFont != nil {
						gdi32.SelectObject(nmlvcd.Nmcd.Hdc, gdi32.HGDIOBJ(tv.itemFont.handleForDPI(tv.DPI())))
					}

					if applyCellStyle() == comctl32.CDRF_SKIPDEFAULT && uxtheme.IsAppThemed() {
						return comctl32.CDRF_SKIPDEFAULT
					}

					return comctl32.CDRF_NEWFONT | comctl32.CDRF_SKIPPOSTPAINT | comctl32.CDRF_NOTIFYPOSTPAINT

				case comctl32.CDDS_ITEMPOSTPAINT | comctl32.CDDS_SUBITEM:
					if applyCellStyle() == comctl32.CDRF_SKIPDEFAULT {
						return comctl32.CDRF_SKIPDEFAULT
					}

					return comctl32.CDRF_NEWFONT | comctl32.CDRF_SKIPPOSTPAINT
				}

				return comctl32.CDRF_SKIPPOSTPAINT
			}

			return comctl32.CDRF_SKIPPOSTPAINT

		case commctrl.LVN_BEGINSCROLL:
			if tv.scrolling {
				break
			}
			tv.scrolling = true
			defer func() {
				tv.scrolling = false
			}()

			var rc gdi32.RECT
			user32.SendMessage(hwnd, commctrl.LVM_GETITEMRECT, 0, uintptr(unsafe.Pointer(&rc)))

			nmlvs := (*commctrl.NMLVSCROLL)(unsafe.Pointer(lp))
			user32.SendMessage(hwndOther, commctrl.LVM_SCROLL, 0, uintptr(nmlvs.Dy*(rc.Bottom-rc.Top)))

		case commctrl.LVN_COLUMNCLICK:
			nmlv := (*commctrl.NMLISTVIEW)(unsafe.Pointer(lp))

			col := tv.fromLVColIdx(hwnd == tv.hwndFrozenLV, nmlv.ISubItem)

			if sorter, ok := tv.model.(Sorter); ok && sorter.ColumnSortable(col) {
				prevCol := sorter.SortedColumn()
				var order SortOrder
				if col != prevCol || sorter.SortOrder() == SortDescending {
					order = SortAscending
				} else {
					order = SortDescending
				}
				tv.sortedColumnIndex = col
				tv.sortOrder = order
				sorter.Sort(col, order)
			}

			tv.columnClickedPublisher.Publish(col)

		case commctrl.LVN_ITEMCHANGED:
			nmlv := (*commctrl.NMLISTVIEW)(unsafe.Pointer(lp))

			if tv.hwndItemChanged != 0 && tv.hwndItemChanged != hwnd {
				break
			}
			tv.hwndItemChanged = hwnd
			defer func() {
				tv.hwndItemChanged = 0
			}()

			tv.copySelectedIndexes(hwndOther, hwnd)

			if nmlv.IItem == -1 && !tv.publishNextSelClear {
				break
			}
			tv.publishNextSelClear = false

			selectedNow := nmlv.UNewState&commctrl.LVIS_SELECTED > 0
			selectedBefore := nmlv.UOldState&commctrl.LVIS_SELECTED > 0
			if tv.itemIndexOfLastMouseButtonDown != -1 && selectedNow && !selectedBefore && ModifiersDown()&(ModControl|ModShift) == 0 {
				tv.prevIndex = tv.currentIndex
				tv.currentIndex = int(nmlv.IItem)
				if tv.itemStateChangedEventDelay > 0 {
					tv.delayedCurrentIndexChangedCanceled = false
					if user32.SetTimer(
						tv.hWnd,
						tableViewCurrentIndexChangedTimerId,
						uint32(tv.itemStateChangedEventDelay),
						0) == 0 {

						errs.LastError("SetTimer")
					}

					tv.SetCurrentIndex(int(nmlv.IItem))
				} else {
					tv.SetCurrentIndex(int(nmlv.IItem))
				}
			}

			if selectedNow != selectedBefore {
				if !tv.inSetSelectedIndexes && tv.MultiSelection() {
					tv.updateSelectedIndexes()
				}
			}

		case commctrl.LVN_ODSTATECHANGED:
			if tv.hwndItemChanged != 0 && tv.hwndItemChanged != hwnd {
				break
			}
			tv.hwndItemChanged = hwnd
			defer func() {
				tv.hwndItemChanged = 0
			}()

			tv.copySelectedIndexes(hwndOther, hwnd)

			tv.updateSelectedIndexes()

		case commctrl.LVN_ITEMACTIVATE:
			nmia := (*commctrl.NMITEMACTIVATE)(unsafe.Pointer(lp))

			if tv.itemStateChangedEventDelay > 0 {
				tv.delayedCurrentIndexChangedCanceled = true
			}

			if int(nmia.IItem) != tv.currentIndex {
				tv.SetCurrentIndex(int(nmia.IItem))
				tv.currentIndexChangedPublisher.Publish()
				tv.currentItemChangedPublisher.Publish()
			}

			tv.itemActivatedPublisher.Publish()

		case commctrl.HDN_ITEMCHANGING:
			tv.updateLVSizes()
		}

	case user32.WM_UPDATEUISTATE:
		switch win.LOWORD(uint32(wp)) {
		case user32.UIS_SET:
			wp |= user32.UISF_HIDEFOCUS << 16

		case user32.UIS_CLEAR, user32.UIS_INITIALIZE:
			wp &^= ^uintptr(user32.UISF_HIDEFOCUS << 16)
		}
	}

	lpFixed := lp
	fixXInLP := func() {
		// fmt.Printf("hwnd == tv.hwndNormalLV: %t, tv.hasFrozenColumn: %t\n", hwnd == tv.hwndNormalLV, tv.hasFrozenColumn)
		if hwnd == tv.hwndNormalLV && tv.hasFrozenColumn {
			var rc gdi32.RECT
			if user32.GetWindowRect(tv.hwndFrozenLV, &rc) {
				x := int(user32.GET_X_LPARAM(lp)) + int(rc.Right-rc.Left)
				y := int(user32.GET_Y_LPARAM(lp))

				lpFixed = uintptr(win.MAKELONG(uint16(x), uint16(y)))
			}
		}
	}

	switch msg {
	case user32.WM_LBUTTONDOWN, user32.WM_MBUTTONDOWN, user32.WM_RBUTTONDOWN:
		fixXInLP()
		tv.publishMouseEvent(&tv.mouseDownPublisher, msg, wp, lpFixed)

	case user32.WM_LBUTTONUP, user32.WM_MBUTTONUP, user32.WM_RBUTTONUP:
		fixXInLP()
		tv.publishMouseEvent(&tv.mouseUpPublisher, msg, wp, lpFixed)

	case user32.WM_MOUSEMOVE:
		fixXInLP()
		tv.publishMouseEvent(&tv.mouseMovePublisher, msg, wp, lpFixed)

	case user32.WM_MOUSEWHEEL:
		fixXInLP()
		tv.publishMouseWheelEvent(&tv.mouseWheelPublisher, wp, lpFixed)
	}

	if maybeStretchLastColumn {
		if tv.lastColumnStretched && !tv.busyStretchingLastColumn {
			if normalVisColCount := tv.visibleColumnCount() - tv.visibleFrozenColumnCount(); normalVisColCount == 0 || normalVisColCount > 0 == (hwnd == tv.hwndNormalLV) {
				tv.busyStretchingLastColumn = true
				defer func() {
					tv.busyStretchingLastColumn = false
				}()
				tv.StretchLastColumn()
			}
		}

		if msg == user32.WM_ERASEBKGND {
			return 1
		}
	}

	return user32.CallWindowProc(origWndProcPtr, hwnd, msg, wp, lp)
}

func tableViewHdrWndProc(hwnd handle.HWND, msg uint32, wp, lp uintptr) uintptr {
	tv := (*TableView)(unsafe.Pointer(windowFromHandle(user32.GetParent(user32.GetParent(hwnd))).AsWindowBase()))

	var origWndProcPtr uintptr
	if hwnd == tv.hwndFrozenHdr {
		origWndProcPtr = tv.frozenHdrOrigWndProcPtr
	} else {
		origWndProcPtr = tv.normalHdrOrigWndProcPtr
	}

	switch msg {
	case user32.WM_NOTIFY:
		switch ((*user32.NMHDR)(unsafe.Pointer(lp))).Code {
		case comctl32.NM_CUSTOMDRAW:
			if tv.customHeaderHeight == 0 {
				break
			}

			nmcd := (*comctl32.NMCUSTOMDRAW)(unsafe.Pointer(lp))

			switch nmcd.DwDrawStage {
			case comctl32.CDDS_PREPAINT:
				return comctl32.CDRF_NOTIFYITEMDRAW

			case comctl32.CDDS_ITEMPREPAINT:
				return comctl32.CDRF_NOTIFYPOSTPAINT

			case comctl32.CDDS_ITEMPOSTPAINT:
				col := tv.fromLVColIdx(hwnd == tv.hwndFrozenHdr, int32(nmcd.DwItemSpec))
				if tv.styler != nil && col > -1 {
					tv.style.row = -1
					tv.style.col = col
					tv.style.bounds = rectangleFromRECT(nmcd.Rc)
					tv.style.dpi = tv.DPI()
					tv.style.hdc = nmcd.Hdc
					tv.style.TextColor = tv.themeNormalTextColor
					tv.style.Font = nil

					tv.styler.StyleCell(&tv.style)

					defer func() {
						tv.style.bounds = Rectangle{}
						if tv.style.canvas != nil {
							tv.style.canvas.Dispose()
							tv.style.canvas = nil
						}
						tv.style.hdc = 0
					}()
				}

				return comctl32.CDRF_DODEFAULT
			}

			return comctl32.CDRF_DODEFAULT
		}

	case commctrl.HDM_LAYOUT:
		if tv.customHeaderHeight == 0 {
			break
		}

		result := user32.CallWindowProc(origWndProcPtr, hwnd, msg, wp, lp)

		hdl := (*commctrl.HDLAYOUT)(unsafe.Pointer(lp))
		hdl.Prc.Top = int32(tv.customHeaderHeight)
		hdl.Pwpos.Cy = int32(tv.customHeaderHeight)

		return result

	case user32.WM_MOUSEMOVE, user32.WM_LBUTTONDOWN, user32.WM_LBUTTONUP, user32.WM_MBUTTONDOWN, user32.WM_MBUTTONUP, user32.WM_RBUTTONDOWN, user32.WM_RBUTTONUP:
		hti := commctrl.HDHITTESTINFO{Pt: gdi32.POINT{X: int32(user32.GET_X_LPARAM(lp)), Y: int32(user32.GET_Y_LPARAM(lp))}}
		user32.SendMessage(hwnd, commctrl.HDM_HITTEST, 0, uintptr(unsafe.Pointer(&hti)))
		if hti.IItem == -1 {
			tv.group.toolTip.setText(hwnd, "")
			break
		}

		col := tv.fromLVColIdx(hwnd == tv.hwndFrozenHdr, hti.IItem)
		text := tv.columns.At(col).TitleEffective()

		var rc gdi32.RECT
		if user32.SendMessage(hwnd, commctrl.HDM_GETITEMRECT, uintptr(hti.IItem), uintptr(unsafe.Pointer(&rc))) == 0 {
			tv.group.toolTip.setText(hwnd, "")
			break
		}

		size := calculateTextSize(text, tv.Font(), tv.DPI(), 0, hwnd)
		if size.Width <= rectangleFromRECT(rc).Width-int(user32.SendMessage(hwnd, commctrl.HDM_GETBITMAPMARGIN, 0, 0)) {
			tv.group.toolTip.setText(hwnd, "")
			break
		}

		if tv.group.toolTip.text(hwnd) == text {
			break
		}

		tv.group.toolTip.setText(hwnd, text)

		m := user32.MSG{
			HWnd:    hwnd,
			Message: msg,
			WParam:  wp,
			LParam:  lp,
			Pt:      hti.Pt,
		}

		tv.group.toolTip.SendMessage(commctrl.TTM_RELAYEVENT, 0, uintptr(unsafe.Pointer(&m)))
	}

	return user32.CallWindowProc(origWndProcPtr, hwnd, msg, wp, lp)
}

func (tv *TableView) WndProc(hwnd handle.HWND, msg uint32, wp, lp uintptr) uintptr {
	switch msg {
	case user32.WM_NOTIFY:
		nmh := (*user32.NMHDR)(unsafe.Pointer(lp))
		switch nmh.HwndFrom {
		case tv.hwndFrozenLV:
			return tableViewFrozenLVWndProc(nmh.HwndFrom, msg, wp, lp)

		case tv.hwndNormalLV:
			return tableViewNormalLVWndProc(nmh.HwndFrom, msg, wp, lp)
		}

	case user32.WM_WINDOWPOSCHANGED:
		wp := (*user32.WINDOWPOS)(unsafe.Pointer(lp))

		if wp.Flags&user32.SWP_NOSIZE != 0 {
			break
		}

		if tv.formActivatingHandle == -1 {
			if form := tv.Form(); form != nil {
				tv.formActivatingHandle = form.Activating().Attach(func() {
					if tv.hwndNormalLV == user32.GetFocus() {
						user32.SetFocus(tv.hwndFrozenLV)
					}
				})
			}
		}

		tv.updateLVSizes()

		// FIXME: The InvalidateRect and redrawItems calls below prevent
		// painting glitches on resize. Though this seems to work reasonably
		// well, in the long run we would like to find the root cause of this
		// issue and come up with a better fix.
		dpi := uint32(tv.DPI())
		var rc gdi32.RECT

		vsbWidth := user32.GetSystemMetricsForDpi(user32.SM_CXVSCROLL, dpi)
		rc = gdi32.RECT{Left: wp.Cx - vsbWidth - 1, Top: 0, Right: wp.Cx, Bottom: wp.Cy}
		user32.InvalidateRect(tv.hWnd, &rc, true)

		hsbHeight := user32.GetSystemMetricsForDpi(user32.SM_CYHSCROLL, dpi)
		rc = gdi32.RECT{Left: 0, Right: wp.Cy - hsbHeight - 1, Top: wp.Cx, Bottom: wp.Cy}
		user32.InvalidateRect(tv.hWnd, &rc, true)

		tv.redrawItems()

	case user32.WM_TIMER:
		if !user32.KillTimer(tv.hWnd, wp) {
			errs.LastError("KillTimer")
		}

		switch wp {
		case tableViewCurrentIndexChangedTimerId:
			if !tv.delayedCurrentIndexChangedCanceled {
				tv.currentIndexChangedPublisher.Publish()
				tv.currentItemChangedPublisher.Publish()
			}

		case tableViewSelectedIndexesChangedTimerId:
			tv.selectedIndexesChangedPublisher.Publish()
		}

	case user32.WM_MEASUREITEM:
		mis := (*user32.MEASUREITEMSTRUCT)(unsafe.Pointer(lp))
		mis.ItemHeight = uint32(tv.customRowHeight)

		ensureWindowLongBits(tv.hwndFrozenLV, user32.GWL_STYLE, commctrl.LVS_OWNERDRAWFIXED, false)
		ensureWindowLongBits(tv.hwndNormalLV, user32.GWL_STYLE, commctrl.LVS_OWNERDRAWFIXED, false)

	case user32.WM_SETFOCUS:
		user32.SetFocus(tv.hwndFrozenLV)

	case user32.WM_DESTROY:
		// As we subclass all windows of system classes, we prevented the
		// clean-up code in the WM_NCDESTROY handlers of some windows from
		// being called. To fix this, we restore the original window
		// procedures here.
		if tv.frozenHdrOrigWndProcPtr != 0 {
			user32.SetWindowLongPtr(tv.hwndFrozenHdr, user32.GWLP_WNDPROC, tv.frozenHdrOrigWndProcPtr)
		}
		if tv.frozenLVOrigWndProcPtr != 0 {
			user32.SetWindowLongPtr(tv.hwndFrozenLV, user32.GWLP_WNDPROC, tv.frozenLVOrigWndProcPtr)
		}
		if tv.normalHdrOrigWndProcPtr != 0 {
			user32.SetWindowLongPtr(tv.hwndNormalHdr, user32.GWLP_WNDPROC, tv.normalHdrOrigWndProcPtr)
		}
		if tv.normalLVOrigWndProcPtr != 0 {
			user32.SetWindowLongPtr(tv.hwndNormalLV, user32.GWLP_WNDPROC, tv.normalLVOrigWndProcPtr)
		}
	}

	return tv.WidgetBase.WndProc(hwnd, msg, wp, lp)
}

func (tv *TableView) updateLVSizes() {
	tv.updateLVSizesWithSpecialCare(false)
}

func (tv *TableView) updateLVSizesWithSpecialCare(needSpecialCare bool) {
	var width int
	for i := tv.columns.Len() - 1; i >= 0; i-- {
		if col := tv.columns.At(i); col.frozen && col.visible {
			width += col.Width()
		}
	}

	dpi := tv.DPI()
	widthPixels := IntFrom96DPI(width, dpi)

	cb := tv.ClientBoundsPixels()

	user32.MoveWindow(tv.hwndNormalLV, int32(widthPixels), 0, int32(cb.Width-widthPixels), int32(cb.Height), true)

	var sbh int
	if hasWindowLongBits(tv.hwndNormalLV, user32.GWL_STYLE, user32.WS_HSCROLL) {
		sbh = int(user32.GetSystemMetricsForDpi(user32.SM_CYHSCROLL, uint32(dpi)))
	}

	user32.MoveWindow(tv.hwndFrozenLV, 0, 0, int32(widthPixels), int32(cb.Height-sbh), true)

	if needSpecialCare {
		tv.updateLVSizesNeedsSpecialCare = true
	}

	if tv.updateLVSizesNeedsSpecialCare {
		user32.ShowWindow(tv.hwndNormalLV, user32.SW_HIDE)
		user32.ShowWindow(tv.hwndNormalLV, user32.SW_SHOW)
	}

	if !needSpecialCare {
		tv.updateLVSizesNeedsSpecialCare = false
	}
}

func (*TableView) CreateLayoutItem(ctx *LayoutContext) LayoutItem {
	return NewGreedyLayoutItem()
}

func (tv *TableView) SetScrollbarOrientation(orientation Orientation) {
	tv.scrollbarOrientation = orientation
}

func (tv *TableView) ScrollbarOrientation() Orientation {
	return tv.scrollbarOrientation
}

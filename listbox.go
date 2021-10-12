// Copyright 2012 The Walk Authors. All rights reserved.
// Use of lb source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"fmt"
	"math/big"
	"reflect"
	"syscall"
	"time"
	"unsafe"

	"github.com/Gipcomp/win32/commctrl"
	"github.com/Gipcomp/win32/gdi32"
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/uxtheme"
	"github.com/Gipcomp/win32/win"
	"github.com/Gipcomp/win32/winuser"
	"github.com/Gipcomp/winapi/errs"
)

type ListBox struct {
	WidgetBase
	bindingValueProvider            BindingValueProvider
	model                           ListModel
	providedModel                   interface{}
	styler                          ListItemStyler
	style                           ListItemStyle
	bindingMember                   string
	displayMember                   string
	format                          string
	precision                       int
	prevCurIndex                    int
	currentValue                    interface{}
	itemsResetHandlerHandle         int
	itemChangedHandlerHandle        int
	itemsInsertedHandlerHandle      int
	itemsRemovedHandlerHandle       int
	maxItemTextWidth                int   // in native pixels
	lastWidth                       int   // in native pixels
	lastWidthsMeasuredFor           []int // in native pixels
	currentIndexChangedPublisher    EventPublisher
	selectedIndexesChangedPublisher EventPublisher
	itemActivatedPublisher          EventPublisher
	themeNormalBGColor              Color
	themeNormalTextColor            Color
	themeSelectedBGColor            Color
	themeSelectedTextColor          Color
	themeSelectedNotFocusedBGColor  Color
	trackingMouseEvent              bool
}

func NewListBox(parent Container) (*ListBox, error) {
	return NewListBoxWithStyle(parent, 0)
}

func NewListBoxWithStyle(parent Container, style uint32) (*ListBox, error) {
	lb := new(ListBox)

	err := InitWidget(
		lb,
		parent,
		"LISTBOX",
		user32.WS_BORDER|user32.WS_TABSTOP|user32.WS_VISIBLE|user32.WS_VSCROLL|user32.WS_HSCROLL|winuser.LBS_NOINTEGRALHEIGHT|winuser.LBS_NOTIFY|style,
		0)
	if err != nil {
		return nil, err
	}

	succeeded := false
	defer func() {
		if !succeeded {
			lb.Dispose()
		}
	}()

	lb.setTheme("Explorer")

	lb.style.dpi = lb.DPI()

	lb.ApplySysColors()

	lb.GraphicsEffects().Add(InteractionEffect)
	lb.GraphicsEffects().Add(FocusEffect)

	lb.MustRegisterProperty("CurrentIndex", NewProperty(
		func() interface{} {
			return lb.CurrentIndex()
		},
		func(v interface{}) error {
			return lb.SetCurrentIndex(assertIntOr(v, -1))
		},
		lb.CurrentIndexChanged()))

	lb.MustRegisterProperty("CurrentItem", NewReadOnlyProperty(
		func() interface{} {
			if i := lb.CurrentIndex(); i > -1 {
				if rm, ok := lb.providedModel.(reflectModel); ok {
					return reflect.ValueOf(rm.Items()).Index(i).Interface()
				}
			}

			return nil
		},
		lb.CurrentIndexChanged()))

	lb.MustRegisterProperty("HasCurrentItem", NewReadOnlyBoolProperty(
		func() bool {
			return lb.CurrentIndex() != -1
		},
		lb.CurrentIndexChanged()))

	lb.MustRegisterProperty("Value", NewProperty(
		func() interface{} {
			index := lb.CurrentIndex()

			if lb.bindingValueProvider == nil || index == -1 {
				return nil
			}

			return lb.bindingValueProvider.BindingValue(index)
		},
		func(v interface{}) error {
			if lb.bindingValueProvider == nil {
				if lb.model == nil {
					return nil
				} else {
					return errs.NewError("Data binding is only supported using a model that implements BindingValueProvider.")
				}
			}

			index := -1

			count := lb.model.ItemCount()
			for i := 0; i < count; i++ {
				if lb.bindingValueProvider.BindingValue(i) == v {
					index = i
					break
				}
			}

			return lb.SetCurrentIndex(index)
		},
		lb.CurrentIndexChanged()))

	succeeded = true

	return lb, nil
}

func (*ListBox) LayoutFlags() LayoutFlags {
	return ShrinkableHorz | ShrinkableVert | GrowableHorz | GrowableVert | GreedyHorz | GreedyVert
}

func (lb *ListBox) ItemStyler() ListItemStyler {
	return lb.styler
}

func (lb *ListBox) SetItemStyler(styler ListItemStyler) {
	lb.styler = styler
}

func (lb *ListBox) ApplySysColors() {
	lb.WidgetBase.ApplySysColors()

	var hc user32.HIGHCONTRAST
	hc.CbSize = uint32(unsafe.Sizeof(hc))
	if user32.SystemParametersInfo(user32.SPI_GETHIGHCONTRAST, hc.CbSize, unsafe.Pointer(&hc), 0) {
		lb.style.highContrastActive = hc.DwFlags&user32.HCF_HIGHCONTRASTON != 0
	}

	lb.themeNormalBGColor = Color(user32.GetSysColor(user32.COLOR_WINDOW))
	lb.themeNormalTextColor = Color(user32.GetSysColor(user32.COLOR_WINDOWTEXT))
	lb.themeSelectedBGColor = Color(user32.GetSysColor(user32.COLOR_HIGHLIGHT))
	lb.themeSelectedTextColor = Color(user32.GetSysColor(user32.COLOR_HIGHLIGHTTEXT))
	lb.themeSelectedNotFocusedBGColor = Color(user32.GetSysColor(user32.COLOR_BTNFACE))
}

func (lb *ListBox) ApplyDPI(dpi int) {
	lb.style.dpi = dpi

	lb.WidgetBase.ApplyDPI(dpi)
}

func (lb *ListBox) applyFont(font *Font) {
	lb.WidgetBase.applyFont(font)

	for i := range lb.lastWidthsMeasuredFor {
		lb.lastWidthsMeasuredFor[i] = 0
	}
}

func (lb *ListBox) itemString(index int) string {
	switch val := lb.model.Value(index).(type) {
	case string:
		return val

	case time.Time:
		return val.Format(lb.format)

	case *big.Rat:
		return val.FloatString(lb.precision)

	default:
		return fmt.Sprintf(lb.format, val)
	}
}

//insert one item from list model
func (lb *ListBox) insertItemAt(index int) error {
	str := lb.itemString(index)
	strPtr, err := syscall.UTF16PtrFromString(str)
	if err != nil {
		errs.NewError(err.Error())
	}
	lp := uintptr(unsafe.Pointer(strPtr))
	ret := int(lb.SendMessage(winuser.LB_INSERTSTRING, uintptr(index), lp))
	if ret == winuser.LB_ERRSPACE || ret == winuser.LB_ERR {
		return errs.NewError("SendMessage(LB_INSERTSTRING)")
	}
	return nil
}

func (lb *ListBox) removeItem(index int) error {
	if winuser.LB_ERR == int(lb.SendMessage(winuser.LB_DELETESTRING, uintptr(index), 0)) {
		return errs.NewError("SendMessage(LB_DELETESTRING)")
	}

	return nil
}

// reread all the items from list model
func (lb *ListBox) resetItems() error {
	lb.SetSuspended(true)
	defer lb.SetSuspended(false)

	lb.SendMessage(winuser.LB_RESETCONTENT, 0, 0)

	lb.maxItemTextWidth = 0

	oldValue := lb.currentValue

	if lb.model == nil {
		lb.SetCurrentIndex(-1)
		return nil
	}

	count := lb.model.ItemCount()

	lb.lastWidthsMeasuredFor = make([]int, count)

	for i := 0; i < count; i++ {
		if err := lb.insertItemAt(i); err != nil {
			return err
		}
	}

	if oldValue != nil {
		lb.Property("Value").Set(oldValue)
	} else {
		lb.SetCurrentIndex(-1)
	}

	if lb.styler == nil {
		// Update the listbox width (this sets the correct horizontal scrollbar).
		sh := lb.idealSize()
		lb.SendMessage(winuser.LB_SETHORIZONTALEXTENT, uintptr(sh.Width), 0)
	}

	return nil
}

func (lb *ListBox) ensureVisibleItemsHeightUpToDate() error {
	if lb.styler == nil {
		return nil
	}

	if !lb.Suspended() {
		lb.SetSuspended(true)
		defer lb.SetSuspended(false)
	}

	topIndex := int(lb.SendMessage(winuser.LB_GETTOPINDEX, 0, 0))
	offset := maxi(0, topIndex-10)
	count := lb.model.ItemCount()
	var rc gdi32.RECT
	lb.SendMessage(winuser.LB_GETITEMRECT, uintptr(offset), uintptr(unsafe.Pointer(&rc)))
	width := int(rc.Right - rc.Left)
	offsetTop := int(rc.Top)
	lbHeight := lb.HeightPixels()

	var pastBottomCount int
	for i := offset; i >= 0 && i < count; i++ {
		if lb.lastWidthsMeasuredFor[i] == lb.lastWidth {
			continue
		}

		lb.SendMessage(winuser.LB_GETITEMRECT, uintptr(i), uintptr(unsafe.Pointer(&rc)))

		if int(rc.Top)-offsetTop > lbHeight {
			if pastBottomCount++; pastBottomCount > 10 {
				break
			}
		}

		height := lb.styler.ItemHeight(i, width)

		lb.SendMessage(winuser.LB_SETITEMHEIGHT, uintptr(i), uintptr(height))

		lb.lastWidthsMeasuredFor[i] = lb.lastWidth
	}

	lb.EnsureItemVisible(topIndex)

	return nil
}

func (lb *ListBox) attachModel() {
	itemsResetHandler := func() {
		lb.resetItems()
	}
	lb.itemsResetHandlerHandle = lb.model.ItemsReset().Attach(itemsResetHandler)

	itemChangedHandler := func(index int) {
		if commctrl.CB_ERR == lb.SendMessage(winuser.LB_DELETESTRING, uintptr(index), 0) {
			errs.NewError("SendMessage(CB_DELETESTRING)")
		}

		lb.insertItemAt(index)

		if lb.styler != nil {
			var rc gdi32.RECT
			lb.SendMessage(winuser.LB_GETITEMRECT, uintptr(index), uintptr(unsafe.Pointer(&rc)))
			width := int(rc.Right - rc.Left)
			height := lb.styler.ItemHeight(index, width)

			lb.SendMessage(winuser.LB_SETITEMHEIGHT, uintptr(index), uintptr(height))

			lb.lastWidthsMeasuredFor[index] = lb.lastWidth
		}

		lb.SetCurrentIndex(lb.prevCurIndex)
	}
	lb.itemChangedHandlerHandle = lb.model.ItemChanged().Attach(itemChangedHandler)

	lb.itemsInsertedHandlerHandle = lb.model.ItemsInserted().Attach(func(from, to int) {
		if !lb.Suspended() {
			lb.SetSuspended(true)
			defer lb.SetSuspended(false)
		}

		for i := from; i <= to; i++ {
			lb.insertItemAt(i)
		}

		lb.lastWidthsMeasuredFor = append(lb.lastWidthsMeasuredFor[:from], append(make([]int, to-from+1), lb.lastWidthsMeasuredFor[from:]...)...)

		lb.ensureVisibleItemsHeightUpToDate()
	})

	lb.itemsRemovedHandlerHandle = lb.model.ItemsRemoved().Attach(func(from, to int) {
		if !lb.Suspended() {
			lb.SetSuspended(true)
			defer lb.SetSuspended(false)
		}

		for i := to; i >= from; i-- {
			lb.removeItem(i)
		}

		lb.lastWidthsMeasuredFor = append(lb.lastWidthsMeasuredFor[:from], lb.lastWidthsMeasuredFor[to:]...)

		lb.ensureVisibleItemsHeightUpToDate()
	})
}

func (lb *ListBox) detachModel() {
	lb.model.ItemsReset().Detach(lb.itemsResetHandlerHandle)
	lb.model.ItemChanged().Detach(lb.itemChangedHandlerHandle)
	lb.model.ItemsInserted().Detach(lb.itemsInsertedHandlerHandle)
	lb.model.ItemsRemoved().Detach(lb.itemsRemovedHandlerHandle)
}

// Model returns the model of the ListBox.
func (lb *ListBox) Model() interface{} {
	return lb.providedModel
}

// SetModel sets the model of the ListBox.
//
// It is required that mdl either implements walk.ListModel or
// walk.ReflectListModel or be a slice of pointers to struct or a []string.
func (lb *ListBox) SetModel(mdl interface{}) error {
	model, ok := mdl.(ListModel)
	if !ok && mdl != nil {
		var err error
		if model, err = newReflectListModel(mdl); err != nil {
			return err
		}

		if _, ok := mdl.([]string); !ok {
			if badms, ok := model.(bindingAndDisplayMemberSetter); ok {
				badms.setBindingMember(lb.bindingMember)
				badms.setDisplayMember(lb.displayMember)
			}
		}
	}
	lb.providedModel = mdl

	if lb.model != nil {
		lb.detachModel()
	}

	lb.model = model
	lb.bindingValueProvider, _ = model.(BindingValueProvider)

	if model != nil {
		lb.attachModel()
	}

	if err := lb.resetItems(); err != nil {
		return err
	}

	return lb.ensureVisibleItemsHeightUpToDate()
}

// BindingMember returns the member from the model of the ListBox that is bound
// to a field of the data source managed by an associated DataBinder.
//
// This is only applicable to walk.ReflectListModel models and simple slices of
// pointers to struct.
func (lb *ListBox) BindingMember() string {
	return lb.bindingMember
}

// SetBindingMember sets the member from the model of the ListBox that is bound
// to a field of the data source managed by an associated DataBinder.
//
// This is only applicable to walk.ReflectListModel models and simple slices of
// pointers to struct.
//
// For a model consisting of items of type S, data source field of type T and
// bindingMember "Foo", this can be one of the following:
//
//	A field		Foo T
//	A method	func (s S) Foo() T
//	A method	func (s S) Foo() (T, error)
//
// If bindingMember is not a simple member name like "Foo", but a path to a
// member like "A.B.Foo", members "A" and "B" both must be one of the options
// mentioned above, but with T having type pointer to struct.
func (lb *ListBox) SetBindingMember(bindingMember string) error {
	if bindingMember != "" {
		if _, ok := lb.providedModel.([]string); ok {
			return errs.NewError("invalid for []string model")
		}
	}

	lb.bindingMember = bindingMember

	if badms, ok := lb.model.(bindingAndDisplayMemberSetter); ok {
		badms.setBindingMember(bindingMember)
	}

	return nil
}

// DisplayMember returns the member from the model of the ListBox that is
// displayed in the ListBox.
//
// This is only applicable to walk.ReflectListModel models and simple slices of
// pointers to struct.
func (lb *ListBox) DisplayMember() string {
	return lb.displayMember
}

// SetDisplayMember sets the member from the model of the ListBox that is
// displayed in the ListBox.
//
// This is only applicable to walk.ReflectListModel models and simple slices of
// pointers to struct.
//
// For a model consisting of items of type S, the type of the specified member T
// and displayMember "Foo", this can be one of the following:
//
//	A field		Foo T
//	A method	func (s S) Foo() T
//	A method	func (s S) Foo() (T, error)
//
// If displayMember is not a simple member name like "Foo", but a path to a
// member like "A.B.Foo", members "A" and "B" both must be one of the options
// mentioned above, but with T having type pointer to struct.
func (lb *ListBox) SetDisplayMember(displayMember string) error {
	if displayMember != "" {
		if _, ok := lb.providedModel.([]string); ok {
			return errs.NewError("invalid for []string model")
		}
	}

	lb.displayMember = displayMember

	if badms, ok := lb.model.(bindingAndDisplayMemberSetter); ok {
		badms.setDisplayMember(displayMember)
	}

	return nil
}

func (lb *ListBox) Format() string {
	return lb.format
}

func (lb *ListBox) SetFormat(value string) {
	lb.format = value
}

func (lb *ListBox) Precision() int {
	return lb.precision
}

func (lb *ListBox) SetPrecision(value int) {
	lb.precision = value
}

// calculateMaxItemTextWidth returns maximum item text width in native pixels.
func (lb *ListBox) calculateMaxItemTextWidth() int {
	hdc := user32.GetDC(lb.hWnd)
	if hdc == 0 {
		errs.NewError("GetDC failed")
		return -1
	}
	defer user32.ReleaseDC(lb.hWnd, hdc)

	hFontOld := gdi32.SelectObject(hdc, gdi32.HGDIOBJ(lb.Font().handleForDPI(lb.DPI())))
	defer gdi32.SelectObject(hdc, hFontOld)

	var maxWidth int

	if lb.model == nil {
		return -1
	}
	count := lb.model.ItemCount()
	for i := 0; i < count; i++ {
		item := lb.itemString(i)
		var s gdi32.SIZE
		str, err := syscall.UTF16FromString(item)
		if err != nil {
			errs.NewError(err.Error())
		}

		if !gdi32.GetTextExtentPoint32(hdc, &str[0], int32(len(str)-1), &s) {
			errs.NewError("GetTextExtentPoint32 failed")
			return -1
		}

		maxWidth = maxi(maxWidth, int(s.CX))
	}

	return maxWidth
}

// idealSize returns listbox ideal size in native pixels.
func (lb *ListBox) idealSize() Size {
	defaultSize := lb.dialogBaseUnitsToPixels(Size{50, 12})

	if lb.maxItemTextWidth <= 0 {
		lb.maxItemTextWidth = lb.calculateMaxItemTextWidth()
	}

	// FIXME: Use GetThemePartSize instead of guessing
	w := maxi(defaultSize.Width, lb.maxItemTextWidth+IntFrom96DPI(24, lb.DPI()))
	h := defaultSize.Height + 1

	return Size{w, h}
}

func (lb *ListBox) ItemVisible(index int) bool {
	topIndex := int(lb.SendMessage(winuser.LB_GETTOPINDEX, 0, 0))
	var rc gdi32.RECT
	lb.SendMessage(winuser.LB_GETITEMRECT, uintptr(index), uintptr(unsafe.Pointer(&rc)))

	return index >= topIndex && int(rc.Top) < lb.HeightPixels()
}

func (lb *ListBox) EnsureItemVisible(index int) {
	lb.SendMessage(winuser.LB_SETTOPINDEX, uintptr(index), 0)
}

func (lb *ListBox) CurrentIndex() int {
	return int(int32(lb.SendMessage(winuser.LB_GETCURSEL, 0, 0)))
}

func (lb *ListBox) SetCurrentIndex(value int) error {
	if value > -1 && winuser.LB_ERR == int(int32(lb.SendMessage(winuser.LB_SETCURSEL, uintptr(value), 0))) {
		return errs.NewError("Invalid index or ensure lb is single-selection listbox")
	}

	if value != lb.prevCurIndex {
		if value == -1 {
			lb.currentValue = nil
		} else {
			lb.currentValue = lb.Property("Value").Get()
		}

		lb.prevCurIndex = value
		lb.currentIndexChangedPublisher.Publish()
	}

	return nil
}

func (lb *ListBox) SelectedIndexes() []int {
	count := int(int32(lb.SendMessage(winuser.LB_GETCOUNT, 0, 0)))
	if count < 1 {
		return nil
	}
	index32 := make([]int32, count)
	if n := int(int32(lb.SendMessage(winuser.LB_GETSELITEMS, uintptr(count), uintptr(unsafe.Pointer(&index32[0]))))); n == winuser.LB_ERR {
		return nil
	} else {
		indexes := make([]int, n)
		for i := 0; i < n; i++ {
			indexes[i] = int(index32[i])
		}
		return indexes
	}
}

func (lb *ListBox) SetSelectedIndexes(indexes []int) {
	var m int32 = -1
	lb.SendMessage(winuser.LB_SETSEL, win.FALSE, uintptr(m))
	for _, v := range indexes {
		lb.SendMessage(winuser.LB_SETSEL, win.TRUE, uintptr(uint32(v)))
	}
	lb.selectedIndexesChangedPublisher.Publish()
}

func (lb *ListBox) CurrentIndexChanged() *Event {
	return lb.currentIndexChangedPublisher.Event()
}

func (lb *ListBox) SelectedIndexesChanged() *Event {
	return lb.selectedIndexesChangedPublisher.Event()
}

func (lb *ListBox) ItemActivated() *Event {
	return lb.itemActivatedPublisher.Event()
}

func (lb *ListBox) WndProc(hwnd handle.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case user32.WM_MEASUREITEM:
		if lb.styler == nil {
			break
		}

		mis := (*user32.MEASUREITEMSTRUCT)(unsafe.Pointer(lParam))

		mis.ItemHeight = uint32(lb.styler.DefaultItemHeight())

		return win.TRUE

	case user32.WM_DRAWITEM:
		dis := (*user32.DRAWITEMSTRUCT)(unsafe.Pointer(lParam))

		if lb.styler == nil || dis.ItemID < 0 || dis.ItemAction != user32.ODA_DRAWENTIRE {
			return win.TRUE
		}

		lb.style.index = int(dis.ItemID)
		lb.style.rc = dis.RcItem
		lb.style.bounds = rectangleFromRECT(dis.RcItem)
		lb.style.dpi = lb.DPI()
		lb.style.state = dis.ItemState
		lb.style.hwnd = lb.hWnd
		lb.style.hdc = dis.HDC
		lb.style.Font = lb.Font()

		if dis.ItemAction == user32.ODA_FOCUS {
			return win.TRUE
		}

		var hTheme uxtheme.HTHEME
		if !lb.style.highContrastActive {
			strPtr, err := syscall.UTF16PtrFromString("Listview")
			if err != nil {
				errs.NewError(err.Error())
			}
			if hTheme = uxtheme.OpenThemeData(lb.hWnd, strPtr); hTheme != 0 {
				defer uxtheme.CloseThemeData(hTheme)
			}
		}
		lb.style.hTheme = hTheme

		if dis.ItemState&user32.ODS_CHECKED != 0 {
			if lb.style.highContrastActive || lb.Focused() {
				lb.style.BackgroundColor = lb.themeSelectedBGColor
				lb.style.TextColor = lb.themeSelectedTextColor
			} else {
				lb.style.BackgroundColor = lb.themeSelectedNotFocusedBGColor
				lb.style.TextColor = lb.themeNormalTextColor
			}
		} else if int(dis.ItemID) == lb.style.hoverIndex {
			if hTheme == 0 {
				lb.style.BackgroundColor = lb.themeNormalBGColor
			} else {
				lb.style.BackgroundColor = lb.themeSelectedBGColor
			}
			lb.style.TextColor = lb.themeNormalTextColor
		} else {
			lb.style.BackgroundColor = lb.themeNormalBGColor
			lb.style.TextColor = lb.themeNormalTextColor
		}
		if lb.themeNormalTextColor == RGB(0, 0, 0) {
			lb.style.LineColor = RGB(0, 0, 0)
		} else {
			lb.style.LineColor = RGB(255, 255, 255)
		}
		lb.style.defaultTextColor = lb.style.TextColor

		lb.style.DrawBackground()

		lb.styler.StyleItem(&lb.style)

		defer func() {
			lb.style.bounds = Rectangle{}
			if lb.style.canvas != nil {
				lb.style.canvas.Dispose()
				lb.style.canvas = nil
			}
			lb.style.hdc = 0
		}()

		return win.TRUE

	case user32.WM_WINDOWPOSCHANGED:
		wp := (*user32.WINDOWPOS)(unsafe.Pointer(lParam))

		if wp.Flags&user32.SWP_NOSIZE != 0 {
			break
		}

		if lb.styler != nil && lb.styler.ItemHeightDependsOnWidth() {
			width := lb.WidthPixels()
			if width != lb.lastWidth {
				lb.lastWidth = width
				lb.lastWidthsMeasuredFor = make([]int, lb.model.ItemCount())
			}
		}

		lb.ensureVisibleItemsHeightUpToDate()

		return user32.CallWindowProc(lb.origWndProcPtr, hwnd, msg, wParam, lParam)

	case user32.WM_VSCROLL:
		lb.ensureVisibleItemsHeightUpToDate()

	case user32.WM_MOUSEWHEEL:
		lb.ensureVisibleItemsHeightUpToDate()

	case user32.WM_LBUTTONDOWN:
		lb.Invalidate()

	case user32.WM_MOUSEMOVE:
		if lb.styler == nil {
			break
		}

		if !lb.trackingMouseEvent {
			var tme user32.TRACKMOUSEEVENT
			tme.CbSize = uint32(unsafe.Sizeof(tme))
			tme.DwFlags = user32.TME_LEAVE
			tme.HwndTrack = lb.hWnd

			lb.trackingMouseEvent = user32.TrackMouseEvent(&tme)
		}

		oldHoverIndex := lb.style.hoverIndex

		result := uint32(lb.SendMessage(winuser.LB_ITEMFROMPOINT, 0, lParam))
		if win.HIWORD(result) == 0 {
			index := int(win.LOWORD(result))

			var rc gdi32.RECT
			lb.SendMessage(winuser.LB_GETITEMRECT, uintptr(index), uintptr(unsafe.Pointer(&rc)))

			lp := uint32(lParam)
			x := int32(win.LOWORD(lp))
			y := int32(win.HIWORD(lp))

			if x >= rc.Left && x <= rc.Right && y >= rc.Top && y <= rc.Bottom {
				lb.style.hoverIndex = index

				user32.InvalidateRect(lb.hWnd, &rc, true)
			}
		}

		if lb.style.hoverIndex != oldHoverIndex {
			if wParam&user32.MK_LBUTTON != 0 {
				lb.Invalidate()
			} else {
				lb.invalidateItem(oldHoverIndex)
				lb.invalidateItem(lb.style.hoverIndex)
			}
		}

	case user32.WM_MOUSELEAVE:
		if lb.styler == nil {
			break
		}

		lb.trackingMouseEvent = false

		index := lb.style.hoverIndex

		lb.style.hoverIndex = -1

		lb.invalidateItem(index)

	case user32.WM_COMMAND:
		switch win.HIWORD(uint32(wParam)) {
		case winuser.LBN_SELCHANGE:
			lb.ensureVisibleItemsHeightUpToDate()
			lb.prevCurIndex = lb.CurrentIndex()
			lb.currentValue = lb.Property("Value").Get()
			lb.currentIndexChangedPublisher.Publish()
			lb.selectedIndexesChangedPublisher.Publish()

		case winuser.LBN_DBLCLK:
			lb.itemActivatedPublisher.Publish()
		}

	case user32.WM_GETDLGCODE:
		if form := ancestor(lb); form != nil {
			if dlg, ok := form.(dialogish); ok {
				if dlg.DefaultButton() != nil {
					// If the ListBox lives in a Dialog that has a DefaultButton,
					// we won't swallow the return key.
					break
				}
			}
		}

		if wParam == user32.VK_RETURN {
			return user32.DLGC_WANTALLKEYS
		}

	case user32.WM_KEYDOWN:
		if uint32(lParam)>>30 == 0 && Key(wParam) == KeyReturn && lb.CurrentIndex() > -1 {
			lb.itemActivatedPublisher.Publish()
		}
	}

	return lb.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}

func (lb *ListBox) invalidateItem(index int) {
	var rc gdi32.RECT
	lb.SendMessage(winuser.LB_GETITEMRECT, uintptr(index), uintptr(unsafe.Pointer(&rc)))

	user32.InvalidateRect(lb.hWnd, &rc, true)
}

func (lb *ListBox) CreateLayoutItem(ctx *LayoutContext) LayoutItem {
	return NewGreedyLayoutItem()
}

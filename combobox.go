// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	"github.com/Gipcomp/win32/commctrl"
	"github.com/Gipcomp/win32/gdi32"
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/win"
	"github.com/Gipcomp/winapi/errs"
)

type ComboBox struct {
	WidgetBase
	bindingValueProvider         BindingValueProvider
	model                        ListModel
	providedModel                interface{}
	bindingMember                string
	displayMember                string
	format                       string
	precision                    int
	itemsResetHandlerHandle      int
	itemChangedHandlerHandle     int
	itemsInsertedHandlerHandle   int
	itemsRemovedHandlerHandle    int
	maxItemTextWidth             int // in native pixels
	currentValue                 interface{}
	prevCurIndex                 int
	selChangeIndex               int
	maxLength                    int
	currentIndexChangedPublisher EventPublisher
	textChangedPublisher         EventPublisher
	editingFinishedPublisher     EventPublisher
	editOrigWndProcPtr           uintptr
	editing                      bool
	persistent                   bool
}

var comboBoxEditWndProcPtr uintptr

func init() {
	AppendToWalkInit(func() {
		comboBoxEditWndProcPtr = syscall.NewCallback(comboBoxEditWndProc)
	})
}

func comboBoxEditWndProc(hwnd handle.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	cb := (*ComboBox)(unsafe.Pointer(user32.GetWindowLongPtr(hwnd, user32.GWLP_USERDATA)))

	switch msg {
	case user32.WM_GETDLGCODE:
		if !cb.editing {
			if form := ancestor(cb); form != nil {
				if dlg, ok := form.(dialogish); ok {
					if dlg.DefaultButton() != nil {
						// If the ComboBox lives in a Dialog that has a
						// DefaultButton, we won't swallow the return key.
						break
					}
				}
			}
		}

		if wParam == user32.VK_RETURN {
			return user32.DLGC_WANTALLKEYS
		}

	case user32.WM_KEYDOWN:
		if wParam != user32.VK_RETURN || cb.SendMessage(commctrl.CB_GETDROPPEDSTATE, 0, 0) == 0 {
			cb.handleKeyDown(wParam, lParam)
		}

		if cb.editing && wParam == user32.VK_RETURN {
			cb.editing = false
			cb.editingFinishedPublisher.Publish()
		}

	case user32.WM_KEYUP:
		if wParam != user32.VK_RETURN || cb.SendMessage(commctrl.CB_GETDROPPEDSTATE, 0, 0) == 0 {
			cb.handleKeyUp(wParam, lParam)
		}

	case user32.WM_SETFOCUS, user32.WM_KILLFOCUS:
		cb.invalidateBorderInParent()

		if cb.editing && msg == user32.WM_KILLFOCUS {
			cb.editing = false
			cb.editingFinishedPublisher.Publish()
		}
	}

	return user32.CallWindowProc(cb.editOrigWndProcPtr, hwnd, msg, wParam, lParam)
}

func NewComboBox(parent Container) (*ComboBox, error) {
	cb, err := newComboBoxWithStyle(parent, commctrl.CBS_AUTOHSCROLL|commctrl.CBS_DROPDOWN)
	if err != nil {
		return nil, err
	}

	editHwnd := user32.GetWindow(cb.hWnd, user32.GW_CHILD)

	user32.SetWindowLongPtr(editHwnd, user32.GWLP_USERDATA, uintptr(unsafe.Pointer(cb)))
	cb.editOrigWndProcPtr = user32.SetWindowLongPtr(editHwnd, user32.GWLP_WNDPROC, comboBoxEditWndProcPtr)

	return cb, nil
}

func NewDropDownBox(parent Container) (*ComboBox, error) {
	return newComboBoxWithStyle(parent, commctrl.CBS_DROPDOWNLIST)
}

func newComboBoxWithStyle(parent Container, style uint32) (*ComboBox, error) {
	cb := &ComboBox{prevCurIndex: -1, selChangeIndex: -1, precision: 2}

	if err := InitWidget(
		cb,
		parent,
		"COMBOBOX",
		user32.WS_TABSTOP|user32.WS_VISIBLE|user32.WS_VSCROLL|style,
		0); err != nil {
		return nil, err
	}

	succeeded := false
	defer func() {
		if !succeeded {
			cb.Dispose()
		}
	}()

	var event *Event
	if style&commctrl.CBS_DROPDOWNLIST == commctrl.CBS_DROPDOWNLIST {
		event = cb.CurrentIndexChanged()
	} else {
		event = cb.TextChanged()
	}

	cb.GraphicsEffects().Add(InteractionEffect)
	cb.GraphicsEffects().Add(FocusEffect)

	cb.MustRegisterProperty("CurrentIndex", NewProperty(
		func() interface{} {
			return cb.CurrentIndex()
		},
		func(v interface{}) error {
			return cb.SetCurrentIndex(assertIntOr(v, -1))
		},
		cb.CurrentIndexChanged()))

	cb.MustRegisterProperty("Text", NewProperty(
		func() interface{} {
			return cb.Text()
		},
		func(v interface{}) error {
			return cb.SetText(assertStringOr(v, ""))
		},
		event))

	cb.MustRegisterProperty("CurrentItem", NewReadOnlyProperty(
		func() interface{} {
			if rlm, ok := cb.providedModel.(ReflectListModel); ok {
				if i := cb.CurrentIndex(); i > -1 {
					return reflect.ValueOf(rlm.Items()).Index(i).Interface()
				}
			}

			return nil
		},
		cb.CurrentIndexChanged()))

	cb.MustRegisterProperty("HasCurrentItem", NewReadOnlyBoolProperty(
		func() bool {
			return cb.CurrentIndex() != -1
		},
		cb.CurrentIndexChanged()))

	cb.MustRegisterProperty("TextNotEmpty", NewReadOnlyBoolProperty(
		func() bool {
			return cb.Text() != ""
		},
		cb.CurrentIndexChanged()))

	cb.MustRegisterProperty("Value", NewProperty(
		func() interface{} {
			if cb.Editable() {
				return cb.Text()
			}

			index := cb.CurrentIndex()

			if cb.bindingValueProvider == nil || index == -1 {
				return nil
			}

			return cb.bindingValueProvider.BindingValue(index)
		},
		func(v interface{}) error {
			if cb.Editable() {
				return cb.SetText(assertStringOr(v, ""))
			}

			if cb.bindingValueProvider == nil {
				if cb.model == nil {
					return nil
				} else {
					return errs.NewError("Data binding is only supported using a model that implements BindingValueProvider.")
				}
			}

			index := -1

			count := cb.model.ItemCount()
			for i := 0; i < count; i++ {
				if cb.bindingValueProvider.BindingValue(i) == v {
					index = i
					break
				}
			}

			return cb.SetCurrentIndex(index)
		},
		event))

	succeeded = true

	return cb, nil
}

func (cb *ComboBox) applyFont(font *Font) {
	cb.WidgetBase.applyFont(font)

	if cb.model != nil {
		cb.maxItemTextWidth = cb.calculateMaxItemTextWidth()
		cb.RequestLayout()
	}
}

func (cb *ComboBox) Editable() bool {
	return !cb.hasStyleBits(commctrl.CBS_DROPDOWNLIST)
}

func (cb *ComboBox) itemString(index int) string {
	switch val := cb.model.Value(index).(type) {
	case string:
		return val

	case time.Time:
		return val.Format(cb.format)

	case *big.Rat:
		return val.FloatString(cb.precision)

	default:
		return fmt.Sprintf(cb.format, val)
	}

	panic("unreachable")
}

func (cb *ComboBox) insertItemAt(index int) error {
	str := cb.itemString(index)
	strPtr, _ := syscall.UTF16PtrFromString(str)
	lp := uintptr(unsafe.Pointer(strPtr))

	if commctrl.CB_ERR == cb.SendMessage(commctrl.CB_INSERTSTRING, uintptr(index), lp) {
		return errs.NewError("SendMessage(CB_INSERTSTRING)")
	}

	return nil
}

func (cb *ComboBox) removeItem(index int) error {
	if commctrl.CB_ERR == cb.SendMessage(commctrl.CB_DELETESTRING, uintptr(index), 0) {
		return errs.NewError("SendMessage(CB_DELETESTRING")
	}

	return nil
}

func (cb *ComboBox) resetItems() error {
	cb.SetSuspended(true)
	defer cb.SetSuspended(false)

	cb.selChangeIndex = -1

	if win.FALSE == cb.SendMessage(commctrl.CB_RESETCONTENT, 0, 0) {
		return errs.NewError("SendMessage(CB_RESETCONTENT)")
	}

	cb.maxItemTextWidth = 0

	oldValue := cb.currentValue

	if cb.model == nil {
		cb.SetCurrentIndex(-1)
		return nil
	}

	count := cb.model.ItemCount()

	for i := 0; i < count; i++ {
		if err := cb.insertItemAt(i); err != nil {
			return err
		}
	}

	if oldValue != nil {
		cb.Property("Value").Set(oldValue)
	} else {
		cb.SetCurrentIndex(-1)
	}

	cb.RequestLayout()

	return nil
}

func (cb *ComboBox) attachModel() {
	itemsResetHandler := func() {
		cb.resetItems()
	}
	cb.itemsResetHandlerHandle = cb.model.ItemsReset().Attach(itemsResetHandler)

	itemChangedHandler := func(index int) {
		if commctrl.CB_ERR == cb.SendMessage(commctrl.CB_DELETESTRING, uintptr(index), 0) {
			errs.NewError("SendMessage(CB_DELETESTRING)")
		}

		cb.insertItemAt(index)

		cb.SetCurrentIndex(cb.prevCurIndex)
	}
	cb.itemChangedHandlerHandle = cb.model.ItemChanged().Attach(itemChangedHandler)

	cb.itemsInsertedHandlerHandle = cb.model.ItemsInserted().Attach(func(from, to int) {
		for i := from; i <= to; i++ {
			cb.insertItemAt(i)
		}
	})

	cb.itemsRemovedHandlerHandle = cb.model.ItemsRemoved().Attach(func(from, to int) {
		for i := to; i >= from; i-- {
			cb.removeItem(i)
		}
	})
}

func (cb *ComboBox) detachModel() {
	cb.model.ItemsReset().Detach(cb.itemsResetHandlerHandle)
	cb.model.ItemChanged().Detach(cb.itemChangedHandlerHandle)
	cb.model.ItemsInserted().Detach(cb.itemsInsertedHandlerHandle)
	cb.model.ItemsRemoved().Detach(cb.itemsRemovedHandlerHandle)
}

// Model returns the model of the ComboBox.
func (cb *ComboBox) Model() interface{} {
	return cb.providedModel
}

// SetModel sets the model of the ComboBox.
//
// It is required that mdl either implements walk.ListModel or
// walk.ReflectListModel or be a slice of pointers to struct or a []string.
func (cb *ComboBox) SetModel(mdl interface{}) error {
	model, ok := mdl.(ListModel)
	if !ok && mdl != nil {
		var err error
		if model, err = newReflectListModel(mdl); err != nil {
			return err
		}

		if _, ok := mdl.([]string); !ok {
			if badms, ok := model.(bindingAndDisplayMemberSetter); ok {
				var bindingMember string
				if cb.Editable() {
					bindingMember = cb.displayMember
				} else {
					bindingMember = cb.bindingMember
				}
				badms.setBindingMember(bindingMember)
				badms.setDisplayMember(cb.displayMember)
			}
		}
	}
	cb.providedModel = mdl

	if cb.model != nil {
		cb.detachModel()
	}

	cb.model = model
	cb.bindingValueProvider, _ = model.(BindingValueProvider)

	if model != nil {
		cb.attachModel()
	}

	if err := cb.resetItems(); err != nil {
		return err
	}

	if !cb.Editable() && model != nil && model.ItemCount() == 1 {
		cb.SetCurrentIndex(0)
	}

	return cb.Invalidate()
}

// BindingMember returns the member from the model of the ComboBox that is bound
// to a field of the data source managed by an associated DataBinder.
//
// This is only applicable to walk.ReflectListModel models and simple slices of
// pointers to struct.
func (cb *ComboBox) BindingMember() string {
	return cb.bindingMember
}

// SetBindingMember sets the member from the model of the ComboBox that is bound
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
func (cb *ComboBox) SetBindingMember(bindingMember string) error {
	if bindingMember != "" {
		if _, ok := cb.providedModel.([]string); ok {
			return errs.NewError("invalid for []string model")
		}
	}

	cb.bindingMember = bindingMember

	if badms, ok := cb.model.(bindingAndDisplayMemberSetter); ok {
		badms.setBindingMember(bindingMember)
	}

	return nil
}

// DisplayMember returns the member from the model of the ComboBox that is
// displayed in the ComboBox.
//
// This is only applicable to walk.ReflectListModel models and simple slices of
// pointers to struct.
func (cb *ComboBox) DisplayMember() string {
	return cb.displayMember
}

// SetDisplayMember sets the member from the model of the ComboBox that is
// displayed in the ComboBox.
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
func (cb *ComboBox) SetDisplayMember(displayMember string) error {
	if displayMember != "" {
		if _, ok := cb.providedModel.([]string); ok {
			return errs.NewError("invalid for []string model")
		}
	}

	cb.displayMember = displayMember

	if badms, ok := cb.model.(bindingAndDisplayMemberSetter); ok {
		badms.setDisplayMember(displayMember)
	}

	return nil
}

func (cb *ComboBox) Format() string {
	return cb.format
}

func (cb *ComboBox) SetFormat(value string) {
	cb.format = value
}

func (cb *ComboBox) Precision() int {
	return cb.precision
}

func (cb *ComboBox) SetPrecision(value int) {
	cb.precision = value
}

func (cb *ComboBox) MaxLength() int {
	return cb.maxLength
}

func (cb *ComboBox) SetMaxLength(value int) {
	cb.SendMessage(commctrl.CB_LIMITTEXT, uintptr(value), 0)

	cb.maxLength = value
}

// calculateMaxItemTextWidth returns maximum item text width in native pixels.
func (cb *ComboBox) calculateMaxItemTextWidth() int {
	hdc := user32.GetDC(cb.hWnd)
	if hdc == 0 {
		errs.NewError("GetDC failed")
		return -1
	}
	defer user32.ReleaseDC(cb.hWnd, hdc)

	hFontOld := gdi32.SelectObject(hdc, gdi32.HGDIOBJ(cb.Font().handleForDPI(cb.DPI())))
	defer gdi32.SelectObject(hdc, hFontOld)

	var maxWidth int

	count := cb.model.ItemCount()
	for i := 0; i < count; i++ {
		var s gdi32.SIZE
		str, _ := syscall.UTF16PtrFromString(cb.itemString(i))

		//str := syscall.StringToUTF16(cb.itemString(i))

		// if !gdi32.GetTextExtentPoint32(hdc, &str[0], int32(len(str)-1), &s) {

		// if !gdi32.GetTextExtentPoint32(hdc, str, int32(len(helpers.UINT16PtrToString(str))), &s) {
		if !gdi32.GetTextExtentPoint32(hdc, str, int32(len(cb.itemString(i))), &s) {
			errs.NewError("GetTextExtentPoint32 failed")
			return -1
		}

		maxWidth = maxi(maxWidth, int(s.CX))
	}

	return maxWidth
}

func (cb *ComboBox) CurrentIndex() int {
	return int(int32(cb.SendMessage(commctrl.CB_GETCURSEL, 0, 0)))
}

func (cb *ComboBox) SetCurrentIndex(value int) error {
	index := int(int32(cb.SendMessage(commctrl.CB_SETCURSEL, uintptr(value), 0)))

	if index != value {
		return errs.NewError("invalid index")
	}

	if value != cb.prevCurIndex {
		if value == -1 {
			cb.currentValue = nil
		} else {
			cb.currentValue = cb.Property("Value").Get()
		}

		cb.prevCurIndex = value
		cb.currentIndexChangedPublisher.Publish()
	}

	return nil
}

func (cb *ComboBox) CurrentIndexChanged() *Event {
	return cb.currentIndexChangedPublisher.Event()
}

func (cb *ComboBox) Text() string {
	return cb.text()
}

func (cb *ComboBox) SetText(value string) error {
	var oldText string
	oldText, _ = cb.currentValue.(string)

	if err := cb.setText(value); err != nil {
		return err
	}

	if value == oldText {
		return nil
	}

	if cb.Editable() {
		cb.currentValue = value
	}

	cb.textChangedPublisher.Publish()

	return nil
}

func (cb *ComboBox) TextSelection() (start, end int) {
	cb.SendMessage(commctrl.CB_GETEDITSEL, uintptr(unsafe.Pointer(&start)), uintptr(unsafe.Pointer(&end)))
	return
}

func (cb *ComboBox) SetTextSelection(start, end int) {
	cb.SendMessage(commctrl.CB_SETEDITSEL, 0, uintptr(win.MAKELONG(uint16(start), uint16(end))))
}

func (cb *ComboBox) TextChanged() *Event {
	return cb.textChangedPublisher.Event()
}

func (cb *ComboBox) EditingFinished() *Event {
	return cb.editingFinishedPublisher.Event()
}

func (cb *ComboBox) Persistent() bool {
	return cb.persistent
}

func (cb *ComboBox) SetPersistent(value bool) {
	cb.persistent = value
}

func (cb *ComboBox) SaveState() error {
	cb.WriteState(strconv.Itoa(cb.CurrentIndex()))

	return nil
}

func (cb *ComboBox) RestoreState() error {
	state, err := cb.ReadState()
	if err != nil {
		return err
	}
	if state == "" {
		return nil
	}

	if i, err := strconv.Atoi(state); err == nil {
		cb.SetCurrentIndex(i)
	}

	return nil
}

func (cb *ComboBox) WndProc(hwnd handle.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case user32.WM_COMMAND:
		code := win.HIWORD(uint32(wParam))
		selIndex := cb.CurrentIndex()

		switch code {
		case commctrl.CBN_EDITCHANGE:
			cb.editing = true
			cb.selChangeIndex = -1
			cb.textChangedPublisher.Publish()

		case commctrl.CBN_SELCHANGE:
			cb.selChangeIndex = selIndex
			cb.currentIndexChangedPublisher.Publish()

		case commctrl.CBN_SELENDCANCEL:
			if cb.selChangeIndex != -1 {
				if cb.selChangeIndex < cb.model.ItemCount() {
					cb.SetCurrentIndex(cb.selChangeIndex)
				}

				cb.selChangeIndex = -1
			}

		case commctrl.CBN_SELENDOK:
			if editable := cb.Editable(); editable || selIndex != cb.prevCurIndex {
				valueProp := cb.Property("Value")
				if editable && selIndex > -1 {
					valueProp.Set(cb.model.Value(selIndex))
				} else {
					cb.currentValue = valueProp.Get()
				}
				cb.currentIndexChangedPublisher.Publish()
				cb.prevCurIndex = selIndex
				return 0
			}

			cb.selChangeIndex = -1
		}

	case user32.WM_MOUSEWHEEL:
		if !cb.Enabled() {
			return 0
		}

	case user32.WM_WINDOWPOSCHANGED:
		wp := (*user32.WINDOWPOS)(unsafe.Pointer(lParam))

		if wp.Flags&user32.SWP_NOSIZE != 0 {
			break
		}

		if cb.Editable() {
			result := cb.WidgetBase.WndProc(hwnd, msg, wParam, lParam)

			cb.SetTextSelection(0, 0)

			return result
		}
	}

	return cb.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}

func (*ComboBox) NeedsWmSize() bool {
	return true
}

func (cb *ComboBox) CreateLayoutItem(ctx *LayoutContext) LayoutItem {
	var layoutFlags LayoutFlags
	if cb.Editable() {
		layoutFlags = GrowableHorz | GreedyHorz
	} else {
		layoutFlags = GrowableHorz
	}

	defaultSize := cb.dialogBaseUnitsToPixels(Size{30, 12})

	if cb.model != nil && cb.maxItemTextWidth <= 0 {
		cb.maxItemTextWidth = cb.calculateMaxItemTextWidth()
	}

	// FIXME: Use GetThemePartSize instead of guessing
	w := maxi(defaultSize.Width, cb.maxItemTextWidth+int(user32.GetSystemMetricsForDpi(user32.SM_CXVSCROLL, uint32(ctx.dpi)))+8)
	h := defaultSize.Height + 1

	return &comboBoxLayoutItem{
		layoutFlags: layoutFlags,
		idealSize:   Size{w, h},
	}
}

type comboBoxLayoutItem struct {
	LayoutItemBase
	layoutFlags LayoutFlags
	idealSize   Size // in native pixels
}

func (li *comboBoxLayoutItem) LayoutFlags() LayoutFlags {
	return li.layoutFlags
}

func (li *comboBoxLayoutItem) IdealSize() Size {
	return li.idealSize
}

func (li *comboBoxLayoutItem) MinSize() Size {
	return li.idealSize
}

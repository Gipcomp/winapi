// Copyright 2010 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"syscall"
	"unsafe"

	"github.com/Gipcomp/win32/comctl32"
	"github.com/Gipcomp/win32/commctrl"
	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/user32"
	"github.com/Gipcomp/win32/win"
	"github.com/Gipcomp/winapi/errs"
)

type treeViewItemInfo struct {
	handle       commctrl.HTREEITEM
	child2Handle map[TreeItem]commctrl.HTREEITEM
}

type TreeView struct {
	WidgetBase
	model                          TreeModel
	lazyPopulation                 bool
	itemsResetEventHandlerHandle   int
	itemChangedEventHandlerHandle  int
	itemInsertedEventHandlerHandle int
	itemRemovedEventHandlerHandle  int
	item2Info                      map[TreeItem]*treeViewItemInfo
	handle2Item                    map[commctrl.HTREEITEM]TreeItem
	currItem                       TreeItem
	hIml                           comctl32.HIMAGELIST
	usingSysIml                    bool
	imageUintptr2Index             map[uintptr]int32
	filePath2IconIndex             map[string]int32
	expandedChangedPublisher       TreeItemEventPublisher
	currentItemChangedPublisher    EventPublisher
	itemActivatedPublisher         EventPublisher
}

func NewTreeView(parent Container) (*TreeView, error) {
	tv := new(TreeView)

	if err := InitWidget(
		tv,
		parent,
		"SysTreeView32",
		user32.WS_TABSTOP|user32.WS_VISIBLE|commctrl.TVS_HASBUTTONS|commctrl.TVS_LINESATROOT|commctrl.TVS_SHOWSELALWAYS|commctrl.TVS_TRACKSELECT,
		user32.WS_EX_CLIENTEDGE); err != nil {
		return nil, err
	}

	succeeded := false
	defer func() {
		if !succeeded {
			tv.Dispose()
		}
	}()

	if hr := win.HRESULT(tv.SendMessage(commctrl.TVM_SETEXTENDEDSTYLE, commctrl.TVS_EX_DOUBLEBUFFER, commctrl.TVS_EX_DOUBLEBUFFER)); win.FAILED(hr) {
		return nil, errs.ErrorFromHRESULT("TVM_SETEXTENDEDSTYLE", hr)
	}

	if err := tv.setTheme("Explorer"); err != nil {
		return nil, err
	}

	tv.GraphicsEffects().Add(InteractionEffect)
	tv.GraphicsEffects().Add(FocusEffect)

	tv.MustRegisterProperty("CurrentItem", NewReadOnlyProperty(
		func() interface{} {
			return tv.CurrentItem()
		},
		tv.CurrentItemChanged()))

	tv.MustRegisterProperty("CurrentItemLevel", NewReadOnlyProperty(
		func() interface{} {
			level := -1
			item := tv.CurrentItem()

			for item != nil {
				level++
				item = item.Parent()
			}

			return level
		},
		tv.CurrentItemChanged()))

	tv.MustRegisterProperty("HasCurrentItem", NewReadOnlyBoolProperty(
		func() bool {
			return tv.CurrentItem() != nil
		},
		tv.CurrentItemChanged()))

	succeeded = true

	return tv, nil
}

func (tv *TreeView) Dispose() {
	tv.WidgetBase.Dispose()

	tv.disposeImageListAndCaches()
}

func (tv *TreeView) SetBackground(bg Brush) {
	tv.WidgetBase.SetBackground(bg)

	color := Color(user32.GetSysColor(user32.COLOR_WINDOW))

	if bg != nil {
		type Colorer interface {
			Color() Color
		}

		if c, ok := bg.(Colorer); ok {
			color = c.Color()
		}
	}

	tv.SendMessage(commctrl.TVM_SETBKCOLOR, 0, uintptr(color))
}

func (tv *TreeView) Model() TreeModel {
	return tv.model
}

func (tv *TreeView) SetModel(model TreeModel) error {
	if tv.model != nil {
		tv.model.ItemsReset().Detach(tv.itemsResetEventHandlerHandle)
		tv.model.ItemChanged().Detach(tv.itemChangedEventHandlerHandle)
		tv.model.ItemInserted().Detach(tv.itemInsertedEventHandlerHandle)
		tv.model.ItemRemoved().Detach(tv.itemRemovedEventHandlerHandle)

		tv.disposeImageListAndCaches()
	}

	tv.model = model

	if model != nil {
		tv.lazyPopulation = model.LazyPopulation()

		tv.itemsResetEventHandlerHandle = model.ItemsReset().Attach(func(parent TreeItem) {
			if parent == nil {
				tv.resetItems()
			} else if tv.item2Info[parent] != nil {
				tv.SetSuspended(true)
				defer tv.SetSuspended(false)

				if err := tv.removeDescendants(parent); err != nil {
					return
				}

				if err := tv.insertChildren(parent); err != nil {
					return
				}
			}
		})

		tv.itemChangedEventHandlerHandle = model.ItemChanged().Attach(func(item TreeItem) {
			if item == nil || tv.item2Info[item] == nil {
				return
			}

			if err := tv.updateItem(item); err != nil {
				return
			}
		})

		tv.itemInsertedEventHandlerHandle = model.ItemInserted().Attach(func(item TreeItem) {
			tv.SetSuspended(true)
			defer tv.SetSuspended(false)

			var hInsertAfter commctrl.HTREEITEM
			parent := item.Parent()
			for i := parent.ChildCount() - 1; i >= 0; i-- {
				if parent.ChildAt(i) == item {
					if i > 0 {
						hInsertAfter = tv.item2Info[parent.ChildAt(i-1)].handle
					} else {
						hInsertAfter = commctrl.TVI_FIRST
					}
				}
			}

			if _, err := tv.insertItemAfter(item, hInsertAfter); err != nil {
				return
			}
		})

		tv.itemRemovedEventHandlerHandle = model.ItemRemoved().Attach(func(item TreeItem) {
			if err := tv.removeItem(item); err != nil {
				return
			}
		})
	}

	return tv.resetItems()
}

func (tv *TreeView) CurrentItem() TreeItem {
	return tv.currItem
}

func (tv *TreeView) SetCurrentItem(item TreeItem) error {
	if item == tv.currItem {
		return nil
	}

	if item != nil {
		if err := tv.ensureItemAndAncestorsInserted(item); err != nil {
			return err
		}
	}

	handle, err := tv.handleForItem(item)
	if err != nil {
		return err
	}

	if tv.SendMessage(commctrl.TVM_SELECTITEM, commctrl.TVGN_CARET, uintptr(handle)) == 0 {
		return errs.NewError("SendMessage(TVM_SELECTITEM) failed")
	}

	tv.currItem = item

	return nil
}

func (tv *TreeView) EnsureVisible(item TreeItem) error {
	handle, err := tv.handleForItem(item)
	if err != nil {
		return err
	}

	tv.SendMessage(commctrl.TVM_ENSUREVISIBLE, 0, uintptr(handle))

	return nil
}

func (tv *TreeView) handleForItem(item TreeItem) (commctrl.HTREEITEM, error) {
	if item != nil {
		if info := tv.item2Info[item]; info == nil {
			return 0, errs.NewError("invalid item")
		} else {
			return info.handle, nil
		}
	}

	return 0, errs.NewError("invalid item")
}

// ItemAt determines the location of the specified point in native pixels relative to the client area of a tree-view control.
func (tv *TreeView) ItemAt(x, y int) TreeItem {
	hti := commctrl.TVHITTESTINFO{Pt: Point{x, y}.toPOINT()}

	tv.SendMessage(commctrl.TVM_HITTEST, 0, uintptr(unsafe.Pointer(&hti)))

	if item, ok := tv.handle2Item[hti.HItem]; ok {
		return item
	}

	return nil
}

// ItemHeight returns the height of each item in native pixels.
func (tv *TreeView) ItemHeight() int {
	return int(tv.SendMessage(commctrl.TVM_GETITEMHEIGHT, 0, 0))
}

// SetItemHeight sets the height of the tree-view items in native pixels.
func (tv *TreeView) SetItemHeight(height int) {
	tv.SendMessage(commctrl.TVM_SETITEMHEIGHT, uintptr(height), 0)
}

func (tv *TreeView) resetItems() error {
	tv.SetSuspended(true)
	defer tv.SetSuspended(false)

	if err := tv.clearItems(); err != nil {
		return err
	}

	if tv.model == nil {
		return nil
	}

	if err := tv.insertRoots(); err != nil {
		return err
	}

	return nil
}

func (tv *TreeView) clearItems() error {
	if tv.SendMessage(commctrl.TVM_DELETEITEM, 0, 0) == 0 {
		return errs.NewError("SendMessage(TVM_DELETEITEM) failed")
	}

	tv.item2Info = make(map[TreeItem]*treeViewItemInfo)
	tv.handle2Item = make(map[commctrl.HTREEITEM]TreeItem)

	return nil
}

func (tv *TreeView) insertRoots() error {
	for i := tv.model.RootCount() - 1; i >= 0; i-- {
		if _, err := tv.insertItem(tv.model.RootAt(i)); err != nil {
			return err
		}
	}

	return nil
}

func (tv *TreeView) ApplyDPI(dpi int) {
	tv.WidgetBase.ApplyDPI(dpi)

	tv.disposeImageListAndCaches()
}

func (tv *TreeView) applyImageListForImage(image interface{}) {
	tv.hIml, tv.usingSysIml, _ = imageListForImage(image, tv.DPI())

	tv.SendMessage(commctrl.TVM_SETIMAGELIST, 0, uintptr(tv.hIml))

	tv.imageUintptr2Index = make(map[uintptr]int32)
	tv.filePath2IconIndex = make(map[string]int32)
}

func (tv *TreeView) disposeImageListAndCaches() {
	if tv.hIml != 0 && !tv.usingSysIml {
		comctl32.ImageList_Destroy(tv.hIml)
	}
	tv.hIml = 0

	tv.imageUintptr2Index = nil
	tv.filePath2IconIndex = nil
}

func (tv *TreeView) setTVITEMImageInfo(tvi *commctrl.TVITEM, item TreeItem) {
	if imager, ok := item.(Imager); ok {
		if tv.hIml == 0 {
			tv.applyImageListForImage(imager.Image())
		}

		// FIXME: If not setting TVIF_SELECTEDIMAGE and tvi.ISelectedImage,
		// some default icon will show up, even though we have not asked for it.

		tvi.Mask |= commctrl.TVIF_IMAGE | commctrl.TVIF_SELECTEDIMAGE
		tvi.IImage = imageIndexMaybeAdd(
			imager.Image(),
			tv.hIml,
			tv.usingSysIml,
			tv.imageUintptr2Index,
			tv.filePath2IconIndex,
			tv.DPI())

		tvi.ISelectedImage = tvi.IImage
	}
}

func (tv *TreeView) insertItem(item TreeItem) (commctrl.HTREEITEM, error) {
	return tv.insertItemAfter(item, commctrl.TVI_FIRST)
}

func (tv *TreeView) insertItemAfter(item TreeItem, hInsertAfter commctrl.HTREEITEM) (commctrl.HTREEITEM, error) {
	var tvins commctrl.TVINSERTSTRUCT
	tvi := &tvins.Item

	tvi.Mask = commctrl.TVIF_CHILDREN | commctrl.TVIF_TEXT
	tvi.PszText = comctl32.LPSTR_TEXTCALLBACK
	tvi.CChildren = comctl32.I_CHILDRENCALLBACK

	tv.setTVITEMImageInfo(tvi, item)

	parent := item.Parent()

	if parent == nil {
		tvins.HParent = commctrl.TVI_ROOT
	} else {
		info := tv.item2Info[parent]
		if info == nil {
			return 0, errs.NewError("invalid parent")
		}
		tvins.HParent = info.handle
	}

	tvins.HInsertAfter = hInsertAfter

	hItem := commctrl.HTREEITEM(tv.SendMessage(commctrl.TVM_INSERTITEM, 0, uintptr(unsafe.Pointer(&tvins))))
	if hItem == 0 {
		return 0, errs.NewError("TVM_INSERTITEM failed")
	}
	tv.item2Info[item] = &treeViewItemInfo{hItem, make(map[TreeItem]commctrl.HTREEITEM)}
	tv.handle2Item[hItem] = item

	if !tv.lazyPopulation {
		if err := tv.insertChildren(item); err != nil {
			return 0, err
		}
	}

	return hItem, nil
}

func (tv *TreeView) insertChildren(parent TreeItem) error {
	info := tv.item2Info[parent]

	for i := parent.ChildCount() - 1; i >= 0; i-- {
		child := parent.ChildAt(i)

		if handle, err := tv.insertItem(child); err != nil {
			return err
		} else {
			info.child2Handle[child] = handle
		}
	}

	return nil
}

func (tv *TreeView) updateItem(item TreeItem) error {
	tvi := &commctrl.TVITEM{
		Mask:    commctrl.TVIF_TEXT,
		HItem:   tv.item2Info[item].handle,
		PszText: comctl32.LPSTR_TEXTCALLBACK,
	}

	tv.setTVITEMImageInfo(tvi, item)

	if tv.SendMessage(commctrl.TVM_SETITEM, 0, uintptr(unsafe.Pointer(tvi))) == 0 {
		return errs.NewError("SendMessage(TVM_SETITEM) failed")
	}

	return nil
}

func (tv *TreeView) removeItem(item TreeItem) error {
	if err := tv.removeDescendants(item); err != nil {
		return err
	}

	info := tv.item2Info[item]
	if info == nil {
		return errs.NewError("invalid item")
	}

	if tv.SendMessage(commctrl.TVM_DELETEITEM, 0, uintptr(info.handle)) == 0 {
		return errs.NewError("SendMessage(TVM_DELETEITEM) failed")
	}

	if parentInfo := tv.item2Info[item.Parent()]; parentInfo != nil {
		delete(parentInfo.child2Handle, item)
	}
	delete(tv.item2Info, item)
	delete(tv.handle2Item, info.handle)

	return nil
}

func (tv *TreeView) removeDescendants(parent TreeItem) error {
	for item := range tv.item2Info[parent].child2Handle {
		if err := tv.removeItem(item); err != nil {
			return err
		}
	}

	return nil
}

func (tv *TreeView) ensureItemAndAncestorsInserted(item TreeItem) error {
	if item == nil {
		return errs.NewError("invalid item")
	}

	tv.SetSuspended(true)
	defer tv.SetSuspended(false)

	var hierarchy []TreeItem

	for item != nil && tv.item2Info[item] == nil {
		item = item.Parent()

		if item != nil {
			hierarchy = append(hierarchy, item)
		} else {
			return errs.NewError("invalid item")
		}
	}

	for i := len(hierarchy) - 1; i >= 0; i-- {
		if err := tv.insertChildren(hierarchy[i]); err != nil {
			return err
		}
	}

	return nil
}

func (tv *TreeView) Expanded(item TreeItem) bool {
	if tv.item2Info[item] == nil {
		return false
	}

	tvi := &commctrl.TVITEM{
		HItem:     tv.item2Info[item].handle,
		Mask:      commctrl.TVIF_STATE,
		StateMask: commctrl.TVIS_EXPANDED,
	}

	if tv.SendMessage(commctrl.TVM_GETITEM, 0, uintptr(unsafe.Pointer(tvi))) == 0 {
		errs.NewError("SendMessage(TVM_GETITEM) failed")
	}

	return tvi.State&commctrl.TVIS_EXPANDED != 0
}

func (tv *TreeView) SetExpanded(item TreeItem, expanded bool) error {
	if expanded {
		if err := tv.ensureItemAndAncestorsInserted(item); err != nil {
			return err
		}
	}

	info := tv.item2Info[item]
	if info == nil {
		return errs.NewError("invalid item")
	}

	var action uintptr
	if expanded {
		action = commctrl.TVE_EXPAND
	} else {
		action = commctrl.TVE_COLLAPSE
	}

	if 0 == tv.SendMessage(commctrl.TVM_EXPAND, action, uintptr(info.handle)) {
		return errs.NewError("SendMessage(TVM_EXPAND) failed")
	}

	return nil
}

func (tv *TreeView) ExpandedChanged() *TreeItemEvent {
	return tv.expandedChangedPublisher.Event()
}

func (tv *TreeView) CurrentItemChanged() *Event {
	return tv.currentItemChangedPublisher.Event()
}

func (tv *TreeView) ItemActivated() *Event {
	return tv.itemActivatedPublisher.Event()
}

func (tv *TreeView) WndProc(hwnd handle.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case user32.WM_GETDLGCODE:
		if wParam == user32.VK_RETURN {
			return user32.DLGC_WANTALLKEYS
		}

	case user32.WM_NOTIFY:
		nmhdr := (*user32.NMHDR)(unsafe.Pointer(lParam))

		switch nmhdr.Code {
		case commctrl.TVN_GETDISPINFO:
			nmtvdi := (*commctrl.NMTVDISPINFO)(unsafe.Pointer(lParam))
			item := tv.handle2Item[nmtvdi.Item.HItem]

			if nmtvdi.Item.Mask&commctrl.TVIF_TEXT != 0 {
				text := item.Text()
				utf16, err := syscall.UTF16FromString(text)
				if err != nil {
					errs.NewError(err.Error())
				}
				buf := (*[264]uint16)(unsafe.Pointer(nmtvdi.Item.PszText))
				max := mini(len(utf16), int(nmtvdi.Item.CchTextMax))
				copy((*buf)[:], utf16[:max])
				(*buf)[max-1] = 0
			}
			if nmtvdi.Item.Mask&commctrl.TVIF_CHILDREN != 0 {
				if hc, ok := item.(HasChilder); ok {
					if hc.HasChild() {
						nmtvdi.Item.CChildren = 1
					} else {
						nmtvdi.Item.CChildren = 0
					}
				} else {
					nmtvdi.Item.CChildren = int32(item.ChildCount())
				}
			}

		case commctrl.TVN_ITEMEXPANDING:
			nmtv := (*commctrl.NMTREEVIEW)(unsafe.Pointer(lParam))
			item := tv.handle2Item[nmtv.ItemNew.HItem]

			if nmtv.Action == commctrl.TVE_EXPAND && tv.lazyPopulation {
				info := tv.item2Info[item]
				if len(info.child2Handle) == 0 {
					tv.insertChildren(item)
				}
			}

		case commctrl.TVN_ITEMEXPANDED:
			nmtv := (*commctrl.NMTREEVIEW)(unsafe.Pointer(lParam))
			item := tv.handle2Item[nmtv.ItemNew.HItem]

			switch nmtv.Action {
			case commctrl.TVE_COLLAPSE:
				tv.expandedChangedPublisher.Publish(item)

			case commctrl.TVE_COLLAPSERESET:

			case commctrl.TVE_EXPAND:
				tv.expandedChangedPublisher.Publish(item)

			case commctrl.TVE_EXPANDPARTIAL:

			case commctrl.TVE_TOGGLE:
			}

		case comctl32.NM_DBLCLK:
			tv.itemActivatedPublisher.Publish()

		case commctrl.TVN_KEYDOWN:
			nmtvkd := (*commctrl.NMTVKEYDOWN)(unsafe.Pointer(lParam))
			if nmtvkd.WVKey == uint16(KeyReturn) {
				tv.itemActivatedPublisher.Publish()
			}

		case commctrl.TVN_SELCHANGED:
			nmtv := (*commctrl.NMTREEVIEW)(unsafe.Pointer(lParam))

			tv.currItem = tv.handle2Item[nmtv.ItemNew.HItem]

			tv.currentItemChangedPublisher.Publish()
		}
	}

	return tv.WidgetBase.WndProc(hwnd, msg, wParam, lParam)
}

func (*TreeView) NeedsWmSize() bool {
	return true
}

func (tv *TreeView) CreateLayoutItem(ctx *LayoutContext) LayoutItem {
	return NewGreedyLayoutItem()
}

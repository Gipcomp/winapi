// Copyright 2011 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package winapi

import (
	"syscall"

	"github.com/Gipcomp/win32/handle"
	"github.com/Gipcomp/win32/shell32"
)

type dropFilesEventHandlerInfo struct {
	handler DropFilesEventHandler
	once    bool
}

type DropFilesEventHandler func([]string)

type DropFilesEvent struct {
	hWnd     handle.HWND
	handlers []dropFilesEventHandlerInfo
}

func (e *DropFilesEvent) Attach(handler DropFilesEventHandler) int {
	if len(e.handlers) == 0 {
		shell32.DragAcceptFiles(e.hWnd, true)
	}

	handlerInfo := dropFilesEventHandlerInfo{handler, false}

	for i, h := range e.handlers {
		if h.handler == nil {
			e.handlers[i] = handlerInfo
			return i
		}
	}

	e.handlers = append(e.handlers, handlerInfo)

	return len(e.handlers) - 1
}

func (e *DropFilesEvent) Detach(handle int) {
	e.handlers[handle].handler = nil

	for _, h := range e.handlers {
		if h.handler != nil {
			return
		}
	}

	shell32.DragAcceptFiles(e.hWnd, false)
}

func (e *DropFilesEvent) Once(handler DropFilesEventHandler) {
	i := e.Attach(handler)
	e.handlers[i].once = true
}

type DropFilesEventPublisher struct {
	event DropFilesEvent
}

func (p *DropFilesEventPublisher) Event(hWnd handle.HWND) *DropFilesEvent {
	p.event.hWnd = hWnd
	return &p.event
}

func (p *DropFilesEventPublisher) Publish(hDrop shell32.HDROP) {
	var files []string

	n := shell32.DragQueryFile(hDrop, 0xFFFFFFFF, nil, 0)
	for i := 0; i < int(n); i++ {
		bufSize := uint(512)
		buf := make([]uint16, bufSize)
		if shell32.DragQueryFile(hDrop, uint(i), &buf[0], bufSize) > 0 {
			files = append(files, syscall.UTF16ToString(buf))
		}
	}
	shell32.DragFinish(hDrop)

	for i, h := range p.event.handlers {
		if h.handler != nil {
			h.handler(files)

			if h.once {
				p.event.Detach(i)
			}
		}
	}
}

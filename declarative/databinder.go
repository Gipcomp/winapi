// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import (
	"time"

	"github.com/Gipcomp/winapi"
)

type DataBinder struct {
	AssignTo            **winapi.DataBinder
	AutoSubmit          bool
	AutoSubmitDelay     time.Duration
	DataSource          interface{}
	ErrorPresenter      ErrorPresenter
	Name                string
	OnCanSubmitChanged  winapi.EventHandler
	OnDataSourceChanged winapi.EventHandler
	OnReset             winapi.EventHandler
	OnSubmitted         winapi.EventHandler
}

func (db DataBinder) create() (*winapi.DataBinder, error) {
	b := winapi.NewDataBinder()

	if db.AssignTo != nil {
		*db.AssignTo = b
	}

	if db.ErrorPresenter != nil {
		ep, err := db.ErrorPresenter.Create()
		if err != nil {
			return nil, err
		}
		b.SetErrorPresenter(ep)
	}

	b.SetDataSource(db.DataSource)

	b.SetAutoSubmit(db.AutoSubmit)
	b.SetAutoSubmitDelay(db.AutoSubmitDelay)

	if db.OnCanSubmitChanged != nil {
		b.CanSubmitChanged().Attach(db.OnCanSubmitChanged)
	}
	if db.OnDataSourceChanged != nil {
		b.DataSourceChanged().Attach(db.OnDataSourceChanged)
	}
	if db.OnReset != nil {
		b.ResetFinished().Attach(db.OnReset)
	}
	if db.OnSubmitted != nil {
		b.Submitted().Attach(db.OnSubmitted)
	}

	return b, nil
}

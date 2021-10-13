// Copyright 2012 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package declarative

import "github.com/Gipcomp/winapi"

type ValidatorRef struct {
	Validator winapi.Validator
}

func (vr ValidatorRef) Create() (winapi.Validator, error) {
	return vr.Validator, nil
}

type Range struct {
	Min float64
	Max float64
}

func (r Range) Create() (winapi.Validator, error) {
	return winapi.NewRangeValidator(r.Min, r.Max)
}

type Regexp struct {
	Pattern string
}

func (re Regexp) Create() (winapi.Validator, error) {
	return winapi.NewRegexpValidator(re.Pattern)
}

type SelRequired struct {
}

func (SelRequired) Create() (winapi.Validator, error) {
	return winapi.SelectionRequiredValidator(), nil
}

type dMultiValidator struct {
	validators []Validator
}

func (av dMultiValidator) Create() (winapi.Validator, error) {
	var validators []winapi.Validator

	for _, dv := range av.validators {
		if wv, err := dv.Create(); err != nil {
			return nil, err
		} else {
			validators = append(validators, wv)
		}
	}

	return &wMultiValidator{validators}, nil
}

type wMultiValidator struct {
	validators []winapi.Validator
}

func (av *wMultiValidator) Validate(v interface{}) error {
	for _, validator := range av.validators {
		if err := validator.Validate(v); err != nil {
			return err
		}
	}

	return nil
}

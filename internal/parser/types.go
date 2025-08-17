// Copyright 2025 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package parser

import "errors"

type Config struct {
	SelectProfile string     `hcl:"select_profile"`
	Profiles      []*Profile `hcl:"profile,block"`
}

type Profile struct {
	Name   string `hcl:"name,label"`
	Region string `hcl:"region,optional"`
	Chain  *Chain `hcl:"chain,block"`
}

type Chain struct {
	TTL      int64       `hcl:"ttl,optional"`
	UseRoles []*UseRoles `hcl:"use,block"`
}

type UseRoles struct {
	ARN  string `hcl:"arn"`
	Skip bool   `hcl:"skip,optional"`
}

const defaultTLL = 12

var (
	ErrParseHCL           = errors.New("ParseHCL return a nil")
	ErrSchemaConfig       = errors.New("unable to configure schema for config file")
	ErrProfileNotSelected = errors.New("select_profile is required")
)

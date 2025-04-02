// Copyright 2025 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"container/ring"
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

var ErrMinRoles = errors.New("at least two roles are required")

type ServiceSTS interface {
	AssumeRole(
		ctx context.Context,
		params *sts.AssumeRoleInput,
		optFns ...func(*sts.Options),
	) (*sts.AssumeRoleOutput, error)
	GetCallerIdentity(
		ctx context.Context,
		params *sts.GetCallerIdentityInput,
		optFns ...func(*sts.Options),
	) (*sts.GetCallerIdentityOutput, error)
}

type App struct {
	client          ServiceSTS
	region          string
	roles           *ring.Ring
	usableRoles     map[string]struct{}
	sessionDuration time.Duration
}

func NewApp(ctx context.Context, region string, roles []string, usableRoles []string) (*App, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %w", err)
	}

	rolesPool, err := setRolePool(roles)
	if err != nil {
		return nil, err
	}

	hMap := make(map[string]struct{})
	for _, role := range usableRoles {
		hMap[role] = struct{}{}
	}

	const maxSessionDuration = 15

	return &App{
		client:          sts.NewFromConfig(cfg),
		region:          region,
		roles:           rolesPool,
		usableRoles:     hMap,
		sessionDuration: maxSessionDuration * time.Minute,
	}, nil
}

type StringSlice []string

func (s *StringSlice) String() string {
	return strings.Join(*s, ", ")
}

func (s *StringSlice) Set(value string) error {
	*s = append(*s, value)

	return nil
}

var _ flag.Value = (*StringSlice)(nil)

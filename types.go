// Copyright 2025 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"container/ring"
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

var (
	ErrMinRoles                = errors.New("at least two roles are required")
	ErrInvalidCredentials      = errors.New("invalid credentials: one or more required fields are nil")
	ErrUsableRoleNotInRoleList = errors.New("usable role is missing from roles list")
)

type ServiceSTS interface {
	stscreds.AssumeRoleAPIClient
	GetCallerIdentity(
		ctx context.Context,
		params *sts.GetCallerIdentityInput,
		optFns ...func(*sts.Options),
	) (*sts.GetCallerIdentityOutput, error)
}

type App struct {
	// client is the AWS STS service client used for role assumptions
	client ServiceSTS
	// profileWriter manages the execution of AWS profile updates
	profileWriter *ProfileWriter
	// region is the AWS region used for IAM operations
	region string
	// roles is a ring buffer containing all roles that can be assumed
	roles *ring.Ring
	// usableRoles is a set of roles with meaningful permissions
	usableRoles map[string]struct{}
	// sessionDuration is the duration for which assumed role credentials are valid
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
		return nil, fmt.Errorf("failed to set role pool: %w", err)
	}

	hPool := make(map[string]struct{})
	for _, role := range roles {
		hPool[role] = struct{}{}
	}

	hMap := make(map[string]struct{})

	for _, role := range usableRoles {
		if _, ok := hPool[role]; !ok {
			slog.Error("usable role is missing from roles list", slog.String("role", role))

			return nil, ErrUsableRoleNotInRoleList
		}

		hMap[role] = struct{}{}
	}

	// maxSessionDuration defines the duration in minutes for which assumed role credentials are valid
	const maxSessionDuration = 15

	return &App{
		client:          sts.NewFromConfig(cfg),
		profileWriter:   NewProfileWriter(nil),
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

type CmdExecutor interface {
	Execute(name string, arg ...string) ([]byte, error)
}

type DefaultCmdExecutor struct{}

var _ CmdExecutor = (*DefaultCmdExecutor)(nil)

func (e *DefaultCmdExecutor) Execute(name string, arg ...string) ([]byte, error) {
	return exec.Command(name, arg...).CombinedOutput() //nolint:wrapcheck
}

type ProfileWriter struct {
	cmdExecutor CmdExecutor
	profileName string
}

func NewProfileWriter(cmdExecutor CmdExecutor) *ProfileWriter {
	if cmdExecutor == nil {
		cmdExecutor = &DefaultCmdExecutor{}
	}

	return &ProfileWriter{
		cmdExecutor: cmdExecutor,
		profileName: defaultProfileName,
	}
}

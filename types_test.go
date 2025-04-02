// Copyright 2025 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type MockSTSClient struct {
	mockGetCallerIdentityOutput sts.GetCallerIdentityOutput
	mockGetCallerIdentityError  error
	mockAssumeRoleOutput        map[string]sts.AssumeRoleOutput
	mockAssumeRoleError         error
}

func (m MockSTSClient) AssumeRole(
	_ context.Context,
	params *sts.AssumeRoleInput,
	_ ...func(*sts.Options),
) (*sts.AssumeRoleOutput, error) {
	v, ok := m.mockAssumeRoleOutput[*params.RoleArn]
	if !ok {
		return nil, m.mockAssumeRoleError
	}

	return &v, nil
}

func (m MockSTSClient) GetCallerIdentity(
	_ context.Context,
	_ *sts.GetCallerIdentityInput,
	_ ...func(*sts.Options),
) (*sts.GetCallerIdentityOutput, error) {
	return &m.mockGetCallerIdentityOutput, m.mockGetCallerIdentityError
}

var _ ServiceSTS = (*MockSTSClient)(nil)

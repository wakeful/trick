// Copyright 2025 variHQ OÜ
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"reflect"
	"testing"

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

func TestNewApp(t *testing.T) {
	pool, err := setRolePool([]string{"role-a", "role-b"})
	if err != nil {
		t.Fatalf("setRolePool failed: %v", err)
	}

	type args struct {
		region      string
		roles       []string
		usableRoles []string
	}

	tests := []struct {
		name    string
		args    args
		want    *App
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				region:      "eu-west-1",
				roles:       []string{"role-a", "role-b"},
				usableRoles: []string{"role-a", "role-b"},
			},
			want: &App{
				client:          MockSTSClient{},
				region:          "eu-west-1",
				roles:           pool,
				usableRoles:     map[string]struct{}{"role-a": {}, "role-b": {}},
				sessionDuration: 0,
			},
			wantErr: false,
		},
		{
			name: "error",
			args: args{
				region:      "eu-west-1",
				roles:       []string{"role-a", "role-b"},
				usableRoles: []string{"role-c"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "no roles",
			args: args{
				region:      "eu-west-1",
				roles:       []string{},
				usableRoles: []string{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewApp(t.Context(), tt.args.region, tt.args.roles, tt.args.usableRoles)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewApp() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if (err != nil) && tt.wantErr {
				return
			}

			if !reflect.DeepEqual(got.roles, tt.want.roles) {
				t.Errorf("NewApp() roles got = %v, want %v", got, tt.want)
			}

			if !reflect.DeepEqual(got.usableRoles, tt.want.usableRoles) {
				t.Errorf("NewApp() usableRoles got = %v, want %v", got, tt.want)
			}
		})
	}
}
